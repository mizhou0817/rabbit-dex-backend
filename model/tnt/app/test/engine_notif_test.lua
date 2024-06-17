local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')

local a = require('app.archiver')
local engine = require('app.engine.engine')
local market = require('app.engine.market')
local notif = require('app.engine.notif')
local o = require('app.engine.order')
local position = require('app.engine.position')
local trade = require('app.engine.trade')
local time = require('app.lib.time')

require('app.config.constants')

local g = t.group('engine.notif')

local work_dir = fio.tempdir()

local mock_rpc = {call={}}
function mock_rpc.callrw_pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    table.insert(mock_rpc.call, {channel, json_data})
end

local mock_time = {}
function mock_time.now()
    return 1681343466169600
end

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
    notif._test_set_rpc(mock_rpc)
    notif._test_set_time(mock_time)
    o._test_set_time(mock_time)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    t.assert_is_not(a.init_sequencer('shard'), nil)
    market.init_spaces({
        id = 'BTC-USD',
        status = 'active',
        min_initial_margin = ONE,
        forced_margin = ONE,
        liquidation_margin = ONE,
        min_tick = ONE,
        min_order = ONE,
    })
    position.init_spaces()
    trade.init_spaces()
    notif.init_spaces()
    o.init_spaces()

    mock_rpc.call = {}
    cg.params = {}
end)

g.after_each(function(cg)
    notif.clear()
    box.sequence.shard_archive_id_sequencer:drop()
end)

g.test_notify = function(cg)
    notif.ask(ONE, ONE)
    notif.bid(ONE, ONE)

    local market_id = 'BTC-USD'
    local tm = mock_time.now()

    --profile
    local profile_id = 234
    notif.add_profile(profile_id)

    -- position
    local trader_id, size, entry_side, price = 123, decimal.new(2), "long", decimal.new(100)
    local pos, err = position.create(market_id, trader_id, size, entry_side, price)
    t.assert_is(err, nil)

    notif.add_private(market_id, pos.id, tm, pos.profile_id, "position", pos)

    --trade'n'fill
    local trade_id = 'tr-1000'
    local maker_fill_id = 'tr-1001'
    local taker_fill_id = 'tr-1002'
    local maker_id, maker_side, maker_fee, maker_order_id = 200, 'long', ZERO, 'order-100'
    local taker_id, taker_side, taker_fee, taker_order_id = 300, 'short', ZERO, 'order-101'
    local is_liquidation, fill_price, fill_size = false, ONE, ONE

    local trade_item, err = a.insert(box.space.trade, {
        trade_id,
        market_id,
        tm,
        fill_price,
        fill_size,
        is_liquidation,
        taker_side,
    })
    t.assert_is(err, nil)

    local maker_fill, err = a.insert(box.space.fill, {
        maker_fill_id,
        maker_id,
        market_id,
        maker_order_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        maker_side,

        true,
        maker_fee,
        is_liquidation,
        "coid_1"
    })
    t.assert_is(err, nil)

    local taker_fill, err = a.insert(box.space.fill, {
        taker_fill_id,
        taker_id,
        market_id,
        taker_order_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        taker_side,

        false,
        taker_fee,
        is_liquidation,
        "coid_2"
    })
    t.assert_is(err, nil)

    notif.add_trade(market_id, trade_item)
    notif.add_private(market_id, tostring(taker_fill_id), tm, taker_id, "fill", taker_fill)
    notif.add_private(market_id, tostring(maker_fill_id), tm, maker_id, "fill", maker_fill)

    local seq = 666
    notif.notify(market_id, seq)

    local expected = {
        {
            'orderbook:BTC-USD',
            '{"data":{"timestamp":1681343466169600,"sequence":666,"market_id":"BTC-USD","asks":[["1","1"]],"bids":[["1","1"]]}}',
        },
        {
            'trade:BTC-USD',
            '{"data":[{"timestamp":1681343466169600,"price":"1","size":"1","id":"tr-1000","liquidation":false,"market_id":"BTC-USD","taker_side":"short"}]}',
        },
        {
            'account@123',
            '{"data":{"positions":[{"size":"2","id":"pos-BTC-USD-tr-123","market_id":"BTC-USD","profile_id":123,"entry_price":"100","unrealized_pnl":"0","liquidation_price":"0","notional":"0","fair_price":"0","side":"long","margin":"0"}],"id":123}}',
        },
        {
            'account@200',
            '{"data":{"fills":[{"trade_id":"tr-1000","price":"1","size":"1","id":"tr-1001","market_id":"BTC-USD","client_order_id":"coid_1","profile_id":200,"timestamp":1681343466169600,"order_id":"order-100","side":"long","is_maker":true,"liquidation":false,"fee":"0","archive_id":4,"shard_id":"shard"}],"id":200}}',
        },
        {
            'account@300',
            '{"data":{"fills":[{"trade_id":"tr-1000","price":"1","size":"1","id":"tr-1002","market_id":"BTC-USD","client_order_id":"coid_2","profile_id":300,"timestamp":1681343466169600,"order_id":"order-101","side":"short","is_maker":false,"liquidation":false,"fee":"0","archive_id":5,"shard_id":"shard"}],"id":300}}',
        },
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end
end

g.test_notify_with_order = function(cg)
    local expected = {
        {
            'account@123456',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0","order_type":"","trigger_price":"0","profile_id":123456,"timestamp":1681343466169600,"total_filled_size":"0","price":"0","size":"0","status":"rejected","market_id":"BTC-USD","client_order_id":"","reason":"ORDER_NOT_FOUND","time_in_force":"","created_at":1681343466169600,"id":"ID-1","side":"","size_percent":"0"}],"profile_notifications":[{"type":"type","title":"title","description":"description"}],"id":123456}}',
        },
        {
            'account@123456',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"1","order_type":"limit","trigger_price":"0","profile_id":123456,"timestamp":1681343466169600,"total_filled_size":"0","price":"1","size":"1","status":"open","market_id":"BTC-USD","client_order_id":"","reason":"","time_in_force":"gtc","created_at":1681343466169600,"id":"BTC-100","side":"long","size_percent":"0"}],"profile_notifications":[{"type":"type","title":"title","description":"description"}],"id":123456}}',
        },
        {
            'account@123456',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"1","order_type":"limit","trigger_price":"0","profile_id":123456,"timestamp":1681343466169600,"total_filled_size":"0","price":"1","size":"1","status":"open","market_id":"BTC-USD","client_order_id":"CUSTOM-100","reason":"","time_in_force":"gtc","created_at":1681343466169600,"id":"","side":"long","size_percent":"0"}],"profile_notifications":[{"type":"type","title":"title","description":"description"}],"id":123456}}',
        },
    }

    local api_order = { order_id = 'ID-1', client_order_id = nil, market_id = 'BTC-USD', profile_id = 123456, status = ''}
    notif.send_profile_notification_with_order(api_order.profile_id, 'type', 'title', 'description', api_order)

    local err, order
    order, err = o.create(
        'BTC-100',
        123456,
        'BTC-USD',
        'limit',
        ONE,
        ONE,
        ONE,
        'long',
        '',
        ZERO,
        ZERO,
        'gtc',
        false
    )
    t.assert_is(err, nil)

    api_order = { order_id = order.id, client_order_id = order.client_order_id, market_id = order.market_id, profile_id = order.profile_id, status = order.status}
    notif.send_profile_notification_with_order(api_order.profile_id, 'type', 'title', 'description', api_order)

    order, err = o.create(
        '',
        123456,
        'BTC-USD',
        'limit',
        ONE,
        ONE,
        ONE,
        'long',
        'CUSTOM-100',
        ZERO,
        ZERO,
        'gtc',
        false
    )
    t.assert_is(err, nil)
    api_order = { order_id = order.id, client_order_id = order.client_order_id, market_id = order.market_id, profile_id = order.profile_id, status = order.status}
    notif.send_profile_notification_with_order(api_order.profile_id, 'type', 'title', 'description', api_order)

    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end

end
