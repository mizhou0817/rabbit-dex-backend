return {
    up = function()
        local ddl = require('app.ddl')

        local sp = box.space['profile']
        if sp == nil then
            error('space `profile` not found')
        end

        ddl.alter_column(sp, {name = 'created_at', type = 'number', is_nullable = true}, { if_not_exists = true },
            function()
                for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
                    sp:update(tuple[1], {{'=', 'created_at', 0}})
                end
            end
        )
        ddl.alter_column(sp, {name = 'created_at', type = 'number'})

        sp:create_index('created_at', {
            unique = false,
            parts = {{field = 'created_at'}},
            if_not_exists = true,
        })
    end
}
