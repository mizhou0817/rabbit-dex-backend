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
local router = require('app.engine.router')
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
local cache = require('app.profile.cache')
local p_cache = require('app.profile')
local util = require('app.util')
local cm_get = require('app.engine.cache_and_meta')
local rpc = require('app.rpc')

require('app.config.constants')

local g = t.group('engine.private_matching')

local work_dir = fio.tempdir()

local mock_rpc = {call={}}
function mock_rpc.callrw_pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    table.insert(mock_rpc.call, {channel, json_data})
end

function mock_rpc.callrw_profile(func_name, params)
    return cache.get_cache_and_meta(params[1], params[2])
end


local function ad_hoc_p_cache_init_spaces()
    local profile_cache, err = a.create('profile_cache', {temporary=true, if_not_exists = true}, {
        {name = 'id', type = 'unsigned'},
        {name = 'profile_type', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'wallet', type = 'string'},

        {name = 'last_update', type = 'number'},
        {name = 'balance', type = 'decimal'},
        {name = 'account_equity', type = 'decimal'},
        {name = 'total_position_margin', type = 'decimal'},
        {name = 'total_order_margin', type = 'decimal'},
        {name = 'total_notional', type = 'decimal'},
        {name = 'account_margin', type = 'decimal'},
        {name = 'withdrawable_balance', type = 'decimal'},
        {name = 'cum_unrealized_pnl', type = 'decimal'},
        {name = 'health', type = 'decimal'},
        {name = 'account_leverage', type = 'decimal'},
        {name = 'cum_trading_volume', type = 'decimal'},
        {name = 'leverage', type = '*'},
        {name = 'last_liq_check', type = 'number'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        error(err)
    end

    profile_cache:create_index('for_liquidation', {
        unique = false,
        parts = {{field = 'status'}, {field = "id"}},
        if_not_exists = true })
end

local function mock_handle_task(order, profile_data, matching_meta)
    notif.clear()
    local res = engine._handle_create(order,profile_data, matching_meta)
    notif.clear()

    return res
end

local mock_time = {}
function mock_time.now()
    return 1681343466169600
end

local risk = {}
function risk.post_match(_market_id, profile_data, profile_id, position)
    return nil
end

local mock_cache = {}
function mock_cache.update(profile_id)
    -- REPLACE old value
    local _, err = a.replace(box.space.profile_cache, {
        profile_id,
        config.params.PROFILE_TYPE.TRADER,
        config.params.PROFILE_STATUS.ACTIVE,
        "",

        time.now(),
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        time.now(),
    })

    return nil
end

function mock_cache.get_meta(profile_id, market_id)
    return {profile_id=profile_id, market_id=market_id}
end



t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }

    rpc.test_set_mock_callrw_profile(mock_rpc.callrw_profile)
    notif._test_set_rpc(mock_rpc)
    notif._test_set_time(mock_time)
    o._test_set_time(mock_time)
    engine._test_set_time(mock_time)
    engine._test_set_post_match(risk.post_match)

    cache._test_set_cache_update(mock_cache.update)
    cache._test_set_get_meta(mock_cache.get_meta)
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
    router.init_spaces()
    market.init_spaces(market_data)
    
    ad_hoc_p_cache_init_spaces()

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
    box.space.router:truncate()
    box.space.unique_cp:truncate()
    box.space.default_cp:truncate()
    box.space.liquidation_cp:truncate()
    box.space.black_list:truncate()
    box.space.router_system:truncate()
    box.space.profile_cache:truncate()
    box.space.fill:truncate()
end)

g.test_private_matching = function(cg)
    local market_id = 'BTC-USD'
    local tm = mock_time.now()

    --[[
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

    --]]
end

g.test_multiple_cache = function(cg)
    local res = cache.get_cache_and_meta({1,2, 3}, "BTC-USD")

    local profile_data, err = cm_get.handle_get_cache_and_meta(nil, {1,2,3}, "BTC-USD", 2)
    t.assert_is(err, nil)
    t.assert_is_not(profile_data, nil)
    t.assert_is(profile_data.cache[1], 2)

    local profile_data, err = cm_get.handle_get_cache_and_meta(router, {4,5,6}, "BTC-USD", 5)
    t.assert_is(err, nil)
    t.assert_is_not(profile_data, nil)
    t.assert_is(profile_data.cache[1], 5)
    t.assert_is(router.only_active_counterparty(5), 5)

    t.assert_is(router.get_profile_data(1), nil)
    t.assert_is(router.get_profile_data(4).cache[1], 4)

    t.assert_is(router.get_profile_data(5).cache[1], 5)
    t.assert_is(router.get_profile_data(6).cache[1], 6)


end


g.test_router = function(cg)
    router.init_spaces()

    local exist = box.space.router_system:get(0)
    t.assert_is_not(exist, nil)
    t.assert_is(exist.is_active, false)
    t.assert_is(exist.range, decimal.new("0.001"))
    t.assert_is(exist.max_size, decimal.new("100000"))
    t.assert_is(exist.allow_liquidation, false)

    -- rewrite router system to not break the test
    -- CREATE deafult router_system
    a.replace(box.space.router_system, {
        0,
        true,
        decimal.new("0.02"),
        decimal.new("300000"),
        true
    })

    -- Check default settings
    exist = box.space.router_system:get(0)
    t.assert_is_not(exist, nil)
    t.assert_is(exist.is_active, true)
    t.assert_is(exist.range, decimal.new("0.02"))
    t.assert_is(exist.max_size, decimal.new("300000"))
    t.assert_is(exist.allow_liquidation, true)

    -- Check routable
    local order = {
        order_type = config.params.ORDER_TYPE.PING_LIMIT
    }
    local res_m = require('app.engine.market').get_market("BTC-USD")
    local market_data = res_m.res
    local res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: order_type")

    order.order_type = config.params.ORDER_TYPE.MARKET
    order.time_in_force = config.params.TIME_IN_FORCE.FOK
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: time_in_force")

    order.order_type = config.params.ORDER_TYPE.LIMIT
    order.time_in_force = config.params.TIME_IN_FORCE.GTC 
    order.size = decimal.new("1000")
    order.price = decimal.new("301")
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: max_size")

    order.size = decimal.new("1") 
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: best_bid")


    -- NOW CREATE BID AND ASK WITH 10 tick spread
    local market_id = "BTC-USD"
    local fair_price = decimal.new(100)
    market.update_fair_price(market_id, fair_price)

    --profile
    local profile_id = 234
    local err = profile.ensure_meta(profile_id, market_id)
    t.assert_is(err, nil)
    
    local start_sequence = tonumber(ob.sequence:current())
    local n_order, res

    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
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
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)

    -- Create best_ask
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@2"),
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
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)
    

    res_m = require('app.engine.market').get_market("BTC-USD")
    market_data = res_m.res
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: exceed range")


    -- CREATE best bid and ask with 1 tick spread

    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@3"),
        true,
        config.params.LONG,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("107"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)

    -- Create best_ask
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@4"),
        true,
        config.params.SHORT,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("109"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)

    -- check that not marketable orders will not pass
    -- best ask = 109 
    -- best bid = 107
    res_m = require('app.engine.market').get_market("BTC-USD")
    market_data = res_m.res
    order.side = config.params.LONG
    order.price = decimal.new("108")
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: LONG not marketable")

    order.side = config.params.SHORT
    res = router.check_routable(order, market_data)
    t.assert_str_contains(res, "NOT_ROUTABLE: SHORT not marketable")


    order.price = decimal.new("105")
    res = router.check_routable(order, market_data)
    t.assert_is(res, nil)


    ------------------------------------------------------
    ------------------------------------------------------
    -- WHICH_COUNTERPARTY TESTS
    box.space.router_system:update(0, {{'=', 'is_active', false}})
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_COUNTERPARTY: ROUTER_NOT_ACTIVE")

    -- switch back to active
    box.space.router_system:update(0, {{'=', 'is_active', true}})
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_COUNTERPARTY: which_counterparty ERROR: profile_id")

    a.replace(box.space.black_list, {12})
    order.profile_id = 12
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_COUNTERPARTY: BLACK_LISTED profile")


    order.profile_id = 13
    order.is_liquidation = true
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_DEFAULT_LIQUIDATION_CP")


    router.test_add_profile_data({
        profile_id=666,
        cache={666, 'trader', 'paused'},
        meta={market_id="BTC-USD", profile_id=666}

    })
    a.replace(box.space.liquidation_cp, {"default", 666})


    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "CP_NOT_ACTIVE")

    a.replace(box.space.liquidation_cp, {"default", 5})
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, 5)
    t.assert_is(err, nil)

    order.is_liquidation = false
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, nil)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_COUNTERPARTY: matching_meta nil")

    local matching_meta = action_creator.wrap_meta({"ios", false, "rbx", true})
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, matching_meta)
    t.assert_is(res, nil)
    t.assert_str_contains(err, "NO_COUNTERPARTY: no default_cp")
    
    a.replace(box.space.default_cp, {"default", 4})
    res, err = router.which_counterparty(order.is_liquidation, order, market_data, matching_meta)
    t.assert_is(res, 4)
    t.assert_is(err, nil)

    -- FILTER conditions start here
    -- test filter conditions based on meta
    -- {id, exchange_id, device, is_api, to_counterparty}
    local affiliate_id = 4
    local vault = 6
    
    -- all  non-api mobile flow  for any exchange we route to affiliate 
    a.replace(box.space.router, {0,"rbx","ios",false,affiliate_id})
    a.replace(box.space.router, {1,"rbx","android",false,affiliate_id})
    a.replace(box.space.router, {2,"bfx","ios",false,affiliate_id})
    a.replace(box.space.router, {3,"bfx","android",false,affiliate_id})


    -- all  non-api desktop flow  for any exchange we route to affiliate 
    a.replace(box.space.router, {4,"rbx","desktop",false,vault})
    a.replace(box.space.router, {5,"rbx","other",false,vault})
    a.replace(box.space.router, {6,"bfx","desktop",false,vault})
    a.replace(box.space.router, {7,"bfx","other",false,vault})

    -- {device, is_api, exchange_id, is_pm}
    matching_meta = action_creator.wrap_meta({"ios", false, "rbx", false})
    res, err = router.filter_conditions(matching_meta)
    t.assert_is(err, nil)
    t.assert_is(res, affiliate_id)
    

    matching_meta = action_creator.wrap_meta({"android", false, "bfx", false})
    res, err = router.filter_conditions(matching_meta)
    t.assert_is(err, nil)
    t.assert_is(res, affiliate_id)
    
    matching_meta = action_creator.wrap_meta({"android", true, "bfx", false})
    res, err = router.filter_conditions(matching_meta)
    t.assert_is(err, nil)
    t.assert_is(res, nil)

    matching_meta = action_creator.wrap_meta({"desktop", false, "bfx", false})
    res, err = router.filter_conditions(matching_meta)
    t.assert_is(err, nil)
    t.assert_is(res, vault)


    mock_rpc.call = {}
    -- try make match with vault long 110
    order = {
        order_id = "BTC-USD@666",
        profile_id = 999,
        size = decimal.new("0.1"),
        side = config.params.LONG,
        price = decimal.new("110"),
        is_liquidation = false,
        order_type = config.params.ORDER_TYPE.LIMIT,
        time_in_force = config.params.TIME_IN_FORCE.GTC,
    }
    matching_meta = action_creator.wrap_meta({"desktop", false, "bfx", false})
    res = mock_handle_task(order, {}, matching_meta)
    t.assert_is(res, nil)

    local expected = {
        {
        'account@999',
        '{"data":{"fills":[{"trade_id":"BTC-USD-0","price":"109","size":"0.1","id":"BTC-USD-2","market_id":"BTC-USD","client_order_id":"","profile_id":999,"timestamp":1681343466169600,"order_id":"BTC-USD@666","side":"long","is_maker":false,"liquidation":false,"fee":"-0.00763","archive_id":38,"shard_id":"shard"}],"orders":[{"updated_at":1681343466169600,"initial_size":"0.1","order_type":"limit","trigger_price":"0","profile_id":999,"timestamp":1681343466169600,"total_filled_size":"0.1","price":"110","size":"0","status":"closed","market_id":"BTC-USD","client_order_id":"","reason":"","time_in_force":"good_till_cancel","size_percent":"0","id":"BTC-USD@666","side":"long","created_at":1681343466169600}],"positions":[{"unrealized_pnl":"-0.9","size":"0.1","take_profit":null,"side":"long","stop_loss":null,"market_id":"BTC-USD","profile_id":999,"entry_price":"109","shard_id":"shard","margin":"10.0","liquidation_price":"0","notional":"10.0","fair_price":"100","archive_id":35,"id":"pos-BTC-USD-tr-999"}],"id":999}}',
        },
        {
        'account@6',
        '{"data":{"fills":[{"trade_id":"BTC-USD-0","price":"109","size":"0.1","id":"BTC-USD-1","market_id":"BTC-USD","client_order_id":"","profile_id":6,"timestamp":1681343466169600,"order_id":"pm","side":"short","is_maker":true,"liquidation":false,"fee":"-0.0","archive_id":37,"shard_id":"shard"}],"positions":[{"unrealized_pnl":"0.9","size":"0.1","take_profit":null,"side":"short","stop_loss":null,"market_id":"BTC-USD","profile_id":6,"entry_price":"109","shard_id":"shard","margin":"10.0","liquidation_price":"109","notional":"10.0","fair_price":"100","archive_id":34,"id":"pos-BTC-USD-tr-6"}],"id":6}}',
        },
    }
    t.assert_equals(#mock_rpc.call, #expected)
    for i = 1, #expected do
        t.assert_equals(mock_rpc.call[i], expected[i])
    end
end

g.test_pm_sequence = function(cg)
    router.init_spaces()

        -- rewrite router system to not break the test
    -- CREATE deafult router_system
    a.replace(box.space.router_system, {
        0,
        true,
        decimal.new("0.02"),
        decimal.new("300000"),
        true
    })

    local bfx_cp = 4
    router.add_desktop("rbx", 1)
    router.add_desktop("rbx", 2)
    router.add_mobile("bfx", 3)
    router.add_mobile("bfx", bfx_cp)
    
    -- Check default settings
    local exist = box.space.router_system:get(0)
    t.assert_is_not(exist, nil)
    t.assert_is(exist.is_active, true)
    t.assert_is(exist.range, decimal.new("0.02"))
    t.assert_is(exist.max_size, decimal.new("300000"))
    t.assert_is(exist.allow_liquidation, true)
    
    -- NOW CREATE BID AND ASK WITH 10 tick spread
    local market_id = "BTC-USD"
    local fair_price = decimal.new(100)
    market.update_fair_price(market_id, fair_price)

    --profile
    local profile_id = 234
    local err = profile.ensure_meta(profile_id, market_id)
    t.assert_is(err, nil)
    
    local n_order, res

    -- Create best_bid
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
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
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)

    -- Create best_ask
    -- is_liquidation = true, cuz we don't want post_match check in this test
    n_order, _ = action_creator.pack_create(
        profile_id,
        market_id,
        market_id .. tostring("@2"),
        true,
        config.params.SHORT,
        config.params.ORDER_TYPE.LIMIT,
        decimal.new("0.1"),
        decimal.new("101"),
        "",
        nil,
        nil,
        config.params.TIME_IN_FORCE.GTC
    )
    res = mock_handle_task(n_order, {}, nil)
    t.assert_is(res, nil)
    


    -- Check routable
    local order = {
        order_id = "BTC-USD@666",
        profile_id = 999,
        size = decimal.new("0.1"),
        side = config.params.LONG,
        price = decimal.new("110"),
        is_liquidation = false,
        order_type = config.params.ORDER_TYPE.LIMIT,
        time_in_force = config.params.TIME_IN_FORCE.GTC,
    }

    mock_rpc.call = {}
    local matching_meta = action_creator.wrap_meta({"ios", false, "bfx", false})
    engine._test_set_post_match(function (_market_id, profile_data, profile_id, position)
        if profile_id == bfx_cp then
            return "FAKE_BFX_ERROR"
        end

        return nil
    end)

    local start_sequence = tonumber(ob.sequence:current())

    local res_m = require('app.engine.market').get_market("BTC-USD")
    local market_data = res_m.res

    local pm, _err = router.which_counterparty(false, order, market_data, matching_meta)
    t.assert_is(_err, nil)
    t.assert_is(pm, bfx_cp)

    res = mock_handle_task(order, {}, matching_meta)
    t.assert_is(res, nil)
    
    local end_sequence = tonumber(ob.sequence:current())

    t.assert_is_not(start_sequence, end_sequence)
end