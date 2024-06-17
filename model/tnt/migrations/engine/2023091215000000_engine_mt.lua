--TODO: received list from MING before mainnet
local ids = {
    14910,
    14919,
    2104,
    12071,
    12072,
    12483,
    13436,
    12472,
    15679
}

return {
    up = function()
        local sp = box.space['affected_profiles']
        if sp == nil then
           return
        end

        if sp:count() == 0 then
            for _, id in ipairs(ids) do 
                box.space.affected_profiles:replace{id}
            end            
        end
    end
}
