local t = require('luatest')

local g = t.group('luaparts')

g.test_xpcall_box_error = function()
    local ok, val = xpcall(function()
        error(box.error.new({code=500, reason='box-error'}))
    end, function(err)
        return debug.traceback(tostring(err))
    end)

    t.assert_is(ok, false)
    t.assert_str_contains(val, 'box-error\nstack traceback:')
end

g.test_xpcall_lua_error = function()
    local ok, val = xpcall(function()
        error('lua-error')
    end, function(err)
        return debug.traceback(tostring(err))
    end)

    t.assert_is(ok, false)
    t.assert_str_contains(val, 'lua-error\nstack traceback:')
end

