local t = require('luatest')
local util = require('app.util')

local g = t.group('util')

g.test_tostring_str = function()
    local v = '123456'
    t.assert_is(util.tostring(v), v)
end

g.test_tostring_number = function()
    local v = 123456
    t.assert_is(util.tostring(v), tostring(v))
end

g.test_tostring_tuple = function()
    local v = box.tuple.new({100, 200, 300, {400, 500, 600}, 700, 800})
    t.assert_is(util.tostring(v), tostring(v))
end

g.test_tostring_table = function(cg)
    local f = function(v)
        return loadstring('return ' .. util.tostring(v))()
    end

    local expected = {
        {100, 200, 300, {400, 500, 600}, 700, 800},
        {100, 200, 300, {a=400, b=500, c=600}, 700, 800},
        {a=100, b=200, c=300, d={400, 500, 600}, e=700, e=800},
        {a=100, b=200, c=300, d={e=400, f=500, g=600}, h=700, i=800},
    }

    for _, v in ipairs(expected) do
        t.assert_items_equals(f(v), v)
    end
end
