local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')

local a = require('app.archiver')
local engine = require('app.engine.engine')
local market = require('app.engine.market')
local notif = require('app.engine.notif')
local position = require('app.engine.position')
local profile = require('app.engine.profile')
local trade = require('app.engine.trade')
local time = require('app.lib.time')
local balance = require('app.balance')
local config = require('app.config')
local o = require('app.engine.order')
local ob = require('app.engine.orderbook')
local ag = require('app.engine.aggregate')
local candles = require('app.engine.candles')
local action_creator = require('app.action')
local json = require("json")
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')
local router = require('app.engine.router')

require('app.config.constants')

local g = t.group('order.ping_limit')

local work_dir = fio.tempdir()

local mock_rpc = {call={}}
function pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    table.insert(mock_rpc.call, {channel, json_data})
end

local function mock_handle_task(order)
    notif.clear()
    local res = engine._handle_create(order,{})
    notif.clear()

    return res
end

local mock_time = {}
function mock_time.now()
    return 1681343466169600
end

local risk = {}
function risk.post_match(_market_id, profile_data, profile_id, position)
    return "FAKE_ERROR"
end

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
    notif._test_set_rpc(mock_rpc)
    notif._test_set_time(mock_time)
    o._test_set_time(mock_time)
    engine._test_set_time(mock_time)
    engine._test_set_post_match(risk.post_match)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    t.assert_is_not(a.init_sequencer('shard'), nil)
    balance.init_spaces("BTC-USD")
    local MIN_TICK = ONE
    local MIN_ORDER = decimal.new("0.1")

    local market_data = {
        id = "BTC-USD",
        status = 'active',
        min_initial_margin = ONE,
        forced_margin = ONE,
        liquidation_margin = ONE,
        min_tick = MIN_TICK,
        min_order = MIN_ORDER,
    }
    market.init_spaces(market_data)
    router.init_spaces()

    position.init_spaces()
    trade.init_spaces()
    notif.init_spaces()
    profile.init_spaces()

    o.init_spaces()
    ob.init_spaces(market_data)

    ag.init_spaces()
    balance.init_spaces(0)
    notif.init_spaces()
    position.init_spaces()
    candles.init_spaces()

    engine.init(market_data.id, MIN_TICK, MIN_ORDER)

    mock_rpc.call = {}
    cg.params = {}
end)

g.after_each(function(cg)
    notif.clear()
    box.sequence.shard_archive_id_sequencer:drop()
    box.space.position:truncate()
    box.space.profile_meta:truncate()
    box.space.order:truncate()
    box.space.orderbook:truncate()
    box.space.bids_to_size:truncate()
    box.space.asks_to_size:truncate()
    box.space.trader_order_to_notional:truncate()
    balance.test_clear_spaces()
end)

g.test_ping_limit = function(cg)
    local market_id = 'BTC-USD'
    local tm = mock_time.now()

    local fair_price = decimal.new(100)
    market.update_fair_price(market_id, fair_price)

    --profile
    local profile_id = 234
    local err = profile.ensure_meta(profile_id, market_id)
    t.assert_is(err, nil)
    
    local start_sequence = tonumber(ob.sequence:current())
    local order, res

    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@1"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("100"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is(res, nil)

    -- Create best_ask
    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@2"),
        true,
        config.params.SHORT,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("104"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is(res, nil)
    

    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@111"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.PING_LIMIT,
        decimal.new("0.1"),
        decimal.new("1000"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    

    res = mock_handle_task(order)
    t.assert_is(res, nil)

    local expected = {
        {
        'orderbook:BTC-USD',
        '{"data":{"timestamp":1681343466169600,"sequence":1,"market_id":"BTC-USD","asks":[],"bids":[["100","0.1"]]}}',
        },
        {
        'market:BTC-USD',
        '{"data":{"id":"BTC-USD","last_trade_price":"0","index_price":"0","best_ask":"0","best_bid":"100","market_price":"100"}}',
        },
        {
        'account@234',
        '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.0","price":"100","size":"0.1","status":"open","market_id":"BTC-USD","client_order_id":"","reason":"liquidation","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@1","side":"long","created_at":1681343466169600}],"id":234}}',
        },
        {
        'orderbook:BTC-USD',
        '{"data":{"timestamp":1681343466169600,"sequence":2,"market_id":"BTC-USD","asks":[["104","0.1"]],"bids":[]}}',
        },
        {
        'market:BTC-USD',
        '{"data":{"id":"BTC-USD","last_trade_price":"0","index_price":"0","best_ask":"104","best_bid":"100","market_price":"102"}}',
        },
        {
        'account@234',
        '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.0","price":"104","size":"0.1","status":"open","market_id":"BTC-USD","client_order_id":"","reason":"liquidation","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@2","side":"short","created_at":1681343466169600}],"id":234}}',
        },
        {
        'orderbook:BTC-USD',
        '{"data":{"timestamp":1681343466169600,"sequence":3,"market_id":"BTC-USD","asks":[],"bids":[["103","0.1"]]}}',
        },
        {
        'market:BTC-USD',
        '{"data":{"id":"BTC-USD","last_trade_price":"0","index_price":"0","best_ask":"104","best_bid":"103","market_price":"103.5"}}',
        },
        {
        'orderbook:BTC-USD',
        '{"data":{"timestamp":1681343466169600,"sequence":4,"market_id":"BTC-USD","asks":[],"bids":[["103","0"]]}}',
        },
        {
        'trade:BTC-USD',
        '{"data":[{"timestamp":1681343466169600,"price":"103","size":"0.1","id":"BTC-USD-0","liquidation":true,"market_id":"BTC-USD","taker_side":"short"}]}',
        },
        {
        'market:BTC-USD',
        '{"data":{"id":"BTC-USD","last_trade_price":"103","index_price":"0","best_ask":"104","best_bid":"100","market_price":"102"}}',
        },
        {
        'account@234',
        '{"data":{"fills":[{"trade_id":"BTC-USD-0","price":"103","size":"0.1","id":"BTC-USD-1","market_id":"BTC-USD","client_order_id":"","profile_id":234,"timestamp":1681343466169600,"order_id":"BTC-USD@111","side":"long","is_maker":true,"liquidation":true,"fee":"-0.0","archive_id":21,"shard_id":"shard"},{"trade_id":"BTC-USD-0","price":"103","size":"0.1","id":"BTC-USD-2","market_id":"BTC-USD","client_order_id":"","profile_id":234,"timestamp":1681343466169600,"order_id":"BTC-USD@111-pong","side":"short","is_maker":false,"liquidation":true,"fee":"-0.00721","archive_id":22,"shard_id":"shard"}],"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"ping_limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.1","price":"103","size":"0.0","status":"closed","market_id":"BTC-USD","client_order_id":"","reason":"liquidation","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@111","side":"long","created_at":1681343466169600},{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"ping_limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.1","price":"103","size":"0","status":"closed","market_id":"BTC-USD","client_order_id":"","reason":"liquidation","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@111-pong","side":"short","created_at":1681343466169600}],"positions":[{"size":"0","id":"pos-BTC-USD-tr-234","market_id":"BTC-USD","profile_id":234,"entry_price":"103","unrealized_pnl":"0","liquidation_price":"0","notional":"0","fair_price":"0","side":"long","margin":"0"}],"id":234}}',
        },
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end

    mock_rpc.call = {}

    local sequence_before = tonumber(ob.sequence:current())

    -- Let's create order with fake post_match
    -- is_liquidation = false, cuz we want post_match
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@113"),
        false,
        config.params.LONG,
        config.params.ORDER_TYPE.PING_LIMIT,
        decimal.new("0.1"),
        decimal.new("1000"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is_not(res, nil)

    expected = {
        {
            'account@234',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"ping_limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.0","price":"103","size":"0.1","status":"rejected","market_id":"BTC-USD","client_order_id":"","reason":"FAKE_ERROR","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@113","side":"long","created_at":1681343466169600}],"id":234}}'
        }
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end
    
    local sequence_after = tonumber(ob.sequence:current())
    t.assert_equals(sequence_before, sequence_after)


    -- let's test revert scenarios
    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@3"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("104"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is(res, nil)

    mock_rpc.call = {}

    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@112"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.PING_LIMIT,
        decimal.new("0.1"),
        decimal.new("1000"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is_not(res, nil)
    

    expected = {
        {
            'account@234',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"ping_limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.0","price":"1000","size":"0.1","status":"rejected","market_id":"BTC-USD","client_order_id":"","reason":"BEST_ASK_ZERO","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@112","side":"long","created_at":1681343466169600}],"id":234}}'
        }
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end
    


    -- let's test revert scenarios
    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@4"),
        true,
        config.params.SHORT,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("110"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is(res, nil)

    mock_rpc.call = {}

    -- is_liquidation = true, cuz we don't want post_match check in this test
    order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@115"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.PING_LIMIT,
        decimal.new("0.1"),
        decimal.new("1000"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(order)
    t.assert_is_not(res, nil)

    log.info(mock_rpc.call)

    expected = {
        {
            'account@234',
            '{"data":{"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"ping_limit","trigger_price":"0","profile_id":234,"timestamp":1681343466169600,"total_filled_size":"0.0","price":"1000","size":"0.1","status":"rejected","market_id":"BTC-USD","client_order_id":"","reason":"RiskManager CheckOrderParams: price=109 should be less\\/equal than= 105","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@115","side":"long","created_at":1681343466169600}],"id":234}}'
        }
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end
    


end
