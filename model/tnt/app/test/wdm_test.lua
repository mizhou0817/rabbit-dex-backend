local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local wdm = require('app.wdm')
local rolling = require('app.rolling')
require('app.config.constants')

local g = t.group('wdm')
g.before_each(function(cg)
    rolling.init_spaces()
    wdm.init_spaces()
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
end)

g.test_wdm = function(cg)
    local res = box.space.wd_tier:select()
    t.assert_is(res[1].min_amount, ZERO)
    t.assert_is(res[2].min_amount, decimal.new("10000"))
    
    local delay = wdm.get_delay("rbx", 1, decimal.new("100"))
    t.assert_is(delay, 1)

    delay = wdm.get_delay("bfx", 1, decimal.new("100"))
    t.assert_is(delay, 1)

    wdm.replace_special_tier("rbx", 1, decimal.new("10001"), decimal.new("10003"), 111)
    delay = wdm.get_delay("bfx", 1, decimal.new("100"))
    t.assert_is(delay, 1)

    delay = wdm.get_delay("rbx", 1, decimal.new("100"))
    t.assert_is(delay, 1)

    delay = wdm.get_delay("rbx", 1, decimal.new("10003"))
    t.assert_is(delay, 111)

    delay = wdm.get_delay("rbx", 1, decimal.new("10001"))
    t.assert_is(delay, 111)
    delay = wdm.get_delay("rbx", 1, decimal.new("10002"))
    t.assert_is(delay, 111)

    delay = wdm.get_delay("bfx", 1, decimal.new("10002"))
    t.assert_is(delay, 100)

    delay = wdm.get_delay("rbx", 1, decimal.new("10004"))
    t.assert_is(delay, 10)

    wdm.replace_tier("bfx", decimal.new("0"), 2)
    delay = wdm.get_delay("bfx", 1, decimal.new("100"))
    t.assert_is(delay, 2)

    -- roll value
    wdm.roll_volume(decimal.new("100"))
    wdm.roll_volume(decimal.new("101"))
    wdm.roll_volume(decimal.new("102"))
    
    local total = wdm.wds_per_24h()
    t.assert_is(total, decimal.new("303"))
end
