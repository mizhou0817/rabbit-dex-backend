local checks = require('checks')
local log = require('log')

local balance = require('app.balance')
local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local cache = require('app.profile.cache')
local integrity = require('app.profile.integrity')
local periodics = require('app.profile.periodics')
local tier = require('app.profile.tier')
local rpc = require('app.rpc')
local util = require('app.util')

require('app.lib.table')
require('app.config.constants')

local ProfileError = errors.new_class("PROFILE")

local SELECT_META_LIMIT = 1000000
local YIELD_LIMIT = 1000

local getters = {}

function getters.profile_meta(profile_id)
    checks("number")

    local res = box.space.profile_meta:select(profile_id, {iterator = 'EQ', limit=SELECT_META_LIMIT})
    return {res = res, error = nil}
end

function getters.get_profile_data(profile_id)
    checks("number")

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end

    local res = box.space.profile_cache:get(profile_id)
    if res == nil then
        return {res = nil, error = "PROFILE_CACHE_NOT_FOUND"}
    end 
    res = res:totable()

    -- TODO: remove that for prod
    res[21] = {}
    local p_res = getters.get_open_positions(profile_id)
    if p_res["error"] == nil and p_res.res ~= nil and #p_res.res > 0 then
        res[21] = p_res.res
    end

    res[22] = {}
    local os = config.params.ORDER_STATUS
    local o_res = getters.get_orders(profile_id, {os.OPEN, os.PLACED})
    if o_res["error"] == nil and o_res.res ~= nil and #o_res.res > 0 then
        res[22] = o_res.res
    end

    res[23] = {}

    return {res = res, error = nil}
end

function getters.get_extended_profile_data(profile_id)
    checks("number")

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end

    local res = box.space.profile_cache:get(profile_id)
    if res == nil  then
        return {res = nil, error = "PROFILE_CACHE_NOT_FOUND"}
    end
    res = res:totable()

    -- TODO: remove that for prod
    res[21] = {}
    local p_res = getters.get_extended_open_positions(profile_id)
    if p_res["error"] == nil and p_res.res ~= nil and #p_res.res > 0 then
        res[21] = p_res.res
    end

    res[22] = {}
    local os = config.params.ORDER_STATUS
    local o_res = getters.get_orders(profile_id, {os.OPEN, os.PLACED})
    if o_res["error"] == nil and o_res.res ~= nil and #o_res.res > 0 then
        res[22] = o_res.res
    end

    res[23] = {}

    return {res = res, error = nil}
end

function getters.get_extended_profiles(profiles_ids)
    checks('?table')

    local count = 0
    local res = {}

    if util.is_nil(profiles_ids) or #profiles_ids == 0 then
        local max = box.space.profile.index.primary:max()
        if max then
            for _, profile in box.space.profile:pairs(max.id, {iterator = box.index.LE}) do
                local p = profile:totable()
                table.insert(p, balance.get_balance(profile.id))
                table.insert(res, p)

                count = util.safe_yield(count, YIELD_LIMIT)
            end
        end
    else
        for _, profile_id in ipairs(profiles_ids) do
            local profile = box.space.profile:get(profile_id)
            if profile == nil then
                return { res = nil, error = "PROFILE_NOT_FOUND" }
            end

            local p = profile:totable()
            table.insert(p, balance.get_balance(profile.id))
            table.insert(res, p)

            count = util.safe_yield(count, YIELD_LIMIT)
        end
    end

    return { res = res, error = nil }
end

function getters.get_open_positions(profile_id, offset, limit)
    checks('number', '?number', '?number')

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end


    local positions = {}
    -- TODO: check for market_not found error. Right now just skip
    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callro_engine(market_id, "get_positions", {profile_id})
        if res["error"] == nil and res.res ~= nil and #res.res > 0 then
            table.extend(positions, res.res)
        end
    end

    -- positions = table.slice(positions, offset, offset + limit)

    return {res = positions, error = nil}
end

function getters.get_requested_unstakes(profile_id)
    checks('number')

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end


    local unstakes = {}

    local res = balance.list_operations_of_type_status(profile_id, config.params.BALANCE_TYPE.VAULT_UNSTAKE_SHARES, config.params.BALANCE_STATUS.REQUESTED, 0, 1000)
    if res.error ~= nil then
        return {res = nil, error = res.error}
    end
    if res.res ~= nil and #res.res > 0 then
        table.extend(unstakes, res.res)
    end

    return {res = unstakes, error = nil}
end

function getters.get_extended_open_positions(profile_id, offset, limit)
    checks('number', '?number', '?number')

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end


    local positions = {}
    -- TODO: check for market_not found error. Right now just skip
    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callro_engine(market_id, "get_extended_position", {profile_id, market_id})
        if res["error"] == nil and res.res ~= nil then
            table.insert(positions, res.res)
        end
    end

    -- positions = table.slice(positions, offset, offset + limit)

    return {res = positions, error = nil}
end

function getters.get_orders(profile_id, statuses)
    checks('number', 'table')

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return {res = nil, error = "PROFILE_NOT_FOUND"}
    end

    local orders = {}

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callro_engine(market_id, "get_orders", {profile_id, statuses})
        if res["error"] == nil and res.res ~= nil and #res.res > 0 then
            table.extend(orders, res.res)
        end
    end

    return {res = orders, error = nil}
end

function getters.get_open_orders(profile_id)
    checks('number')

    return getters.get_orders(profile_id, {config.params.ORDER_STATUS.OPEN})
end

function getters.get_exchange_data()
    local data = box.space.exchange_total:get(EXCHANGE_ID)
    if data == nil then
        return {res = nil, error = "EXCHANGE_TOTAL_NOT_FOUND"}
    end

    return {res = data, error = nil}
end


function getters.liquidation_batch(last_id_checked, limit)
    checks("?number", "number")
    
    local is_valid = integrity.is_valid()
    if not is_valid then
        log.error("is_inv3_valid: not valid")

        -- we just return nothing
        return {res = nil, error = "INV3_DATA_NOT_INITIALIZED"}  
    end

    local res
    if last_id_checked == nil then
        res = box.space.profile_cache.index.for_liquidation:select(
            {config.params.PROFILE_STATUS.LIQUIDATING, 0}, 
            {iterator = 'GE', limit = limit}
        )
    else
        res = box.space.profile_cache.index.for_liquidation:select(
            {config.params.PROFILE_STATUS.LIQUIDATING, last_id_checked}, 
            {iterator = 'GT', limit = limit}
        )    
    end

    return {res = res, error = nil}
end

function getters.is_inv3_valid(inv3_buffer)
    checks("number")

    local is_valid = integrity.is_valid()
    if not is_valid then
        log.error("is_inv3_valid: not valid")

        -- we just return nothing
        return {res = nil, error = "INV3_DATA_NOT_INITIALIZED"}  
    end

    local res = periodics.update_profiles_meta(false)
    if res ~= nil then
        return {res = nil, error = res}    
    end

    res = periodics.update_exchange_data()
    if res ~= nil then
        return {res = nil, error = res}
    end

    local inv3 = box.space.inv3_data:get(0)
    if inv3 == nil then
        return {res = nil, error = "INV3_DATA_EMPTY"}    
    end

    return {res = inv3.valid, error = nil}
end

function getters.cached_is_inv3_valid(inv3_buffer)
    checks("number")

    local is_valid = integrity.is_valid()
    if not is_valid then
        log.error("cached_is_inv3_valid: not valid")

        -- we just return nothing
        return {res = nil, error = "INV3_DATA_NOT_INITIALIZED"}  
    end

    local inv3 = box.space.inv3_data:get(0)
    if inv3 == nil then
        getters.is_inv3_valid(inv3_buffer)

        inv3 = box.space.inv3_data:get(0)
        if inv3 == nil then
            return {res = nil, error = "INV3_DATA_EMPTY"}    
        end
    end


    return {res = inv3.valid, error = nil}
end


function getters.list_operations(profile_id, offset, limit)
    checks('number', 'number', 'number')

    local ops = {}

    local res = balance.list_operations(profile_id, offset, limit)
    if res["error"] == nil and res.res ~= nil and #res.res > 0 then
        table.extend(ops, res.res)
    end

    for _, market in pairs(config.markets) do
        local market_id = market.id

        res = rpc.callro_engine(market_id, "list_balance_ops", {profile_id, offset, limit})

        if res["error"] == nil and res.res ~= nil and #res.res > 0 then
            table.extend(ops, res.res)
        end
    end

    return {res = ops, error = nil}
end

function getters.get_tiers()
    local res = tier.get_tiers()
    return {res = res, error = nil}
end

function getters.get_profiles_to_tiers()
    local res = tier.get_profiles_to_tiers()
    return {res = res, error = nil}
end

function getters.get_profiles_special_tiers()
    local res = tier.get_profiles_special_tiers()
    return {res = res, error = nil}
end

function getters.get_affiliate_profiles_tiers()
    local res = tier.get_affiliate_profiles_tiers()
    return {res = res, error = nil}
end

return getters
