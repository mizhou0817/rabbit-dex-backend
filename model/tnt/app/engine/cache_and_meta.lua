local log = require('log')

local errors = require('app.lib.errors')
local rpc = require('app.rpc')

require("app.config.constants")

local CMERROR = errors.new_class("cache_and_meta")

return {
    handle_get_cache_and_meta = function(router, profile_ids, market_id, return_data_for_id)
        local data = rpc.callrw_profile("get_cache_and_meta", {profile_ids, market_id})
        if data["error"] ~= nil then
            log.error(CMERROR:new(data["error"]))
            return nil, data["error"]
        end
    
        local _data = data["res"]
        if _data == nil then
            local text = "NO_RES_RETURNED_FOR_CACHE_AND_META market_id=" .. tostring(market_id) .. " ids=" .. tostring(profile_ids)
            return nil, text
        end
    
        local profile_data = nil
        for _, d in pairs(_data) do
            if d.profile_id == return_data_for_id then
                profile_data = {
                    cache = d.cache,
                    meta = d.meta
                }
                break
            end
        end
        if profile_data == nil then
            local text = "NO_CACHE market_id=" .. tostring(market_id) .. " profile_id=" .. tostring(profile_ids)
            log.error(CMERROR:new(text))
            return nil, text
        end
        
        if router ~= nil then
            router.save_profile_data(_data)
        end
        
        return profile_data, nil
    end,
}


