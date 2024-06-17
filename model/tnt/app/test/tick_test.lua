local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')

--local a = require('app.archiver')
--local engine = require('app.engine.engine')
--local market = require('app.engine.market')
--local notif = require('app.engine.notif')
--local position = require('app.engine.position')
--local trade = require('app.engine.trade')
local tick = require('app.lib.tick')

--require('app.config.constants')

local g = t.group('lib.tick')

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
    cg.params = {}
end)

g.after_each(function(cg)
end)

g.test_is_valid_rounding = function(cg)
    t.assert_equals(tick.is_valid_rounding(decimal.new(0.100000001), decimal.new(0.1)), false)
    t.assert_equals(tick.is_valid_rounding(decimal.new(0.1), decimal.new(0.100000001)), true)
end

g.test_round_to_nearest = function(cg)
    local zero = decimal.new(0)
    local min_tick = decimal.new(0.1)

    t.assert_equals(tick.round_to_nearest_tick(zero, min_tick), zero)

    for val = 0, 0.2, 0.01 do
        t.assert_equals(tick.round_to_nearest_tick(val, min_tick, min_tick), min_tick)
    end
end
