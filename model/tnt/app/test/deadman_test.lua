local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local config = require('app.config')
local equeue = require('app.enginequeue')
local api = require('app.api')
local api_notif = require('app.api.notif')
local fiber = require('fiber')

local g = t.group('deadman')
local action = require('app.action')

local api_public = api.public

local work_dir = fio.tempdir()

local mock_rpc = {call={}}
function pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    t.assert_is_not(json_data, nil)
  --  log.info(json_data)
end

function mock_rpc.callrw_profile(command, params)
    
end

local mock_getters = {call={}}
function mock_getters.load_profile_and_market(profile_id, market_id)
    return {}, {}, nil
end

local mock_risk = {call={}}
function mock_risk.check_market(market)
    return nil
end

function mock_risk.check_profile(profile)
    return nil
end


-- need this to support imports inside tqueue module
fio.copytree("./app", fio.pathjoin(work_dir, 'app'))

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }

    api_public._test_set_rpc(mock_rpc)
    api_public._test_set_getters(mock_getters)
    api_public._test_set_risk(mock_risk)
    api_notif._test_set_rpc(mock_rpc)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    api.init_spaces()
    equeue.init_spaces()
end)

g.after_each(function(cg)
end)

g.test_deadman= function(cg)
    local profile_id = 999
    local market_id = "BTC-USD"

    local res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    t.assert_is(res["error"], nil)
    local qname = res["res"]

    --

    res = api_public.deadman_get(profile_id)
    t.assert_is(res["res"], nil)

    --[[
        return box.space.deadman:replace({
            profile_id, 
            timeout,
            time.now_milli(),
            DEADMAN_ACTIVE,
        })
    --]]
    local cur, interval
    local timeout = 900
    res = api_public.deadman_create(profile_id, timeout)
    t.assert_is_not(res["res"], nil)
    cur = res["res"]
    t.assert_is(cur[1], profile_id)
    t.assert_is(cur[2], timeout)
    t.assert_is_not(cur[3], 0)
    t.assert_is(cur[4], DEADMAN_ACTIVE)
    
    interval = box.space.deadman_interval:get("deadman")
    t.assert_is(interval.timeout, timeout)

    timeout = timeout + 1
    res = api_public.deadman_create(profile_id, timeout)
    t.assert_is_not(res["res"], nil)
    cur = res["res"]
    t.assert_is(cur[1], profile_id)
    t.assert_is(cur[2], timeout)

    interval = box.space.deadman_interval:get("deadman")
    t.assert_is_not(interval.timeout, timeout)
    
    res = api_public.deadman_get(profile_id)
    t.assert_is_not(res["res"], nil)
    t.assert_is(cur[1], res["res"][1])
    t.assert_is(cur[2], res["res"][2])
    

    res = api_public.deadman_delete(profile_id)
    t.assert_is_not(res["res"], nil)

    -- will set to default
    interval = box.space.deadman_interval:get("deadman")
    t.assert_is(interval.timeout, 5000)

    res = api_public.deadman_get(profile_id)
    t.assert_is(res["res"], nil)

    timeout = 9000
    res = api_public.deadman_create(profile_id, timeout)
    t.assert_is_not(res["res"], nil)

    interval = box.space.deadman_interval:get("deadman")
    t.assert_is(interval.timeout, 5000)

    fiber.sleep(5)
    res = api_public.deadman_get(profile_id)
    t.assert_is_not(res["res"], nil)
    t.assert_is(res["res"][4], DEADMAN_ACTIVE)


    fiber.sleep(7)
    res = api_public.deadman_get(profile_id)
    t.assert_is_not(res["res"], nil)
    t.assert_is(res["res"][4], DEADMAN_CANCELED)

    -- CANCEL ALL MUST BE in the queue
    local r1, qname, oid, r2
    r1 = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    t.assert_is_not(r1["res"], nil)
    qname = r1["res"]
    oid = "cancelall"
    r2 = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is_not(r2["res"], nil)
    equeue.delete(qname, r2["res"][1])


    -- NOW let's do the same test but with touch in the middle
    timeout = 9000
    res = api_public.deadman_create(profile_id, timeout)
    t.assert_is_not(res["res"], nil)

    fiber.sleep(5)
    res = api_public.deadman_get(profile_id)
    t.assert_is_not(res["res"], nil)
    t.assert_is(res["res"][4], DEADMAN_ACTIVE)
    api_public.deadman_touch(profile_id)

    fiber.sleep(7)
    res = api_public.deadman_get(profile_id)
    t.assert_is_not(res["res"], nil)
    t.assert_is(res["res"][4], DEADMAN_ACTIVE)

    -- CANCEL ALL MUST BE in the queue
    r1 = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    t.assert_is_not(r1["res"], nil)
    qname = r1["res"]
    oid = "cancelall"
    r2 = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is(r2["res"], nil)
end
