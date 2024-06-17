local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local vaults_migration = require('migrations.profile.2024021290000000_vaults_fields')
local config = require('app.config')
local ddl = require('app.ddl')


require('app.config.constants')
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

local g = t.group('vaults_migration')
g.before_each(function(cg)
    log.info("**** started:")
    -- Create old market and init
    archiver.init_sequencer("BTC-USD")

    -- CREATE old schema and add some records
    local old_vaults, err = archiver.create('vaults', { if_not_exists = true }, {
        { name = 'vault_profile_id',     type = 'unsigned' },
        { name = 'manager_profile_id',   type = 'unsigned' },
        { name = 'treasurer_profile_id', type = 'unsigned' },
        { name = 'performance_fee',      type = 'decimal' },
        { name = 'status',               type = 'string' },
        { name = 'total_shares',         type = 'decimal' },
    }, {
        unique = true,
        parts = { { field = 'vault_profile_id' } },
        if_not_exists = true,
    })
    t.assert_is_not(old_vaults, nil)
    t.assert_is(err, nil)

    for _, val in pairs({1,2,3,4,5}) do 
        local res, err = archiver.insert(box.space.vaults, {
            val,
            val,
            val,
            decimal.new(0),
            "active",
            decimal.new(0)
        })
        t.assert_is(err, nil)
        log.info(res)
    end
end)

g.after_each(function(cg)
    box.space.vaults:drop()
end)

g.test_vaults_migration = function(cg)

    local vaults_new_format = {
        { name = 'vault_profile_id',     type = 'unsigned' },
        { name = 'manager_profile_id',   type = 'unsigned' },
        { name = 'treasurer_profile_id', type = 'unsigned' },
        { name = 'performance_fee',      type = 'decimal' },
        { name = 'status',               type = 'string' },
        { name = 'total_shares',         type = 'decimal' },
 
        { name = 'vault_name',           type = 'string', is_nullable = false},
        { name = 'manager_name',         type = 'string', is_nullable = false},
        { name = 'initialised_at',       type = 'number', is_nullable = false },

        {name = "shard_id", type = "string"},
        {name = "archive_id", type = "number"},
    }
    
    -- let's try to do it twice

    for _, step in pairs{1,2} do
        -- Migrate to new schema
        vaults_migration.up()

        local sp = box.space['vaults']
        t.assert_is_not(sp, nil)
    
        -- Check that format migrated
        t.assert_equals(sp:format(), vaults_new_format)
    
        -- check that vaues initialized correclty
        local some_item = sp:get(1)
        t.assert_is(some_item[7], 'elixir')
    end
end