local fio = require('fio')
local t = require('luatest')

local a = require('app.archiver')

local g = t.group('archiver')

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
    local seq = a.init_sequencer('shard')
    t.assert_is_not(seq, nil)

    local space, err = a.create('test.archiver', {temporary = true}, {
        {name = 'id', type = 'unsigned'},
        {name = 'val', type = 'number'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
    })
    t.assert_is(err, nil)
    t.assert_is_not(space, nil)

    cg.params = {
        seq = seq,
        space = space,
    }
end)

g.after_each(function(cg)
    cg.params.space:drop()
    cg.params.seq:drop()
end)

g.test_create_success = function(cg)
    local space = cg.params.space

    t.assert_equals(space:format(), {
        {name = 'id', type = 'unsigned'},
        {name = "val", type = "number"},
        {name = "shard_id", type = "string"},
        {name = "archive_id", type = "number"},
    })
end

g.test_create_bad_primary_index = function(cg)
    local space, err = a.create('test.archiver.create_bad_primary', {temporary = true}, {
        {name = 'id', type = 'unsigned'},
    }, {
        unique = false,
        parts = {{field = 'id'}},
    })

    t.assert_str_contains(tostring(err), 'primary key must be unique')
end

g.test_create_already_exists = function(cg)
    local space, err
    for i = 0, 1, 1 do
        space, err = a.create('test.archiver.create_exists', {temporary = true}, {
            {name = 'id', type = 'unsigned'},
        }, {
            unique = true,
            parts = {{field = 'id'}},
        })
    end

    t.assert_str_contains(tostring(err), 'already exists')
end

g.test_create_dup_archive_id = function(cg)
    local space, err = a.create('test.archiver.create_dup_archive_id', {temporary = true}, {
        {name = 'id', type = 'unsigned'},
        {name = "archive_id", type = "number"},
    }, {
        unique = true,
        parts = {{field = 'id'}},
    })

    t.assert_str_contains(tostring(err), "Space field 'archive_id' is duplicate")
end

g.test_insert_success = function(cg)
    local space = cg.params.space

    for _, rec in pairs({{1, 1}, {2, 1}}) do
        local _, err = a.insert(space, rec)
        t.assert_is(err, nil)
    end

    t.assert_equals(
        space:select(0, {iterator = 'GT', offset = 0}),
        {{1, 1, 'shard', 1}, {2, 1, 'shard', 2}}
    )
end

g.test_insert_dup = function(cg)
    local space = cg.params.space

    local _, err = a.insert(space, {1, 1})
    t.assert_is(err, nil)

    _, err = a.insert(space, {1, 2})
    t.assert_str_contains(tostring(err), 'Duplicate key exists')
end

g.test_replace_success = function(cg)
    local space = cg.params.space

    for _, rec in pairs({{1, 1}, {2, 1}, {1, 2}}) do
        local _, err = a.replace(space, rec)
        t.assert_is(err, nil)
    end

    t.assert_equals(
        space:select(0, {iterator = 'GT', offset = 0}),
        {{1, 2, 'shard', 3}, {2, 1, 'shard', 2}}
    )
end

g.test_update_success = function(cg)
    local space = cg.params.space

    for _, rec in pairs({{1, 1}, {2, 2}, {3, 3}}) do
        local _, err = a.insert(space, rec)
        t.assert_is(err, nil)
    end

    local res, err = a.update(space, 2, {{'=', 'val', 22}})
    t.assert_is(err, nil)
    t.assert_is_not(res, nil)
    t.assert_equals(
        space:select(0, {iterator = 'GT', offset = 0}),
        {{1, 1, 'shard', 1}, {2, 22, 'shard', 4}, {3, 3, 'shard', 3}}
    )
end

g.test_update_non_existent = function(cg)
    local space = cg.params.space

    local res, err = a.update(space, 22, {{'=', 'val', 22}})
    t.assert_is(err, nil)
    t.assert_is(res, nil) -- record not found
end

g.test_upsert_success = function(cg)
    local space = cg.params.space

    local _, err
    _, err = a.upsert(space, {1, 1}, {{'=', 'val', 11}})
    t.assert_is(err, nil)
    _, err = a.upsert(space, {2, 2}, {{'=', 'val', 22}})
    t.assert_is(err, nil)
    _, err = a.upsert(space, {3, 3}, {{'=', 'val', 33}})
    t.assert_is(err, nil)

    t.assert_equals(
        space:select(0, {iterator = 'GT', offset = 0}),
        {{1, 1, 'shard', 1}, {2, 2, 'shard', 2}, {3, 3, 'shard', 3}}
    )

    _, err = a.upsert(space, {2, 2}, {{'=', 'val', 22}})
    t.assert_is(err, nil)
    t.assert_equals(
        space:select(0, {iterator = 'GT', offset = 0}),
        {{1, 1, 'shard', 1}, {2, 22, 'shard', 4}, {3, 3, 'shard', 3}}
    )
end

g.test_format = function(cg)
    local fmt, err = a.format(box.space['test.archiver'])
    t.assert_is(err, nil)
    t.assert_equals(
        fmt.columns,
        {{name = 'id', type = 'unsigned'}, {name = 'val', type = 'number'}}
    )
    t.assert_equals(
        fmt.options,
        {temporary = true, is_local = false, is_sync = false, field_count = 0}
    )
    t.assert_equals(
        fmt.indices,
        {{name = 'primary', parts = {{field = 1, type = 'unsigned'}}, unique = true}}
    )
end
