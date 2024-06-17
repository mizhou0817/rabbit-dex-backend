local checks = require('checks')
local log = require('log')

local engine = require('app.engine')
local client_methods = require('app.engine.methods').client_methods
local errors = require('app.lib.errors')
local p = require('app.profile')
local rpc = require('app.rpc')

local APISpacesError = errors.new_class("API_SPACES")

local getters = {}

function getters.get_profile(profile_id)
    local res = rpc.callro_profile("get", {profile_id})

    if res["error"] ~= nil then
        return nil, res["error"]
    end

    return p.bind(res["res"]), nil
end

function getters.get_market(market_id)
    checks('string')

    local res = rpc.callro_engine(market_id, client_methods.GET_MARKET, {market_id})
    if res["error"] ~= nil then
        return nil, res["error"]
    end

    return engine.market.bind(res["res"]), nil
end

function getters.load_profile_and_market(profile_id, market_id)
    checks('number', 'string')

    local profile, err = getters.get_profile(profile_id)
    if err ~= nil then
        return nil, nil, err
    end
    if profile == nil then
        local text = "profile_id=" .. tostring(profile_id) .. "not found"
        return nil, nil, text
    end

    local market, err = getters.get_market(market_id)
    if err ~= nil then
        log.error(APISpacesError:new(err))
        return nil, nil, err
    end

    return profile, market, nil
end

return getters