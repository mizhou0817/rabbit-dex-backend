local checks = require('checks')
local fiber = require('fiber')

local util = {}

function util.safe_yield(count, limit)
    count = count + 1
    if count >= limit then
        local alreadyInTx = box.is_in_txn()

        -- commit current transaction before yield
        if alreadyInTx then
            box.commit()
        end

        fiber.yield()

        -- if we were in TX just start again
        if alreadyInTx then
            box.begin()
        end

        return 0
    end

    return count
end

function util.retry(count, fn)
    checks('number', 'function')
    local err

    for i = 0, count do
        local status, res = pcall(function() return fn() end)
        if status == true then
            return res, nil
        end
        err = res
    end

    return nil, err
end

function util.migrator_upgrade(path)
    checks('string')

    local migrator = require('migrator')
    local loader = require('migrator.directory-loader').new(path)

    migrator.set_loader(loader)
    return migrator.upgrade()
end

function util.return_not_nil(value, default)
    return (value == nil or value == box.NULL) and default or value
end

function util.is_value_in(value, array)
    checks('?', 'table')
    for _, elem in ipairs(array) do
        if value == elem then
            return true
        end
    end

    return false
end

function util.tostring(value)
    if type(value) == 'table' then
        local str = ''
        for i, v in ipairs(value) do
            if i > 1 then
                str = str .. ', '
            end
            str = str .. util.tostring(v)
        end
        if str == '' then
            local i = 1
            for k, v in pairs(value) do
                if i > 1 then
                    str = str .. ', '
                end
                str = str .. tostring(k) .. ' = ' .. util.tostring(v)
                i = i + 1
            end
        end
        return '{' .. str .. '}'
    end
    return tostring(value)
end

function util.is_nil(v)
    return (v == nil or v == box.NULL)
end

return util
