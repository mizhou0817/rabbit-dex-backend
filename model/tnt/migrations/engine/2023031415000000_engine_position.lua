return {
    up = function()
        local archiver = require('app.archiver')
        local ddl = require('app.ddl')

        local sp = box.space['position']
        if sp == nil then
            error('space `position` not found')
        end

        local shard_id = archiver.get_shard_id()

        ddl.alter_column(sp, {name = 'shard_id', type = 'string', is_nullable = true}, {if_not_exists = true},
            function()
                for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
                    sp:update(tuple[1], {{'=', 'shard_id', shard_id}})
                end
            end
        )
        ddl.alter_column(sp, {name = 'shard_id', type = 'string'})

        ddl.alter_column(sp, {name = 'archive_id', type = 'number', is_nullable = true}, {if_not_exists = true},
            function()
                for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
                    sp:update(tuple[1], {{'=', 'archive_id', archiver.next_archive_id()}})
                end
            end
        )
        ddl.alter_column(sp, {name = 'archive_id', type = 'number'})

        sp:create_index('archive_id', {
            unique = false,
            parts = {{field = 'archive_id'}},
            if_not_exists = true,
        })
    end
}
