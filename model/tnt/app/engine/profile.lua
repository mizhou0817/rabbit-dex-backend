local checks = require('checks')
local decimal = require("decimal")
local log = require('log')

local config = require('app.config')
local ddl = require('app.ddl')
local en_meta = require('app.engine.enricher_meta')
local risk = require('app.engine.risk')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rolling = require('app.rolling')
local rpc = require('app.rpc')
local tuple = require('app.tuple')
local balance = require('app.balance')
local periodics = require('app.engine.periodics')
local tier = require('app.engine.tier')
local util = require('app.util')
local cm_get = require('app.engine.cache_and_meta')

require("app.config.constants")

local ProfileError = errors.new_class("PROFILE")

local SELECT_META_LIMIT = 1000000
local YIELD_LIMIT = 1000

local PM = {
    format = {
        {name = 'profile_id', type = 'unsigned'},
        {name = 'market_id', type = 'string'},
        {name = 'status', type = 'string'},

        {name = 'cum_unrealized_pnl', type = 'decimal'},
        {name = 'total_notional', type = 'decimal'},
        {name = 'total_position_margin', type = 'decimal'},
        {name = 'total_order_margin', type = 'decimal'},
        {name = 'initial_margin', type = 'decimal'},
        {name = 'market_leverage', type = 'decimal'},
        {name = 'balance', type = 'decimal'},
        {name = 'cum_trading_volume', type = 'decimal'},
        {name = 'timestamp', type = 'number'},
    },
    strict_type = 'engine_profile_meta',
}

function PM.bind(meta)
    return tuple.new(meta, PM.format, PM.strict_type)
end

function PM.init_spaces()
    rolling.init_spaces()
    tier.init_spaces()

    local profile_meta = ddl.create_space('profile_meta', {if_not_exists = true}, PM.format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'profile_id'}},
            if_not_exists = true,
        },
        {
            name = 'timestamp',
            unique = false,
            parts = {{field = 'timestamp'}},
            if_not_exists = true,
        }
    })

    local insurance_data = box.schema.space.create('insurance_data', {if_not_exists = true})
    insurance_data:format({
        {name = 'title', type = 'string'},
        {name = 'insurance_id', type = 'number'},
    })
    insurance_data:create_index('primary', {
        unique = true,
        parts = {{field = 'title'}},
        if_not_exists = true })
end

-- TODO: remove <cum_trading_volume> and all code around it
-- because no need of it (but still required in Cheng tests)
function PM.update_volume(market_id, taker_id, maker_id, notion)
    local status, res

    PM.ensure_meta(taker_id, market_id)
    PM.ensure_meta(maker_id, market_id)

    status, res = pcall(function() return
        box.space.profile_meta:update(taker_id, {
            {'+', "cum_trading_volume", notion},
            {'=', 'timestamp', time.now()},
        })
    end)
    if status == false then
        log.error(ProfileError:new("can't update_volume for profile_id=%d error=%s", taker_id, res))
    end


    status, res = pcall(function() return
        box.space.profile_meta:update(maker_id, {
            {'+', "cum_trading_volume", notion},
            {'=', 'timestamp', time.now()},
        })
    end)
    if status == false then
        log.error(ProfileError:new("can't update_volume for profile_id=%d error=%s", maker_id, res))
    end

end

function PM.ensure_meta(profile_id, market_id)
    local exist = box.space.profile_meta:get(profile_id)
    if exist ~= nil then
        return nil
    end

    local balance = ZERO
    local b_sum = box.space.balance_sum:get(profile_id)
    if b_sum ~= nil then
        balance = b_sum.balance
    end

    local status, res = pcall(function() return box.space.profile_meta:insert{
        profile_id,
        market_id,
        config.params.PROFILE_STATUS.ACTIVE,
        ZERO, ZERO, ZERO, ZERO,
        ONE,
        ONE,
        balance,
        ZERO,
        time.now(),
    } end)

    if status == false then
        return res
    end

    return nil
end

function PM.get_meta_for_profile(profile_id)
    checks("number")

    local meta = box.space.profile_meta:get(profile_id)
    if meta == nil then
        local text = "META_NOT_FOUND for profile_id=" .. tostring(profile_id)
        log.error(ProfileError:new(text))
        return {res=nil, error=text}
    end

    return {res = PM.bind(meta), error = nil}
end

function PM.get_profile_meta(profiles, fullscan)
    if  #profiles == 0 and fullscan == true then
        local res = box.space.profile_meta:select(nil, {iterator = 'ALL', limit=SELECT_META_LIMIT, fullscan = true})
        return {res = res, error = nil}
    end

    if #profiles == 0 then
        -- SEND profiles which has positions only
        local metas = {}
        for _, pos in box.space.position:pairs() do
            local meta = box.space.profile_meta:get(pos.profile_id)
            if meta ~= nil then
                table.insert(metas, meta:totable())
            end
        end

        return {res = metas, error = nil}
    end


    local res = {}
    for _, profile_id in pairs(profiles) do
        local meta = box.space.profile_meta:get(profile_id)
        if meta ~= nil then
            table.insert(res, meta:totable())
        end
    end

    return {res = res, error = nil}
end

function PM.get_profiles_meta_after_ts(ts)
    checks('number')

    local it = (ts == 0 and box.index.GE or box.index.GT)
    local now = time.now()
    local metas = {}
    local count = 0

    for _, mt in box.space.profile_meta.index.timestamp:pairs(ts, {iterator = it}) do
        if mt.timestamp > now then
            break
        end
        table.insert(metas, mt:totable())
        count = util.safe_yield(count, YIELD_LIMIT)
    end

    return {res = metas, error = nil}
end

function PM.update_balance(profile_id, market_id)
    local balance = ZERO
    local b_sum = box.space.balance_sum:get(profile_id)
    if b_sum == nil then
        log.warn("balance_sum not found for profile_id=%d", profile_id)
    else
        balance = b_sum.balance
    end

    local new_item = {
        profile_id,
        market_id,
        config.params.PROFILE_STATUS.ACTIVE,

        ZERO, ZERO, ZERO, ZERO,
        ONE,
        ONE,
        balance,
        ZERO
    }

    local status, res = pcall(function() return
        box.space.profile_meta:upsert(new_item, {
            {'=', "balance", balance},
            {'=', 'timestamp', time.now()},
        })
    end)
    if status == false then
        log.error(ProfileError:new("can't update_balance error=%s", res))
        return res
    end

    return nil
end

function PM.update_roll_value(title, profile_id, new_value, period_sec, max_values, is_replace)
    return rolling.update_roll_value(title, tostring(profile_id), new_value, period_sec, max_values, is_replace)
end


function PM.get_roll_avg(title, profile_id)
    return rolling.get_roll_avg(title, tostring(profile_id))
end

function PM.update_leverage(profile_id, market_id, market_leverage)
    checks("number", "string", "decimal")

    local err = PM.ensure_meta(profile_id, market_id)
    if err ~= nil then
        return {res = nil, error = err}
    end

    local profile_data
    profile_data, err = cm_get.handle_get_cache_and_meta(nil, {profile_id}, market_id, profile_id)
    if err ~= nil then
        local text = "NO_CACHE to update_leverage profile_id=" .. tostring(profile_id)
        log.error(ProfileError:new(err))
        return {res=nil, error=text}
    end

    local market = box.space.market:get(market_id)
    if market == nil then
        return {res = nil, error = "NO_MARKET"}
    end

    box.begin()

    --xxx ???
    local initial_margin = ZERO
    if market_leverage > 0 then
        initial_margin = ONE / market_leverage
    end

    local status, res = pcall(function() return
        box.space.profile_meta:update(profile_id, {
            {'=', "initial_margin", initial_margin},
            {'=', "market_leverage", market_leverage},
            {'=', 'timestamp', time.now()},
        })
    end)
    if status == false then
        log.error(ProfileError:new("can't update_leverage error=%s", res))
        box.rollback()
        return {res = nil, error = err}
    end

    err = risk.post_match(market_id, profile_data, profile_id, nil)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = err}
    end

    local profile_meta = box.space.profile_meta:get(profile_id)
    en_meta.update_profile_meta(market, profile_id, profile_meta)

    box.commit()

    return {res = market_leverage, error = nil}
end

function PM.which_tier(profile_id)
if true then box.space.tier:get(0) end
    checks('number')

    local p2t = box.space.profile_to_tier:get(profile_id)
    if p2t then
        if p2t.special_tier_id and p2t.special_tier_id > 0 then
            local tier = box.space.special_tier:get(p2t.special_tier_id)
            if tier then
                return tier
            end
        end
        if p2t.tier_id then
            local tier = box.space.tier:get(p2t.tier_id)
            if tier then
                return tier
            end
        end
    end

    return assert(box.space.tier:get(0), 'No default tier')
end

-- Transform data to one format
function PM.getter_which_tier(profile_id)
    checks('number')

    local which = PM.which_tier(profile_id)

    local common_tier = {
        tier = which.tier,
        title = which.title,
        maker_fee = which.maker_fee,
        taker_fee = which.taker_fee
    }

    return {res = common_tier, error = nil}
end

function PM.set_insurance_id(market_id)
    checks("string")

    local exist = box.space.insurance_data:get(config.params.PROFILE_TYPE.INSURANCE)
    if exist == nil then
        return PM.update_insurance_id(0, market_id)
    end

    return {res = exist, error = nil}
end

function PM.update_insurance_id(insurance_id, market_id)
    checks("number", "string")

    local res = box.space.insurance_data:replace{
        config.params.PROFILE_TYPE.INSURANCE,
        insurance_id
    }

    local err = PM.ensure_meta(insurance_id, market_id)
    if err ~= nil then
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

function PM.get_insurance_id()
    local res = box.space.insurance_data:get(config.params.PROFILE_TYPE.INSURANCE)
    if res == nil then
        return nil
    end

    return res.insurance_id
end


function PM.add_tier(tier_id, title, maker_fee, taker_fee, min_volume, min_assets)
    checks("number", "string", "decimal", "decimal", "decimal", "decimal")

    local status, res = pcall(function() return box.space.tier:insert{
        tier_id,
        title,
        maker_fee,
        taker_fee,
        min_volume,
        min_assets
    } end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end

function PM.edit_tier(tier_id, title, maker_fee, taker_fee, min_volume, min_assets)
    checks("number", "string", "decimal", "decimal", "decimal", "decimal")

    local status, res = pcall(function() return box.space.tier:update(tier_id, {
        {'=', 'title', title},
        {'=', 'maker_fee', maker_fee},
        {'=', 'taker_fee', taker_fee},
        {'=', 'min_volume', min_volume},
        {'=', 'min_assets', min_assets},
    }) end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end


function PM.remove_tier(tier_id)
    checks("number")

    if tier_id == 0 then
        return {res = nil, error = "ZERO_TIER_CANT_BE_REMOVED"}
    end

    local status, res = pcall(
        function()
           return box.space.tier:delete(tier_id)
        end
     )
     if status == false then
        return {res = nil, error = tostring(res)}
     end

    return {res = nil, error = nil}
end


function PM.add_special_tier(tier_id, title, maker_fee, taker_fee)
    checks("number", "string", "decimal", "decimal")

    local status, res = pcall(function() return box.space.special_tier:insert{
        tier_id,
        title,
        maker_fee,
        taker_fee
    } end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end

function PM.edit_special_tier(tier_id, title, maker_fee, taker_fee)
    checks("number", "string", "decimal", "decimal")


    local status, res = pcall(function() return box.space.special_tier:update(tier_id, {
        {'=', 'title', title},
        {'=', 'maker_fee', maker_fee},
        {'=', 'taker_fee', taker_fee},
    }) end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end


function PM.remove_special_tier(tier_id)
    checks("number")

    local status, res = pcall(
        function()
           return box.space.special_tier:delete(tier_id)
        end
     )
     if status == false then
        return {res = nil, error = tostring(res)}
     end

    return {res = nil, error = nil}
end


function PM.add_profile_to_special_tier(profile_id, tier_id)
    checks("number", "number")

    local exist = box.space.special_tier:get(tier_id)
    if exist == nil then
        return {res = nil, error = "NO_SPECIAL_TIER_FOUND"}
    end

    local status, res = pcall(function() return box.space.profile_to_tier:replace{
        profile_id,
        tier_id
    } end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end


function PM.remove_profile_from_tier(profile_id)
    checks("number")

    local status, res = pcall(function() return box.space.profile_to_tier:delete(profile_id) end)

    if status == false then
        return {res = nil, error = tostring(res)}
    end

    return {res = res, error = nil}
end

function PM.get_tiers()
    local res = box.space.tier:select()
    return {res = res, error = nil}
end

function PM.get_special_tiers()
    local res = box.space.special_tier:select()
    return {res = res, error = nil}
end

function PM.get_profile_tiers()
    local res = box.space.profile_to_tier:select()
    return {res = res, error = nil}
end

function PM.process_referral_payout(market_id)
    checks("string")

    box.begin()

    local res = balance.process_referral_payout()
    if res['error'] ~= nil then
        box.rollback()
        return {res = nil, error = res['error']}
    end

    local count = 0
    local profiles = res['res']
    for _, profile_id in pairs(profiles) do
        PM.ensure_meta(profile_id, market_id)
        count = util.safe_yield(count, 1000)
    end

    local err = periodics.update_profiles(profiles)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = err}
    end

    box.commit()

    return {res = true, error = nil}
end

return PM
