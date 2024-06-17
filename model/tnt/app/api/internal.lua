local checks = require('checks')
local log = require('log')

local action = require('app.action')
local getters = require('app.api.getters')
local risk = require('app.api.risk')
local setters = require('app.api.setters')
local config = require('app.config')
local d = require('app.data')
local equeue = require('app.enginequeue')
local errors = require('app.lib.errors')
local rpc = require('app.rpc')
local util = require('app.util')

local InternalError = errors.new_class("API_INTERNAL")

local i = {}

local DEFAULT_TIMEOUT = 0

function i.next_task(params)
    checks('table')

    local e, res
    if params.market_id == nil or params.queue_type == nil then
        return {res = nil, error = "market_id and queue_type required"}
    end

    res = equeue.which_qname(params.market_id, params.queue_type)
    if res["error"] ~= nil then
        log.error(params)
        log.error(InternalError:new(res["error"]))
        return {res = nil, error=res["error"]}
    end

    if params.task_id ~= nil then
        local r = equeue.ack(res["res"], params.task_id)
        if r["error"] ~= nil then
           log.error(params)
           log.error(InternalError:new(r["error"]))
            return {res=nil, error=r["error"]}
        end
    end

    res = equeue.take(res["res"])
    if res["error"] ~= nil then
        log.error(params)
        log.error(InternalError:new(res["error"]))
        return {res = nil, error = res["error"]}
    end
 
    return res
end

function i.ack_task(params)
    checks('table')

    local e, res

    if params.market_id == nil or params.queue_type == nil or params.task_id == nil then
        return {res = nil, error = "market_id, queue_type, task_id required"}
    end

    res = equeue.which_qname(params.market_id, params.queue_type)
    if res["error"] ~= nil then
        return {res=nil, error=res["error"]}
    end

    res = equeue.ack(res["res"], params.task_id)
    if res["error"] ~= nil then
        return {res=nil, error=res["error"]}
    end

    return res
end

function i.create_profile(profile_type, status, wallet)
    checks("string", "string", "string")

    local res = rpc.callrw_profile("create", {profile_type, status, wallet})
    if res["error"] ~= nil then
        return {res = nil, error = res["error"]}
    end

    return res
end


function i.update_index_price(market_id, index_price)
    checks('string', 'number')

    local res = rpc.callrw_engine(market_id, "update_index_price", {market_id, index_price})
    if res["error"] ~= nil then
        return {res = nil, error = res["error"]}
    end

    return res
end


function i.white_list(profile_id)
    checks('number')

    local res = equeue.white_list(profile_id)

    return {res=res, error=nil}
end


function i.high_priority_cancell_all(
    profile_id,
    is_liquidation
)
    checks('number', 'boolean')

    local res, e, profile, market, task, qname

    profile, e = getters.get_profile(profile_id)
    if e ~= nil then
        return {task = nil, order = nil, error = e}
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

            res = equeue.put_with_highest_priority(qname, task_data, tostring(profile_id), oid, config.params.ORDER_ACTION.CANCEL)
            if res["error"] ~= nil then
                return {task = nil, order = nil, error = res["error"]}
            end
        end
    end

    equeue.inc_count(profile_id)
    return {task = nil, order = nil, error = nil}    
end
    

function i.queue_liq_actions(actions)
    checks('table')

    local res, qname

    if #actions == 0 then
        return {res=nil, error=nil}
    end

    local first_action = actions[1]
    local profile_id = first_action[d.liq_action_trader_id]

    -- Caller need to cancel it, as it's hard to give 100% guarantee of cancel before without sync request
    --[[
    res = i.high_priority_cancell_all(profile_id, true)
    if res["error"] ~= nil then
        return {res=nil, error=res["error"]}
    end
    --]]


    for _, a in ipairs(actions) do    
        local market_id = a[d.liq_action_market_id]

        -- risk check of the liquidation order
        local market, e
        market, e = getters.get_market(market_id)
        if e ~= nil then
            return {res = nil, error = e}
        end
    
        e = risk.liquidation_order_check(market, a[d.liq_action_kind])
        if e == nil then
            res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
            if res["error"] ~= nil then
                log.error(InternalError:new(res["error"]))
                return {res = nil, error = res["error"]}
            end
            qname = res["res"]
            
            local oid = setters.next_order_id(market_id)    
            local task_data = action.pack_liquidation(a, oid)
        
            res = equeue.put_with_highest_priority(qname, task_data, tostring(profile_id), oid, "liquidation")
            if res["error"] ~= nil then
                log.error(InternalError:new(res["error"]))
                return {res = nil, error = res["error"]}
            end    
        else 
            log.warn(e)
        end
    end


    return {res=res, error=nil}
end


function i.is_cancel_all_accepted(profile_id)
    checks('number')

    local is_accepted, res, qname

    is_accepted = true

    -- If any cancelall order still in the queue then it's not accepted
    for _, market in pairs(config.markets) do
        local market_id = market.id

        res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
        if res["error"] ~= nil then
            return {res=nil, error = res["error"]}
        end
        qname = res["res"]

        local oid = "cancelall"
        res = equeue.find_task(qname, tostring(profile_id), oid)
        if res["res"] ~= nil then
            is_accepted = false
        end
    end

    return {res=is_accepted, error=nil}
end

function i.clear_coid_table()

    local count = 0

    for _, pt in box.space.used_client_order_id:pairs() do
        if pt.market_id ~= "" then
            local res = rpc.callro_engine(pt.market_id, "check_coid_for_reuse", {pt.profile_id, pt.client_order_id})
            if res["error"] == nil and res["res"] == pt.client_order_id then
                box.space.used_client_order_id:delete{pt.profile_id, pt.client_order_id}
            end
        end

        count = util.safe_yield(count, 1000)
    end
end

-- execute conditional order with <placed> state
function i.execute_order(profile_id, market_id, order_id)
    checks('number', 'string', 'string')
    local  err
 
    local profile, market
    profile, market, err = getters.load_profile_and_market(profile_id, market_id)
    if err ~= nil then
        return {task = nil, order = nil, error = err}
    end

    err = risk.check_profile(profile)
    if err ~= nil then
        return {task = nil, order = nil, error = err}
    end

    err = risk.check_market(market)
    if err ~= nil then
        return {task = nil, order = nil, error = err}
    end

    local res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end
    local qname = res["res"]

    local order, task_data = action.pack_execute(profile_id, market_id, order_id)

    res = equeue.put(qname, task_data, profile_id, order_id, config.params.ORDER_ACTION.EXECUTE)
    if res["error"] ~= nil then
        return {task = nil, order = nil, error = res["error"]}
    end

    equeue.inc_count(profile_id)
    return {task = task_data, order = order, error = nil}
end

function i._test_set_rpc(override_rpc)
    rpc = override_rpc
end

return i
