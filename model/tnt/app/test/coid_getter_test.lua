local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')

local a = require('app.archiver')
local engine = require('app.engine')
local market = require('app.engine.market')
local order = require('app.engine.order')
local config = require('app.config')
local archiver = require('app.archiver')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')

local g = t.group('engine.getters')

local work_dir = fio.tempdir()

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)

    local MARKET = config.markets["BTC-USD"]
    archiver.init_sequencer(MARKET.id)
    engine.init_spaces(MARKET)

    cg.params = {}
end)

g.after_each(function(cg)
    box.space.order:drop()
    box.space.orderbook:drop()
    box.space.profile_meta:drop()
    box.space.market:drop()
end)

g.test_coid_getters = function(cg)
    local market_id = 'BTC-USD'
    local res, err

    local profile_id = 1

    res, err = order.create("BTC-USD@1", 
        profile_id, 
        market_id, 
        "limit", 
        decimal.new(1000.0), 
        decimal.new(1.0),
        decimal.new(1.0), 
        "long",
        "client_order_id-1",
        decimal.new(0),
        decimal.new(0),
        "good_till_cancel",
        false)
        
    t.assert_is(err, nil)


    res, err = order.create("BTC-USD@2", 
        profile_id, 
        market_id, 
        "limit", 
        decimal.new(1000.0), 
        decimal.new(1.0),
        decimal.new(1.0), 
        "long",
        "client_order_id-2",
        decimal.new(0),
        decimal.new(0),
        "good_till_cancel",
        false)
        
    t.assert_is(err, nil)


    res, err = order.create("BTC-USD@3", 
        profile_id, 
        market_id, 
        "limit", 
        decimal.new(1000.0), 
        decimal.new(1.0),
        decimal.new(1.0), 
        "long",
        "client_order_id-3",
        decimal.new(0),
        decimal.new(0),
        "good_till_cancel",
        false)

    res, err = order.create("BTC-USD@4", 
        profile_id, 
        market_id, 
        "limit", 
        decimal.new(1000.0), 
        decimal.new(1.0),
        decimal.new(1.0), 
        "long",
        "client_order_id-3",
        decimal.new(0),
        decimal.new(0),
        "good_till_cancel",
        false)

        
    t.assert_is(err, nil)

    -- only open orders: nothing to remove
    local r_res = order.check_coid_for_reuse(profile_id, "client_order_id-1")
    t.assert_is(r_res["error"], "OPEN_FOUND")

    -- no open orders with this coid, can be reused
    box.space.order:update("BTC-USD@1", {{'=', 'status', 'closed'}})
    r_res = order.check_coid_for_reuse(profile_id, "client_order_id-1")
    t.assert_is(r_res["error"], nil)
    t.assert_is(r_res["res"], "client_order_id-1")

    -- one closed and one open: can't be reused
    box.space.order:update("BTC-USD@3", {{'=', 'status', 'closed'}})
    r_res = order.check_coid_for_reuse(profile_id, "client_order_id-3")
    t.assert_is(r_res["error"], "OPEN_FOUND")

    -- 2nd rejected, can be reused
    box.space.order:update("BTC-USD@4", {{'=', 'status', 'closed'}})
    r_res = order.check_coid_for_reuse(profile_id, "client_order_id-3")
    t.assert_is(r_res["error"], nil)
    t.assert_is(r_res["res"], "client_order_id-3")


end
