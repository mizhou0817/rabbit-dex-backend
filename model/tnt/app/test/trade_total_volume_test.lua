local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local trade = require('app.engine.trade')
require('app.config.constants')

local g = t.group('trade')
g.before_each(function(cg)
    trade.init_spaces()
end)

local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
    box.space.fill:drop()
end)

g.test_total_volume = function(cg)
    -- Create fills

    local shard_id = "shard_id"
    local archive_id = 1
    

    -- 3 traders
    -- 2000 trades for each
    for i=1,3 do
        local tm = 1
        for j=1,2000 do
            local fill_id = "fill-" .. tostring(i) .. "-tr-" .. tostring(j)
            local res = box.space.fill:insert{
                fill_id,
                i,
                "btc",
                "order_id",
                tm,
                "trade_id",
                decimal.new(1),
                decimal.new(10),
                "long",
                false,
                decimal.new(0),
                false,
                "coid_22",

                shard_id,
                archive_id
            }
            tm = tm + 1
            t.assert_is_not(res, nil)
        end
    end

    local notional = decimal.new(10)

    -- empty value for non-existance
    local res = trade.total_volume(5, 10, 100)
    t.assert_is(res['error'], nil)
    t.assert_is(res['res'][1], decimal.new(0))
    t.assert_is(res['res'][2], 10)

    -- yield should happen
    res = trade.total_volume(1, 0, 3000)
    t.assert_is(res['error'], nil)
    t.assert_is(res['res'][1], notional * 2000)
    t.assert_is(res['res'][2], 2000)


    -- check that conditions work correct
    local profile_id = 2
    local from = 0
    local to = 100
    local total = decimal.new(0)

    for i =1,2000,100 do
        res = trade.total_volume(profile_id, from, to)
        t.assert_is(res['error'], nil)
        t.assert_is(res['res'][1], notional * 100)
        t.assert_le(res['res'][2], to)
        
        from = to
        to = from + 100

        total = total + res['res'][1]
    end
    t.assert_is(total, notional * 2000)



end
