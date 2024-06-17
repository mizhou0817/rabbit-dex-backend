local checks = require('checks')
local log = require('log')

local config = require('app.config')
local errors = require('app.lib.errors')
local tick = require('app.lib.tick')
local util = require('app.util')

require("app.config.constants")
require('app.errcodes')

local RiskError = errors.new_class("RISK")

-- all orders after PLACED status look like limit ones
local function pre_amend_as_limit_order(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if amend_order.price == nil or amend_order.price <= 0 then
        return "bad order price"
    end
    if amend_order.size ~= nil and amend_order.size <= 0 then
        return "bad order size"
    end

    local near_price = tick.round_to_nearest_tick(amend_order.price, market.min_tick)
    if near_price ~= amend_order.price then
        local text = "bad_price has="..tostring(amend_order.price).." required="..tostring(near_price)
        return text
    end

    local fair_price = market.fair_price
    if order.side == config.params.LONG then
        local limit_price = tick.round_to_nearest_tick(market_config.limit_buy_ratio * fair_price, market.min_tick)
        if amend_order.price > limit_price then
            local text = "RiskManager CheckOrderParams: price=" .. tostring(amend_order.price) .. " should be less/equal than= " .. tostring(limit_price)
            return text
        end
    else
        local limit_price = tick.round_to_nearest_tick(market_config.limit_sell_ratio * fair_price, market.min_tick)
        if amend_order.price < limit_price then
            local text = "RiskManager CheckOrderParams: price=" .. tostring(amend_order.price) .. " should be greater/equal than= " .. tostring(limit_price)
            return text
        end
    end

    local new_size = order.size
    
    if amend_order.size ~= nil then
        new_size = amend_order.size

        local near_size = tick.round_to_nearest_tick(amend_order.size, market.min_order)
        if near_size ~= amend_order.size then
            local text = "bad_order_size has="..tostring(amend_order.size).." required="..tostring(near_size)
            return text
        end
    end


    local notional = amend_order.price * new_size
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

local market_order = {}

function market_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.MARKET then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local tif = config.params.TIME_IN_FORCE
    if not util.is_value_in(order.time_in_force, {tif.GTC, tif.IOC, tif.FOK}) then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    local ratio = (order.side == config.params.LONG)
        and market_config.limit_buy_ratio
        or market_config.limit_sell_ratio
    order.price = tick.round_to_nearest_tick(ratio * market.fair_price, market.min_tick)

    if order.price <= 0 then
        return "bad order price"
    end
    if order.size == nil or order.size <= 0 then
        return "bad order size"
    end

    local near_size = tick.round_to_nearest_tick(order.size, market.min_order)
    if near_size ~= order.size then
        local text = "bad_order_size has=" .. tostring(order.size) .. " required=" .. tostring(near_size)
        return text
    end

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

function market_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.MARKET then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return pre_amend_as_limit_order(amend_order, order, market, market_config)
end

local limit_order = {}

function limit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local tif = config.params.TIME_IN_FORCE
    if not util.is_value_in(order.time_in_force, {tif.GTC, tif.IOC, tif.FOK, tif.POST_ONLY}) then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    if order.price == nil or order.price <= 0 then
        return "bad order price"
    end
    if order.size == nil or order.size <= 0 then
        return "bad order size"
    end

    local near_price = tick.round_to_nearest_tick(order.price, market.min_tick)
    if near_price ~= order.price then
        local text = "bad_price has=" .. tostring(order.price) .. " required=" .. tostring(near_price)
        return text
    end

    local near_size = tick.round_to_nearest_tick(order.size, market.min_order)
    if near_size ~= order.size then
        local text = "bad_order_size has=" .. tostring(order.size) .. " required=" .. tostring(near_size)
        return text
    end

    local fair_price = market.fair_price
    if order.side == config.params.LONG then
        local limit_price = tick.round_to_nearest_tick(market_config.limit_buy_ratio * fair_price, market.min_tick)
        if order.price > limit_price then
            local text = "RiskManager CheckOrderParams: price=" .. tostring(order.price) .. " should be less/equal than= " .. tostring(limit_price)
            return text
        end
    else
        local limit_price = tick.round_to_nearest_tick(market_config.limit_sell_ratio * fair_price, market.min_tick)
        if order.price < limit_price then
            local text = "RiskManager CheckOrderParams: price=" .. tostring(order.price) .. " should be greater/equal than= " .. tostring(limit_price)
            return text
        end
    end

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

function limit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return pre_amend_as_limit_order(amend_order, order, market, market_config)
end

local stop_loss_common = {}

function stop_loss_common.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.time_in_force ~= config.params.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end
    if order.trigger_price == nil or order.trigger_price <= 0 then
        return "bad order trigger price"
    end
    if order.size_percent == nil or order.size_percent <= 0 or order.size_percent > 1.0 then
        return "bad order size percent"
    end

    local near_price = tick.round_to_nearest_tick(order.trigger_price, market.min_tick)
    if near_price ~= order.trigger_price then
        local text = "bad_trigger_price has=" .. tostring(order.trigger_price) .. " required=" .. tostring(near_price)
        return text
    end

    order.size = ZERO

    return nil
end

function stop_loss_common.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        return pre_amend_as_limit_order(amend_order, order, market, market_config)
    else
        if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
            return "bad order trigger price"
        end
        if amend_order.size_percent == nil or amend_order.size_percent <= 0 or amend_order.size_percent > 1.0 then
            return "bad order percent size"
        end

        local near_price = tick.round_to_nearest_tick(amend_order.trigger_price, market.min_tick)
        if near_price ~= amend_order.trigger_price then
            local text = "bad_trigger_price has="..tostring(amend_order.trigger_price).." required="..tostring(near_price)
            return text
        end
    end

    return nil
end

local stop_loss_order = {}

function stop_loss_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = stop_loss_common.pre_create(order, market, market_config)
    if err ~= nil then
        return err
    end

    order.price = ZERO
    return nil
end

function stop_loss_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = stop_loss_common.pre_amend(amend_order, order, market, market_config)
    if err ~= nil then
        return err
    end

    return nil
end

local stop_loss_limit_order = {}

function stop_loss_limit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = stop_loss_common.pre_create(order, market, market_config)
    if err ~= nil then
        return err
    end

    if order.price == nil or order.price <= 0 then
        return RiskError:new("bad order price")
    end

    local near_price = tick.round_to_nearest_tick(order.price, market.min_tick)
    if near_price ~= order.price then
        return RiskError:new('bad_price has=%s required=%s', order.price, near_price)
    end

    return nil
end

function stop_loss_limit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if amend_order.price == nil or amend_order.price <= 0 then
        amend_order.price = order.price
    end
    if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
        amend_order.trigger_price = order.trigger_price
    end
    if amend_order.size_percent == nil or amend_order.size_percent <= 0 then
        amend_order.size_percent = order.size_percent
    end

    local err = stop_loss_common.pre_amend(amend_order, order, market, market_config)
    if err ~= nil then
        return err
    end

    if amend_order.price <= 0 then
        return RiskError:new("bad order price")
    end

    local near_price = tick.round_to_nearest_tick(amend_order.price, market.min_tick)
    if near_price ~= amend_order.price then
        return RiskError:new('bad_price has=%s required=%s', amend_order.price, near_price)
    end

    if order.side == config.params.LONG then
        local max_price = amend_order.trigger_price * market_config.sltp_limit_buy_ratio
        if amend_order.price > max_price then
            return RiskError:new('RiskManager CheckOrderParams: price=%s should be less/equal than=%s', amend_order.price, max_price)
        end
    else
        local min_price = amend_order.trigger_price * market_config.sltp_limit_sell_ratio
        if amend_order.price < min_price then
            return RiskError:new('RiskManager CheckOrderParams: price=%s should be greater/equal than=%s', amend_order.price, min_price)
        end
    end

    return nil
end

local take_profit_common = {}

function take_profit_common.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.time_in_force ~= config.params.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end
    if order.trigger_price == nil or order.trigger_price <= 0 then
        return "bad order trigger price"
    end
    if order.size_percent == nil or order.size_percent <= 0 or order.size_percent > 1.0 then
        return "bad order size percent"
    end

    local near_price = tick.round_to_nearest_tick(order.trigger_price, market.min_tick)
    if near_price ~= order.trigger_price then
        local text = "bad_trigger_price has=" .. tostring(order.trigger_price) .. " required=" .. tostring(near_price)
        return text
    end

    order.size = ZERO

    return nil
end

function take_profit_common.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        return pre_amend_as_limit_order(amend_order, order, market, market_config)
    else
        if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
            return "bad order trigger price"
        end
        if amend_order.size_percent == nil or amend_order.size_percent <= 0 or amend_order.size_percent > 1.0 then
            return "bad order percent size"
        end

        local near_price = tick.round_to_nearest_tick(amend_order.trigger_price, market.min_tick)
        if near_price ~= amend_order.trigger_price then
            local text = "bad_trigger_price has="..tostring(amend_order.trigger_price).." required="..tostring(near_price)
            return text
        end
    end

    return nil
end

local take_profit_order = {}

function take_profit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = take_profit_common.pre_create(order, market, market_config)
    if err ~= nil then
        return RiskError:new(err)
    end

    order.price = ZERO

    return nil
end

function take_profit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = take_profit_common.pre_amend(amend_order, order, market, market_config)
    if err ~= nil then
        return err
    end

    return nil
end

local take_profit_limit_order = {}

function take_profit_limit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = take_profit_common.pre_create(order, market, market_config)
    if err ~= nil then
        return RiskError:new(err)
    end

    if order.price == nil or order.price <= 0 then
        return RiskError:new("bad order price")
    end

    local near_price = tick.round_to_nearest_tick(order.price, market.min_tick)
    if near_price ~= order.price then
        return RiskError:new('bad_price has=%s required=%s', order.price, near_price)
    end

    return nil
end

function take_profit_limit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if amend_order.price == nil or amend_order.price <= 0 then
        amend_order.price = order.price
    end
    if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
        amend_order.trigger_price = order.trigger_price
    end
    if amend_order.size_percent == nil or amend_order.size_percent <= 0 then
        amend_order.size_percent = order.size_percent
    end

    local err = take_profit_common.pre_amend(amend_order, order, market, market_config)
    if err ~= nil then
        return RiskError:new(err)
    end

    if amend_order.price <= 0 then
        return RiskError:new("bad order price")
    end

    local near_price = tick.round_to_nearest_tick(amend_order.price, market.min_tick)
    if near_price ~= amend_order.price then
        return RiskError:new('bad_price has=%s required=%s', amend_order.price, near_price)
    end

    if order.side == config.params.LONG then
        local max_price = amend_order.trigger_price * market_config.sltp_limit_buy_ratio
        if amend_order.price > max_price then
            return RiskError:new('RiskManager CheckOrderParams: price=%s should be less/equal than=%s', amend_order.price, max_price)
        end
    else
        local min_price = amend_order.trigger_price * market_config.sltp_limit_sell_ratio
        if amend_order.price < min_price then
            return RiskError:new('RiskManager CheckOrderParams: price=%s should be greater/equal than=%s', amend_order.price, min_price)
        end
    end

    return nil
end

local stop_limit_order = {}

function stop_limit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.time_in_force ~= config.params.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end
    if order.trigger_price == nil or order.trigger_price <= 0 then
        return "bad order trigger price"
    end
    if order.price == nil or order.price <= 0 then
        return "bad order price"
    end
    if order.size == nil or order.size <= 0 then
        return "bad order size"
    end

    local near_trigger_price = tick.round_to_nearest_tick(order.trigger_price, market.min_tick)
    if near_trigger_price ~= order.trigger_price then
        local text = "bad_trigger_price has=" .. tostring(order.trigger_price) .. " required=" .. tostring(near_trigger_price)
        return text
    end

    local near_price = tick.round_to_nearest_tick(order.price, market.min_tick)
    if near_price ~= order.price then
        local text = "bad_price has=" .. tostring(order.price) .. " required=" .. tostring(near_price)
        return text
    end

    local near_size = tick.round_to_nearest_tick(order.size, market.min_order)
    if near_size ~= order.size then
        local text = "bad_order_size has=" .. tostring(order.size) .. " required=" .. tostring(near_size)
        return text
    end

    local fair_price = market.fair_price

    if order.side == config.params.LONG then
        local max_price = order.trigger_price * config.params.LIMIT_BUY_RATIO
        if order.price > max_price then
            --ERR_ORDER_PRICE_OVERFLOW
            return string.format("RiskManager CheckOrderParams: price=%s should be less/equal than=%s", order.price, max_price)
        end
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be greater than=%s", order.trigger_price, fair_price)
        end
    else
        local min_price = order.trigger_price * config.params.LIMIT_SELL_RATIO
        if order.price < min_price then
            --ERR_ORDER_PRICE_OVERFLOW
            return string.format("RiskManager CheckOrderParams: price=%s should be greater/equal than=%s", order.price, min_price)
        end
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be less than=%s", order.trigger_price, fair_price)
        end
    end

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

function stop_limit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        return pre_amend_as_limit_order(amend_order, order, market, market_config)
    else
        if amend_order.price == nil or amend_order.price <= 0 then
            return "bad order price"
        end
        if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
            return "bad order trigger price"
        end
        if amend_order.size == nil or amend_order.size <= 0 then
            return "bad order size"
        end

        local near_price = tick.round_to_nearest_tick(amend_order.price, market.min_tick)
        if near_price ~= amend_order.price then
            local text = "bad_price has=" .. tostring(amend_order.price) .. " required=" .. tostring(near_price)
            return text
        end

        local trigger_price = tick.round_to_nearest_tick(amend_order.trigger_price, market.min_tick)
        if trigger_price ~= amend_order.trigger_price then
            local text = "bad_trigger_price has=" .. tostring(amend_order.trigger_price) .. " required=" .. tostring(near_price)
            return text
        end

        local near_size = tick.round_to_nearest_tick(amend_order.size, market.min_order)
        if near_size ~= amend_order.size then
            local text = "bad_order_size has=" .. tostring(amend_order.size) .. " required=" .. tostring(near_size)
            return text
        end

        local fair_price = market.fair_price

        if order.side == config.params.LONG then
            local max_price = order.trigger_price * config.params.LIMIT_BUY_RATIO
            if amend_order.price > max_price then
                --ERR_ORDER_PRICE_OVERFLOW
                return string.format("RiskManager CheckOrderParams: price=%s should be less/equal than=%s", amend_order.price, max_price)
            end
            if  order.trigger_price <= fair_price then
                -- immediate buying is forbidden
                --ERR_ORDER_IMMEDIATE_EXECUTION
                return string.format("RiskManager CheckOrderParams: trigger-price=%s should be greater than=%s", order.trigger_price, fair_price)
            end
        else
            local min_price = order.trigger_price * config.params.LIMIT_SELL_RATIO
            if amend_order.price < min_price then
                --ERR_ORDER_PRICE_OVERFLOW
                return string.format("RiskManager CheckOrderParams: price=%s should be greater/equal than=%s", amend_order.price, min_price)
            end
            if  order.trigger_price >= fair_price then
                -- immediate selling is forbidden
                --ERR_ORDER_IMMEDIATE_EXECUTION
                return string.format("RiskManager CheckOrderParams: trigger-price=%s should be less than=%s", order.trigger_price, fair_price)
            end
        end

        local notional = amend_order.price * amend_order.size
        -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
        -- After we will have mechanics to deliver adv to tarantool
        local max_order = market_config.adv_constant
        if notional >= max_order then
            local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
            return text
        end
    end

    return nil
end

local stop_market_order = {}

function stop_market_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_MARKET then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.time_in_force ~= config.params.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end
    if order.trigger_price == nil or order.trigger_price <= 0 then
        return "bad order trigger price"
    end
    if order.size == nil or order.size <= 0 then
        return "bad order size"
    end

    local near_price = tick.round_to_nearest_tick(order.trigger_price, market.min_tick)
    if near_price ~= order.trigger_price then
        local text = "bad_trigger_price has=" .. tostring(order.trigger_price) .. " required=" .. tostring(near_price)
        return text
    end

    local near_size = tick.round_to_nearest_tick(order.size, market.min_order)
    if near_size ~= order.size then
        local text = "bad_order_size has=" .. tostring(order.size) .. " required=" .. tostring(near_size)
        return text
    end

    local fair_price = market.fair_price

    if order.side == config.params.LONG then
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be greater than=%s", order.trigger_price, fair_price)
        end
    else
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be less than=%s", order.trigger_price, fair_price)
        end
    end
    local ratio = (order.side == config.params.LONG)
        and market_config.limit_buy_ratio
        or market_config.limit_sell_ratio
    local order_price = tick.round_to_nearest_tick(ratio * market.fair_price, market.min_tick)

    local notional = order_price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

function stop_market_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_MARKET then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        return pre_amend_as_limit_order(amend_order, order, market, market_config)
    else
        if amend_order.trigger_price == nil or amend_order.trigger_price <= 0 then
            return "bad order trigger price"
        end
        if amend_order.size == nil or amend_order.size <= 0 then
            return "bad order size"
        end

        local near_price = tick.round_to_nearest_tick(amend_order.trigger_price, market.min_tick)
        if near_price ~= amend_order.trigger_price then
            local text = "bad_trigger_price has=" .. tostring(amend_order.trigger_price) .. " required=" .. tostring(near_price)
            return text
        end

        local near_size = tick.round_to_nearest_tick(amend_order.size, market.min_order)
        if near_size ~= amend_order.size then
            local text = "bad_order_size has=" .. tostring(amend_order.size) .. " required=" .. tostring(near_size)
            return text
        end
    end

    local fair_price = market.fair_price

    if order.side == config.params.LONG then
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be greater than=%s", order.trigger_price, fair_price)
        end
    else
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            --ERR_ORDER_IMMEDIATE_EXECUTION
            return string.format("RiskManager CheckOrderParams: trigger-price=%s should be less than=%s", order.trigger_price, fair_price)
        end
    end

    return nil
end

local ping_limit_order = {}

function ping_limit_order.pre_create(order, market, market_config)
    checks('table|api_create_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.PING_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.size == nil or order.size <= 0 then
        return "bad order size"
    end

    local near_size = tick.round_to_nearest_tick(order.size, market.min_order)
    if near_size ~= order.size then
        local text = "bad_order_size has=" .. tostring(order.size) .. " required=" .. tostring(near_size)
        return text
    end

    if market.best_ask == nil or market.best_ask <= 0 or 
        market.best_bid == nil or market.best_bid <= 0 then
        local text = "mid_price not possible best_ask=" .. tostring(market.best_ask) .. " best_bid=" .. tostring(market.best_bid)
        return text
    end

    local mid_price = tick.calc_mid_price(market.best_ask, market.min_tick)

    if not tick.is_mid_price_valid(mid_price, market.best_ask, market.best_bid) then
        local text = "mid_price not possible has=" .. tostring(mid_price) .. " best_ask=" .. tostring(market.best_ask) .. " best_bid=" .. tostring(market.best_bid)
        return text
    end

    local fair_price = market.fair_price
    local limit_price_long = tick.round_to_nearest_tick(market_config.limit_buy_ratio * fair_price, market.min_tick)
    if mid_price > limit_price_long then
        local text = "RiskManager CheckOrderParams: price=" .. tostring(mid_price) .. " should be less/equal than= " .. tostring(limit_price_long)
        return text
    end

    local limit_price_short = tick.round_to_nearest_tick(market_config.limit_sell_ratio * fair_price, market.min_tick)
    if mid_price < limit_price_short then
        local text = "RiskManager CheckOrderParams: price=" .. tostring(mid_price) .. " should be greater/equal than= " .. tostring(limit_price_short)
        return text
    end


    order.price = mid_price

    local notional = order.price * order.size
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "RiskManager CheckOrderParams: Notional more than allowed=" .. tostring(max_order)
        return text
    end

    return nil
end

function ping_limit_order.pre_amend(amend_order, order, market, market_config)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market', 'table')

    if order.order_type ~= config.params.ORDER_TYPE.PING_LIMIT then
        local err = RiskError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return ERR_ACTION_NOT_AVAILABLE
end


local risk = {
    _order_metatypes = {
        limit = limit_order,
        market = market_order,
        stop_loss = stop_loss_order,
        take_profit = take_profit_order,
        stop_loss_limit = stop_loss_limit_order,
        take_profit_limit = take_profit_limit_order,
        stop_limit = stop_limit_order,
        stop_market = stop_market_order,
        ping_limit = ping_limit_order,
    },
}

function risk.check_profile(profile)
    checks('table|profile')

    if profile.status ~= config.params.PROFILE_STATUS.ACTIVE then
        return ERR_PROFILE_NOT_ACTIVE
    end

    return nil
end

function risk.check_market(market)
    checks('table|engine_market')

    if market.status ~= config.params.MARKET_STATUS.ACTIVE then
        log.info("NOT_ACTIVE_MARKET_REQUEST: market=%s status=%s", market.id, market.status)
        return ERR_MARKET_NOT_ACTIVE
    end

    return nil
end

function risk.pre_create_order(order, market)
    checks('table|api_create_order', 'table|engine_market')

    local market_config = config.markets[market.id]
    if market_config == nil then
        log.error("market config not found for market_id=%s", market.id)
        return ERR_WRONG_MARKET_ID
    end

    if order.time_in_force == nil or order.time_in_force == '' then
        order.time_in_force = config.params.TIME_IN_FORCE.GTC
    end

    if order.client_order_id ~= nil then
        local l = string.len(order.client_order_id)
        if l > UUID_LEN then
            return ERR_CLIENT_ORDER_ID_TOO_LARGE
        end
    end

    local order_mtype = risk._order_metatypes[order.order_type]
    if order_mtype == nil then
        log.error("order handler not found: order_id=%s, order_type=%s", order.id, order.order_type)
        return ERR_WRONG_ORDER_TYPE
    end

    local err = order_mtype.pre_create(order, market, market_config)
    if err ~= nil then
        log.error(RiskError:new(err))
        return err
    end

    return nil
end

function risk.pre_amend_order(amend_order, order, market)
    checks('table|api_amend_order', 'table|engine_order', 'table|engine_market')

    local market_config = config.markets[market.id]
    if market_config == nil then
        log.error(RiskError:new("market config not found for market_id=%s", market.id))
        return ERR_WRONG_MARKET_ID
    end

    local order_mtype = risk._order_metatypes[order.order_type]
    if order_mtype == nil then
        log.error(RiskError:new("order handler not found for order_type=%s", order.order_type))
        return ERR_WRONG_ORDER_TYPE
    end

    local err = order_mtype.pre_amend(amend_order, order, market, market_config)
    if err ~= nil then
        log.error(RiskError:new(err))
        return err
    end

    return nil
end

function risk.liquidation_order_check(market, liquidate_kind)
    if market.status ~= config.params.MARKET_STATUS.ACTIVE then
        if liquidate_kind == config.params.LIQUIDATE_KIND.APLACESELLORDERS then
            return "paused_market"
        end
    end

    return nil
end

return risk