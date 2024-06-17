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
function mock_rpc.callrw_pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    t.assert_is_not(json_data, nil)
    log.info(json_data)
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

g.test_cancel_from_the_queue = function(cg)
    local market_id = "BTC-USD"

    local res = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
    t.assert_is(res["error"], nil)
    local qname = res["res"]

    -- SETUP priorities
    local tup = equeue.add_priority_by_profile_id(0, 100, 100, true, -1, -1)
    t.assert_is_not(tup, nil)

    -- profile 111 has unlimited priority
    tup = equeue.add_priority_by_profile_id(111, 1, 1, true, 0, -1)
    t.assert_is_not(tup, nil)

    -- cancel has the highest priority
    tup = equeue.add_priority_by_order_type("cancel", 0, 0, true, -1, 1)
    t.assert_is_not(tup, nil)



    --[[
        Test put_after

        put:
            1) 10, limit            (oid-1)
            2) 10, cancel           (cancel-oid-1)
            3) 0, market            (oid-2)
            4) 0, cancel            (cancel-oid-2)
            5) 11, cancel           (cancel-oid-11)
            6) 111, cancel          (cancel-oid-111)
    --]]


    local order_sequence = {
        {profile_id=10, oid="oid-1", cid="cid-1", order_type="limit"},
        {profile_id=10, oid="oid-1", cid="cid-1", order_type="cancel"},
        {profile_id=0, oid="oid-2", cid="cid-2", order_type="market"},
        {profile_id=0, oid="oid-2", cid="cid-2", order_type="cancel"},
        {profile_id=11, oid="oid-11", cid="cid-11", order_type="cancel"},
        {profile_id=111, oid="oid-111", cid="cid-111", order_type="cancel"},
    }

    for _, o in ipairs(order_sequence) do 
        if o.order_type == "cancel" then
            api_public.cancel_order(o.profile_id, market_id,  o.oid, o.cid)
        else
            local order, task_data = action.pack_create(
                o.profile_id,
                market_id,
                o.oid,
                false,
                "long",
                o.order_type,
                decimal.new("1"),
                decimal.new("1"),
                o.cid,
                ZERO,
                ZERO,
                "gtc"
            )
        
            local res = equeue.put(qname, task_data, o.profile_id, o.oid, o.order_type)
            t.assert_is(res["error"], nil)                
        end
    end

    res = equeue.take(qname, 0)
    t.assert_is(#res, 0)
end