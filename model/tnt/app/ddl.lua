local checks = require('checks')

local ddl = {}

function ddl.alter_column(space, col_fmt, opts, on_alter_cb)
    checks('table', 'table', '?table', '?function')
    if opts == nil then
        opts = {if_not_exists = false}
    end

    local fmt = space:format()
    local idx

    for i, c in ipairs(fmt) do
        if col_fmt.name == c.name then
            idx = i
            break
        end
    end

    if opts.if_not_exists == true and idx ~= nil then
        return
    end

    if idx ~= nil then
        fmt[idx] = col_fmt
    else
        table.insert(fmt, col_fmt)
    end

    space:format(fmt)

    if on_alter_cb ~= nil then
        on_alter_cb()
    end
end

function ddl.has_column(fmt, col_name)
    checks('table', 'string')

    for _, col in pairs(fmt) do
        if col_name == col.name then
            return true
        end
    end

    return false
end

-- returns (space format{columns,indices,options}, error)
function ddl.format(space, options)
    checks('table', '?table')
    local space_fmt = {}

    local fmt = space:format()

    space_fmt.columns = fmt
    space_fmt.options = {
        temporary   = space.temporary,
        is_local    = space.is_local,
        is_sync     = space.is_sync,
        field_count = space.field_count,
    }

    local indices = {}

    local idx_no = 0
    while true do
        local idx = space.index[idx_no]
        if idx == nil then
            break
        end
        idx_no = idx_no + 1

        local parts = {}
        for _, part in pairs(idx.parts) do
            table.insert(parts, {field = part.fieldno, type = part.type})
        end

        if #parts > 0 then
            table.insert(indices, {name = idx.name, unique = idx.unique, parts = parts})
        end
    end

    space_fmt.indices = indices

    return space_fmt
end

-- returns (space object, error)
function ddl.create_space(space_name, options, format, index)
    checks('string', 'table', 'table', 'table')

    local indices = index
    if index.parts ~= nil then
        -- backward compat for just primary index
        index.name = 'primary'
        indices = {index}
    end
    options.format = format

    local status, res

    if options.if_not_exists == true and box.space[space_name] ~= nil then
        return box.space[space_name], nil
    end

    status, res = pcall(function() return box.schema.space.create(space_name, options) end)
    if status == false then
        log.error("failed to create space: %s error: %s", space_name, ArchiverError:new(res))
        return nil, res
    end
    local space = res

    for _, idx in pairs(indices) do
        status, res = pcall(function() return space:create_index(idx.name, {
            unique = idx.unique,
            parts = idx.parts,
            if_not_exists = options.if_not_exists,
            sequence = idx.sequence,
        }) end)
        if status == false then
            log.error("failed to create index for space: %s", space_name)
            return space, res
        end
    end

    return space, nil
end

return ddl
