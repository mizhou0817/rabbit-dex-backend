return {
    up = function()
        local ddl = require('app.ddl')

        if box.space.profile_to_tier_tmp ~= nil then
            box.space.profile_to_tier_tmp:drop()
        end

        local sp = box.space['profile_to_tier']
        if sp == nil then
            error('space `profile_to_tier` not found')
        end

        local fmt = ddl.format(sp)
        if ddl.has_column(fmt.columns, 'tier_id') then
            return
        end
        local last_field_no = #fmt.columns
        table.extend(fmt.columns, {
            {name = 'tier_id', type = 'unsigned', is_nullable = true},
        })

        local tmp_sp, err = ddl.create_space('profile_to_tier_tmp', fmt.options, fmt.columns, fmt.indices)
        if err ~= nil then
            error(err)
        end

        for _, tuple in sp:pairs(nil, {iterator = box.index.ALL}) do
            -- raises error
            tmp_sp:insert(tuple:transform(last_field_no+1, 0, 0))
        end

        ddl.alter_column(tmp_sp, {name = 'tier_id', type = 'unsigned'})

        sp:drop()
        tmp_sp:rename(sp.name)
    end
}
