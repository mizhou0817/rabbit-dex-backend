local checks = require('checks')
local log = require('log')
local decimal = require('decimal')

local archiver = require('app.archiver')
local balance = require('app.balance')
local config = require('app.config')
local d = require('app.data')
local ag = require('app.engine.aggregate')
local errors = require('app.lib.errors')
local tick = require('app.lib.tick')
local util = require('app.util')
local time = require('app.lib.time')

require('app.config.constants')
require('app.errcodes')

local RouterError = errors.new_class("ENGINE_ROUTER_ERROR")

local DEFAULT_ROUTER = 0

local router = {}

local global_profile_data = {}

local routable_order_types = {
    config.params.ORDER_TYPE.LIMIT,
    config.params.ORDER_TYPE.MARKET
}
local routable_order_time_in_force = {
    config.params.TIME_IN_FORCE.GTC
}

function router.init_spaces()
    local _

    local router, err = archiver.create('router', { if_not_exists = true }, {
        { name = 'id',           type = 'number' }, -- unique for fast deletion
        { name = 'exchange_id',  type = 'string' },
        { name = 'device',       type = 'string' },
        { name = 'is_api',       type = 'boolean' },        
        { name = 'to_counterparty', type = 'unsigned' },
    }, {
        unique = true,
        parts = { { field = 'id' } },
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(RouterError:new(err))
        error(err)
    end

    router:create_index('by_all', {
        parts = { { field = 'exchange_id' }, { field = 'device' }, { field = 'is_api' },  },
        unique = true,
        if_not_exists = true
    })

    router:create_index('by_counterparty',
        {
            parts = { { field = 'to_counterparty' } },
            unique = false,
            if_not_exists = true
        })

    local unique_cp, err = archiver.create('unique_cp', { if_not_exists = true }, {
        { name = 'to_counterparty', type = 'unsigned' },
    }, {
        unique = true,
        parts = { { field = 'to_counterparty' } },
        if_not_exists = true,
    })

    if err ~= nil then
        log.error(RouterError:new(err))
        error(err)
    end

    local default_cp, err = archiver.create('default_cp', { if_not_exists = true }, {
        { name = 'default', type = 'string' },
        { name = 'to_counterparty', type = 'unsigned' },
    }, {
        unique = true,
        parts = { { field = 'default' } },
        if_not_exists = true,
    })

    local liquidation_cp, err = archiver.create('liquidation_cp', { if_not_exists = true }, {
        { name = 'default', type = 'string' },
        { name = 'to_counterparty', type = 'unsigned' },
    }, {
        unique = true,
        parts = { { field = 'default' } },
        if_not_exists = true,
    })

    local black_list, err = archiver.create('black_list', { if_not_exists = true }, {
        { name = 'profile_id', type = 'unsigned' },
    }, {
        unique = true,
        parts = {{ field = 'profile_id'} },
        if_not_exists = true,
    })

    if err ~= nil then
        log.error(RouterError:new(err))
        error(err)
    end

    local router_system, err = archiver.create('router_system', { if_not_exists = true }, {
        { name = 'id', type = 'unsigned' },
        { name = 'is_active', type = 'boolean' },
        { name = 'range', type = 'decimal' },
        { name = 'max_size', type = 'decimal' },
        { name = 'allow_liquidation', type = 'boolean' },
    }, {
        unique = true,
        parts = {{ field = 'id'} },
        if_not_exists = true,
    })

    if err ~= nil then
        log.error(RouterError:new(err))
        error(err)
    end

    local c_exist = box.space.router_system:get(DEFAULT_ROUTER)
    if c_exist == nil then
        -- CREATE deafult router_system
        _, err = archiver.replace(box.space.router_system, {
            DEFAULT_ROUTER,
            false,
            decimal.new("0.001"),
            decimal.new("100000"),
            false
        })
        if err ~= nil then
            log.error(RouterError:new(err))
            error(err)
        end
    end

end

function router.filter_conditions(matching_meta)
    checks("table")

    local exist = box.space.router.index.by_all:get({
        matching_meta.exchange_id,
        matching_meta.device,
        matching_meta.is_api
    })

    if exist ~= nil then
        return router.only_active_counterparty(exist.to_counterparty)
    end
    
    return nil, nil
end

function router.get_all_counterparties()
    local ids = {}

    for _, item in box.space.unique_cp:pairs() do
        table.insert(ids, item.to_counterparty)
    end

    return ids
end

function router.check_routable(order, market_data)
    checks("table|api_create_order", "table|engine_market")

    local s = box.space.router_system:get(DEFAULT_ROUTER)
    if s == nil then
        local text = "NO_SYSTEM_ROUTER"
        return text
    end

    if not util.is_value_in(order.order_type, routable_order_types) then
        local text = "NOT_ROUTABLE: order_type = " .. tostring(order.order_type)
        return text
    end

    if not util.is_value_in(order.time_in_force, routable_order_time_in_force) then
        local text = "NOT_ROUTABLE: time_in_force = " .. tostring(order.time_in_force)
        return text
    end

    if order.size * order.price > s.max_size then
        local text = "NOT_ROUTABLE: max_size = " .. tostring(s.max_size) .. " exceed = ".. tostring(order.size * order.price)
        return text
    end

    if market_data.best_bid == nil or market_data.best_ask == nil then
        local text = "NOT_ROUTABLE: nil best_bid=" .. tostring(market_data.best_bid) .. "or best_ask".. tostring(market_data.best_ask)
        return text    
    end

    if market_data.best_bid <= 0 or market_data.best_ask <= 0 then
        local text = "NOT_ROUTABLE: best_bid=" .. tostring(market_data.best_bid) .. "best_ask".. tostring(market_data.best_ask)
        return text
    end

    local mid_price = tick.calc_middle_price(market_data.best_ask, market_data.best_bid, market_data.min_tick)
    if mid_price <= 0 then
        local text = "NOT_ROUTABLE: negative_or_zero mid_price=" .. tostring(mid_price)
        return text
    end

    local range = (market_data.best_ask - market_data.best_bid) / mid_price
    if range > s.range then
        local text = "NOT_ROUTABLE: exceed range=" .. tostring(range)
        return text
    end

    -- Is marketable immidiatly
    if order.side == config.params.LONG then
        if order.price < market_data.best_ask then
            local text = "NOT_ROUTABLE: LONG not marketable order price=" .. tostring(order.price) .. " best_ask=" .. tostring(market_data.best_ask)
            return text
        end
    else
        if order.price > market_data.best_bid then
            local text = "NOT_ROUTABLE: SHORT not marketable order price=" .. tostring(order.price) .. " best_bid=" .. tostring(market_data.best_bid)
            return text
        end
    end

    return nil
end

function router.only_active_counterparty(cp_id)
    checks('number')
    local data = router.get_profile_data(cp_id)
    if data == nil or data.cache == nil then
        return nil, "NO_DATA"
    end

    if data.cache[3] ~= config.params.PROFILE_STATUS.ACTIVE then
        return nil, "CP_NOT_ACTIVE"
    end

    return cp_id, nil
end

function router.which_counterparty(is_liquidation, order, market_data, matching_meta)
    checks("?boolean", "table|api_create_order", "table|engine_market", "?table")

    local s = box.space.router_system:get(DEFAULT_ROUTER)
    if s == nil then
        local text = "NO_COUNTERPARTY: NO_SYSTEM_ROUTER"
        log.error(RouterError:new(text))
        return nil, text
    end

    if s.is_active == false then
        local text = "NO_COUNTERPARTY: ROUTER_NOT_ACTIVE"
        log.error(RouterError:new(text))
        return nil, text
    end

    if order.profile_id == nil then
        local text = "NO_COUNTERPARTY: which_counterparty ERROR: profile_id is NIL"
        log.error(RouterError:new(text))
        return nil, text
    end

    local listed = box.space.black_list:get(order.profile_id)
    if listed ~= nil then
        local text = "NO_COUNTERPARTY: BLACK_LISTED profile=" .. tostring(order.profile_id)
        log.error(RouterError:new(text))
        return nil, text
    end

    local err = router.check_routable(order, market_data)
    if err ~= nil then
        log.error(RouterError:new(err))
        return nil, err
    end


    -- we control where to route liquidation flow
    if is_liquidation ~= nil and is_liquidation == true then
        local exist = box.space.liquidation_cp:get("default")
        if exist ~= nil then
            return router.only_active_counterparty(exist.to_counterparty)
        end

        return nil, "NO_DEFAULT_LIQUIDATION_CP"
    end

    -- by liquidation meta will be nil, we can't filter means we don't route
    if matching_meta == nil then
        local text = "NO_COUNTERPARTY: matching_meta nil"
        log.error(RouterError:new(text))
        return nil, text
    end

    if matching_meta.is_pm == true then
        local default = box.space.default_cp:get("default")
        if default == nil then
            local text = "NO_COUNTERPARTY: no default_cp for is_pm = true"
            log.error(RouterError:new(text))
            return nil, text
        end

        return router.only_active_counterparty(default.to_counterparty)
    end

    return router.filter_conditions(matching_meta)
end

function router.save_profile_data(profile_data_ids)
    checks("table")

    global_profile_data = {}

    for _, item in pairs(profile_data_ids) do
        table.insert(global_profile_data, {
            profile_id = item.profile_id,
            cache = item.cache,
            meta = item.meta
        })
    end
end


function router.get_profile_data(profile_id)
    checks("number")

    local res = nil
    for _, item in pairs(global_profile_data) do
        if item.profile_id == profile_id then
            res = {
                cache = item.cache,
                meta = item.meta
            }
            break
        end
    end

    return res
end


function router.all_profile_data()
    return global_profile_data
end

function router.test_add_profile_data(data)
    checks("table")

    table.insert(global_profile_data, {
        profile_id = data.profile_id,
        cache = data.cache,
        meta = data.meta
    })
end

function router.add_route(id, exchange_id, device, is_api, profile_id)
    checks("number", "string", "string", "boolean", "number")

    box.space.unique_cp:delete(profile_id)
    local _, err = archiver.insert(box.space.unique_cp, {profile_id})
    if err ~= nil then
        error(err)
    end

    local _, err = archiver.insert(box.space.router, {
        id,
        exchange_id,
        device,
        is_api,
        profile_id
    })

    if err ~= nil then
        error(err)
    end

    return nil
end

function router.add_desktop(exchange_id, profile_id)
    return router.replace_cp(exchange_id, "other", false, profile_id)
end

function router.add_mobile(exchange_id, profile_id)
    for _, device in pairs({"ios", "android"}) do
        router.replace_cp(exchange_id, device, false, profile_id)
    end

    return nil
end


function router.replace_cp(exchange_id, device, is_api, profile_id)
    checks("string", "string", "boolean", "number")

    local _id, err, exist

    _id = box.space.router.index.primary:max()
    if _id == nil then
        _id = tonumber(1)
    else
        _id = tonumber(_id.id) + 1
    end

    exist = box.space.router.index.by_all:get({exchange_id, device, is_api})
    if exist ~= nil then
        _id = exist.id
        box.space.router:delete(_id)
    end

    err = router.add_route(tonumber(_id), exchange_id, device, is_api, profile_id)
    if err ~= nil then
            error(err)
    end
    
    return nil
end

return router