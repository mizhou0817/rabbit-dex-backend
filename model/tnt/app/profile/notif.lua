local json = require('json')
local log = require('log')

local errors = require('app.lib.errors')
local rpc = require('app.rpc')

local err = errors.new_class("notif_error")

local notif = {}

function notif.notify_profiles(profiles)
    -- UPDATE all by default
    local iter = function(do_notify) 
        for _, cache in box.space.profile_cache:pairs(nil, {iterator = box.index.ALL}) do
            do_notify(cache.id, cache)
        end
    end

    if #profiles ~= 0 then
        iter = function(do_notify) 
            for _, profile_id in pairs(profiles) do
                local cache = box.space.profile_cache:get(profile_id)
                if cache == nil then
                    local text = "NO CACHE for profile=" .. tostring(profile_id)
                    log.error("%s", err:new(text))
                else
                    do_notify(profile_id, cache)
                end
            end
        end
    end

    -- SEND notif
    iter(
        function(profile_id, cache)
            local channel = "account@" .. tostring(profile_id)
            local json_update = json.encode({data=cache:tomap({names_only=true})})
            rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
            json_update = nil
        end)
end

--TODO: refactor as notif:new(rpc), notif:notify_profiles() and remove this func
function notif._test_set_rpc(override_rpc)
    rpc = override_rpc
end

return notif
