local checks = require('checks')
local log = require('log')

local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')

require('app.lib.table')
require('app.config.constants')

local ProfileError = errors.new_class("PROFILE")

local YIELD_LIMIT = 1000

local setters = {}

function setters.add_affected_profile(profile_id)
    checks('number')

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callrw_engine(market_id, "mt_add_profile_id", {profile_id})
        if res["error"] ~= nil then
            log.error({
                message = string.format("add_affected_profile: error: profile_id=%d market_id=%s err=%s", profile_id, tostring(market_id), tostring(res["error"])),
                [ALERT_TAG] = ALERT_CRIT,
            })
        end
    end

    return {res = nil, error = nil}
end

function setters.remove_affected_profile(profile_id)
    checks('number')

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callrw_engine(market_id, "mt_remove_profile_id", {profile_id})
        if res["error"] ~= nil then
            log.error({
                message = string.format("remove_affected_profile: error: profile_id=%d market_id=%s err=%s", profile_id, tostring(market_id), tostring(res["error"])),
                [ALERT_TAG] = ALERT_CRIT,
            })
        end
    end

    return {res = nil, error = nil}
end

function setters.cancel_all_listed()
    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callrw_engine(market_id, "mt_cancel_all_listed", {})
        if res["error"] ~= nil then
            log.error({
                message = string.format("cancel_all_listed: market_id=%s err=%s", tostring(market_id), tostring(res["error"])),
                [ALERT_TAG] = ALERT_CRIT,
            })
        end
    end

    return {res = nil, error = nil}
end

function setters.update_profiles_caches_and_metas(data)
    checks('table|profiles_caches_metas')
    local count = 0

    box.begin()
    for _, p in ipairs(data) do
        local
            cache,
            metas = unpack(p)

        local
            profile_id,
            profile_type,
            status,
            wallet,
            last_update,
            balance,
            account_equity,
            total_position_margin,
            total_order_margin,
            total_notional,
            account_margin,
            withdrawable_balance,
            cum_unrealized_pnl,
            health,
            account_leverage,
            cum_trading_volume,
            leverage,
            last_liq_check = unpack(cache)

        last_update = time.now()
        last_liq_check = ZERO

        local _, err = archiver.upsert(box.space.profile_cache, {
            profile_id,
            profile_type,
            status,
            wallet,

            last_update,
            balance,
            account_equity,
            total_position_margin,
            total_order_margin,
            total_notional,
            account_margin,
            withdrawable_balance,
            cum_unrealized_pnl,
            health,
            account_leverage,
            cum_trading_volume,
            leverage,
            last_liq_check,
        }, {
            {'=', 'status', status},
            {'=', 'wallet', wallet},
            {'=', 'last_update', last_update},
            {'=', 'balance', balance},
            {'=', 'account_equity', account_equity},
            {'=', 'total_position_margin', total_position_margin},
            {'=', 'total_order_margin', total_order_margin},
            {'=', 'total_notional', total_notional},
            {'=', 'account_margin', account_margin},
            {'=', 'withdrawable_balance', withdrawable_balance},
            {'=', 'cum_unrealized_pnl', cum_unrealized_pnl},
            {'=', 'health', health},
            {'=', 'account_leverage', account_leverage},
            {'=', 'cum_trading_volume', cum_trading_volume},
            {'=', 'leverage', leverage},
        })
        if err ~= nil then
            box.rollback()
            local msg = string.format("%s: profile.id=%d: %s", ERR_PROFILE_CACHE_ERROR, profile_id, err)
            return {res = nil, error = msg}
        end

        -- liquidation service activates profile
        if status ~= config.params.PROFILE_STATUS.ACTIVE then
            local _, err = archiver.update(box.space.profile, profile_id, {
                {'=', 'status', status},
            })
            if err ~= nil then
                box.rollback()
                local msg = string.format("%s: profile.id=%d: %s", ERR_PROFILE_LIQUIDATION_ERROR, profile_id, err)
                return {res = nil, error = msg}
            end

            if profile_type == config.params.PROFILE_TYPE.VAULT and account_margin < config.params.LIQUIDATION.LIQUIDATION_MARGIN then
                local _, err = archiver.update(
                    box.space.vaults, profile_id,
                    {
                        { '=', 'status', config.params.VAULT_STATUS.SUSPENDED }
                    }
                )
                if err ~= nil then
                    box.rollback()
                    local msg = string.format("%s: vault-profile.id=%d: %s", ERR_VAULT_SUSPENSION_ERROR, profile_id, res)
                    return {res = nil, error = msg}
                end
            end
        end

       for _, mt in pairs(metas) do
           local ok, res = pcall(function()
                return box.space.profile_meta:replace(mt)
            end)
            if not ok then
                box.rollback()
                local msg = string.format("%s: profile.id=%d: %s", ERR_PROFILE_META_ERROR, profile_id, err)
                return {res = nil, error = msg}
            end
           count = count + 1
       end

        count = util.safe_yield(count, YIELD_LIMIT)
    end
    box.commit()

    return {res = nil, error = nil}
end

return setters
