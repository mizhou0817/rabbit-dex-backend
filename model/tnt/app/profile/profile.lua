local checks = require('checks')
local log = require('log')
local string = require('string')

local archiver = require('app.archiver')
local balance = require('app.balance')
local config = require('app.config')
local d = require('app.data')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local cache = require('app.profile.cache')
local wdm = require('app.wdm')

local ProfileError = errors.new_class("PROFILE")

local profile = {}

function profile.create(profile_type, status, wallet, exchange_id)
    checks("string", "string", "string", "string")

    if string.startswith(wallet, "0x") == false and string.startswith(wallet, "0X") == false then
        return { res = nil, error = "BROKEN_WALLET" }
    end

    local l_wallet = string.lower(wallet)
    local l_exchange_id = string.lower(exchange_id)
    local res
    status, res = pcall(function() return archiver.insert(box.space.profile, {
        box.NULL,
        profile_type,
        status,
        l_wallet,
        time.now(),
        l_exchange_id,
    }) end)
    if status == false then
        return { res = nil, error = res }
    end

    if res == nil then
        log.error({
            message = tostring(res),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = "CANT_INSERT_PROFILE" }
    end
    
    local success, err = pcall(balance.resolve_all_unknown, res.id, l_wallet, l_exchange_id)
    if not success then
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = tostring(err) }
    end

    return { res = res, error = nil }
end

function profile.create_insurance(exchange_id)
    checks('string')


    --only one insurace can exist, no matter which exchange_id
    local found = nil
    for _, p in box.space.profile.index.profile_type:pairs(config.params.PROFILE_TYPE.INSURANCE, {iterator = box.index.EQ}) do
        found = p
        break
    end

    if found ~= nil then
        return {res = found, error = nil}
    end

    local exist = box.space.profile:get(0)
    if exist == nil then
        return profile.create(config.params.PROFILE_TYPE.INSURANCE, "active", "0xinsurance", exchange_id)
    end

    return { res = exist, error = nil }
end

function profile.get(profile_id)
    checks("number")

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return { res = nil, error = "PROFILE_NOT_FOUND" }
    end

    return { res = exist, error = nil }
end

function profile.get_by_wallet_for_exchange_id(wallet, exchange_id)
    checks("string", 'string')

    local l_wallet = string.lower(wallet)
    local l_exchange_id = string.lower(exchange_id)

    local exist = box.space.profile.index.exchange_id_wallet:get({l_exchange_id, l_wallet})
    if exist == nil then
        return { res = nil, error = "PROFILE_NOT_FOUND" }
    end

    return { res = exist, error = nil }
end

function profile.withdraw_credit(profile_id, amount)
    checks("number", "decimal")

    if amount <= 0 then
        return { res = nil, error = "WRONG_AMOUNT" }
    end

    local res = cache.get_cache(profile_id)
    if res["error"] ~= nil then
        return { res = nil, error = res["error"] }
    end

    local profile_cache = res["res"]

    if profile_cache[d.cache_withdrawable_balance] - amount < 0 then
        return { res = nil, error = "NOT_ENOUGH_WB" }
    end

    res = balance.withdraw_credit(profile_id, amount)
    if res["error"] ~= nil then
        return { res = nil, error = res["error"] }
    end

    return { res = res["res"], error = nil }
end

function profile.liquidated_vaults(profile_ids)
    checks("table")

    for _, vault_profile_id in pairs(profile_ids) do
        local res, err = archiver.update(
            box.space.vaults, tonumber(vault_profile_id),
            {
                { '=', 'status', config.params.VAULT_STATUS.SUSPENDED },
                { '=', 'total_shares', ZERO }
            }
        )
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: : vault_prodile_id=%d",
                    ERR_VAULT_SUSPENSION_ERROR,
                    vault_profile_id
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
        end
        for _, holding in box.space.vault_holdings.index.vault:pairs(
            vault_profile_id, { iterator = box.index.EQ }
        )
        do
            local success, result = pcall(
                function()
                    return archiver.update(
                        box.space.vault_holdings,  {
                            holding.vault_profile_id,
                            holding.staker_profile_id
                        },
                        {
                            { '=', 'shares', ZERO },
                            { '=', 'entry_nav', ZERO },
                            { '=', 'entry_price', ONE },
                        })
                end
            )

            if not success then
                log.error({
                    message = string.format(
                        "%s: : vault_profile_id=%d, staker_profile_id=%d",
                        ERR_HOLDING_DELETION_FAILURE,
                        holding.vault_profile_id,
                        holding.staker_profile_id
                    ),
                    [ALERT_TAG] = ALERT_CRIT,
                })
            end
        end
    end
end

function profile.update_status(profile_id, new_status)
    checks("number", "string")

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return { res = nil, error = "PROFILE_NOT_FOUND" }
    end

    local cur_cache = box.space.profile_cache:get(profile_id)
    if cur_cache == nil then
        local e = cache.update(profile_id)
        if e ~= nil then
            return { res = nil, error = e }
        end
    end

    box.begin()
    local _, err = archiver.update(box.space.profile, profile_id, {
        { "=", "status", new_status }
    })
    if err ~= nil then
        error(err)
    end

    _, err = archiver.update(box.space.profile_cache, profile_id, {
        { "=", "status", new_status }
    })
    if err ~= nil then
        error(err)
    end
    box.commit()

    return { res = nil, error = nil }
end

function profile.update_last_checked(profile_id)
    checks("number")

    local exist = box.space.profile:get(profile_id)
    if exist == nil then
        return { res = nil, error = "PROFILE_NOT_FOUND" }
    end

    local cur_cache = box.space.profile_cache:get(profile_id)
    if cur_cache == nil then
        local e = cache.update(profile_id)
        if e ~= nil then
            return { res = nil, error = e }
        end
    end

    local tm = time.now()
    local res, err = archiver.update(box.space.profile_cache, profile_id, {
        { "=", "last_liq_check", tm }
    })
    if err ~= nil then
        log.error(ProfileError:new(err))
        return { res = nil, error = err }
    end

    if res == nil then
        return { res = nil, error = "CACHE_NOT_EXIST" }
    end

    return { res = res.last_liq_check, error = nil }
end

function profile.merge_unknown_ops()
    local total = 0

    for _, ops in box.space.balance_operations.index.type_status_exchange_id_chain_id:pairs(
        {
            config.params.BALANCE_TYPE.DEPOSIT,
            config.params.BALANCE_STATUS.UNKNOWN,
        },
        {iterator = "EQ"}
   )
   do
        if ops.status == config.params.BALANCE_STATUS.UNKNOWN then  
            local exist = profile.get_by_wallet_for_exchange_id(ops.wallet, ops.exchange_id)
            local p = exist["res"]
            if p == nil then
                log.info("MERGE skip, profile not found for ops.wallet=%s ops.exchange_id=%s", ops.wallet, ops.exchange_id)
            else
                local res = balance.resolve_unknown(p.id, ops.id)
                if res["error"] ~= nil then
                    log.error("MERGE: %s", tostring(res["error"]))
                else
                    total = total + 1
                end
            end
        else
            log.error("MERGE integrity error status = %s", ops.status)
        end
    end

    return { res = total, error = nil }
end


function profile.is_valid_signer(vault, wallet, required_role)
    checks("string", "string", "number")

    local l_vault = string.lower(vault)
    local l_wallet = string.lower(wallet)


    local exist = box.space.vault_permissions:get({l_vault, l_wallet, required_role})
    if exist == nil then
        return {res = false, error = nil}
    end

    return {res = true, error = nil}
end

function profile.add_permission(vault, wallet, required_role)
    checks("string", "string", "number")

  
    local l_vault = string.lower(vault)
    local l_wallet = string.lower(wallet)

    local res = box.space.vault_permissions:replace{l_vault, l_wallet, required_role}
    return {res = res, error = nil}
end

function profile.wds_per_24h()
    local res = wdm.wds_per_24h()
    
    return {res = res, error = nil}
end


return profile
