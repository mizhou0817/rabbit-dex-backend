return {
    up = function()
        if box.space.profile_cache ~= nil then
            box.space.profile_cache:truncate()
            box.space.profile_cache:alter({temporary = false})
        end

        if box.space.profile_meta ~= nil then
            box.space.profile_meta:truncate()
            box.space.profile_meta:alter({temporary = false})
        end

        if box.space.exchange_total ~= nil then
            box.space.exchange_total:truncate()
            box.space.exchange_total:alter({temporary = false})
        end
    end
}
