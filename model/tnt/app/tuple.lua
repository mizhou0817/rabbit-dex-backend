local checks = require('checks')

local errors = require('app.lib.errors')
local util = require('app.util')

local TupleError = errors.new_class("TUPLE_ERROR")

local T = {}

local function cached_mt(format, typ)
    checks('table', '?string')
    local mt = T[format]

    if mt == nil then
        mt = {}
        --checks module type support
        mt.__type = 'table' -- so can do smooth migration to strict types

        -- tuple helpers
        mt.__strict_type = typ or 'tuple'
        mt.__format = format

        -- table overloaded funcs
        mt.__index = function(self, key)
            return T[key] or rawget(
                self,
                mt.index[key] or error(TupleError:new('%s has no %s field to get', mt.__strict_type, key))
            )
        end
        mt.__newindex = function(self, key, val)
            return rawset(
                self,
                mt.index[key] or error(TupleError:new('%s has no %s field to set', mt.__strict_type, key)),
                val
            )
        end
        mt.__tostring = util.tostring

        -- mapper of named fields to positional ones
        mt.index = {}
        for i, field in ipairs(format) do
            mt.index[field.name] = i
        end

        T[format] = mt
    end

    return mt
end

-- opts can be {names_only = true} in future to simulate box.tuple:tomap()
function T.tomap(self, opts)
    local res = {}
    local mt = getmetatable(self)

    for i, f in ipairs(mt.__format) do
        local data = self[i]
        if type(data) == 'table' and data.tomap ~= nil then
            res[f.name] = data:tomap(opts)
        else
            res[f.name] = data
        end
    end

    return res
end

function T.totable(self)
    local mt = getmetatable(self)
    error('totable not supported yet for type ' .. mt.__strict_type)
end

function T.new(value, format, typ)
    -- table|cdata should be array|box.tuple
    checks('table|cdata', 'table', '?string')

    -- TODO: add field types validation

    -- for prod the most part of logic can be removed to speedup code
    -- but for dev it's useful

    if box.tuple.is(value) then
        -- this code faster than tuple:totable()
        local new_value = {}
        for _, v in value:ipairs() do
            table.insert(new_value, v)
        end
        value = new_value
    end

    return setmetatable(value, cached_mt(format, typ))
end

return T
