local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local config = require('app.config')
local equeue = require('app.enginequeue')
local api = require('app.api')
local api_notif = require('app.api.notif')
local g = t.group('pqueue')
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
function mock_getters.get_profile(profile_id)
    return {}, nil
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

g.test_cancel_queue_no_priority = function(cg)
    local profile_id = 11
    local market_id = "BTC-USD"

    local res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    t.assert_is(res["error"], nil)
    local qname = res["res"]

    -- cancel by oid
    -- cancel by cid
    -- cancel when order not in the queue

    -- cancel by oid
    local oid = "oid-1"
    local cid = "cid-1"
    local order, task_data = action.pack_create(
        profile_id,
        market_id,
        oid,
        false,
        "long",
        "limit",
        decimal.new("1"),
        decimal.new("1"),
        cid,
        ZERO,
        ZERO,
        "gtc"
    )

    local res = equeue.put(qname, task_data, profile_id, oid, "limit")
    t.assert_is(res["error"], nil)

    box.space.used_client_order_id:insert{profile_id, cid, oid, market_id}

    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is_not(res["res"], nil)


    api_public.cancel_order(profile_id, market_id,  oid)
    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is(res["res"], nil)


    -- cancel by cid
    oid = "oid-2"
    cid = "cid-2"
    _, task_data = action.pack_create(
        profile_id,
        market_id,
        oid,
        false,
        "long",
        "limit",
        decimal.new("1"),
        decimal.new("1"),
        cid,
        ZERO,
        ZERO,
        "gtc"
    )

    local res = equeue.put(qname, task_data, profile_id, oid, "limit")
    t.assert_is(res["error"], nil)

    box.space.used_client_order_id:insert{profile_id, cid, oid, market_id}

    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is_not(res["res"], nil)


    api_public.cancel_order(profile_id, market_id, "", cid)
    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is(res["res"], nil)


    -- cancel when order not in the queue
    oid = "oid-3"
    cid = "cid-3"
    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is(res["res"], nil)

    box.space.used_client_order_id:insert{profile_id, cid, oid, market_id}

    api_public.cancel_order(profile_id, market_id, oid)
    res = equeue.find_task(qname, tostring(profile_id), "cancel@" .. oid)
    t.assert_is_not(res["res"], nil)


    -- cancel by cid, but not order_id found
    oid = "oid-4"
    cid = "cid-4"
    res = equeue.find_task(qname, tostring(profile_id), oid)
    t.assert_is(res["res"], nil)


    api_public.cancel_order(profile_id, market_id, "", cid)
    res = equeue.find_task(qname, tostring(profile_id), "cancel@" .. cid)
    t.assert_is_not(res["res"], nil)
end
