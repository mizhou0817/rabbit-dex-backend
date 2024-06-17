return {
    up = function()
        local archiver = require('app.archiver')
        local ddl = require('app.ddl')

        local sp = box.space['order']
        if sp == nil then
            error('space `order` not found')
        end

        local fmt, err = archiver.format(sp)
        if err ~= nil then
            error(err)
        end
        local client_order_id = 'client_order_id'
        if ddl.has_column(fmt.columns, client_order_id) then
            return
        end
        table.extend(fmt.columns, {
            {name = client_order_id, type = 'string', is_nullable = true},
        })
        local last_field_no = #fmt.columns

        local tmp_sp, err = archiver.create('order_tmp', fmt.options, fmt.columns, fmt.indices)
        if err ~= nil then
            error(err)
        end

        for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
            -- raises error
            tmp_sp:insert(tuple:transform(last_field_no, 0, ''))
        end

        ddl.alter_column(tmp_sp, {name = client_order_id, type = 'string'})

        sp:drop()
        tmp_sp:rename(sp.name)
    end
}
