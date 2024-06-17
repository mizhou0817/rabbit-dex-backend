local DEFAULT_EXCHANGE_ID = 'rbx'

local function drop_migration(ddl)
    local create_processed_blocks = function ()
        local processed_blocks = box.schema.space.create('processed_blocks', { if_not_exists = true })
        processed_blocks:format({
            { name = 'contract_address', type = 'string' },
            { name = 'chain_id',  type = 'unsigned'},
            { name = 'event_type',  type = 'string'},
            { name = 'last_processed_block',  type = 'string' },
        })

        processed_blocks:create_index('primary', {
            unique = true,
            parts = { { field = 'contract_address' }, {field = 'chain_id'},  {field = 'event_type'} },
            if_not_exists = true
        })        
    end

    local sp = box.space['processed_blocks']
    if sp == nil then
        create_processed_blocks()
    end
end

local function new_indexes(sp)
    sp:create_index('exchange_id',
    {
        parts = { {field = 'exchange_id'} },
        unique = false,
        if_not_exists = true
    })

    sp:create_index('chain_id_contract_address',
    {
        parts = { {field = 'chain_id'}, {field = 'contract_address'} },
        unique = false,
        if_not_exists = true
    })

    sp:create_index('exchange_id_wallet_type_status',
        {
            parts = { {field = 'exchange_id'}, { field = 'wallet' }, { field = 'ops_type' }, { field = 'status' } },
            unique = false,
            if_not_exists = true
        })    

    sp:create_index('type_contract_due_block', 
    {parts = {
        {field = 'ops_type'}, 
        {field = 'contract_address'},
        {field = 'due_block'}},
        unique = false,
        if_not_exists = true })

    sp:create_index('type_status_exchange_id_chain_id', {
        parts = { { field = 'ops_type' }, { field = 'status' }, { field = "exchange_id"}, { field = "chain_id" } },
        unique = false,
        if_not_exists = true
    })
end

local function eid_balance_migration()
    local archiver = require('app.archiver')
    local ddl = require('app.ddl')

    drop_migration(ddl)

    local sp = box.space['balance_operations']
    if sp == nil then
        error('space `balance_operations` not found')
    end

    local fmt, err = archiver.format(sp)
    if err ~= nil then
        error(err)
    end
    local balance_operations_eid = 'exchange_id'
    if ddl.has_column(fmt.columns, balance_operations_eid) then
        return new_indexes(sp)
    end
    table.extend(fmt.columns, {
        { name = 'exchange_id',  type = 'string' , is_nullable = true},
        { name = 'chain_id',  type = 'unsigned' , is_nullable = true},
        { name = 'contract_address',  type = 'string' , is_nullable = true},
    })
    local ca_field_no = #fmt.columns
    local chain_id_field_no = #fmt.columns - 1
    local eid_field_no = #fmt.columns - 2
    

    local tmp_sp, err = archiver.create('balance_operations_tmp', fmt.options, fmt.columns, fmt.indices)
    if err ~= nil then
        error(err)
    end

    for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
        -- raises error
        tmp_sp:insert(tuple:transform(eid_field_no, 0, DEFAULT_EXCHANGE_ID, 1, "0xrbxl1"))
    end

    ddl.alter_column(tmp_sp, { name = 'exchange_id',  type = 'string' , is_nullable = false})
    ddl.alter_column(tmp_sp, { name = 'chain_id',  type = 'unsigned' , is_nullable = false})
    ddl.alter_column(tmp_sp, { name = 'contract_address',  type = 'string' , is_nullable = false})


    sp:drop()
    tmp_sp:rename(sp.name)

    return new_indexes(tmp_sp)
end

return {
    up = eid_balance_migration,
    drop_migration = drop_migration,
}