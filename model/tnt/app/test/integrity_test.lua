local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local integrity = require('app.profile.integrity')
local config = require('app.config')
local getters = require('app.profile.getters')


require('app.config.constants')

local g = t.group('integrity')

local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    integrity.init_spaces()
end)

g.after_each(function(cg)
    box.space.init_lock:drop()
end)

g.test_basic = function(cg)
    local is_valid = integrity.is_valid()
    t.assert_is(is_valid, false)

    --check getters only part for checking lock
    local res = getters.cached_is_inv3_valid(0)
    t.assert_is(res["error"], "INV3_DATA_NOT_INITIALIZED")

    res = getters.is_inv3_valid(0)
    t.assert_is(res["res"], nil)

    res = getters.liquidation_batch(1, 1)
    t.assert_is(res["error"], "INV3_DATA_NOT_INITIALIZED")

    integrity.make_valid()
    is_valid = integrity.is_valid()
    t.assert_is(is_valid, true)


    integrity.make_invalid()
    is_valid = integrity.is_valid()
    t.assert_is(is_valid, false)
end