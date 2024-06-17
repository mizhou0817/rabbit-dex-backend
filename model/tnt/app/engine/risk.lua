local checks = require('checks')
local log = require('log')

local balance = require('app.balance')
local config = require('app.config')
local d = require('app.data')
local ag = require('app.engine.aggregate')
local errors = require('app.lib.errors')
local tick = require('app.lib.tick')

require('app.config.constants')
require('app.errcodes')

local RiskError = errors.new_class("ENGINE_RISK_ERROR")

local risk = {}

-- AFTER create/amend order this check should be done
-- CHECK passed only if:
-- withdrawble_balance >= 0
-- account_margin >= forced_margin
function risk.post_match(market_id, profile_data, profile_id, position_before)
    local market_config = config.markets[market_id]
    if market_config == nil then
        local text = "NO_CONFIG  _post_match for market_id=" .. tostring(market_id)
        log.error(RiskError:new(text))
        return text
    end

    local profile_cache = profile_data["cache"]
    if profile_cache == nil then
        local err = RiskError:new('INTEGRITY_ERROR_NO_CACHE: for %s', profile_id)
        log.error({message = err:backtrace(), [ALERT_TAG] = ALERT_CRIT})
        return err
    end

    local current_meta = box.space.profile_meta:get(profile_id)
    if current_meta == nil then
        local err = RiskError:new('INTEGRITY_ERROR_NO_CURRENT_META: for %s', profile_id)
        log.error({message = err:backtrace(), [ALERT_TAG] = ALERT_CRIT})
        return err
    end

    local _market = box.space.market:get(market_id)
    if _market == nil then
        local err = RiskError:new('INTEGRITY_ERROR_NO_MARKET_DATA: for %s', profile_id)
        log.error({message = err:backtrace(), [ALERT_TAG] = ALERT_CRIT})
        return err
    end

    -- WE NEED to substract previous value for this market from totals
    local prev_meta = profile_data["meta"]
    if prev_meta ~= nil then
        profile_cache[d.cache_balance] = profile_cache[d.cache_balance] - prev_meta[d.meta_balance]
        profile_cache[d.cache_cum_unrealized_pnl] = profile_cache[d.cache_cum_unrealized_pnl] - prev_meta[d.meta_cum_unrealized_pnl]
        profile_cache[d.cache_total_notional] = profile_cache[d.cache_total_notional] - prev_meta[d.meta_total_notional]
        profile_cache[d.cache_total_position_margin] = profile_cache[d.cache_total_position_margin] - prev_meta[d.meta_total_position_margin]
        profile_cache[d.cache_total_order_margin] = profile_cache[d.cache_total_order_margin] - prev_meta[d.meta_total_order_margin]
    end

    -- RECALC new values for:
    -- balance
    -- position_margin
    -- order_margin
    local new_balance = balance.get_balance(profile_id)

    local new_unrealized_pnl = ZERO
    local new_position_margin = ZERO
    local new_notional = ZERO

    local position_after = nil

    for _, position in box.space.position.index.profile_id:pairs(profile_id, {iterator = 'EQ'}) do                     
        position_after = position

        local sign = 1
        if position.side == config.params.SHORT then
            sign = -1
        end

        new_unrealized_pnl = position.size * (_market.fair_price - position.entry_price) * sign
        new_notional = position.size * _market.fair_price
        new_position_margin = current_meta.initial_margin * position.size * _market.fair_price
    end

    local total_order_notional = ag.get_order_total_notional(profile_id)
    local new_order_margin = current_meta.initial_margin * total_order_notional

    
    -- UPDATE CACHE values with new data for this market
    profile_cache[d.cache_balance] = profile_cache[d.cache_balance] + new_balance
    profile_cache[d.cache_cum_unrealized_pnl] = profile_cache[d.cache_cum_unrealized_pnl] + new_unrealized_pnl
    profile_cache[d.cache_total_notional] = profile_cache[d.cache_total_notional] + new_notional
    profile_cache[d.cache_total_position_margin] = profile_cache[d.cache_total_position_margin] + new_position_margin
    profile_cache[d.cache_total_order_margin] = profile_cache[d.cache_total_order_margin] + new_order_margin

    local new_account_equity = profile_cache[d.cache_balance] + profile_cache[d.cache_cum_unrealized_pnl]
    local new_account_margin = ONE

    if profile_cache[d.cache_total_notional] ~= 0 then 
        new_account_margin = new_account_equity / profile_cache[d.cache_total_notional]
    elseif new_account_equity <= 0 then
        new_account_margin = ZERO
    end

    local new_withdrawable_balance = tick.min(new_account_equity, profile_cache[d.cache_balance]) 
                - profile_cache[d.cache_total_position_margin] 
                - profile_cache[d.cache_total_order_margin]

    -- CHECK passed only if:
    -- reduce position scenario: need to check margin only
    if position_before ~= nil then

        --MUST BE: account_margin >= forced_margin
        if new_account_margin < market_config.forced_margin then
            return string.format('POST_MATCH_ERROR_MARGIN: account margin(%s) less than allowed margin(%s)', new_account_margin, market_config.forced_margin)
        end

        -- User closed or reduced position size, and didn't flip the side, margin is ok
        if position_after == nil or
            (position_before.side == position_after.side and
            position_before.size > position_after.size) then
                return nil
        end
    end

    --MUST BE: withdrawble_balance >= 0
    if new_withdrawable_balance < 0 then
        return string.format('POST_MATCH_ERROR_WB: withdrawable balance(%s) is negative', new_withdrawable_balance)
    end

    --TODO: refactoring is required, need remove code duplication
    --MUST BE: account_margin >= forced_margin
    if new_account_margin < market_config.forced_margin then
        return string.format('POST_MATCH_ERROR_MARGIN: account margin(%s) less than allowed margin(%s)', new_account_margin, market_config.forced_margin)
    end

    return nil
end

function risk.check_market(market)
    checks('table|engine_market')

    if market.status ~= config.params.MARKET_STATUS.ACTIVE then
        return ERR_MARKET_NOT_ACTIVE
    end

    return nil
end

function risk.calc_sltp_execution_size(position, order)
    checks('table|engine_position', 'table|engine_order')

    if position.size == 0 then
        log.error(RiskError:new('%s: %s', ERR_POSITION_NOT_FOUND, position.id))
        return ZERO, ERR_POSITION_NOT_FOUND
    end

    if order.size_percent == ONE then
        return position.size, nil
    end

    local desired_size = position.size * order.size_percent
    local rounded_size = tick.round_to_nearest_tick(desired_size, engine._min_order, engine._min_order)
    if rounded_size <= ZERO then
        local err = RiskError:new('%s: order(%s) size is negative or zero', ERR_WRONG_ORDER_SIZE, order.id)
        log.error(err:backtrace())
        return ZERO, err
    end
    if rounded_size > position.size then
        local err = RiskError:new('%s: order(%s) size(%s) is greater than position(%s) size(%s)', ERR_WRONG_ORDER_SIZE, order.id, order.size, position.id, position.size)
        log.error(err:backtrace())
        return ZERO, err
    end
    if not tick.is_valid_rounding(rounded_size - desired_size, engine._min_order) then
        local err = RiskError:new('%s: invalid rounding: max delta: %s, actual: %s, desired: %s', ERR_INTEGRITY_ERROR, engine._min_order, order.size, desired_size)
        log.error({message = err:backtrace(), [ALERT_TAG] = ALERT_CRIT})
        return ZERO, err
        end

    return rounded_size, nil
end

function risk.check_ping_limit_price(market_id, min_tick, mid_price, fair_price)
    checks('string', 'decimal', 'decimal', 'decimal')

    local market_config = config.markets[market_id]
    if market_config == nil then
        local text = "NO_CONFIG  check_ping_limit_price for market_id=" .. tostring(market_id)
        log.error(RiskError:new(text))
        return text
    end

    local limit_price_long = tick.round_to_nearest_tick(market_config.limit_buy_ratio * fair_price, min_tick)
    if mid_price > limit_price_long then
        local text = "RiskManager CheckOrderParams: price=" .. tostring(mid_price) .. " should be less/equal than= " .. tostring(limit_price_long)
        return text
    end

    local limit_price_short = tick.round_to_nearest_tick(market_config.limit_sell_ratio * fair_price, min_tick)
    if mid_price < limit_price_short then
        local text = "RiskManager CheckOrderParams: price=" .. tostring(mid_price) .. " should be greater/equal than= " .. tostring(limit_price_short)
        return text
    end

    return nil
end


function risk._test_set_post_match(fn)
    risk.post_match = fn
end

return risk