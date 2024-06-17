local fio = require('fio')
local t = require('luatest')
local log = require('log')
local config = require('app.config')
local equeue = require('app.enginequeue')
local api = require('app.api')
local g = t.group('pqueue')

local work_dir = fio.tempdir()

-- need this to support imports inside tqueue module
fio.copytree("./app", fio.pathjoin(work_dir, 'app'))

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
    equeue.init_spaces()
end)

g.after_each(function(cg)
end)

g.test_pqueue = function(cg)
    local qname = "testqueue"
    local res = equeue.get_or_create(qname)
    t.assert_is(res.error, nil)


    -- put with priorities
    res = equeue._put_with_priority(qname, {data="data1"}, "profile-1", "order-1", "cancel", -1, -1, 100)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-1")
    t.assert_is(res.res[7], -1)
    t.assert_is(res.res[8], -1)
    t.assert_is(res.res[9], 100)

    -- put with no priority
    res = equeue.put_with_lowest_priority(qname, {data="data3"}, "profile-1", "order-3", "cancel")
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-3")
    t.assert_is(res.res[7], equeue.lowest_priority)
    t.assert_is(res.res[8], equeue.lowest_priority)
    t.assert_is_not(res.res[9], 0)
    
    -- put the same priority but later timestamp
    res = equeue._put_with_priority(qname, {data="data2"}, "profile-1", "order-2", "cancel", -1, -1, 200)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-2")
    t.assert_is(res.res[7], -1)
    t.assert_is(res.res[8], -1)
    t.assert_is(res.res[9], 200)


    res = equeue.take(qname, 0)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-1")

    res = equeue.take(qname, 0)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-2")

    res = equeue.take(qname, 0)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "order-3")

    res = equeue.take(qname, 0)
    t.assert_is(#res, 0)

    -- insurance has unlimited and the highest priority
    local tup = equeue.add_priority_by_profile_id(0, 100, 100, true, -1, -1)
    t.assert_is_not(tup, nil)

    -- profile 1 has 1 attempt priority
    tup = equeue.add_priority_by_profile_id(1, 1, 1, false, 0, -1)
    t.assert_is_not(tup, nil)
    
    -- profile 2 limit order 
    tup = equeue.add_priority_by_profile_id_order_type(2, "limit", 1, 1, false, -1, 0)
    t.assert_is_not(tup, nil)
    
    -- profile 3 limit order
    tup = equeue.add_priority_by_profile_id_order_type(3, "limit", 1, 1, false, -1, 0)
    t.assert_is_not(tup, nil)

    -- profile 111 has unlimited priority
    tup = equeue.add_priority_by_profile_id(111, 1, 1, true, 0, -1)
    t.assert_is_not(tup, nil)

    --[[
        if we see orders commint like: market, limit, cancel

        cancel - is the more priority than market/limit
        market/limit - the same priority 

        they will be executed in the sequnce: 
        cancel
        market 
        limit
    --]]
    
    tup = equeue.add_priority_by_order_type("cancel", 0, 0, true, -1, 1)
    t.assert_is_not(tup, nil)

    --[[
    tup = equeue.add_priority_by_order_type("limit", 0, 0, true, 0, 1)
    t.assert_is_not(tup, nil)
    
    tup = equeue.add_priority_by_order_type("market", 0, 0, true, 0, 1)
    t.assert_is_not(tup, nil)
    --]]


    --[[
        Test the next sequence

        put:
            1) insurance, limit     (data4)
            2) insurance, market    (data5) 
            3) 5, market            (data6)
            4) 6, limit             (data7)
            5) insurance, cancel    (data8)
            6) 7, cancel            (data9)
            7) 1, limit             (data10)
            8) 1, market            (data11)
            9) 1, market            (data12)
            10) 2, limit            (data13)
            11) 9, market           (data14)
            12) 2, market           (data15)

        take must return:
            1) data4
            2) data5 
            3) data8
            4) data10
            5) data12 
            6) data13
            7) data9
            8) data6
            9) data7
            10) data11
            11) data14
            12) data15
    --]]
    
    res = equeue.put(qname, {data="data4"}, 0, "order-4", "limit")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data5"}, 0, "order-5", "market")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data6"}, 5, "order-6", "market")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data7"}, 6, "order-7", "limit")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data8"}, 0, "order-8", "cancel")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data9"}, 7, "order-9", "cancel")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data10"}, 1, "order-10", "limit")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data11"}, 1, "order-11", "market")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data12"}, 1, "order-12", "market")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data13"}, 2, "order-13", "limit")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data14"}, 9, "order-14", "market")
    t.assert_is(res.error, nil)

    res = equeue.put(qname, {data="data15"}, 2, "order-15", "market")
    t.assert_is(res.error, nil)


    --[[
            take must return:
            1) data4  - cuz insurance
            2) data5  - cuz insurance
            3) data8  - cuz insurance 
            4) data10 - 1st order from user 1
            5) data12 - 3rd order from user 1  (cuz wait is 1,  1, 0, -1)
            6) data13 - 1st limit from user 2
            7) data9  - cuz cancel is the highest priority

            -- till this moment we just have FIFO, no priority
            8) data6
            9) data7
            10) data11
            11) data14
            12) data15
    --]]    
    
    --]]
    local expected_data = {
        "data4", 
        "data5", 
        "data8", 
        "data10", 
        "data12", 
        "data13",
        "data9",
        "data6",
        "data7",
        "data11",
        "data14",
        "data15"}
    for _, data in ipairs(expected_data) do
        res = equeue.take(qname, 0)
        t.assert_is(res.error, nil)
    
        local has_data = res.res[3].data
        t.assert_is(has_data, data)
        log.info(has_data)
    end

    res = equeue.take(qname, 0)
    t.assert_is(#res, 0)


    res = equeue._put_with_priority(qname, {data="lowest"}, "lowest", "lowest", "lowest", 1, 1, 200)
    t.assert_is(res.error, nil)

    res = equeue.put_with_highest_priority(qname, {data="highest"}, "highest", "highest", "highest")
    t.assert_is(res.error, nil)

    res = equeue.take(qname, 0)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "highest")

    res = equeue.take(qname, 0)
    t.assert_is(res.error, nil)
    t.assert_is(res.res[4], "lowest")

    res = equeue.take(qname, 0)
    t.assert_is(#res, 0)


    --[[
        Test put_after

        put:
            1) 10, limit            (oid-1)
            2) 10, cancel           (cancel-1)
            3) 0, market            (oid-2)
            4) 0, cancel            (cancel-2)
            5) 11, cancel           (cancel-11)
            6) 111, cancellall      (cancel-111)
    --]]

    -- TEST the public api part
    local wrap_cancel = function (qname, data, profile_id, oid, order_type, create_task_oid)
            -- if we are in the queue - cancel shold have less priority 
            local r = equeue.find_task(qname, tostring(profile_id), tostring(create_task_oid))
            local existing_task = r["res"]

            log.info("wrap_cancel: create_task_oid=%s  existing_task=%s  not task=%s",tostring(create_task_oid), tostring(existing_task), tostring(not existing_task))

            r = equeue.put_after(qname, data, profile_id, oid, order_type, existing_task)
            t.assert_is(r["error"], nil)
    end

    res = equeue.put(qname, {data="oid-1"}, 10, "oid-1", "limit", nil)
    t.assert_is(res.error, nil)

    wrap_cancel(qname, {data="cancel-1"}, 10, "cancel-1", "cancel", "oid-1")

    res = equeue.put(qname, {data="oid-2"}, 0, "oid-2", "market", nil)
    t.assert_is(res.error, nil)

    wrap_cancel(qname, {data="cancel-2"}, 0, "cancel-2", "cancel", "oid-2")

    -- Check that just cancel work for non-priority
    wrap_cancel(qname, {data="cancel-11"}, 11, "cancel-11", "cancel", "oid-11")

    -- Check that just cancel work for priority
    wrap_cancel(qname, {data="cancel-111"}, 111, "cancel-111", "cancel", "oid-111")


    expected_data = {
        "oid-2", 
        "cancel-2", 
        "cancel-111", 
        "cancel-11",
        "oid-1", 
        "cancel-1", 
    }
    for _, data in ipairs(expected_data) do
        res = equeue.take(qname, 0)
        t.assert_is(res.error, nil)

        log.info(res.res)

        local has_data = res.res[3].data
        t.assert_is(has_data, data)
        log.info(has_data)
    end

    res = equeue.take(qname, 0)
    t.assert_is(#res, 0)


end
