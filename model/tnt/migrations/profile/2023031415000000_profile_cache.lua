return {
    up = function()
        local ddl = require('app.ddl')

        local sp = box.space['profile_cache']
        if sp == nil then
            error('space `profile_cache` not found')
        end

        ddl.alter_column(sp, {name = 'shard_id', type = 'string'}, {if_not_exists = true})
        ddl.alter_column(sp, {name = 'archive_id', type = 'number'}, {if_not_exists = true})

        sp:create_index('archive_id', {
            unique = false,
            parts = {{field = 'archive_id'}},
            if_not_exists = true,
        })
    end
}
