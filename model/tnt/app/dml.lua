local checks = require('checks')

local dml = {}

function dml.atomic2(rollback, fn, ...)
    checks('function', 'function')
    local status, res = pcall(function(...)
        box.begin()
        local res, err = fn(...)
        if err ~= nil then
            error(err)
        end

        box.commit()
        return res
    end, ...)
    if status == false then
        rollback()
        return nil, res
    end

    return res, nil
end

function dml.atomic(fn, ...)
    checks('function')
    return dml.atomic2(box.rollback, fn, ...)
end

function dml.begin(isolation)
    local status, res = pcall(function()
        box.begin(isolation)
    end)
    if status == false then
        return nil, res
    end

    return nil, nil
end

function dml.commit()
    local status, res = pcall(function()
        box.commit()
    end)
    if status == false then
        return nil, res
    end

    return nil, nil
end

function dml.insert(obj, tuple)
    local status, res = pcall(function()
        return obj:insert(tuple)
    end)
    if status == false then
        return nil, res
    end

    return res, nil
end

function dml.update(obj, key, ...)
    local status, res = pcall(function(...)
        return obj:update(key, ...)
    end, ...)
    if status == false then
        return nil, res
    end

    return res, nil
end

function dml.get(obj, key)
    local status, res = pcall(function()
        return obj:get(key)
    end)
    if status == false then
        return nil, res
    end

    return res, nil
end

function dml.select(obj, key, ...)
    local status, res = pcall(function(...)
        return obj:select(key, ...)
    end, ...)
    if status == false then
        return nil, res
    end

    return res, nil
end

-- tarantool box.space.truncate() can yield!
-- can throw error
function dml.truncate(space)
    local primary = space.index[0]

    for _, t in space:pairs(nil, {iterator = box.ALL}) do
        local key = {}
        for _, p in pairs(primary.parts) do
            table.insert(key, t[p.fieldno])
        end
        space:delete(key)
    end
end

return dml
