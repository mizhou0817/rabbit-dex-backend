local checks = require('checks')
local log = require('log')

local errors = require('app.lib.errors')
local time = require('app.lib.time')
local util = require('app.util')

require('app.lib.table')

local ArchiverError = errors.new_class("ARCHIVER")

local A = {}
local G_KEY = "G_ARCHIVER"

local GET_BATCH_LIMIT = 1000
local FIRST_COLUMN = 'shard_id'

local function shard_to_sequencer(shard_id)
    return tostring(shard_id) .. "_archive_id_sequencer"
end

function A.get_shard_id()
    local obj = rawget(_G, G_KEY)
    if obj == nil then
        log.error("get_shard_id error: obj is nil for G_KEY=%s", G_KEY)
        return nil
    end

    return obj.shard_id
end

function A.next_archive_id()
    local obj = rawget(_G, G_KEY)
    if obj == nil then
        log.error("next_archive_id error: obj is nil for G_KEY=%s", G_KEY)
        return nil
    end

    return obj.next_archive_id()
end

function A.get_next_batch(space, from_archive_id, limit)
    checks('string', 'uint64', 'uint64')

    local s = box.space[space]
    if s == nil then
        return {res = nil, error = "space '" .. space .. "' does not exist"}
    end

    local idx = s.index['archive_id']
    if idx == nil then
        return {res = nil, error = "archive_id index is missing on space '" .. space .. "'"}
    end

    local res = {
        timestamp = time.now(),
        columns = s:format(),
        data = {}
    }
    if limit == 0 then
        limit = s:len()
    end
    local processed_records = 0
    local count = 0
    local data = {}

    for _, rec in idx:pairs(from_archive_id, {iterator=box.index.GT}) do
        table.insert(data, rec)
        count = util.safe_yield(count, GET_BATCH_LIMIT)
        processed_records = processed_records + 1

        if processed_records >= limit then
            break
        end
    end
    res.data = data

    return {res = res, error = nil}
end

-- DEVELOPER control this - should be called as a first init per role
function A.init_sequencer(shard_id)
    checks('string')

    local archive_id_sequencer = shard_to_sequencer(shard_id)
    local seq = box.schema.sequence.create(archive_id_sequencer,{start=1, min=1, if_not_exists = true})

    rawset(_G, G_KEY, {
        shard_id = shard_id,
        next_archive_id = function() return box.sequence[archive_id_sequencer]:next() end
    })

    return seq
end

function A.insert(space, row)
    checks('table', 'table')

    local shard_id = A.get_shard_id()
    if shard_id == nil then
        log.error("shard_id is required for space_id=%d", space.id)
        return nil, "SHARD_ID_REQUIRED"
    end
    local archive_id = A.next_archive_id()

    table.extend(row, {
        shard_id,
        archive_id,
    })
    local status, res = pcall(function() return space:insert(row) end)

    if status == false then
        log.error(ArchiverError:new(res))
        return nil, res
    end

    return res, nil
end

function A.update(space, id, ops)
    checks('table', 'string|number|table', 'table')

    local shard_id = A.get_shard_id()
    if shard_id == nil then
        log.error("shard_id is required for space_id=%d", space.id)
        return nil, "SHARD_ID_REQUIRED"
    end
    local archive_id = A.next_archive_id()

    table.extend(ops, {
        {'=', 'shard_id', shard_id},
        {'=', 'archive_id', archive_id}
    })
    local status, res = pcall(function() return space:update(id, ops) end)

    if status == false then
        log.error(ArchiverError:new(res))
        return nil, res
    end

    return res, nil
end

function A.upsert(space, row, ops)
    checks('table', 'table', 'table')

    local shard_id = A.get_shard_id()
    if shard_id == nil then
        log.error("shard_id is required for space_id=%d", space.id)
        return nil, "SHARD_ID_REQUIRED"
    end
    local archive_id = A.next_archive_id()

    table.extend(row, {
        shard_id,
        archive_id,
    })
    table.extend(ops, {
        {'=', 'shard_id', shard_id},
        {'=', 'archive_id', archive_id}
    })
    local status, res = pcall(function() return space:upsert(row, ops) end)

    if status == false then
        log.error(ArchiverError:new(res))
        return nil, res
    end

    return res, nil
end

function A.replace(space, row)
    checks('table', 'table')

    local shard_id = A.get_shard_id()
    if shard_id == nil then
        log.error("shard_id is required for space_id=%d", space.id)
        return nil, "SHARD_ID_REQUIRED"
    end
    local archive_id = A.next_archive_id()

    table.extend(row, {
        shard_id,
        archive_id,
    })
    local status, res = pcall(function() return space:replace(row) end)

    if status == false then
        log.error(ArchiverError:new(res))
        return nil, res
    end

    return res, nil
end

-- returns (space object, error)
function A.create(space_name, options, format, index)
    checks('string', 'table', 'table', 'table')

    format = table.copy(format)
    table.extend(format, {
        {name = 'shard_id', type = 'string'},
        {name = 'archive_id', type = 'number'},
    })

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
            log.error("failed to create primary index for space: %s", space_name)
            return space, res
        end
    end

    status, res = pcall(function() return space:create_index('archive_id', {
        unique = false,
        parts = {{field = 'archive_id'}},
        if_not_exists = options.if_not_exists,
    }) end)
    if status == false then
        log.error("failed to create archive_id index for space: %s error: %s", space_name, ArchiverError:new(res))
        return space, res
    end

    return space, nil
end

-- returns (space format{columns,indices,options}, error)
function A.format(space, options)
    checks('table', '?table')
    local space_fmt = {}

    local arch_fmt = space:format()
    local fmt, is_arch, shard_id_no = {}, false, 0

    for icol, col in pairs(arch_fmt) do
        -- archiver columns should be the last and first of archiver columns is shard_id
        -- so cut off all archiver columns
        if col.name == FIRST_COLUMN then
            is_arch = true
            shard_id_no = icol
            break
        end
        table.insert(fmt, col)
    end

    if not is_arch then
        return nil, 'NOT_ARCHIVER_SPACE'
    end

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
            -- archiver columns should be the last, so skip indices for these columns
            if part.fieldno >= shard_id_no then
                parts = {}
                break
            end
            table.insert(parts, {field = part.fieldno, type = part.type})
        end

        if #parts > 0 then
            table.insert(indices, {name = idx.name, unique = idx.unique, parts = parts})
        end
    end

    space_fmt.indices = indices

    return space_fmt, nil
end

return A
