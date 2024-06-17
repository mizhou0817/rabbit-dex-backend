return {
    up = function()
        local archiver = require('app.archiver')
        local ddl = require('app.ddl')
        local d = require('decimal')

        if box.space.order_tmp ~= nil then
            box.space.order_tmp:drop()
        end

        local sp = box.space['order']
        if sp == nil then
            error('space `order` not found')
        end

        local idx = sp.profile_status
        if idx ~= nil then
            idx:drop()
        end

        local fmt, err = archiver.format(sp)
        if err ~= nil then
            error(err)
        end
        if ddl.has_column(fmt.columns, 'trigger_price') then
            return
        end
        local last_field_no = #fmt.columns
        table.extend(fmt.columns, {
            {name = 'trigger_price', type = 'decimal', is_nullable = true},
            {name = 'size_percent', type = 'decimal', is_nullable = true},
            {name = 'time_in_force', type = 'string', is_nullable = true},
            {name = 'created_at', type = 'number', is_nullable = true},
            {name = 'updated_at', type = 'number', is_nullable = true},
        })

        local tmp_sp, err = archiver.create('order_tmp', fmt.options, fmt.columns, fmt.indices)
        if err ~= nil then
            error(err)
        end

        for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
            -- raises error
            tmp_sp:insert(tuple:transform(last_field_no+1, 0, d.new(0), d.new(0), 'good_till_cancel', 0, 0))
        end

        ddl.alter_column(tmp_sp, {name = 'trigger_price', type = 'decimal'})
        ddl.alter_column(tmp_sp, {name = 'size_percent', type = 'decimal'})
        ddl.alter_column(tmp_sp, {name = 'time_in_force', type = 'string'})
        ddl.alter_column(tmp_sp, {name = 'created_at', type = 'number'})
        ddl.alter_column(tmp_sp, {name = 'updated_at', type = 'number'})

        sp:drop()
        tmp_sp:rename(sp.name)
    end
}
