local log = require('log')
local DEFAULT_EXCHANGE_ID = 'rbx'

local function new_indexes(sp)
    sp:create_index('exchange_id',
    {
        parts = { {field = 'exchange_id'} },
        unique = false,
        if_not_exists = true
    })

    sp:create_index('exchange_id_wallet', {
        unique = true,
        parts = {{field = 'exchange_id'}, {field = 'wallet'}},
        if_not_exists = true 
    })
end

local function eid_profile_migration()
    local archiver = require('app.archiver')
    local ddl = require('app.ddl')

    local sp = box.space['profile']
    if sp == nil then
        error('space `balance_operations` not found')
    end

    local fmt, err = archiver.format(sp)
    if err ~= nil then
        error(err)
    end

    local add_field = false

    
    local exchange_id = 'exchange_id'
    if ddl.has_column(fmt.columns, exchange_id) then
        local sequence_exist = false
        for i, idx in pairs(fmt.indices) do
            if idx.name == 'primary' and fmt.indices[i].sequence_fieldno == 1 then
                sequence_exist = true
                break
            end
        end

        if sequence_exist == true then
            return new_indexes(sp)
        end
    else
        add_field = true
        table.extend(fmt.columns, {
            { name = 'exchange_id',  type = 'string' , is_nullable = true},
        })
    end
    local eid_field_no = #fmt.columns
    
    -- ADD sequence to primary index
    -- remove unique wallet index

    local removed_index = nil
    for i, idx in pairs(fmt.indices) do
        if idx.name == 'primary' then
            fmt.indices[i].sequence = "PID"
        end

        if idx.name == "wallet" then 
            removed_index = i
        end
    end

    if removed_index ~= nil then
        table.remove(fmt.indices, removed_index)
    end

    --IMPORTANT: for profile we need SEQUENCE, which is by name only
    local tmp_sp, err = archiver.create('profile_tmp', fmt.options, fmt.columns, fmt.indices)
    if err ~= nil then
        error(err)
    end

    for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
        -- raises error
        if add_field == true then
            tmp_sp:insert(tuple:transform(eid_field_no, 0, DEFAULT_EXCHANGE_ID))
        else
            tmp_sp:insert(tuple)
        end
    end

    if add_field == true then
        ddl.alter_column(tmp_sp, { name = 'exchange_id',  type = 'string' , is_nullable = false})
    end 

    sp:drop()
    tmp_sp:rename(sp.name)

    return new_indexes(tmp_sp)
end

return {
    up = eid_profile_migration
}