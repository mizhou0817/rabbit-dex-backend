local checks = require('checks')
local json = require('json')
local log = require('log')

local config = require('app.config')
local dml = require('app.dml')
local o = require('app.engine.order')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local tuple = require('app.tuple')
local util = require('app.util')

local err = errors.new_class("notif_error")

local channel_size = 1000

local conditional_orders = {
    config.params.ORDER_TYPE.STOP_LOSS,
    config.params.ORDER_TYPE.TAKE_PROFIT,
    config.params.ORDER_TYPE.STOP_LOSS_LIMIT,
    config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT,
    config.params.ORDER_TYPE.STOP_LIMIT,
    config.params.ORDER_TYPE.STOP_MARKET,
}

local notif = {
    nf_profiles_changed_format = {
        {name = 'profile_id', type = 'number'}
    },
    nf_profiles_changed_type = 'nf_profiles_changed'
}

local function bind_nf_profiles_changed(value)
    return tuple.new(value, notif.nf_profiles_changed_format, notif.nf_profiles_changed_type)
end

local function send_conditional_order(order)
    checks('table|engine_order')

    if util.is_value_in(order.order_type, conditional_orders) then
        local channel = "conditional:" .. order.market_id
        local json_update = json.encode({
            data = {orders = {order}},
        })
        rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    end
end

function notif.init_spaces()
    local nf_private = box.schema.space.create('nf_private', {temporary = true, if_not_exists = true})
    nf_private:format({
        {name = 'data_key', type = 'string'},
        {name = 'timestamp', type = 'number'},
        {name = 'market_id', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'nf_type', type = 'string'},
        {name = 'data', type = '*'}
    })
    nf_private:create_index('primary', {
        unique = true,
        parts = {{field = 'data_key'}},
        if_not_exists = true })

    nf_private:create_index('profile_id', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })
    

    local nf_trades = box.schema.space.create('nf_trades', {temporary = true, if_not_exists = true})
    nf_trades:format({
        {name = 'id', type = 'string'},
        {name = 'market_id', type = 'string'},        
        {name = 'timestamp', type = 'number'},
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
        {name = 'liquidation', type = 'boolean'},
        {name = 'taker_side', type = 'string'}
    })
    nf_trades:create_index('primary', {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true })    

    nf_trades:create_index('timestamp', {
        unique = false,
        parts = {{field = 'timestamp'}},
        if_not_exists = true })     
    

    local nf_bids = box.schema.space.create('nf_bids', {temporary = true, if_not_exists = true})
    nf_bids:format({
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
    })
    nf_bids:create_index('primary', {
        unique = true,
        parts = {{field = 'price'}},
        if_not_exists = true })    

        
    local nf_asks = box.schema.space.create('nf_asks', {temporary = true, if_not_exists = true})
    nf_asks:format({
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
    })
    nf_asks:create_index('primary', {
        unique = true,
        parts = {{field = 'price'}},
        if_not_exists = true }) 


    local nf_profiles_changed = box.schema.space.create('nf_profiles_changed', {temporary = true, if_not_exists = true})
    nf_profiles_changed:format(notif.nf_profiles_changed_format)
    nf_profiles_changed:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })    
end

function notif.clear()
    -- TEMPORARY SPACES that are used for notifications
    dml.truncate(box.space.nf_private)
    dml.truncate(box.space.nf_trades)
    dml.truncate(box.space.nf_bids)
    dml.truncate(box.space.nf_asks)
    dml.truncate(box.space.nf_profiles_changed)
end

function notif.add_trade(market_id, trade)
    checks("string", "cdata")

    box.space.nf_trades:replace(trade)
end

-- FOR any profiles whose order/entry/position we change, we add for updating meta
function notif.add_profile(profile_id)
    box.space.nf_profiles_changed:replace({profile_id})
end

-- not used?
function notif.add_market(market_id, data_key, nf_type, data)
    local id = tostring(nf_type) .. tostring(data_key)

    box.space.nf_market:replace({id, market_id, nf_type, data})
end

function notif.add_private(market_id, data_key, timestamp, profile_id, nf_type, data)
    -- data should implement tomap()
    checks("string", "string", "number", "number", "string", "?")
    local id = tostring(nf_type) .. tostring(data_key)

    box.space.nf_private:replace({id, timestamp, market_id, profile_id, nf_type, data:tomap({names_only=true})})
end

function notif.notify_market(market_id)
    local market = box.space.market:get(market_id)
    if market == nil then
        local text = "NO market space for market_id=" .. tostring(market_id)
        log.error(err:new(text))
        return
    end

    local channel = "market:" .. tostring(market_id)
    local json_update = json.encode({data=market:tomap({names_only=true})})
    
    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    
    json_update = nil
end

function notif.notify_account(market_id)
    checks('string')

    local id_to_data = {}
    local conditional = {
        orders = {},
    }
    for _, item in box.space.nf_private:pairs(nil, {iterator = box.index.ALL}) do
        local profile_id = tostring(item.profile_id)

        id_to_data[profile_id] = id_to_data[profile_id] or {id = item.profile_id}

        local update_type = item.nf_type

        if update_type == "position" then
            id_to_data[profile_id].positions = id_to_data[profile_id].positions or {}
            table.insert(id_to_data[profile_id].positions, item.data)
        elseif update_type == "order" then
            id_to_data[profile_id].orders = id_to_data[profile_id].orders or {}
            table.insert(id_to_data[profile_id].orders, item.data)

            local ot = config.params.ORDER_TYPE
            if util.is_value_in(item.data.order_type, conditional_orders) then
                table.insert(conditional.orders, item.data)
            end
        elseif update_type == "fill" then
            id_to_data[profile_id].fills = id_to_data[profile_id].fills or {}
            table.insert(id_to_data[profile_id].fills, item.data)
        else
            log.error("notify: unknown type=%s", update_type)
        end
    end

    for profile_id, data in pairs(id_to_data) do
        local channel = "account@" .. tostring(profile_id)
        local json_update = json.encode({data=data})
        rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    end

    --TODO: move conditional part to some common part
    if #conditional.orders > 0 then
        local channel = "conditional:" .. market_id
        local json_update = json.encode({data=conditional})
        rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
        conditional.orders = {}
    end
end

function notif.notify_to_table(market_id, sequence)
    checks("string", "number")

    local tm = time.now()

    local update = {
        orderbook = {
            market_id=market_id,
            timestamp=tm,
            sequence=sequence,
            bids={},
            asks={},    
        },
        trades={},
        market={},
    }

    for _, bid in box.space.nf_bids:pairs(nil, {iterator="ALL"}) do
        table.insert(update.orderbook.bids, bid:totable())
    end

    for _, ask in box.space.nf_asks:pairs(nil, {iterator="ALL"}) do
        table.insert(update.orderbook.asks, ask:totable())
    end

    for _, trade in box.space.nf_trades.index.timestamp:pairs(nil, {iterator="REQ"}) do
        table.insert(update.trades, trade:tomap({names_only=true}))
    end

    -- Send market updates
    local market = box.space.market:get(market_id)
    if market ~= nil and market.last_update_sequence >= sequence then
        update.market = {
            id = market_id,
            best_bid = market.best_bid,
            best_ask = market.best_ask,
            market_price = market.market_price,
            index_price = market.index_price,
            last_trade_price = market.last_trade_price,
            last_trade_price_24high = market.last_trade_price_24high,
            last_trade_price_24low = market.last_trade_price_24low,
            adv = market.average_daily_volume
        }
    end

    return update
end

function notif.notify_from_table(notifications, market_id)
    checks("table", "string")

    local channel, json_update
    for _, nf in ipairs(notifications) do
        if nf.orderbook ~= nil and (#nf.orderbook.bids > 0 or #nf.orderbook.asks > 0) then
            channel = "orderbook:" .. market_id

            json_update = json.encode({data=nf.orderbook})
            rpc.callrw_pubsub_publish(channel, json_update, 100, channel_size, 100)
        end

        if nf.trades ~= nil and #nf.trades > 0 then
            channel = "trade:" .. market_id
            json_update = json.encode({data=nf.trades})
            rpc.callrw_pubsub_publish(channel, json_update, 100, channel_size, 100)
        end

        if nf.market ~= nil and next(nf.market) ~= nil then
            channel = "market:" .. market_id    
            json_update = json.encode({data=nf.market})
            rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
        end
    end
end

function notif.notify(market_id, sequence)
    --[[
        Json format
        1. Send bid/asks changes for orderbook
        2. Send trades
        3. Send market updates
        4. Send individual account changes  
    --]]

    local tm = time.now()
    -- Send bid/asks changes for orderbook
    local channel = "orderbook:" .. market_id
    local update = {
        market_id=market_id,
        timestamp=tm,
        sequence=sequence,
        bids={},
        asks={}
    }

    for _, bid in box.space.nf_bids:pairs(nil, {iterator="ALL"}) do
        table.insert(update.bids, bid:totable())
    end

    for _, ask in box.space.nf_asks:pairs(nil, {iterator="ALL"}) do
        table.insert(update.asks, ask:totable())
    end

    local json_update = json.encode({data=update})
    rpc.callrw_pubsub_publish(channel, json_update, 100, channel_size, 100)
    update = nil

    update = {}
    for _, trade in box.space.nf_trades.index.timestamp:pairs(nil, {iterator="REQ"}) do
        table.insert(update, trade:tomap({names_only=true}))
    end

    if #update > 0 then
        channel = "trade:" .. market_id
        json_update = json.encode({data=update})
        rpc.callrw_pubsub_publish(channel, json_update, 100, channel_size, 100)
    end
    update = nil

    -- Send market updates
    local market = box.space.market:get(market_id)
    if market ~= nil and market.last_update_sequence >= sequence then
        channel = "market:" .. market_id
        update = {
            id = market_id,
            best_bid = market.best_bid,
            best_ask = market.best_ask,
            market_price = market.market_price,
            index_price = market.index_price,
            last_trade_price = market.last_trade_price,
            last_trade_price_24high = market.last_trade_price_24high,
            last_trade_price_24low = market.last_trade_price_24low,
            adv = market.average_daily_volume
        }

        json_update = json.encode({data=update})
        rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
        update = nil
    end

    notif.notify_account(market_id)
end

function notif.notify_position(profile_id, position_id)
    local exist = box.space.position:get(position_id)
    if exist == nil then
        local text = "not found position_id=" .. tostring(position_id)
        return text
    end

    local channel = "account@" .. tostring(profile_id)
    local update = {
        id = profile_id,
        positions = {exist:tomap({names_only=true})}
    }

    local json_update = json.encode({data=update})
    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)

    update = nil

    return nil
end

function notif.notify_order(profile_id, order_id)
    local exist = box.space.order:get(order_id)

    if exist == nil then
        local text = "not found order_id=" .. tostring(order_id)
        return text
    end

    local channel = "account@" .. tostring(profile_id)
    local order = exist:tomap({names_only=true})
    local update = {
        id = profile_id,
        orders = {order}
    }
    local json_update = json.encode({data=update})

    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    update = nil

    send_conditional_order(order)

    return nil
end

function notif.bid(price, size)
    local res, e = box.space.nf_bids:replace{price, size}
    if e ~= nil then
        log.error(err:new(e))
    end
end

function notif.ask(price, size)
    local res, e = box.space.nf_asks:replace{price, size}
    if e ~= nil then
        log.error(err:new(e))
    end
end

function notif.send_profile_notification(profile_id, n_type, n_title, n_description)
    local channel = "account@" .. tostring(profile_id)
    local update = {
        profile_notifications = {{
            type = n_type,
            title = n_title,
            description = n_description    
        }}
    }
    local json_update = json.encode({data=update})

    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    update = nil

    return nil
end

function notif.send_profile_notification_with_order(profile_id, n_type, n_title, n_description, api_order)
    checks('number', '?', '?', '?', 'table')
    -- <order> can be any type of api_?_order, need some solution to deal with it
    -- while using minimal subset of order object

    local res = o.get_order_by_oneof(api_order.order_id, api_order.client_order_id)
    local order = res.res
    if res.error ~= nil then
        local timestamp = time.now()
        order = o.new(
            api_order.order_id or '',
            api_order.profile_id,
            api_order.market_id,
            '',
            config.params.ORDER_STATUS.REJECTED,
            ZERO,
            ZERO,
            ZERO,
            ZERO,
            '',
            timestamp,
            tostring(res.error),
            api_order.client_order_id or '',
            ZERO,
            ZERO,
            '',
            timestamp,
            timestamp
        )
    end

    local channel = "account@" .. tostring(profile_id)
    local update = {
        profile_notifications = {{
            type = n_type,
            title = n_title,
            description = n_description,
        }},
        id = profile_id,
        orders = {order:tomap({names_only=true})}
    }
    local json_update = json.encode({data=update})

    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    update = nil

    if order ~= nil then
        send_conditional_order(order)
    end

    return nil

end

function notif.changed_profiles_iterator()
    local iter, iter_param, iter_state = box.space.nf_profiles_changed.index.primary:pairs()

    return function(param, state)
        local state, value = iter(param, state)
        return state, value and bind_nf_profiles_changed(value)
    end,
    iter_param,
    iter_state
end

--TODO: refactor as notif:new(rpc), notif:notify_xxx() and remove this func
function notif._test_set_rpc(override_rpc)
    rpc = override_rpc
end

function notif._test_set_time(override_time)
    time = override_time
end

return notif
