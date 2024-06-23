local checks = require('checks')
local json = require('json')
local log = require('log')

local archiver = require('app.archiver')
local balance = require('app.balance')
local config = require('app.config')
local errors = require('app.lib.errors')
local tick = require("app.lib.tick")
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')
local decimal = require('decimal')
local uuid = require('uuid')

require("app.config.constants")

local ProfileError = errors.new_class("PROFILE")

local cache = {}

function cache.update(profile_id)
    checks("number")

    local profile = box.space.profile:get(profile_id)
    if profile == nil then
        local text = "INTERGRITY ERROR - no profile"
        log.error(ProfileError:new(text))
        return text
    end

    local last_liq_check = 0
    local cur_cache = box.space.profile_cache:get(profile_id)
    if cur_cache ~= nil then
        last_liq_check = cur_cache.last_liq_check
    end


    local profile_totals = {
        balance = balance.get_balance(profile_id),

        account_equity = ZERO,
        total_position_margin = ZERO,
        total_order_margin = ZERO,
        total_notional = ZERO,
        account_margin = ONE,
        withdrawable_balance = ZERO,
        cum_unrealized_pnl = ZERO,
        
        health = ONE,
        account_leverage = ONE,
        cum_trading_volume = ZERO,
        leverage = {},
        last_liq_check
    }
    profile_totals.leverage["BTC-USD"] = ONE
    profile_totals.leverage["ETH-USD"] = ONE
    profile_totals.leverage["SOL-USD"] = ONE
    profile_totals.leverage["ARB-USD"] = ONE
    profile_totals.leverage["DOGE-USD"] = ONE
    profile_totals.leverage["LDO-USD"] = ONE
    profile_totals.leverage["SUI-USD"] = ONE
    profile_totals.leverage["PEPE1000-USD"] = ONE
    profile_totals.leverage["BCH-USD"] = ONE
    profile_totals.leverage["XRP-USD"] = ONE
    profile_totals.leverage["WLD-USD"] = ONE
    profile_totals.leverage["TON-USD"] = ONE
    profile_totals.leverage["STX-USD"] = ONE
    profile_totals.leverage["MATIC-USD"] = ONE
    profile_totals.leverage["TRB-USD"] = ONE
    profile_totals.leverage["APT-USD"] = ONE
    profile_totals.leverage["INJ-USD"] = ONE
    profile_totals.leverage["AAVE-USD"] = ONE
    profile_totals.leverage["LINK-USD"] = ONE
    profile_totals.leverage["BNB-USD"] = ONE
    profile_totals.leverage["RNDR-USD"] = ONE
    profile_totals.leverage["MKR-USD"] = ONE
    profile_totals.leverage["RLB-USD"] = ONE
    profile_totals.leverage["ORDI-USD"] = ONE
    profile_totals.leverage["STG-USD"] = ONE
    profile_totals.leverage["SATS1000000-USD"] = ONE
    profile_totals.leverage["TIA-USD"] = ONE
    profile_totals.leverage["BLUR-USD"] = ONE
    profile_totals.leverage["JTO-USD"] = ONE
    profile_totals.leverage["MEME-USD"] = ONE
    profile_totals.leverage["SEI-USD"] = ONE
    profile_totals.leverage["YES-USD"] = ONE
    profile_totals.leverage["WIF-USD"] = ONE
    profile_totals.leverage["STRK-USD"] = ONE
    profile_totals.leverage["SHIB1000-USD"] = ONE
    profile_totals.leverage["BOME-USD"] = ONE
    profile_totals.leverage["SLERF-USD"] = ONE
    profile_totals.leverage["W-USD"] = ONE
    profile_totals.leverage["ENA-USD"] = ONE
    profile_totals.leverage["PAC-USD"] = ONE
    profile_totals.leverage["MAGA-USD"] = ONE
    profile_totals.leverage["TRUMP-USD"] = ONE
    profile_totals.leverage["MOG1000-USD"] = ONE
    profile_totals.leverage["NOT-USD"] = ONE
    profile_totals.leverage["MOTHER-USD"] = ONE
    profile_totals.leverage["BONK1000-USD"] = ONE
    profile_totals.leverage["TAIKO-USD"] = ONE
    profile_totals.leverage["FLOKI1000-USD"] = ONE

    for _, meta in box.space.profile_meta:pairs(profile_id, {iterator="EQ"}) do
        profile_totals.balance = profile_totals.balance + meta.balance
        profile_totals.cum_unrealized_pnl = profile_totals.cum_unrealized_pnl + meta.cum_unrealized_pnl
        profile_totals.total_notional = profile_totals.total_notional + meta.total_notional
        profile_totals.total_position_margin = profile_totals.total_position_margin + meta.total_position_margin
        profile_totals.total_order_margin = profile_totals.total_order_margin + meta.total_order_margin
        profile_totals.leverage[meta.market_id] = meta.market_leverage
        profile_totals.cum_trading_volume = profile_totals.cum_trading_volume + meta.cum_trading_volume
    end

    -- CALC aggregated value
    profile_totals.account_equity = profile_totals.balance + profile_totals.cum_unrealized_pnl
    if profile_totals.total_notional ~= 0 then 
        profile_totals.account_margin = profile_totals.account_equity / profile_totals.total_notional
    elseif profile_totals.account_equity == 0 then
        profile_totals.account_margin = ONE
    end

    profile_totals.account_leverage = ONE
    profile_totals.health = ZERO
    if profile_totals.account_margin ~= 0 then
        profile_totals.account_leverage = 1 / profile_totals.account_margin
        if profile_totals.account_margin > 0 then
            profile_totals.health = tick.min(ONE, profile_totals.account_margin)
        end
    end

    profile_totals.withdrawable_balance = tick.min(profile_totals.account_equity, profile_totals.balance) 
                                                    - profile_totals.total_position_margin 
                                                    - profile_totals.total_order_margin

    local new_status = profile.status

    -- Don't liquidate accounts for rebalancing
    if profile_totals.account_margin < config.params.LIQUIDATION.FORCED_MARGIN and profile.profile_type ~= config.params.PROFILE_TYPE.INSURANCE then
        new_status = config.params.PROFILE_STATUS.LIQUIDATING

        box.space.profile:update(profile.id, {
            {"=", "status", new_status}
        })
        if profile.profile_type == config.params.PROFILE_TYPE.VAULT and profile_totals.account_margin < config.params.LIQUIDATION.LIQUIDATION_MARGIN then

            local res, err = archiver.update(
                box.space.vaults, profile.id,
                {
                    { '=', 'status', config.params.VAULT_STATUS.SUSPENDED }
                }
            )
            if err ~= nil then
                log.error({
                    message = string.format(
                        "%s: : vault-profile.id=%d",
                        ERR_VAULT_SUSPENSION_ERROR,
                        profile.id
                    ),
                    [ALERT_TAG] = ALERT_CRIT,
                })
            end
        end
    end
    
    -- REPLACE old value
    local _, err = archiver.replace(box.space.profile_cache, {
        profile.id,
        profile.profile_type,
        new_status,
        profile.wallet,

        time.now(),
        profile_totals.balance,
        profile_totals.account_equity,
        profile_totals.total_position_margin,
        profile_totals.total_order_margin,
        profile_totals.total_notional,
        profile_totals.account_margin,
        profile_totals.withdrawable_balance,
        profile_totals.cum_unrealized_pnl,
        profile_totals.health,
        profile_totals.account_leverage,
        profile_totals.cum_trading_volume,
        profile_totals.leverage,
        last_liq_check,
    })
    if err ~= nil then
        error(err)
    end

    profile_totals = nil

    return nil
end

function cache.update_all() 

    for _, profile in box.space.profile:pairs(nil, {iterator="ALL"}) do
        cache.update(profile.id)
    end
end

function cache.ensure_cache(profile_id)
    local exist = box.space.profile_cache:get(profile_id)

    if exist == nil then
        cache.update(profile_id)
    end

    return {res = nil, error = nil}
end

function cache.get_meta(profile_id, market_id)
    return box.space.profile_meta:get({profile_id, market_id})
end

function cache.get_cache_and_meta(profile_ids, market_id)
    checks("table", "string")

    local res = {}
    for _, profile_id in pairs(profile_ids) do
        local pc = box.space.profile_cache:get(profile_id)
        if pc == nil then
            local text = "NO_CACHE_FOR_PROFILE_" .. tostring(profile_id)
            log.error({
                message = string.format(
                    "%s: profile_id=%s error=%s",
                    ERR_GET_CACHE,
                    tostring(profile_id),
                    text
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return {res=nil, error=text}
        end

        local mt = cache.get_meta(profile_id, market_id)

        table.insert(res, {
            profile_id = profile_id,
            cache = pc,
            meta = mt,
        })
    end

    return {res=res, error=nil}
end

function cache.get_cache(profile_id)
    checks("number")

    local exist = box.space.profile_cache:get(profile_id)
    if exist == nil then
        local text = "NO_CACHE_FOR_PROFILE_" .. tostring(profile_id)
        return {res=nil, error=text}
    end

    return {res = exist, error = nil}
end

--TODO: rewrite it to the good version
--Need to structure periodics better
function cache.invalidate_cache(profile_id)
    checks("number")

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callro_engine(market_id, "get_profile_meta", {{profile_id}, false})
        if res["error"] ~= nil then
            local text = "invalidate_cache market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
            log.error(text)
            return {res = nil, error = text}
        end

        local metas = res["res"]
        for _, profile_meta in pairs(metas) do
            box.space.profile_meta:replace(profile_meta)
        end    
    end


    local err = cache.update(profile_id)
    if err ~= nil then
        return {res=nil, error=err}
    end

    local exist = box.space.profile_cache:get(profile_id)
    return {res = exist, error = nil}
end

function cache.invalidate_cache_and_notify(profile_id)
    checks("number")

    local res = cache.invalidate_cache(profile_id)
    if res["error"] ~= nil then
        local text = "invalidate_cache_and_notify  invalidate_cache error=" .. tostring(res["error"])
        log.error(text)
        return {res = nil, error = text}
    end

    res = cache.get_cache(profile_id)
    if res["error"] ~= nil then
        local text = "invalidate_cache_and_notify get_cache  error=" .. tostring(res["error"])
        log.error(text)
        return {res = nil, error = text}
    end

    local updated_cache = res["res"]
    if updated_cache ~= nil then
        local channel = "account@" .. tostring(profile_id)        
        local json_update = json.encode(updated_cache:tomap({names_only=true}))
        
        pubsub_publish(channel, json_update)
        json_update = nil
    end

    return {res = updated_cache, error = nil}
end

function cache.process_yield_and_invalidate(yield_id, amount, tx_hash, exchange_id, chain_id, exchange_address)
    checks('string', 'decimal', 'string', 'string', 'number', 'string')

    if amount <= ZERO then
        return { res = nil, error = "NEGATIVE_OR_ZERO_YIELD_AMOUNT" }
    end

    local tm = time.now()

    box.begin()

    local affected_profile_ids = {}

    local total_balance = decimal.new(0)
    local profile_balances = {}
    local count = 0
    for _, profile in box.space.profile.index.exchange_id:pairs(exchange_id) do
        local res = cache.get_cache(profile.id)
        local p_cache = res.res
        if p_cache ~= nil then
            if p_cache.balance > 0 then
                total_balance = total_balance + p_cache.balance
                profile_balances[profile.id] = p_cache.balance
            end
        end
        count = util.safe_yield(count, 1000)
    end

    if total_balance == ZERO then
        box.rollback()
        return { res = nil, error = "ZERO_BALANCE_FOR_YIELD" }
    end

    local already_paid = decimal.new(0)
    local recipients = 0;
    for _, paid in box.space.balance_operations.index.reason:pairs(yield_id, { iterator = 'EQ' }) do
        already_paid = already_paid + paid['amount']
        recipients = recipients + 1
        total_balance = total_balance - profile_balances[paid['profile_id']]
        profile_balances[paid['profile_id']] = ZERO
    end

    local remaining = amount - already_paid
    if total_balance == ZERO or remaining == ZERO then
        box.rollback()
        return { res = nil, error = "ALL_YIELD_PAID" }
    end 

    local frac = (remaining / total_balance) * ROUND_DOWN_FACTOR

    count = 0
    for profile_id, bsum in pairs(profile_balances) do
        local payment_amount = bsum * frac

        if payment_amount ~= ZERO then

            local new_sum = {
                profile_id,
                payment_amount,
                tm
            }
            local status, res, err
            status, res = pcall(
                function()
                    return box.space.balance_sum:upsert(
                        new_sum,
                        {
                            { '+', "balance",      payment_amount },
                            { '=', 'last_updated', tm }
                        })
                end
            )
            if status == false then
                box.rollback()
                log.error(ProfileError:new(res))
                return { res = nil, error = res }
            end

            local bop_id = uuid.str()
            local _
            _, err = archiver.insert(box.space.balance_operations, {
                bop_id,
                config.params.BALANCE_STATUS.SUCCESS,
                yield_id,
                tx_hash,
                profile_id,
                "",
                config.params.BALANCE_TYPE.YIELD_PAYOUT,
                bop_id,
                payment_amount,
                tm,
                0,
                exchange_id,
                chain_id,
                exchange_address,
            })
            if err ~= nil then
                box.rollback()
                log.error(ProfileError:new(err))
                return { res = nil, error = err }
            end
            already_paid = already_paid + payment_amount
            recipients = recipients + 1

            table.insert(affected_profile_ids, profile_id)
        end

        if already_paid > amount + TENTH_OF_A_CENT then
            box.rollback()
            return { res = nil, error = "YIELD_OVERPAID" }
        end    

        count = util.safe_yield(count, 1000)
    end

    local status, res, err
    res, err = archiver.upsert(box.space.yield,
        {
            yield_id,
            amount,
            already_paid,
            recipients,
            tx_hash,
            tm,
            exchange_id,
            chain_id,
            exchange_address,
        },
        {
            { '=', "amount",     amount },
            { '=', "paid",       already_paid },
            { '=', "recipients", recipients },
            { '=', 'timestamp',  tm }
        })
    if err ~= nil then
        box.rollback()
        log.error(ProfileError:new(res))
        return { res = nil, error = res }
    end

    box.commit()

    count = 0
    for _, id in pairs(affected_profile_ids) do
        cache.update(id)

        count = util.safe_yield(count, 1000)
    end

    return { res = nil, error = nil }
end

function cache._test_set_cache_update(_update_cache)
    cache.update = _update_cache
end

function cache._test_set_get_meta(_get_meta)
    cache.get_meta = _get_meta
end

return cache