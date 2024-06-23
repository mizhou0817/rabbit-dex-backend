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
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')

local g = t.group('position.enrich')

local work_dir = fio.tempdir()

local mock_rpc = {call={}}
function pubsub_publish(channel, json_data, ttl, size, meta_ttl)
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
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    t.assert_is_not(a.init_sequencer('shard'), nil)
    balance.init_spaces("BTC-USD")
    local market_data = {
        id = "BTC-USD",
        status = 'active',
        min_initial_margin = ONE,
        forced_margin = ONE,
        liquidation_margin = ONE,
        min_tick = ONE,
        min_order = ONE,
    }
    market.init_spaces(market_data)

    position.init_spaces()
    trade.init_spaces()
    notif.init_spaces()
    profile.init_spaces()

    o.init_spaces()
    ob.init_spaces(market_data)

    engine.init(market_data.id, ONE, ONE)

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
    balance.test_clear_spaces()
end)

g.test_update_position = function(cg)
    local market_id = 'BTC-USD'
    local tm = mock_time.now()

    local fair_price = decimal.new(111)
    market.update_fair_price(market_id, fair_price)

    --profile
    local profile_id = 234
    local err = profile.ensure_meta(profile_id, market_id)
    t.assert_is(err, nil)

    local fill_price = decimal.new(200)
    local fill_size = decimal.new(2)

    err = engine._test_update_position(profile_id, config.params.SHORT, fill_price, fill_size)
    t.assert_is(err, nil)

    -- check that space updated with rt values
    --[[
            position.unrealized_pnl,
            position.notional,
            position.margin,
            position.liquidation_price,
            market.fair_price,
        })
    --]]
    --[[
    local pos = box.space.position.index.pos_by_market_profile:get({market_id, profile_id})
    t.assert_is_not(pos.unrealized_pnl, ZERO)
    t.assert_is(pos.notional, fill_size * fair_price )
    t.assert_is_not(pos.margin, ZERO)
    t.assert_is_not(pos.liquidation_price, ZERO)
    t.assert_is(pos.fair_price, fair_price)
        --]]

    local check_found_with_data = function (size, notional)
        local found = false
        for _, item in box.space.nf_private:pairs(nil, {iterator = box.index.ALL}) do
            local profile_id = tostring(item.profile_id)
            local update_type = item.nf_type

            if update_type == "position" then
                log.info(item.data)
                t.assert_is(item.data["size"], size)
                t.assert_is(item.data["notional"], notional)
                found = true
            end
        end
        t.assert_is(found, true)    
    end

    check_found_with_data(fill_size, fill_size * fair_price)



    -- close position
    err = engine._test_update_position(profile_id, config.params.LONG, fill_price, fill_size)
    t.assert_is(err, nil)

    --[[
    pos = box.space.position.index.pos_by_market_profile:get({market_id, profile_id})
    t.assert_is(pos, nil)
    --]]

    -- if we delete position, we don't clear enrich data
    check_found_with_data(ZERO, ZERO)
end
