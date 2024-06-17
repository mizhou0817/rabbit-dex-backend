return {
    up = function()
        local ddl = require('app.ddl')

        local sp = box.space['inv3_data']
        if sp == nil then
            error('space `inv3_data` not found')
        end

        ddl.alter_column(sp, {name = 'margined_ae_sum', type = 'decimal'}, {if_not_exists = true})
        ddl.alter_column(sp, {name = 'exchange_balance', type = 'decimal'}, {if_not_exists = true})
        ddl.alter_column(sp, {name = 'insurance_balance', type = 'decimal'}, {if_not_exists = true})

        ddl.alter_column(sp, {name = 'shard_id', type = 'string'}, {if_not_exists = true})
        ddl.alter_column(sp, {name = 'archive_id', type = 'number'}, {if_not_exists = true})

        sp:create_index('archive_id', {
            unique = false,
            parts = {{field = 'archive_id'}},
            if_not_exists = true,
        })
    end
}
