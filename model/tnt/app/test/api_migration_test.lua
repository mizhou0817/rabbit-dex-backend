local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local migration = require('migrations.api.2023060616450000_api_coid')
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

local g = t.group('api_migration')
g.before_each(function(cg)
    local used_client_order_id = box.schema.space.create('used_client_order_id', {temporary=false, if_not_exists = true})
    used_client_order_id:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'client_order_id', type = 'string'},
        {name = 'market_id', type = 'string'},      -- we use market_id for invalidate later
    })

    used_client_order_id:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'client_order_id'}},
        if_not_exists = true })

    used_client_order_id:create_index('profile_market', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true })    

end)

g.after_each(function(cg)
    box.space.used_client_order_id:drop()
end)

g.test_market_migration = function(cg)
    
    -- Migrate to new schema
    migration.up()

    local sp = box.space['used_client_order_id']
    t.assert_is_not(sp, nil)

    log.info(sp:format())
end
