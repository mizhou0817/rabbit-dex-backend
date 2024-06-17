local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')

local a = require('app.archiver')
local engine = require('app.engine')
local o = require('app.engine.order')
local ob = require('app.engine.orderbook')
local notif = require('app.engine.notif')
local mt = require('app.engine.maintenance')
local matching = engine.engine
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')

local g = t.group('engine.maintenance')

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
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    t.assert_is_not(a.init_sequencer('shard'), nil)
    engine.init_spaces({
            id = 'BTC-USD',
            status = 'active',
            min_initial_margin = ONE,
            forced_margin = ONE,
            liquidation_margin = ONE,
            min_tick = ONE,
            min_order = ONE,
    })
    matching.init("BTC-USD", decimal.new("1"), decimal.new("1"))
    mock_rpc.call = {}
    cg.params = {}
end)

g.after_each(function(cg)
    notif.clear()
    box.sequence.shard_archive_id_sequencer:drop()
    box.space.affected_profiles:drop()
    box.space.order:drop()
    box.space.orderbook:drop()
end)

g.test_maintenance = function(cg)
    local market_id = 'BTC-USD'
    local tm = mock_time.now()

    --profile
    local profile_id = 234
    local profile_id2 = 235
    local orders = {"order-1", "order-2", "order-3"}
    for _, oid in ipairs(orders) do
        local _, err = o.create(
            oid,
            profile_id,
            market_id,
            "limit",
            decimal.new("1"),
            decimal.new("1"),
            decimal.new("1"),
            "long",
            "",
            ZERO,
            ZERO,
            "gtc",
            false
        )
        t.assert_is(err, nil)

        local _, err = ob.create(
            oid,
            market_id,
            profile_id,
            decimal.new("1"),
            decimal.new("1"),
            "long"
        )
        t.assert_is(err, nil)    
    end
    
    mt.add_profile_id(profile_id)
    mt.add_profile_id(profile_id2)

    t.assert_is(box.space.orderbook:count(), #orders)
    for _, o in box.space.order:pairs() do
        log.info("profile_id = %s  order_id = %s", tostring(o.profile_id), tostring(o.id))
        t.assert_is(o.status, "open")
    end

    mt.cancel_all_listed()

    t.assert_is(box.space.orderbook:count(), 0)
    local found = false
    for _, o in box.space.order:pairs() do 
        t.assert_is(tostring(o.status), "canceled")
        found = true
    end
    t.assert_is(found, true)

end
