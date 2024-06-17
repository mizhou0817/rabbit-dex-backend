local fiber = require('fiber')
local json = require("json")
local log = require('log')

local archiver = require('app.archiver')
local balance = require('app.balance')
local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local cache = require('app.profile.cache')
local integrity = require('app.profile.integrity')
local notif = require('app.profile.notif')
local rpc = require('app.rpc')
local util = require('app.util')
local dynamic = require('app.profile.dynamic')


require("app.config.constants")

local PeriodicsError = errors.new_class("PERIODICS")

local periodics = {
    _interval = 7,
    _fibers = {},
    _loop = nil,
    _connection_interval = 1250,
    _meta_interval = 3600,
    _exchange_interval = 20,
    _inv3_interval = 80,
    conns = {},
}

function _watcher_callback(key, value)
    if key ~= config.sys.EVENTS.PROFILE_UPDATE then
        local text = "unknow event=" .. tostring(key)
        log.warn(text)
        return text
    elseif value == nil then
        local text = "nil value=" .. tostring(value)
        log.warn(text)
        return text
    end

    local fullscan = value.fullscan or false

    local market_id = value.market_id
    if market_id == nil then
        local text = "broken market_id = " .. tostring(market_id)
        log.error(PeriodicsError:new(text))
        return text
    end
    local res = rpc.callro_engine(market_id, "get_profile_meta", { value.profiles, fullscan })
    if res["error"] ~= nil then
        local text = "_watcher_callback market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
        log.error(text)
        return text
    end

    local metas = res["res"]

    local count = 0
    box.begin()
    if metas ~= nil then
        for _, profile_meta in pairs(metas) do
            box.space.profile_meta:replace(profile_meta)

            cache.update(profile_meta[1])

            count = util.safe_yield(count, 1000)
        end
    end
    box.commit()

    notif.notify_profiles(value.profiles)

    return nil
end

function periodics._do_profile_periodics()
    local count = 0

    for _, profile_cache in box.space.profile_cache:pairs(nil, { iterator = "ALL" }) do
        local profile_id = profile_cache.id

        local profile = box.space.profile:get(profile_id)
        if profile == nil then
            log.warn("no profile for profile_id = %d", profile_cache.id)
        else
            local channel = "account@" .. tostring(profile_id)
            local json_update = json.encode({ data = profile_cache:tomap({ names_only = true }) })

            rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
            json_update = nil

            count = count + 1
        end
    end
end

function periodics._reconnect_to_market(market_id)
    local res = rpc.watcher_engine_subscribe(market_id, config.sys.EVENTS.PROFILE_UPDATE, _watcher_callback)
    local data = res["res"]
    if res["error"] ~= nil then
        local text = "watcher market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
        log.error(text)
        return text
    else
        local conn = res["res"]["conn"]
        if conn == nil or not conn:is_connected() then
            local text = "conn error for market_id=" .. tostring(market_id)
            log.error(text)
            return text
        else
            if periodics.conns[market_id] == nil then
                periodics.conns[market_id] = {}
            end
            periodics.conns[market_id].conn = data["conn"]
            periodics.conns[market_id].handler = data["handler"]

            log.info("***** connected to market_id=%s", market_id)
        end
    end

    return nil
end

function periodics._update_connections()
    local attempts = 0

    while true do
        local total, connected = 0, 0
        for _, market in pairs(config.markets) do
            total = total + 1

            local market_id = market.id
            local market_conn = periodics.conns[market_id]

            local conn = market_conn and market_conn.conn
            if conn == nil or not conn:is_connected() then
                local err = periodics._reconnect_to_market(market_id)
                if err ~= nil then
                    log.error("_update_connections market_id=%s attempt=%d error=%s", market_id, attempts, err)
                    break
                end
            end
            connected = connected + 1
        end
        if connected >= total then
            break
        end
        fiber.sleep(config.params.MARKET_UPDATE_RETRY_INTERVAL)
        attempts = attempts + 1
    end

    return nil
end

function periodics.update_profiles_meta(fullscan)
    periodics._update_connections()

    local start_time = time.now()

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local err = _watcher_callback(config.sys.EVENTS.PROFILE_UPDATE, { market_id = market_id, profiles = {}, fullscan=fullscan })
        if err ~= nil then
            return err
        end
    end

    local end_time = time.now()
    local diff_seconds = (end_time - start_time) / 1e6
    
    if diff_seconds > 30 then
        log.info({
            message = string.format('update_profiles_meta fullscan=%s: start_time=%d end_time=%d diff_seconds=%d', tostring(fullscan), start_time, end_time, diff_seconds),
            [ALERT_TAG] = ALERT_CRIT,
        })
    end

    return nil
end

function periodics.test_update_profiles_meta()
    local res = periodics.update_profiles_meta(false)

    return { res = nil, error = res }
end

function periodics._update_market_status(new_status)
    local markets = {}
    for _, market in pairs(config.markets) do
        table.insert(markets, market)
    end

    local markets_count, markets_ok = #markets, 0
    while true do
        for i = 1, markets_count do
            if markets[i] ~= nil then
                local market_id = markets[i].id

                local res = rpc.callrw_engine(market_id, "change_status", { market_id, new_status })
                if res["error"] == nil then
                    markets[i] = nil
                    markets_ok = markets_ok + 1
                else
                    local text = "change_status market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
                    log.error(text)
                end
            end
        end
        if markets_ok == markets_count then
            break
        end
        fiber.sleep(config.params.MARKET_UPDATE_RETRY_INTERVAL)
    end
end

function periodics.update_exchange_data()
    periodics._update_connections()

    local exchangeBalance = ZERO

    local _update_exchange_totals = function(res_list)
        for _, wallet in pairs(res_list) do
            local new_value = nil
            local update = nil
            if wallet[1] == EXCHANGE_WALLET_ID then
                new_value = {
                    EXCHANGE_ID,
                    ZERO,
                    wallet[2]
                }
                update = {
                    { '+', 'total_balance', wallet[2] }
                }

                exchangeBalance = exchangeBalance + wallet[2]
            elseif wallet[1] == FEE_WALLET_ID then
                new_value = {
                    EXCHANGE_ID,
                    wallet[2],
                    ZERO
                }
                update = {
                    { '+', 'trading_fee', wallet[2] }
                }
            else
                log.warn("unknown exchange wallet_id = %d", wallet[1])
            end

            if new_value ~= nil then
                local status, res = pcall(function() return box.space.exchange_total:upsert(new_value, update) end)
                if status == false then
                    log.error(PeriodicsError:new(res))
                end
            end
        end
    end

    box.space.exchange_total:truncate()

    -- UPDATE data from current instance balance
    local res = balance.get_exchange_wallets_data()
    if res["error"] ~= nil then
        local text = "poll_exchange_data market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
        log.error(text)
        return text
    end
    _update_exchange_totals(res["res"])

    for _, market in pairs(config.markets) do
        local market_id = market.id

        local res = rpc.callro_engine(market_id, "get_exchange_wallets_data")
        if res["error"] ~= nil then
            local text = "poll_exchange_data market_id=" .. tostring(market_id) .. " error=" .. tostring(res["error"])
            log.error(text)
            return text
        end

        _update_exchange_totals(res["res"])
    end


    -- RECALC INV3
    local tradersSumAe = ZERO
    local insuranceBalance = ZERO

    for _, p_cache in box.space.profile_cache:pairs() do


        if p_cache.profile_type ~= config.params.PROFILE_TYPE.INSURANCE then
            if p_cache.account_equity > 0 and p_cache.account_margin >= config.params.LIQUIDATION.LIQUIDATION_MARGIN then
                tradersSumAe = tradersSumAe + p_cache.account_equity
            end
        else
            insuranceBalance = p_cache.balance
        end

        -- Phisical money of each trader
        exchangeBalance = exchangeBalance + p_cache.balance
    end

    local def_inv3_id = 0
    local new_valid = true
    local prev_valid

    local exist = box.space.inv3_data:get(def_inv3_id)
    if exist ~= nil then
        prev_valid = exist.valid
    end

    if tradersSumAe ~= ZERO and tradersSumAe > exchangeBalance then
        new_valid = false
        log.error({
            message = string.format('INV3_BROKEN: inv3_valid=%s tradersSumAe=%s exchangeAndInsurance=%s insuranceBalance=%s',
                new_valid, tradersSumAe, exchangeBalance, insuranceBalance),
            [ALERT_TAG] = ALERT_CRIT,
        })
    end

    local tm = time.now()
    local _, err = archiver.upsert(box.space.inv3_data, {
        def_inv3_id,
        new_valid,
        tm,
        tradersSumAe,
        exchangeBalance,
        insuranceBalance,
    }, {
        { '=', "valid",             new_valid },
        { '=', 'last_updated',      tm },
        { '=', 'margined_ae_sum',   tradersSumAe },
        { '=', 'exchange_balance',  exchangeBalance },
        { '=', 'insurance_balance', insuranceBalance },
    })
    if err ~= nil then
        log.error('update_exchange_data: update inv3_data space: error=%s', PeriodicsError:new(err))
        return err
    end

    if new_valid ~= prev_valid then
        local status = new_valid == true
            and config.params.MARKET_STATUS.ACTIVE
            or config.params.MARKET_STATUS.PAUSED

        periodics._update_market_status(status)
    end

    return nil
end

function periodics._wait_for_markets()
    local connected, total = 0, 0
    for _, market in pairs(config.markets) do
        local market_id = market.id

        local role_name = config.sys.ID_TO_ROLES[market_id]
        if role_name == nil then
            local errmsg = string.format('NOT_FOUND: role_name for market_id = %s', tostring(market_id))
            log.error('PROFILE on start: %s', errmsg)
            return errmsg
        else
            total = total + 1
            fiber.create(function()
                local attempt = 0
                while true do
                    attempt = attempt + 1
                    local res = rpc.wait_for_role(role_name)
                    if res.error == nil then
                        connected = connected + 1
                        break
                    end
                    log.warn("waiting for role_name = %s  attempt = %d, error = %s", role_name, attempt,
                        tostring(res.error))
                    fiber.sleep(config.params.WAIT_ROLE_RETRY_INTERVAL)
                end
            end)
        end
    end
    while connected < total do
        fiber.sleep(config.params.WAIT_ROLE_RETRY_INTERVAL)
    end
    return nil
end

function periodics.validate_markets()
    for _, market in pairs(config.markets) do
        local market_id = market.id

        local role_name = config.sys.ID_TO_ROLES[market_id]
        if not role_name then
            return string.format('role not found for market-id=%s', market_id)
        end
        if not rpc.is_role_present(role_name) then
            return string.format('role %s not present for market-id=%s', role_name, market_id)
        end
    end

    return nil
end

function periodics.start()
    log.info("PERIODICS_LOG: called profile.periodics.start ")
    local err

    integrity.make_invalid()

    -- wait for all markets go live
    err = periodics._wait_for_markets()
    if err ~= nil then
        log.error(PeriodicsError:new(err))
        return
    end

    -- Update connections
    err = periodics._update_connections()
    if err ~= nil then
        log.error(PeriodicsError:new(err))
        return
    end

    cache.update_all()

    -- Update all caches on start
    periodics.update_profiles_meta(true)

    integrity.make_valid()

    -- Update inv3
    local exchange_fiber = fiber.create(function()
        while true do
            fiber.testcancel()

            periodics.update_exchange_data()

            fiber.sleep(periodics._exchange_interval)
        end
    end)
    table.insert(periodics._fibers, exchange_fiber)

    log.info("PERIODICS_LOG: profile.periodics.start started success")
end

function periodics.killall()
    if #periodics._fibers <= 0 then
        log.warn("No active fibers found")
        return
    end

    for _, f in pairs(periodics._fibers) do
        local status, res = fiber.kill(f.id())
        if status == false then
            log.warn("error: kill fiber id=%d error=%s", f.id(), res)
        else
            log.info("success: kill fiber id=%d", f.id())
        end
    end
end

return periodics
