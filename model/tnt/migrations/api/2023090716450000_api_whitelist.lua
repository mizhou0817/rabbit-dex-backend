local ids = {
    13,
    14,
    15,
    19,
    20,
    704,
    1561,
    2104,
    11002,
    11007,
    11831,
    12071,
    12072,
    12472,
    12483,
    12505,
    13436,
    14919,
    15679,
}
return {
    up = function()
        local ddl = require('app.ddl')

        local sp = box.space['white_list']
        if sp == nil then
            return
        end


        if sp.temporary == false then
            return
        end

        -- we have old version of white_list which is temporary
        -- 1. need to create persistent
        -- 2. add ids to white_list
        box.space.white_list:drop()

        local white_list = box.schema.space.create('white_list', {if_not_exists = true})
        white_list:format({
            {name = 'profile_id', type = 'unsigned'}
        })
        white_list:create_index('primary', {
            unique = true,
            parts = {{field = 'profile_id'}},
            if_not_exists = true })
    
        for _, id in ipairs(ids) do
            box.space.white_list:insert{id}
        end
    end
}
