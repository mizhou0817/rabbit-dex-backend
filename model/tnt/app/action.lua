local msgpack = require('msgpack')
local config = require('app.config')
local checks = require('checks')
local decimal = require('decimal')

local action = {}

function action.wrap_meta(matching_meta)
    if matching_meta == nil then
        return matching_meta
    end

    local res = {
        device = matching_meta[1],
        is_api = matching_meta[2],
        exchange_id = matching_meta[3],
        is_pm = matching_meta[4]
    }
    
    return res
end

function action.pack_execute(
    profile_id,
    market_id,
    order_id)
    checks('number', 'string', 'string')

    local order = {
        order_id = order_id,
        market_id = market_id,
        profile_id = profile_id,
    }
    local task_data = {
        action = config.params.ORDER_ACTION.EXECUTE,
        order = order
    }
    return order, task_data

end

function action.pack_create(
    profile_id,
    market_id,
    order_id,
    is_liquidation,
    order_side,
    order_type,
    order_size,
    order_price,
    client_order_id,
    trigger_price,
    size_percent,
    time_in_force,

    matching_meta
)
    checks('number', 'string', 'string', 'boolean', 'string', 'string', '?decimal', '?decimal', '?string', '?decimal', '?decimal', 'string', '?table|matching_meta')

    -- order in response format
    local order = {
        order_id = order_id,
        market_id = market_id,
        profile_id = profile_id,
        status = config.params.ORDER_STATUS.PROCESSING,
        size = order_size,
        price = order_price,
        side = order_side,
        order_type = order_type,
        is_liquidation = is_liquidation,
        client_order_id = client_order_id,
        trigger_price = trigger_price,
        size_percent = size_percent,
        time_in_force = time_in_force,
    }

    local task_data = {
        action = config.params.ORDER_ACTION.CREATE,
        order = order,
        matching_meta = action.wrap_meta(matching_meta),
    }

    return order, task_data
end

function action.pack_cancel(profile_id, market_id, order_id, client_order_id)
    checks('number', 'string', '?string', '?string')

    if order_id == nil  then
        order_id = ""
    end

    if client_order_id == nil then
        client_order_id = ""
    end

    local order = {
        order_id = order_id,
        market_id = market_id,
        profile_id = profile_id,
        status = config.params.ORDER_STATUS.CANCELING,
        client_order_id = client_order_id
    }

    local task_data = {
        action = config.params.ORDER_ACTION.CANCEL,
        order = order
    }

    return order, task_data
end

function action.pack_amend(
    profile_id,
    market_id,
    order_id,
    new_size,
    new_price,
    new_trigger_price,
    new_size_percent
)
    checks('number', 'string', 'string', '?decimal', '?decimal', '?decimal', '?decimal')

    local order = {
        order_id = order_id,
        market_id = market_id,
        profile_id = profile_id,
        status = config.params.ORDER_STATUS.AMENDING,
        size = new_size,
        price = new_price,
        trigger_price = new_trigger_price,
        size_percent = new_size_percent,
    }

    local task_data = {
        action = config.params.ORDER_ACTION.AMEND,
        order = order
    }

    return order, task_data
end

function action.pack_cancelall(profile_id, market_id)
    checks('number', 'string')

    local order = {
        market_id = market_id,
        profile_id = profile_id,
        status = config.params.ORDER_STATUS.CANCELINGALL
    }

    local task_data = {
        action = config.params.ORDER_ACTION.CANCELALL,
        order = order
    }

    return order, task_data
end

function action.pack_liquidation(liq_action, oid)
    checks("table", "string")

    local task_data = {
        action = config.params.ORDER_ACTION.LIQUIDATE,
        order = liq_action,
        order_id = oid
    }

    return task_data
end


return action