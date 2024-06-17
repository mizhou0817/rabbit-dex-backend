local checks = require('checks')
local log = require('log')

local action = require('app.action')
local engine = require('app.engine.engine')
local errors = require('app.lib.errors')

require("app.config.constants")
require('app.errcodes')

local MError = errors.new_class("MAINTENANCE_ERROR")

local MT = {}

function MT.init_spaces()
    local affected_profiles = box.schema.space.create('affected_profiles', {if_not_exists = true})
    affected_profiles:format({
        {name = 'profile_id', type = 'unsigned'}
    })
    affected_profiles:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })


    return nil
end

function MT.add_profile_id(profile_id)
    checks('number')

    local res = box.space.affected_profiles:insert{profile_id}

    return {res = res, error = nil}
end

function MT.remove_profile_id(profile_id)
    checks('number')

    local res = box.space.affected_profiles:delete{profile_id}

    return {res = res, error = nil}
end


-- will cancel all orders for profiles listed in affected_profiles
function MT.cancel_all_listed()
    local market = box.space.market.index.primary:min()
    if market == nil then
        return {res = nil, error = tostring(MError:new('MAINTENANCE_ERROR: cancel_all_listed market not exist'))}
    end
    
    local total = 0
    for _, ap in box.space.affected_profiles:pairs() do 
        local order, _ = action.pack_cancelall(ap.profile_id, market.id)
        if order ~= nil then
            -- cancel_all without notif
            local status, val = xpcall(function()
                    return engine._handle_cancelall(order)        
            end, debug.traceback)

            if status == false then
                log.error(val)
            else
                if val ~= nil then
                    log.error(val)
                end
            end                
        else
            log.error("no order")
        end

        total = total + 1
    end

    log.info({
        message = string.format("cancel_all_listed: market_id=%s profiles total=%d", tostring(market.id), total),
        [ALERT_TAG] = ALERT_LOW,
    })

    return {res = nil, error = nil}
end

return MT