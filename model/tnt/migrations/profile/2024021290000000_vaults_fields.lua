local log = require('log')


local function vaults_field_migration()
    local archiver = require('app.archiver')
    local ddl = require('app.ddl')

    local sp = box.space['vaults']
    if sp == nil then
        error('space `vaults` not found')
    end

    local fmt, err = archiver.format(sp)
    if err ~= nil then
        error(err)
    end
    local vault_name = 'vault_name'
    if ddl.has_column(fmt.columns, vault_name) then
        return

    end
    table.extend(fmt.columns, {
        { name = 'vault_name',  type = 'string' , is_nullable = true},
        { name = 'manager_name',  type = 'string' , is_nullable = true},
        { name = 'initialised_at',  type = 'number' , is_nullable = true},
    })

    local field_no = #fmt.columns - 2
    

    local tmp_sp, err = archiver.create('vaults_tmp', fmt.options, fmt.columns, fmt.indices)
    if err ~= nil then
        error(err)
    end

    for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
        -- raises error
        tmp_sp:insert(tuple:transform(field_no, 0, "elixir", "elixir", tonumber(1707736522000000)))
    end

    ddl.alter_column(tmp_sp, { name = 'vault_name',  type = 'string' , is_nullable = false})
    ddl.alter_column(tmp_sp, { name = 'manager_name',  type = 'string' , is_nullable = false})
    ddl.alter_column(tmp_sp, { name = 'initialised_at',  type = 'number' , is_nullable = false})


    sp:drop()
    tmp_sp:rename(sp.name)

    return

end

return {
    up = vaults_field_migration
}