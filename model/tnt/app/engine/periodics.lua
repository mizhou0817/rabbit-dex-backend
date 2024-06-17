local decimal = require('decimal')
local fiber = require('fiber')
local log = require('log')
local os = require("os")

local config = require('app.config')
local en_meta = require('app.engine.enricher_meta')
local m = require('app.engine.market')
local notif = require('app.engine.notif')
local position = require('app.engine.position')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local util = require('app.util')

require("app.config.constants")

local PeriodicsError = errors.new_class("ENGINE_PERIODICS")

local FUNDING_UPDATE_DIFF = 3600000000 -- each hour (microseconds)
local TEN_MINUTES = 600000000
local UPDATE_FUNDING_MAX_MINUTES = 4

local periodics = {
    _market_id = nil,
    _min_tick = nil,
    _min_order = nil,

    _profile_batch_size = 1000,
    _funding_low_border = decimal.new(-0.0025),
    _funding_up_border = decimal.new(0.0025),

    _basis_low_border = decimal.new(-0.01),
    _basis_up_border = decimal.new(0.01),

    _funding_point_cap_high = decimal.new(0.02),
    _funding_point_cap_low = decimal.new(-0.02),

    _funding_constant = decimal.new(0.0000125),

    _market_interval = 5,  -- PROD: 60
    _profile_interval = 20, -- PROD: 10

    -- TO BE ABLE TO TEST
    _30m_rolling_period = 60, -- PROD: 60
    _30m_rolling_max = 30, -- PROD: 30

    _60m_rolling_period = 60, -- PROD: 60
    _60m_rolling_max = 60, -- PROD: 60

    _fibers = {}

}

function periodics._do_market_periodics()
    if periodics._market_id == nil then
        local err = PeriodicsError:new('INTEGRITY_ERROR: _do_profile_periodics empty market_id=%s', periodics._market_id)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local market = box.space.market:get(periodics._market_id)
    if market == nil then
        local err = PeriodicsError:new('INTEGRITY_ERROR: _do_profile_periodics market not found for market_id=%s', periodics._market_id)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end
    if market.index_price == 0 then
        local text = "_do_market_periodics no index_price for market_id=" .. tostring(periodics._market_id)
        log.error(PeriodicsError:new(text))
        return text
    end

    local insideTx = box.is_in_txn()


    local e

    local time_split = os.date ("*t")
    local current_minute = time_split.min

    if insideTx == false then
        box.begin()
    end

        local basis = (market.market_price - market.index_price) / market.index_price

        local capped_basis = basis
        if capped_basis < periodics._basis_low_border then
            capped_basis = periodics._basis_low_border
        elseif capped_basis > periodics._basis_up_border then
            capped_basis = periodics._basis_up_border
        end


        -- 60 seconds, 30 max values
        e = m.update_roll_value("30m_basis",
                                periodics._market_id,
                                capped_basis,
                                periodics._30m_rolling_period,
                                periodics._30m_rolling_max,
                                true)
        if e ~= nil then
            log.error("30m_basis update_roll_value error=%s", PeriodicsError:new(e))
            box.rollback()
            return e
        end

        -- 60 seconds, 60 max values
        e = m.update_roll_value("60m_basis",
                                periodics._market_id,
                                capped_basis,
                                periodics._60m_rolling_period,
                                periodics._60m_rolling_max,
                                true)
        if e ~= nil then
            log.error("60m_basis update_roll_value error=%s", PeriodicsError:new(e))
            box.rollback()
            return e
        end

        -- we cap funding for 6% per data point
        -- DISCUSSION on May 2023, to protect funding from situations with no orders change in the book
        local funding_capped_basis = basis
        if funding_capped_basis < periodics._funding_point_cap_low then
            funding_capped_basis = periodics._funding_point_cap_low
        elseif funding_capped_basis > periodics._funding_point_cap_high then
            funding_capped_basis = periodics._funding_point_cap_high
        end


        -- 60 max values per hour - but used for funding because it should be reset
        e = m.update_roll_value("funding_60m_basis",
                                periodics._market_id,
                                funding_capped_basis,
                                periodics._60m_rolling_period,
                                periodics._60m_rolling_max,
                                true)
        if e ~= nil then
            log.error("funding_60m_basis update_roll_value error=%s", PeriodicsError:new(e))
            box.rollback()
            return e
        end


        local price1 = market.index_price * (1 + m.get_roll_avg("60m_basis", periodics._market_id))
        local price2 = market.index_price * (1 + m.get_roll_avg("30m_basis", periodics._market_id))

        local median = {price1, price2, market.market_price}
        table.sort(median)
        local new_fair_price = median[2]

        if new_fair_price ~= market.fair_price and new_fair_price ~= 0 then
            e = m.update_fair_price(periodics._market_id, new_fair_price)
            if e ~= nil then
                log.error("update_fair_price error=%s", PeriodicsError:new(e))
                box.rollback()
                return e
            end
        end


        --
        -- UPDATE funding rate
        --
        local tm = time.now()
        local need_update_funding = false
        local funding_update_diff = tm - market.last_funding_update_time

        if funding_update_diff >= TEN_MINUTES  and current_minute < UPDATE_FUNDING_MAX_MINUTES then
            need_update_funding = true
        end

        -- Funding rate need to be feat to 1 hour, so we divide to 24
        local instant_funding_rate = m.get_roll_avg("funding_60m_basis", periodics._market_id) / 8 + periodics._funding_constant
        if instant_funding_rate < periodics._funding_low_border then
            instant_funding_rate = periodics._funding_low_border
        elseif instant_funding_rate > periodics._funding_up_border then
            instant_funding_rate = periodics._funding_up_border
        end

        --we always update instant funding

        e = m.update_funding(periodics._market_id, instant_funding_rate, instant_funding_rate, tm, need_update_funding)
        if e ~= nil then
            log.error("update_funding error=%s", PeriodicsError:new(e))
            box.rollback()
            return e
        end

        if need_update_funding == true then
            m.reset_roll_value("funding_60m_basis", periodics._market_id)
        end

    if insideTx == false then
        box.commit()
    end


    notif.notify_market(market.id)
    return nil
end

function periodics._do_profile_periodics()
    local market = box.space.market:get(periodics._market_id)
    if market == nil then
        local err = PeriodicsError:new("INTEGRITY_ERROR: _do_profile_periodics market not found market_id=%s", periodics._market_id)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if market.fair_price == 0 then
        local err = PeriodicsError:new("INTEGRITY_ERROR: _do_profile_periodics market.fair_price=0 market_id=%s", periodics._market_id)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local count = 0

    box.begin()
    for _, pos in position.iterator_by_market_profile(market.id) do
        local meta_item = box.space.profile_meta:get(pos.profile_id)
        if meta_item == nil then
            local err = PeriodicsError:new("INTEGRITY_ERROR: _do_profile_periodics: profile meta not found: profile_id=%d market_id=%s", pos.profile_id, market.id)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            break
        end
        if meta_item.status == config.params.PROFILE_STATUS.ACTIVE then
            en_meta.update_profile_meta(market, meta_item.profile_id, meta_item)

            count = util.safe_yield(count, 1000)
        end
    end
    box.commit()

    -- NOTIFY all positions
    for _, position in box.space.position:pairs(nil, {iterator = 'ALL'}) do
        notif.notify_position(position.profile_id, position.id)
    end


    -- box.broadcast(config.sys.EVENTS.PROFILE_UPDATE, {market_id=periodics._market_id, profiles={}})
    return nil
end

function periodics.update_profiles(profiles)
    if periodics._market_id == nil then
        local text = "periodics._market_id=" .. tostring(periodics._market_id)
        log.error(PeriodicsError:new(text))
        return text
    end
    local market = box.space.market:get(periodics._market_id)
    if market == nil then
        local err = PeriodicsError:new("INTEGRITY_ERROR: periodics.update_profiles market not found market_id=%s", periodics._market_id)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local insideTx = box.is_in_txn()

    if insideTx == false then
        box.begin()
    end

    local count = 0
    for _, profile_id in pairs(profiles) do
        local profile_meta = box.space.profile_meta:get(profile_id)
        if profile_meta == nil then
            local text = "INTEGRITY error periodics.update_profiles profile_meta nil for profile_id=" .. tostring(profile_id)
            -- log.error(err:new(text))
        else
            en_meta.update_profile_meta(market, profile_id, profile_meta)

            count = util.safe_yield(count, 1000)

            -- TODO: once pub-sub on the same machine we can notify from transaction
            --[[
            local pos_id = "pos-" .. tostring(market.id) .. "-tr-" .. tostring(profile_id)
            notif.notify_position(profile_id, pos_id)
            --]]
        end
    end

    if insideTx == false then
        box.commit()
    end

    if profiles ~= nil and #profiles > 0 then
        box.broadcast(config.sys.EVENTS.PROFILE_UPDATE, {market_id=periodics._market_id, profiles=profiles})
    end

    return nil
end

function periodics.start(market_id, min_tick, min_order)
    periodics._market_id = market_id
    periodics._min_tick = min_tick
    periodics._min_order = min_order

    log.info("PERIODICS_LOG: called engine.periodics.start")

    if config.sys.MODE == "sync" then
         log.warn("*** SYNC MODE ON: not starting periodics")
         return
    end

    local market_fiber = fiber.create(function()
         while true do
             fiber.testcancel()

             local status, err = pcall(function()
                    local e = periodics._do_market_periodics()
                    if e ~= nil then
                        log.error("_market_periodics error=%s", PeriodicsError:new(e))
                    end
             end)

             if status == false then
                 log.error({
                    message = string.format('ENGINE_MARKET_PERIODICS_SYSTEM_ERROR: %s', PeriodicsError:new(err)),
                    [ALERT_TAG] = ALERT_CRIT,
                 })
             end
             fiber.sleep(periodics._market_interval)
         end
     end)
     table.insert(periodics._fibers, market_fiber)

     local profile_fiber = fiber.create(function()
         while true do
             fiber.testcancel()

             local status, err = pcall(function()
                local e = periodics._do_profile_periodics()
                if e ~= nil then
                    log.error("_profile_periodics error=%s", PeriodicsError:new(e))
                end
            end)

            if status == false then
                log.error({
                    message = string.format('ENGINE_MARKET_PERIODICS_SYSTEM_ERROR: %s', PeriodicsError:new(err)),
                    [ALERT_TAG] = ALERT_CRIT,
                 })
            end

             fiber.sleep(periodics._profile_interval)
         end
     end)
     table.insert(periodics._fibers, profile_fiber)

     log.info("PERIODICS_LOG: engine.periodics.start success")

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
