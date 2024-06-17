local fio = require('fio')
local t = require('luatest')
local log = require('log')

local config = require('app.config')
local periodics = require('app.profile.periodics')
local rpc = require('app.rpc')

local g = t.group('profile.conn')

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
    local cnt = 0
    for _ in pairs(config.params.MARKETS) do
        cnt = cnt + 1
    end

    cg.params = {
        markets_count = cnt,
    }
end)

g.after_each(function(cg)
end)

g.test_update_connections = function(cg)
    local called = 0
    local cnt, cnt2 = 0, 0
    local conn = {is_connected = function() return true end}

    -- all ok
    periodics.conns = {} -- impl magic
    periodics._reconnect_to_market = function(market_id)
        called = called + 1
        periodics.conns[market_id] = conn
        return nil
    end
    periodics._update_connections()
    t.assert_equals(called, cg.params.markets_count)

    -- broken multiple times
    called = 0
    cnt, cnt2 = 2, 0
    periodics.conns = {} -- impl magic
    periodics._reconnect_to_market = function(market_id)
        called = called + 1
        if cnt == 0 or market_id ~= config.params.MARKETS.LDOUSDT then
            periodics.conns[market_id] = conn
            return nil
        end
        cnt = cnt - 1
        return 'reconnect error'
    end
    periodics._update_connections()

    for _ in pairs(periodics.conns) do
        cnt2 = cnt2 + 1
    end
    t.assert_gt(called, cg.params.markets_count)
    t.assert_equals(cnt2, cg.params.markets_count)
end

g.test_wait_for_markets = function(cg)
    local called, called_ok = 0, 0
    local cnt, cnt2 = 0, 0

    -- all ok
    rpc.wait_for_role = function(role_name)
        called = called + 1
        return {res = {}, error = nil}
    end
    periodics._wait_for_markets()
    t.assert_equals(called, cg.params.markets_count)

    -- broken multiple times
    called, called_ok = 0, 0
    cnt, cnt2 = 2, 2
    rpc.wait_for_role = function(role_name)
        called = called + 1
        if cnt == 0 then
            called_ok = called_ok + 1
            return {res = {}, error = nil}
        end
        cnt = cnt - 1
        return {res = nil, error = 'rpc error'}
    end
    periodics._wait_for_markets()
    t.assert_equals(called, cnt2 + cg.params.markets_count)
    t.assert_equals(called_ok, cg.params.markets_count)

    called, called_ok = 0, 0
    cnt, cnt2 = 2, 2
    rpc.wait_for_role = function(role_name)
        called = called + 1
        if cnt == 0 or role_name ~= 'market-ldo' then
            called_ok = called_ok + 1
            return {res = {}, error = nil}
        end
        cnt = cnt - 1
        return {res = nil, error = 'rpc error'}
    end
    t.assert_equals(periodics._wait_for_markets(), nil)
    t.assert_equals(called, cnt2 + cg.params.markets_count)
    t.assert_equals(called_ok, cg.params.markets_count)
end

g.test_update_markets_status = function(cg)
    local called, called_ok = 0, 0
    local cnt, cnt2 = 0, 0

    -- all ok
    rpc.callrw_engine = function(market_id)
        called = called + 1
        return {res = {}, error = nil}
    end
    periodics._update_market_status(config.params.MARKET_STATUS.PAUSED)

    -- broken multiple times
    called, called_ok = 0, 0
    cnt, cnt2 = 2, 2
    rpc.callrw_engine = function(market_id)
        called = called + 1
        if cnt == 0 then
            called_ok = called_ok + 1
            return {res = {}, error = nil}
        end
        cnt = cnt - 1
        return {res = nil, error = 'rpc error'}
    end
    periodics._update_market_status(config.params.MARKET_STATUS.PAUSED)
    t.assert_equals(called, cnt2 + cg.params.markets_count)
    t.assert_equals(called_ok, cg.params.markets_count)

    called, called_ok = 0, 0
    cnt, cnt2 = 2, 2
    rpc.callrw_engine = function(market_id)
        called = called + 1
        if cnt == 0 or market_id ~= config.params.MARKETS.LDOUSDT then
            called_ok = called_ok + 1
            return {res = {}, error = nil}
        end
        cnt = cnt - 1
        return {res = nil, error = 'rpc error'}
    end
    periodics._update_market_status(config.params.MARKET_STATUS.PAUSED)
    t.assert_equals(called, cnt2 + cg.params.markets_count)
    t.assert_equals(called_ok, cg.params.markets_count)
end
