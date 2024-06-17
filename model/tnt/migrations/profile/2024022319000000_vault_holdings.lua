local log = require('log')
local decimal = require('decimal')


local function vault_holdings_field_migration()
    local archiver = require('app.archiver')
    local ddl = require('app.ddl')

    local sp = box.space['vault_holdings']
    if sp == nil then
        error('space `vault_holdings` not found')
    end

    local fmt, err = archiver.format(sp)
    if err ~= nil then
        error(err)
    end
    local entry_price = 'entry_price'
    if ddl.has_column(fmt.columns, entry_price) then
        return
    end
    table.extend(fmt.columns, {
        { name = 'entry_price',  type = 'decimal' , is_nullable = true},
    })

    local field_no = #fmt.columns
    

    local tmp_sp, err = archiver.create('vault_holdings_tmp', fmt.options, fmt.columns, fmt.indices)
    if err ~= nil then
        error(err)
    end

    for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
        -- raises error
        tmp_sp:insert(tuple:transform(field_no, 0, decimal.new(1)))
    end

    ddl.alter_column(tmp_sp, { name = 'entry_price',  type = 'decimal' , is_nullable = false})

    sp:drop()
    tmp_sp:rename(sp.name)

    return

end

return {
    up = vault_holdings_field_migration
}