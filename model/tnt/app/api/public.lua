local checks = require('checks')
local json = require('json')
local log = require('log')
local metrics = require('metrics')
local d = require('app.data')

local action = require('app.action')
local getters = require('app.api.getters')
local risk = require('app.api.risk')
local setters = require('app.api.setters')
local config = require('app.config')
local order = require('app.engine.order')
local equeue = require('app.enginequeue')
local errors = require('app.lib.errors')
local rpc = require('app.rpc')
local notif = require('app.api.notif')
local deadman = require('app.api.deadman')


local PublicAPIError = errors.new_class("PUBLIC_API")

local p = {}

function is_client_order_id_available(profile_id, client_order_id)
    if client_order_id == nil or client_order_id == "" then
        return true
    end

    local exist = box.space.used_client_order_id:get{profile_id, client_order_id}
    if exist ~= nil then
        return false
    end

    return true
end

function get_both(profile_id, order_id, client_order_id)
    if client_order_id ~= nil and client_order_id ~= "" then
        local exist = box.space.used_client_order_id:get{profile_id, client_order_id}
        if exist ~= nil then
            return exist.order_id, exist.client_order_id
        end
    end

    if order_id ~= nil and order_id ~= "" then
        local exist = box.space.used_client_order_id.index.profile_order_id:min{profile_id, order_id}
        if exist ~= nil then
            return exist.order_id, exist.client_order_id
        end

    end

    return order_id, client_order_id
end

-- we use market_id, to easier check market status later
function save_client_order_id(profile_id, client_order_id, order_id, market_id)
    if client_order_id == nil or client_order_id == "" then
        return nil
    end

    local status, res = pcall(function() return box.space.used_client_order_id:insert{profile_id, client_order_id, order_id, market_id} end)

    if status == false then
        return tostring(res)
    end

    return nil
end

--[[
    1. Generate new order_id
    2. order.status = PROCESSING
    3. put to the queue
    
    return: {task, order, err}
--]]
function p.new_order(
    profile_id,
    market_id,
    order_type,
    order_side,
    order_price,
    order_size,
    client_order_id,
    trigger_price,
    size_percent,
    time_in_force,
    custom_order_id,

    matching_meta
)
    checks('number', 'string', 'string', 'string', '?decimal', '?decimal', '?string', '?decimal', '?decimal', '?string', '?string', '?table|matching_meta')
    local res, err

    deadman.touch(profile_id)

    res = equeue.check_limit(profile_id)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end

    -- Check that client_order_id was not used:
    if is_client_order_id_available(profile_id, client_order_id) == false then
        return {task = nil, order = nil, error = ERR_CLIENT_ORDER_ID_DUPLICATE}
    end
    -- PUT ORDER TO THE QUEUE --

    res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    if res["error"] ~= nil then
        log.error(PublicAPIError:new(res["error"]))
        return {task = nil, order = nil, error = tostring(res["error"])}
    end
    local qname = res["res"]

    res = equeue.check_queue_limit(qname)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end

    local profile, market
    profile, market, err = getters.load_profile_and_market(profile_id, market_id)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    err = risk.check_profile(profile)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    err = risk.check_market(market)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    local order_req = { -- make it later as api_create_order struct 
        profile_id = profile_id,
        market_id = market_id,
        order_type = order_type,
        side = order_side,
        price = order_price,
        size = order_size,
        client_order_id = client_order_id,
        trigger_price = trigger_price,
        size_percent = size_percent,
        time_in_force = time_in_force,
    }
    err = risk.pre_create_order(order_req, market)
    if err ~= nil then
        log.error(PublicAPIError:new('%s: slog=%s', err, json.encode(order_req)))
        return {task = nil, order = nil, error = tostring(err)}
    end


    local order_id = custom_order_id
    if order_id == nil or order_id == '' then
        order_id = setters.next_order_id(market_id)
    end

    local order, task_data = action.pack_create(
        order_req.profile_id,
        order_req.market_id,
        order_id,
        false,
        order_req.side,
        order_req.order_type,
        order_req.size,
        order_req.price,
        order_req.client_order_id,
        order_req.trigger_price,
        order_req.size_percent,
        order_req.time_in_force,

        matching_meta
    )

    res = equeue.put(qname, task_data, profile_id, order.order_id, tostring(order.order_type))
    if res["error"] ~= nil then
        log.error(PublicAPIError:new(res["error"]))
        return {task = nil, order = nil, error = tostring(res["error"])}
    end
    equeue.inc_count(profile_id)

    save_client_order_id(profile_id, client_order_id, order_id, market_id)

    local c = metrics.counter('rabbitx_new_order_counter', 'Count the number of incomming orders')
    c:inc(1)

    local order_res = {
        order_id = order.order_id,
        market_id = order.market_id,
        profile_id = order.profile_id,
        status = order.status,
        order_size = order.size,
        order_price = order.price,
        order_side = order.side,
        order_type = order.order_type,
        is_liquidation = order.is_liquidation,
        client_order_id = order.client_order_id,
        trigger_price = order.trigger_price,
        size_percent = order.size_percent,
        time_in_force = order.time_in_force,
    }

    return {task = res["res"], order = order_res, error = nil}
end


--[[
    1. If order_id in queue cancel imidiatly (CANCELED)
    2. cancel action duplicates check 
    3. Put to the queue (CANCELING)
    
    return: {task, order, err}
--]]
function p.cancel_order(
    profile_id,
    market_id,
    order_id,
    client_order_id
)
    checks('number', 'string', '?string', '?string')
    
    deadman.touch(profile_id)

    if (order_id == nil or order_id == "") and 
        (client_order_id == nil or client_order_id == "") then
            return {task = nil, order = nil, error = "ORDER_ID_OR_CLIENT_ORDER_ID_REQUIRED"}
    end

    local res, e, task, qname

    res = equeue.check_limit(profile_id)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end

    -- IF ORDER STILL IN THE QUEUE - cancel it --
    res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end
    qname = res["res"]

    res = equeue.check_queue_limit(qname)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end


    order_id, client_order_id = get_both(profile_id, order_id, client_order_id)

    -- if we are in the queue - just cancel from the queue
    if order_id ~= nil and order_id ~= "" then
        box.begin()
        res = equeue.find_task(qname, tostring(profile_id), tostring(order_id))
        local existing_task = res["res"]
        local task_order = nil

        local canceled = false
        if existing_task ~= nil and existing_task[d.task_status] == "r" then
            res = equeue.delete(qname, existing_task[1])
            if res["error"] ~= nil then
                box.rollback()
                return {task = nil, order = nil, error = res["error"]}
            end
            canceled = true
            task_order = existing_task["data"] and existing_task["data"]["order"]
        end
        box.commit()

        if canceled == true then
            -- WE CANCEL ORDER
            local canceled_order = {
                order_id = order_id,
                client_order_id = tostring(client_order_id),
                market_id = market_id,
                profile_id = profile_id,
                status = config.params.ORDER_STATUS.CANCELED
            }

            if task_order ~= nil and task_order ~= {} then
                notif.notify_order(profile_id, task_order, config.params.ORDER_STATUS.CANCELED)
            end
            return {task = nil, order = canceled_order, error = nil}    
        end
    end

    -- if we are in the queue - cancel shold have less priority 
    res = equeue.find_task(qname, tostring(profile_id), tostring(order_id))
    local existing_task = res["res"]

    -- If we have both not nil then will cancel by order_id
    local oid
    if order_id ~= nil and order_id ~= "" then
        oid = "cancel@" .. order_id
    else
        oid = "cancel@" .. client_order_id
    end

    res = equeue.find_task(qname, tostring(profile_id), oid)
    if res["res"] ~= nil then
        return {task = nil, order = nil, error = "duplicated cancel action"}
    end

    local profile
    profile, e = getters.get_profile(profile_id)
    if e ~= nil then
        log.error(PublicAPIError:new(e))
        return {task = nil, order = nil, error = tostring(e)}
    end
    if profile == nil then
        local text = "profile_id=" .. tostring(profile_id) .. "not found"
        log.error(PublicAPIError:new(text))
        return {task = nil, order = nil, error = text}
    end

    e = risk.check_profile(profile)
    if e ~= nil then
        log.error(PublicAPIError:new(e))
        return {task = nil, order = nil, error = tostring(e)}
    end

    local order, task_data = action.pack_cancel(profile_id, market_id, order_id, client_order_id)

    -- IF create order exist we put cancel with less priority
    -- it will allow to still have the right priorities for cancel 
    res = equeue.put_after(qname, task_data, profile_id, oid, config.params.ORDER_ACTION.CANCEL, existing_task)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end

    local c = metrics.counter('rabbitx_cancel_order_counter')
    c:inc(1)

    equeue.inc_count(profile_id)
    return {task = res["res"], order = order, error = nil}
end

--[[
    1. If order_id in queue error
    2. Amend action duplicates check 
    3. Put to the queue (AMENDING)
    
    return: {task, order, error}
--]]
function p.amend_order(
    profile_id,
    market_id,
    order_id,
    new_price,
    new_size,
    new_trigger_price,
    new_size_percent
)
    checks('number', 'string', 'string', '?decimal', '?decimal', '?decimal', '?decimal')

    deadman.touch(profile_id)

    local res, err, profile, market

    res = equeue.check_limit(profile_id)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end

    -- FIND ORDER and CHECK PARAMS
    res = rpc.callro_engine(market_id, "get_order_by_id", {order_id})
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end
    local curr_order = order.bind(res["res"])
    if curr_order == nil or #curr_order == 0 then
        return {task = nil, order = nil, error = ERR_ORDER_NOT_FOUND}
    end
    if profile_id ~= curr_order.profile_id then
        return {task = nil, order = nil, error = ERR_NOT_YOUR_ORDER}
    end

    -- IF ORDER STILL IN THE QUEUE - it can't be amended --
    res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end
    local qname = res["res"]

    res = equeue.check_queue_limit(qname)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end

    res = equeue.find_task(qname, tostring(profile_id), order_id)
    if res["res"] ~= nil then
        return {task = nil, order = nil, error = "Processing order can't be amended"}
    end

    -- CAN'T SEND TWO amend for the same order --
    local order_task_id = "amend@" .. order_id
    res = equeue.find_task(qname, tostring(profile_id), order_task_id)
    if res["res"] ~= nil then
        return {task = nil, order = nil, error = "Duplicate amend action"}
    end

    local profile, market
    profile, market, err = getters.load_profile_and_market(profile_id, market_id)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    err = risk.check_profile(profile)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    err = risk.check_market(market)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    local order_req = { -- make it later as api_amend_order struct 
        profile_id = profile_id,
        market_id = market_id,
        price = new_price,
        size = new_size,
        trigger_price = new_trigger_price,
        size_percent = new_size_percent,
    }
    err = risk.pre_amend_order(order_req, curr_order, market)
    if err ~= nil then
        log.error(PublicAPIError:new(err))
        return {task = nil, order = nil, error = tostring(err)}
    end

    -- ORDER NOT IN THE QUEUE - let's create the task ----
    local task_order, task_data = action.pack_amend(
        order_req.profile_id,
        order_req.market_id,
        order_id,
        order_req.size,
        order_req.price,
        order_req.trigger_price,
        order_req.size_percent
    )

    res = equeue.put(qname, task_data, profile_id, order_task_id, config.params.ORDER_ACTION.AMEND)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = tostring(res["error"])}
    end
    equeue.inc_count(profile_id)

    local c = metrics.counter('rabbitx_amend_order_counter')
    c:inc(1)

    local order_res = {
        order_id = task_order.order_id,
        market_id = task_order.market_id,
        profile_id = task_order.profile_id,
        status = task_order.status,
        order_size = task_order.size,
        order_price = task_order.price,
        trigger_price = task_order.trigger_price,
        size_percent = task_order.size_percent,
    }
    return {task = res["res"], order = order_res, error = nil}
end

--[[
    --Dumb implementation of CANCEL ALL
    - Remove all orders from all queues for each market
    - Send cancelall to each market
--]]

function p.cancel_all(
    profile_id,
    is_liquidation
)
    checks('number', 'boolean')

    deadman.touch(profile_id)

    local res, e, profile, market, task, qname

    profile, e = getters.get_profile(profile_id)
    if e ~= nil then
        return {task = nil, order = nil, error = e}
    end

    if is_liquidation == false then
        res = equeue.check_limit(profile_id)
        if res["error"] ~= nil then
            return {task = nil, order = nil, error = res["error"]}
        end

        e = risk.check_profile(profile)
        if e ~= nil then
            return {task = nil, order = nil, error = e}
        end
    end

    for _, market in pairs(config.markets) do
        local market_id = market.id

        res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
        if res["error"] ~= nil then
            return {task = nil, order = nil, error = res["error"]}
        end
        qname = res["res"]

        -- CAN'T SEND TWO cancellall orders --
        local oid = "cancelall"
        res = equeue.find_task(qname, tostring(profile_id), oid)
        if res["res"] == nil then
            local order, task_data = action.pack_cancelall(profile_id, market_id)

            if is_liquidation == true then
                res = equeue.put_with_highest_priority(qname, task_data, tostring(profile_id), oid, config.params.ORDER_ACTION.CANCEL)
                if res["error"] ~= nil then
                    return {task = nil, order = nil, error = res["error"]}
                end
            else
                res = equeue.put(qname, task_data, profile_id, oid, config.params.ORDER_ACTION.CANCEL)
                if res["error"] ~= nil then
                    return {task = nil, order = nil, error = res["error"]}
                end            
            end
        end
    end

    equeue.inc_count(profile_id)
    return {task = nil, order = nil, error = nil}    
end

function p.deadman_create(
    profile_id,
    timeout
)
    checks('number', 'number')

    local res = deadman.replace(profile_id, timeout)

    return {res = res, error = nil}
end

function p.deadman_delete(
    profile_id
)
    checks('number')

    local res = deadman.delete(profile_id)

    return {res = res, error = nil}
end

function p.deadman_get(
    profile_id
)
    checks('number')

    local res = deadman.list(profile_id)

    return {res = res, error = nil}
end

function p.deadman_touch(
    profile_id
)
    checks('number')

    local res = deadman.touch(profile_id)

    return {res = res, error = nil}
end


function p._test_set_rpc(override_rpc)
    rpc = override_rpc
end

function p._test_set_getters(override_getters)
    getters = override_getters    
end

function p._test_set_risk(override_risk)
    risk = override_risk
end


return p