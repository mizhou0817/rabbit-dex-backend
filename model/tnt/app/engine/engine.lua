local checks = require('checks')
local fiber = require('fiber')
local json = require('json')
local log = require('log')
local metrics = require('metrics')

local action_creator = require('app.action')
local archiver = require('app.archiver')
local balance = require('app.balance')
local config = require('app.config')
local d = require('app.data')
local dml = require('app.dml')
local ag = require('app.engine.aggregate')
local candles = require('app.engine.candles')
local extended = require('app.engine.extended_data')
local market = require('app.engine.market')
local notif = require('app.engine.notif')
local o = require('app.engine.order')
local ob = require('app.engine.orderbook')
local periodics = require('app.engine.periodics')
local p = require('app.engine.position')
local profile = require('app.engine.profile')
local revert = require('app.revert')
local risk = require('app.engine.risk')
local trade = require('app.engine.trade')
local errors = require('app.lib.errors')
local tick = require('app.lib.tick')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')
local router = require('app.engine.router')
local cm_get = require('app.engine.cache_and_meta')

require("app.config.constants")
require('app.errcodes')

local EngineError = errors.new_class("ENGINE_ERROR")

local custom_execute_types = {
    config.params.ORDER_TYPE.PING_LIMIT,
}

local position_dep_order_types = {
    config.params.ORDER_TYPE.STOP_LOSS,
    config.params.ORDER_TYPE.STOP_LOSS_LIMIT,
    config.params.ORDER_TYPE.TAKE_PROFIT,
    config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT,
}

local engine = {
    _market_id = nil,
    _min_tick = nil,
    _min_order = nil,
    _loop = nil,
    _old_sequence = 0,

    _order_metatypes = {
    },
}

local function _notify_extended_position(profile_id)
    checks('number')

    local res = extended.get_extended_position(profile_id, engine._market_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return res.error
    end
    local pos = res.res
    if pos ~= nil then
        local timestamp = time.now()
        notif.add_private(engine._market_id, pos.id, timestamp, pos.profile_id, "position", pos)
    end

    return nil
end

local function _update_bid_ask(price, side, sequence)
    local new_size = ag.get_size(price, side)

    -- UPDATE PRICE LEVELS
    if side == config.params.LONG then
        notif.bid(price, new_size)
    else
        notif.ask(price, new_size)
    end


    local best_ask = ZERO
    local ask = box.space.orderbook.index.short:min({engine._market_id, config.params.SHORT})
    if ask ~= nil then
        best_ask = ask[d.entry_price]
    end

    local best_bid = ZERO
    local bid = box.space.orderbook.index.long:max({engine._market_id, config.params.LONG})
    if bid ~= nil then
        best_bid = bid[d.entry_price]
    end

    market.bid_ask_update(engine._market_id, best_bid, best_ask, sequence)
end

local function _open_order(order_id, price, size)
    checks('string', '?decimal', '?decimal')

    local order, err = o.open(order_id, price, size)
    if err ~= nil then
        return err
    end

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, 'order', order)

    return nil
end

local function _amend_order(order_id, new_price, new_size, new_trigger_price, new_size_percent)
    checks('string', '?decimal', '?decimal', '?decimal', '?decimal')

    local order, err = o.amend(order_id, new_price, new_size, new_trigger_price, new_size_percent)
    if err ~= nil then
        log.error(EngineError:new(err))
        return nil, err
    end

    local timestamp = time.now()
    notif.add_private(engine._market_id, tostring(order.id), timestamp, order.profile_id, "order", order)

    if util.is_value_in(order.order_type, position_dep_order_types) then
        local err = _notify_extended_position(order.profile_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return order, nil
end

local function _amend_entry(entry, new_size, sequence)
    checks('table|orderbook_entry', 'decimal', 'number')

    if new_size <= 0 then
        return "can't amend to 0 size entry_id=" .. tostring(entry.order_id)
    end

    local timestamp = time.now()
    --TODO: refactor it as orderbook.update()
     local res, err = archiver.update(box.space.orderbook, entry.order_id, {
        {'=', 'size', new_size},
        {'=', 'timestamp', timestamp},
        {'=', 'reverse', -timestamp},
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end
    if res == nil then
        return EngineError:new('ERR_ORDERBOOK_ENTRY_NOT_FOUND: orderbook entry not found: entry_id=%s', entry.id)
    end

    local _
    _, err = _amend_order(entry.order_id, nil, new_size, nil, nil)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    local diff = new_size - entry.size
    ag.add_price_level(entry.price, diff, entry.side)
    ag.add_order_notional(entry.trader_id, entry.price * diff)

    _update_bid_ask(entry.price, entry.side, sequence)

    return nil
end

local function _cancel_entry(entry, sequence)
    checks('table|engine_ob_entry', 'number')

    ob.delete(entry.order_id)

    local order, err = o.cancel(entry.order_id)
    if err ~= nil then
        return err
    end

    local diff = -entry.size
    ag.add_price_level(entry.price, diff, entry.side)
    ag.add_order_notional(entry.trader_id, entry.price * diff)

    _update_bid_ask(entry.price, entry.side, sequence)

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, "order", order)

    return nil
end

local function _limit_amend_modifier(amend, order)
    checks('table|api_amend_order', 'table|engine_order')

    if amend.size == nil or amend.size == 0 then
        amend.size = order.size
    end

    return amend
end

local function _amend_as_limit_order(entry, order, sequence, new_price, new_size)
    checks('?table|engine_ob_entry', 'table|engine_order', 'number', 'decimal', 'decimal')
    local err

    if entry == nil then
        log.error(EngineError:new('%s: orderbook entry is absent: order(%s)', ERR_ORDERBOOK_ENTRY_NOT_FOUND, order.id))
        return nil, false, ERR_ORDERBOOK_ENTRY_NOT_FOUND
    end

    --TODO: add validation as in api

    if new_price == entry.price then
        -- JUST CHANGE THE SIZE
        err = _amend_entry(entry, new_size, sequence)
        if err ~= nil then
            return nil, false, err
        end

        return order, false, nil
    end

    err = _cancel_entry(entry, sequence)
    if err ~= nil then
        return nil, false, err
    end

    local amended_order
    amended_order, err = _amend_order(entry.order_id, new_price, new_size, nil, nil)
    if err ~= nil then
        log.error(EngineError:new(err))
        return nil, false, err
    end

    return amended_order, true, nil
end

local limit_order = {}

function limit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local tif = config.params.TIME_IN_FORCE
    if not util.is_value_in(order.time_in_force, {tif.GTC, tif.IOC, tif.FOK, tif.POST_ONLY}) then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    return nil
end

function limit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')
    local cp = config.params

    if order.status ~= cp.ORDER_STATUS.OPEN then
        return ERR_WRONG_ORDER_STATUS
    end

    return nil
end

function limit_order.post_execute(order, left_size, sequence)
    checks('table|engine_order', 'decimal', 'number')

    if order.time_in_force == config.params.TIME_IN_FORCE.FOK then
        if left_size ~= 0 then
            return ERR_TIME_IN_FORCE_FOK_ERROR
        end
    elseif order.time_in_force == config.params.TIME_IN_FORCE.IOC then
        if left_size == order.initial_size then
            return ERR_TIME_IN_FORCE_IOC_ERROR
        elseif left_size ~= 0 and left_size < order.initial_size then
            local entry = ob.get(order.id)
            if entry == nil then
                return ERR_ORDERBOOK_ENTRY_NOT_FOUND
            end
            local err = _cancel_entry(entry, sequence)
            if err ~= nil then
                log.error(EngineError:new(err))
                return err
            end
        end
    elseif order.time_in_force == config.params.TIME_IN_FORCE.POST_ONLY then
        if left_size ~= order.initial_size then
            return ERR_TIME_IN_FORCE_POSTONLY_ERROR
        end
    end

    return nil
end

function limit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    amend = _limit_amend_modifier(amend, order)

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function limit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil
end

local market_order = {}

function market_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local tif = config.params.TIME_IN_FORCE
    if not util.is_value_in(order.time_in_force, {tif.GTC, tif.IOC, tif.FOK}) then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    return nil
end

function market_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')
    local cp = config.params

    if order.status ~= cp.ORDER_STATUS.OPEN then
        return ERR_WRONG_ORDER_STATUS
    end

    return nil
end

function market_order.post_execute(order, left_size, sequence)
    checks('table|engine_order', 'decimal', 'number')

    if order.time_in_force == config.params.TIME_IN_FORCE.FOK then
        if left_size ~= 0 then
            return ERR_TIME_IN_FORCE_FOK_ERROR
        end
    elseif order.time_in_force == config.params.TIME_IN_FORCE.IOC then
        if left_size == order.initial_size then
            return ERR_TIME_IN_FORCE_IOC_ERROR
        elseif left_size ~= 0 and left_size < order.initial_size then
            local entry = ob.get(order.id)
            if entry == nil then
                return ERR_ORDERBOOK_ENTRY_NOT_FOUND
            end
            local err = _cancel_entry(entry, sequence)
            if err ~= nil then
                log.error(EngineError:new(err))
                return err
            end
        end
    end

    return nil
end

function market_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.MARKET then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    amend = _limit_amend_modifier(amend, order)

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function market_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.MARKET then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil
end

local stop_loss_common = {}

function stop_loss_common.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price

    if position == nil or position.size == 0 then
        log.error('position not found: %s', util.tostring(order))
        return ERR_POSITION_NOT_FOUND
    end
    if position.side == cp.LONG then
        order.side = cp.SHORT
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    else
        order.side = cp.LONG
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    end
    if order.time_in_force ~= cp.TIME_IN_FORCE.GTC then
        log.error('wrong time in force: %s', util.tostring(order))
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    if not o.is_new_stop_loss_allowed(order.profile_id) then
        local errmsg = string.format('%s: out of allowed open %s orders', ERR_ORDER_LIMIT_REACHED, order.order_type)
        log.error(EngineError:new(errmsg))
        return errmsg
    end

    order.size = ZERO

    return nil
end

function stop_loss_common.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price

    if order.status ~= cp.ORDER_STATUS.PLACED then
        return ERR_WRONG_ORDER_STATUS
    end

    if order.side == cp.LONG then
        if order.trigger_price > fair_price then
            return ERR_NO_CONDITION_MET
        end
    else
        if order.trigger_price < fair_price then
            return ERR_NO_CONDITION_MET
        end
    end

    if position == nil or position.size == 0 then
        return ERR_POSITION_NOT_FOUND
    end

    return nil
end

function stop_loss_common.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price
    local err

    if order.status == config.params.ORDER_STATUS.PLACED then
        if entry ~= nil then
            local err = EngineError:new(ERR_INTEGRITY_ERROR)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return nil, false, err
        end

        if position == nil or position.size == 0 then
            log.error('%s: %s', ERR_POSITION_NOT_FOUND, util.tostring(order))
            return nil, false, ERR_POSITION_NOT_FOUND
        end
        if position.side == cp.LONG then
            if  amend.trigger_price >= fair_price then
                -- immediate selling is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        else
            if  amend.trigger_price <= fair_price then
                -- immediate buying is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        end

        --TODO: add validation as in api

        local amended_order
        amended_order, err = _amend_order(order.id, amend.price, nil, amend.trigger_price, amend.size_percent)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, false, err
        end

        return amended_order, false, nil
    end

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function stop_loss_common.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
    end

    return nil
end

local stop_loss_order = {}

function stop_loss_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = stop_loss_common.pre_create(order, position, market_data)
    if err ~= nil then
        return err
    end

    order.price = ZERO
    return nil
end

function stop_loss_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end
    if order.status == config.params.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end

    local err
    err = stop_loss_common.pre_execute(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[engine._market_id]

    local ratio = (order.side == config.params.LONG)
        and market_config.limit_buy_ratio
        or market_config.limit_sell_ratio

    order.price = tick.round_to_nearest_tick(ratio * market_data.fair_price, market_data.min_tick)
    local exec_size
    exec_size, err = risk.calc_sltp_execution_size(position, order)
    if err ~= nil then
        return err
    end
    order.size = exec_size
    order.initial_size = exec_size

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        return EngineError:new(ERR_ORDER_NOTIONAL_EXCEEDED)
    end

    return _open_order(order.id, order.price, order.size)
end

function stop_loss_order.post_execute(order, left_size, sequence)
end

function stop_loss_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    if order.status == config.params.ORDER_STATUS.PLACED then
        amend.price = ZERO
    end

    return stop_loss_common.amend(amend, entry, order, position, sequence, market_data)
end

function stop_loss_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return stop_loss_common.post_amend(profile_data, order, position_before)
end

local stop_loss_limit_order = {}

function stop_loss_limit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = stop_loss_common.pre_create(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[market_data.id]

    if order.side == config.params.LONG then
        local max_price = tick.round_to_nearest_tick(order.trigger_price * market_config.sltp_limit_buy_ratio, market_data.min_tick)
        if order.price < market_data.min_tick then
            return EngineError:new('RiskManager CheckOrderParams: price=%s should be greater/equal than=%s', order.price, market_data.min_tick)
        end
        if order.price > max_price then
            return EngineError:new('RiskManager CheckOrderParams: price=%s should be less/equal than=%s', order.price, max_price)
        end
    else
        local min_price = tick.round_to_nearest_tick(order.trigger_price * market_config.sltp_limit_sell_ratio, market_data.min_tick)
        if order.price < min_price then
            return EngineError:new('RiskManager CheckOrderParams: price=%s should be greater/equal than=%s', order.price, min_price)
        end
    end

    return nil
end

function stop_loss_limit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end
    if order.status == config.params.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end

    local err
    err = stop_loss_common.pre_execute(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[engine._market_id]
    local exec_size

    exec_size, err = risk.calc_sltp_execution_size(position, order)
    if err ~= nil then
        return err
    end
    order.size = exec_size
    order.initial_size = exec_size

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        return EngineError:new(ERR_ORDER_NOTIONAL_EXCEEDED)
    end

    return _open_order(order.id, order.price, order.size)
end

function stop_loss_limit_order.post_execute(order, left_size, sequence)
end

function stop_loss_limit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    return stop_loss_common.amend(amend, entry, order, position, sequence, market_data)
end

function stop_loss_limit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LOSS_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return stop_loss_common.post_amend(profile_data, order, position_before)
end

local take_profit_common = {}

function take_profit_common.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price

    if position == nil or position.size == 0 then
        log.error('position not found: %s', util.tostring(order))
        return ERR_POSITION_NOT_FOUND
    end
    if position.side == cp.LONG then
        order.side = cp.SHORT
        if  order.trigger_price <= fair_price then
            -- immediate selling is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    else
        order.side = cp.LONG
        if  order.trigger_price >= fair_price then
            -- immediate buying is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    end
    if order.time_in_force ~= cp.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    if not o.is_new_take_profit_allowed(order.profile_id) then
        local errmsg = string.format('%s: out of allowed open %s orders', ERR_ORDER_LIMIT_REACHED, order.order_type)
        log.error(EngineError:new(errmsg))
        return errmsg
    end

    order.size = ZERO

    return nil
end

function take_profit_common.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price

    if order.status ~= cp.ORDER_STATUS.PLACED then
        return ERR_WRONG_ORDER_STATUS
    end

    if order.side == cp.LONG then
        if order.trigger_price < fair_price then
            return ERR_NO_CONDITION_MET
        end
    else
        if order.trigger_price > fair_price then
            return ERR_NO_CONDITION_MET
        end
    end

    if position == nil or position.size == 0 then
        return ERR_POSITION_NOT_FOUND
    end

    return nil
end

function take_profit_common.post_execute(order, left_size, sequence)
end

function take_profit_common.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    local cp = config.params
    local fair_price = market_data.fair_price
    local err

    if order.status == config.params.ORDER_STATUS.PLACED then
        if entry ~= nil then
            local err = EngineError:new(ERR_INTEGRITY_ERROR)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return nil, false, err
        end

        if position == nil or position.size == 0 then
            log.error('%s: %s', ERR_POSITION_NOT_FOUND, util.tostring(order))
            return nil, false, ERR_POSITION_NOT_FOUND
        end
        if position.side == cp.LONG then
            if  amend.trigger_price <= fair_price then
                -- immediate selling is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        else
            if  amend.trigger_price >= fair_price then
                -- immediate buying is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        end

        --TODO: add validation as in api

        local amended_order
        amended_order, err = _amend_order(order.id, amend.price, nil, amend.trigger_price, amend.size_percent)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, false, err
        end

        return amended_order, false, nil
    end

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function take_profit_common.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
    end

    return nil
end

local take_profit_order = {}

function take_profit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = take_profit_common.pre_create(order, position, market_data)
    if err ~= nil then
        return err
    end

    order.price = ZERO
    return nil
end

function take_profit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end
    if order.status == config.params.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end

    local err
    err = take_profit_common.pre_execute(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[engine._market_id]

    local ratio = (order.side == config.params.LONG)
        and market_config.limit_buy_ratio
        or market_config.limit_sell_ratio

    order.price = tick.round_to_nearest_tick(ratio * market_data.fair_price, market_data.min_tick)
    local exec_size
    exec_size, err = risk.calc_sltp_execution_size(position, order)
    if err ~= nil then
        return err
    end
    order.size = exec_size
    order.initial_size = exec_size

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        return EngineError:new(ERR_ORDER_NOTIONAL_EXCEEDED)
    end

    return _open_order(order.id, order.price, order.size)
end

function take_profit_order.post_execute(order, left_size, sequence)
end

function take_profit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    if order.status == config.params.ORDER_STATUS.PLACED then
        amend.price = ZERO
    end

    return take_profit_common.amend(amend, entry, order, position, sequence, market_data)
end

function take_profit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return take_profit_common.post_amend(profile_data, order, position_before)
end

local take_profit_limit_order = {}

function take_profit_limit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local err = take_profit_common.pre_create(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[market_data.id]

    if order.side == config.params.LONG then
        local max_price = tick.round_to_nearest_tick(order.trigger_price * market_config.sltp_limit_buy_ratio, market_data.min_tick)
        if order.price < market_data.min_tick then
            return EngineError:new('RiskManager CheckOrderParams: price=%s should be greater/equal than=%s', order.price, market_data.min_tick)
        end
        if order.price > max_price then
            return EngineError:new('RiskManager CheckOrderParams: order price=%s should be less/equal than=%s', order.price, max_price)
        end
    else
        local min_price = tick.round_to_nearest_tick(order.trigger_price * market_config.sltp_limit_sell_ratio, market_data.min_tick)
        if order.price < min_price then
            return EngineError:new('RiskManager CheckOrderParams: order price=%s should be greater/equal than=%s', order.price, min_price)
        end
    end

    return nil
end

function take_profit_limit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end
    if order.status == config.params.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end

    local err
    err = take_profit_common.pre_execute(order, position, market_data)
    if err ~= nil then
        return err
    end

    local market_config = config.markets[engine._market_id]
    local exec_size

    exec_size, err = risk.calc_sltp_execution_size(position, order)
    if err ~= nil then
        return err
    end
    order.size = exec_size
    order.initial_size = exec_size

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        return EngineError:new(ERR_ORDER_NOTIONAL_EXCEEDED)
    end

    return _open_order(order.id, order.price, order.size)
end

function take_profit_limit_order.post_execute(order, left_size, sequence)
end

function take_profit_limit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    return take_profit_common.amend(amend, entry, order, position, sequence, market_data)
end

function take_profit_limit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    return take_profit_common.post_amend(profile_data, order, position_before)
end

local stop_limit_order = {}

function stop_limit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local order_type = order.order_type
    local fair_price = market_data.fair_price

    if order_type ~= cp.ORDER_TYPE.STOP_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.side == config.params.LONG then
        local max_price = tick.round_to_nearest_tick(order.trigger_price * config.params.LIMIT_BUY_RATIO, market_data.min_tick)
        if order.price > max_price then
            --ERR_ORDER_PRICE_OVERFLOW
            return string.format("RiskManager CheckOrderParams: price=%s should be less/equal than=%s", order.price, max_price)
        end
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    else
        local min_price = tick.round_to_nearest_tick(order.trigger_price * config.params.LIMIT_SELL_RATIO, market_data.min_tick)
        if order.price < min_price then
            --ERR_ORDER_PRICE_OVERFLOW
            return string.format("RiskManager CheckOrderParams: price=%s should be greater/equal than=%s", order.price, min_price)
        end
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    end
    if order.time_in_force ~= cp.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    if not o.is_new_stop_order_allowed(order.profile_id) then
        local errmsg = string.format('%s: out of allowed conditional stop orders', ERR_ORDER_LIMIT_REACHED)
        log.error(EngineError:new(errmsg))
        return errmsg
    end

    return nil
end

function stop_limit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local err
    local cp = config.params
    local fair_price = market_data.fair_price

    if order.status == cp.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end
    if order.status ~= cp.ORDER_STATUS.PLACED then
        return ERR_WRONG_ORDER_STATUS
    end

    if order.side == cp.LONG then
        if order.trigger_price > fair_price then
            return ERR_NO_CONDITION_MET
        end
    else
        if order.trigger_price < fair_price then
            return ERR_NO_CONDITION_MET
        end
    end

    order.initial_size = order.size

    err = _open_order(order.id, nil, order.size)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil
end

function stop_limit_order.post_execute(order, left_size, sequence)
end

function stop_limit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    local err
    local fair_price = market_data.fair_price

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    if order.status == config.params.ORDER_STATUS.PLACED then
        if entry ~= nil then
            local err = EngineError:new(ERR_INTEGRITY_ERROR)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return nil, false, err
        end

        --TODO: add validation as in api

        if order.side == config.params.LONG then
            if  amend.trigger_price <= fair_price then
                -- immediate buying is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        else
            if  amend.trigger_price >= fair_price then
                -- immediate selling is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        end

        local amended_order
        amended_order, err = _amend_order(order.id, amend.price, amend.size, amend.trigger_price, nil)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, false, err
        end

        return amended_order, true, nil
    end

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function stop_limit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_LIMIT then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
    end

    return nil
end

local stop_market_order = {}

function stop_market_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')

    local cp = config.params
    local order_type = order.order_type
    local fair_price = market_data.fair_price

    if order_type ~= cp.ORDER_TYPE.STOP_MARKET then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.side == config.params.LONG then
        if  order.trigger_price <= fair_price then
            -- immediate buying is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    else
        if  order.trigger_price >= fair_price then
            -- immediate selling is forbidden
            return ERR_ORDER_IMMEDIATE_EXECUTION
        end
    end

    if order.time_in_force ~= cp.TIME_IN_FORCE.GTC then
        --ERR_WRONG_TIME_IN_FORCE
        return string.format("RiskManager CheckOrderParams: invalid time-in-force=%s", order.time_in_force)
    end

    if not o.is_new_stop_order_allowed(order.profile_id) then
        local errmsg = string.format('%s: out of allowed conditional stop orders', ERR_ORDER_LIMIT_REACHED)
        log.error(EngineError:new(errmsg))
        return errmsg
    end

    return nil
end

function stop_market_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local err
    local cp = config.params
    local fair_price = market_data.fair_price

    if order.status == cp.ORDER_STATUS.OPEN then
        -- behaves like normal order
        return nil
    end
    if order.status ~= cp.ORDER_STATUS.PLACED then
        return ERR_WRONG_ORDER_STATUS
    end

    if order.side == cp.LONG then
        if order.trigger_price > fair_price then
            return ERR_NO_CONDITION_MET
        end
    else
        if order.trigger_price < fair_price then
            return ERR_NO_CONDITION_MET
        end
    end

    local market_config = config.markets[engine._market_id]

    local ratio = (order.side == config.params.LONG)
        and market_config.limit_buy_ratio
        or market_config.limit_sell_ratio

    order.price = tick.round_to_nearest_tick(ratio * market_data.fair_price, market_data.min_tick)
    order.initial_size = order.size

    local notional = order.price * order.size
    -- TODO: make it tick.max(market_config.adv_constant,market_config.adv_ratio  * market[d.market_average_daily_volume])
    -- After we will have mechanics to deliver adv to tarantool
    local max_order = market_config.adv_constant
    if notional >= max_order then
        local text = "Notional more than allowed=" .. tostring(max_order)
        return text
    end

    err = _open_order(order.id, order.price, order.size)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil
end

function stop_market_order.post_execute(order, left_size, sequence)
end

function stop_market_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    local err
    local fair_price = market_data.fair_price

    if order.order_type ~= config.params.ORDER_TYPE.STOP_MARKET then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, false, err
    end

    if order.status == config.params.ORDER_STATUS.PLACED then
        if entry ~= nil then
            local err = EngineError:new(ERR_INTEGRITY_ERROR)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return nil, false, err
        end

        --TODO: add validation as in api

        if order.side == config.params.LONG then
            if  amend.trigger_price <= fair_price then
                -- immediate buying is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        else
            if  amend.trigger_price >= fair_price then
                -- immediate selling is forbidden
                log.error('%s: %s', ERR_ORDER_IMMEDIATE_EXECUTION, util.tostring(order))
                return nil, false, ERR_ORDER_IMMEDIATE_EXECUTION
            end
        end

        local amended_order
        amended_order, err = _amend_order(order.id, nil, amend.size, amend.trigger_price, nil)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, false, err
        end

        return amended_order, true, nil
    end

    return _amend_as_limit_order(entry, order, sequence, amend.price, amend.size)
end

function stop_market_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    if order.order_type ~= config.params.ORDER_TYPE.STOP_MARKET then
        local err = EngineError:new(ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    if order.status ~= config.params.ORDER_STATUS.PLACED then
        local err = risk.post_match(order.market_id, profile_data, order.profile_id, position_before)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
    end

    return nil
end

local ping_limit_order = {}

function ping_limit_order.pre_create(order, position, market_data)
    checks('table|api_create_order', '?table|engine_position', 'table|engine_market')
    local cp = config.params

    local res = market.get_mid_price(engine._market_id)
    if res.error ~= nil then
        return res.error
    end

    local mid_price = res.res

    local err = risk.check_ping_limit_price(engine._market_id, engine._min_tick, mid_price, market_data.fair_price)
    if err ~= nil then
        return err
    end

    -- Check that all others orders not exist in orderbook on that price
    local exist = ob.get_by_price(engine._market_id, mid_price)
    if exist ~= nil then
        return ERR_PING_LIMIT_PRICE_NOT_UNQIUE
    end

    -- Now modify the current order
    order.price = mid_price
    if order.side == nil or order.side == "" then
        order.side = cp.LONG
    end

    return nil
end

function ping_limit_order.pre_execute(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')
    local cp = config.params

    if order.status ~= cp.ORDER_STATUS.OPEN then
        return ERR_WRONG_ORDER_STATUS
    end

    if order.id == nil or order.id == "" then
        return ERR_WRONG_ORDER_ID
    end
    return nil
end

function ping_limit_order.post_execute(order, left_size, sequence)
    checks('table|engine_order', 'decimal', 'number')

    if left_size ~= 0 then
        return ERR_PING_LIMIT_POST_EXECUTE
    end

    -- Check that nothing in orderbook left
    local exist = ob.get_by_price(engine._market_id, order.price)
    if exist ~= nil then
        return ERR_PING_LIMIT_POST_EXECUTE
    end

    return nil
end

function ping_limit_order.amend(amend, entry, order, position, sequence, market_data)
    checks('table|api_amend_order', '?table|engine_ob_entry', 'table|engine_order', '?table|engine_position', 'number', 'table|engine_market')

    return ERR_ACTION_NOT_AVAILABLE
end

function ping_limit_order.post_amend(profile_data, order, position_before)
    checks('table', 'table|engine_order', '?table|engine_position')

    return ERR_ACTION_NOT_AVAILABLE
end

function ping_limit_order.execute(ping_order, is_liquidation, no_post_match, profile_data, position)
    checks('table|engine_order', 'boolean', 'boolean', 'table', '?table|engine_position')

    local err, left_size, sequence
    local notifications = {}

    --[[
        EXECUTE ping order
    --]]
    sequence = engine._next_sequence()
    left_size, err = engine._match_order(
        ping_order.side,
        ping_order.price,
        ping_order.size,
        ping_order.id,
        ping_order.profile_id,
        is_liquidation,
        sequence
    )
    if err ~= nil then
        return nil, err
    end

    table.insert(notifications, notif.notify_to_table(engine._market_id, sequence))
    notif.clear()

    --[[
        NOW create pong
    --]]
    local pong_order_id = ping_order.id .. tostring("-pong")
    local pong_order_side = config.params.SHORT
    if ping_order.side == config.params.SHORT then
        pong_order_side = config.params.LONG
    end

    local pong_order
    pong_order, err = engine._create_order(
        pong_order_id,
        ping_order.profile_id,
        ping_order.order_type,
        ping_order.price,
        ping_order.size,
        ping_order.size,
        pong_order_side,
        "",
        ping_order.trigger_price,
        ping_order.size_percent,
        ping_order.time_in_force,
        is_liquidation
    )
    if err ~= nil then
        return nil, err
    end


    --[[
        NOW execute pong
    --]]
    sequence = engine._next_sequence()
    left_size, err = engine._match_order(
        pong_order.side,
        pong_order.price,
        pong_order.size,
        pong_order.id,
        pong_order.profile_id,
        is_liquidation,
        sequence
    )
    if err ~= nil then
        return nil, err
    end

    table.insert(notifications, notif.notify_to_table(engine._market_id, sequence))
    -- no need to clean notifications cuz we want to save account

    if not no_post_match then
        err = risk.post_match(engine._market_id, profile_data, pong_order.profile_id, position)
        if err ~= nil then
            return nil, err
        end
    end

    err = ping_limit_order.post_execute(pong_order, left_size, sequence)
    if err ~= nil then
        return nil, err
    end

    return notifications, nil
end


engine._order_metatypes = {
    limit = limit_order,
    market = market_order,
    stop_loss = stop_loss_order,
    take_profit = take_profit_order,
    stop_loss_limit = stop_loss_limit_order,
    take_profit_limit = take_profit_limit_order,
    stop_limit = stop_limit_order,
    stop_market = stop_market_order,
    ping_limit = ping_limit_order,
}

function engine._current_sequence()
    return tonumber(ob.sequence:current())
end

function engine._next_sequence()
    engine._old_sequence = tonumber(ob.sequence:current())
    return ob.sequence:next()
end

function engine._rollback_with_sequence(svp, to_sequence)
    checks('?table', '?number')

    if to_sequence ~= nil and to_sequence > 0 then
        ob.sequence:set(to_sequence)
    else
        ob.sequence:set(engine._old_sequence)
    end

    if svp == nil then
        box.rollback()
    else
        box.rollback_to_savepoint(svp)
    end

    return engine._old_sequence
end

local function _broadcast_changes()
    -- BROADCAST profiles that was changed
    local changed_ids = {}

    for _, profile in box.space.nf_profiles_changed:pairs(nil, {iterator = box.index.ALL}) do
        table.insert(changed_ids, profile.profile_id)
    end
    if #changed_ids > 0 then
        periodics.update_profiles(changed_ids)
    end

    changed_ids = nil
end

local function _update_position(trader_id, entry_side, fill_price, fill_size)
    local e, new_position

    local position = box.space.position.index.pos_by_market_profile:get({engine._market_id, trader_id})
    if position == nil then
        if fill_size == 0 then
            return ERR_DIV_ZERO
        end
        local entry_price = (fill_price * fill_size) / fill_size

        local price = entry_price -- tick.round_to_nearest_tick(entry_price, engine._min_tick)
        local size = tick.round_to_nearest_tick(fill_size, engine._min_order)

        new_position, e = p.create(engine._market_id, trader_id, size, entry_side, price)
        if e ~= nil then
            return e
        end
    elseif entry_side == position.side then
        local new_price = (fill_price * fill_size + position.size * position.entry_price) / (fill_size + position.size)
        local new_size = position.size + fill_size

        new_size = tick.round_to_nearest_tick(new_size, engine._min_order)

        new_position, e = p.update(position.id, position.side, new_size, new_price)
        if e ~= nil then
            return e
        end
    else
        local mulSide = 1
        if position.side == config.params.SHORT then
            mulSide = -1
        end

        local realized_pnl = ZERO
        local new_size = position.size
        local new_side = position.side
        local new_price = position.entry_price

        if fill_size < position.size then
            realized_pnl =  fill_size * (fill_price - position.entry_price) * mulSide
            new_size = position.size - fill_size
        else
            realized_pnl = position.size * (fill_price - position.entry_price) * mulSide
            new_price = fill_price
            new_side = entry_side
            new_size = fill_size - position.size
        end

        if realized_pnl ~= 0 then
            local balanceUpdate = balance.pay_realized_pnl(trader_id, realized_pnl)
            if balanceUpdate['error'] ~= nil then
                local text = "ERROR profile_pay_realized: " .. tostring(balanceUpdate['error'])
                return EngineError:new(text)
            end
        end

        new_size = tick.round_to_nearest_tick(new_size, engine._min_order)

        new_position, e = p.update(position.id, new_side, new_size, new_price)
        if e ~= nil then
            return e
        end
    end

    if new_position.size == ZERO then
        local tm = time.now()
        notif.add_private(engine._market_id, new_position.id, tm, new_position.profile_id, "position", new_position)
    else
        local err = _notify_extended_position(trader_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return nil
end

function engine._reject_order(order_id, reason)
    local order, err = o.reject(order_id, reason)
    if err ~= nil then
        return err
    end

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, "order", order)

    if util.is_value_in(order.order_type, position_dep_order_types) then
        local err = _notify_extended_position(order.profile_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return nil
end

function engine._create_order(
    order_id,
    trader_id,
    order_type,
    price,
    size,
    initial_size,
    side,
    client_order_id,
    trigger_price,
    size_percent,
    time_in_force,
    is_liquidation
)
    checks('string', 'number', 'string', '?decimal', '?decimal', '?decimal', 'string',
        '?string', '?decimal', '?decimal', 'string', 'boolean')

    local order, err = o.create(
        order_id,
        trader_id,
        engine._market_id,
        order_type,
        price,
        size,
        initial_size,
        side,
        client_order_id,
        trigger_price,
        size_percent,
        time_in_force,
        is_liquidation
    )
    if err ~= nil then
        log.error(EngineError:new(err))
        return nil, err
    end

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, "order", order)

    if util.is_value_in(order.order_type, position_dep_order_types) then
        local err = _notify_extended_position(order.profile_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return order, nil
end

local function _update_order(order_id, new_price, new_size)
    local price = tick.round_to_nearest_tick(new_price, engine._min_tick)
    local size = tick.round_to_nearest_tick(new_size, engine._min_order)

    local order, e = o.update(order_id, price, size)
    if e ~= nil then
        return e
    end

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, "order", order)

    if util.is_value_in(order.order_type, position_dep_order_types) then
        local err = _notify_extended_position(order.profile_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return nil
end

function engine._cancel_order(order_id)
    checks('string')

    local order, err = o.cancel(order_id)
    if err ~= nil then
        return err
    end

    local tm = time.now()
    notif.add_private(engine._market_id, tostring(order.id), tm, order.profile_id, 'order', order)

    if util.is_value_in(order.order_type, position_dep_order_types) then
        local err = _notify_extended_position(order.profile_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
    end

    return nil
end

function engine._create_entry(entry_id, side, trader_id, new_price, new_size, sequence)
    local price = tick.round_to_nearest_tick(new_price, engine._min_tick)
    local size = tick.round_to_nearest_tick(new_size, engine._min_order)

    local res, err = ob.create(
        entry_id,
        engine._market_id,
        trader_id,
        price,
        size,
        side
    )
    if err ~= nil then
        return EngineError:new(err)
    end

    ag.add_price_level(price, size, side)
    ag.add_order_notional(trader_id, price * size)

    _update_bid_ask(price, side, sequence)

    -- WE WILL request meta only for updated profiles
    notif.add_profile(trader_id)
    return nil
end

local function _update_entry(entry_id, side, price, new_size, sequence, trader_id)
    new_size = tick.round_to_nearest_tick(new_size, engine._min_order)

    local entry = ob.get(entry_id)
    if entry == nil then
        local err = EngineError:new("%s: cant be nil", ERR_INTEGRITY_ERROR)
        log.error({
            message = err:backtrace(),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return err
    end

    local diff = new_size - entry.size
    ag.add_price_level(price, diff, side)
    ag.add_order_notional(trader_id, price * diff)

    if new_size == 0 then
        ob.delete(entry_id)
    else
        local res, err = ob.update_size(entry_id, new_size)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
        if res == nil then
            log.error(EngineError:new("entry_id=%s not found", entry_id))
            return ERR_ORDERBOOK_ENTRY_NOT_FOUND
        end
    end

    local err = _update_order(entry_id, price, new_size)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    _update_bid_ask(price, side, sequence)

    -- WE WILL request meta only for updated profiles
    notif.add_profile(trader_id)
    return nil
end

local function _fill_wf3(
    fill_size,
    fill_price,
    insurance_side,
    insurance_id,
    trader_side,
    trader_id,
    trade_id_prefix,
    is_liquidation_insurance,
    is_liquidation_trader)

    local e

    local notion = fill_price * fill_size

    -- UPDATE POSITIONS for maker and taker ---
    if insurance_side == config.params.LONG then
        e = _update_position(insurance_id, insurance_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end

        e = _update_position(trader_id, trader_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end
    else
        e = _update_position(trader_id, trader_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end

        e = _update_position(insurance_id, insurance_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end
    end

    local tm = time.now()
    local trade_id = trade_id_prefix
    local trader_fill_id = trade.next_trade_id(engine._market_id)
    local insurance_fill_id = trade.next_trade_id(engine._market_id)

    local trader_fill, err = archiver.insert(box.space.fill, {
        trader_fill_id,
        trader_id,
        engine._market_id,
        trade_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        trader_side,

        true,
        ZERO,
        is_liquidation_trader,

        "", -- client_order_id is empty
    })
    if err ~= nil then
        return EngineError:new(err)
    end

    local insurance_fill, err = archiver.insert(box.space.fill, {
        insurance_fill_id,
        insurance_id,
        engine._market_id,
        trade_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        insurance_side,

        false,
        ZERO,
        is_liquidation_insurance,

        "", -- client_order_id is empty
    })
    if err ~= nil then
        return EngineError:new(err)
    end

    return nil
end

-- TODO: it seems just need to pass orders here
local function _fill(
    maker_client_order_id,
    taker_client_order_id,

    fill_size,
    fill_price,
    taker_side,
    taker_id,
    taker_order_id,
    maker_side,
    maker_id,
    maker_order_id,
    is_liquidation,
    maker_is_liquidation,
    sequence)

    if maker_client_order_id == nil or maker_client_order_id == box.NULL then
        maker_client_order_id = ''
    end

    if taker_client_order_id == nil or taker_client_order_id == box.NULL then
        taker_client_order_id = ''
    end

    local e

    -- Generate required fills and trades
    -- TODO: implement for production
    --[[
    local maker = box.space.profile_tier.index.primary:get(maker_id)
    if maker ~= nil then
        maker_tier = maker.tier
    end

    local taker = box.space.profile_tier.index.primary:get(taker_id)
    if taker ~= nil then
        taker_tier = taker.tier
    end
    --]]

    local makerTier = profile.which_tier(maker_id)
    local takerTier = profile.which_tier(taker_id)

    local notion = fill_price * fill_size
    local makerFee = notion * makerTier.maker_fee * -1
    local takerFee = notion * takerTier.taker_fee * -1

    -- UPDATE POSITIONS for maker and taker ---
    if taker_side == config.params.LONG then
        e = _update_position(taker_id, taker_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end

        e = _update_position(maker_id, maker_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end
    else
        e = _update_position(maker_id, maker_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end

        e = _update_position(taker_id, taker_side, fill_price, fill_size)
        if e ~= nil then
            return EngineError:new(e)
        end
    end

    local tm = time.now()
    local trade_id = trade.next_trade_id(engine._market_id)
    local maker_fill_id = trade.next_trade_id(engine._market_id)
    local taker_fill_id = trade.next_trade_id(engine._market_id)

    local trade_item, err = archiver.insert(box.space.trade, {
        trade_id,
        engine._market_id,
        tm,
        fill_price,
        fill_size,
        is_liquidation,
        taker_side,
    })
    if err ~= nil then
        return EngineError:new(err)
    end

    local maker_fill, err = archiver.insert(box.space.fill, {
        maker_fill_id,
        maker_id,
        engine._market_id,
        maker_order_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        maker_side,

        true,
        makerFee,
        maker_is_liquidation,  -- we inherit is_liquidation from maker order (using reason field for that)

        maker_client_order_id
    })

    if err ~= nil then
        return err
    end

    local taker_fill, err = archiver.insert(box.space.fill, {
        taker_fill_id,
        taker_id,
        engine._market_id,
        taker_order_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        taker_side,

        false,
        takerFee,
        is_liquidation,

        taker_client_order_id,
    })
    if err ~= nil then
        return err
    end

    -- Create updates
    notif.add_trade(engine._market_id, trade_item)
    notif.add_private(engine._market_id, tostring(taker_fill_id), tm, taker_id, "fill", taker_fill)
    notif.add_private(engine._market_id, tostring(maker_fill_id), tm, maker_id, "fill", maker_fill)

    -- GENERATE all candles for this trade on the fly
    candles.add_all_periods(fill_price, fill_size, tm)

    -- PAY maker and taker FEE
    if makerFee ~= 0 then
        local balanceUpdate = balance.pay_fee(maker_id, makerFee)
        if balanceUpdate['error'] ~= nil then
            err = "ERROR makerFee profile_pay_fee: " .. tostring(balanceUpdate['error'])
            return err
        end
    end

    if takerFee ~= 0 then
        local balanceUpdate = balance.pay_fee(taker_id, takerFee)
        if balanceUpdate['error'] ~= nil then
            err = "ERROR takerFee profile_pay_fee: " .. tostring(balanceUpdate['error'])
            return err
        end
    end


    -- UPDATE marketRT info
    local best_ask = ZERO
    local ask = box.space.orderbook.index.short:min({engine._market_id, config.params.SHORT})
    if ask ~= nil then
        best_ask = ask[d.entry_price]
    end

    local best_bid = ZERO
    local bid = box.space.orderbook.index.long:max({engine._market_id, config.params.LONG})
    if bid ~= nil then
        best_bid = bid[d.entry_price]
    end

    market.on_trade_update(engine._market_id,
                    best_bid,
                    best_ask,
                    fill_price,
                    notion,
                    fill_size,
                    sequence)

    profile.update_volume(engine._market_id, taker_id, maker_id, notion)

    return nil
end

function engine._match_order(side, price, size, order_id, trader_id, is_liquidation, sequence)
    local start_condition, end_condition, which_index, which_iter, which_cond
    local left_size = size
    local e

    if side == config.params.LONG then
        local ask = box.space.orderbook.index.short:min({engine._market_id, config.params.SHORT})
        start_condition = function()
            if ask ~= nil and price >= ask[d.entry_price] then
                return true
            end

            return false
        end

        end_condition = function(next_price)
            if next_price > price then
                return true
            end

            return false
        end

        which_index = "short"
        which_cond = {engine._market_id, config.params.SHORT}
        which_iter = {iterator='EQ'}
    else
        local bid = box.space.orderbook.index.long:max({engine._market_id, config.params.LONG})
        start_condition = function()
            if bid ~= nil and price <= bid[d.entry_price] then
                return true
            end

            return false
        end

        end_condition = function(next_price)
            if next_price < price then
                return true
            end

            return false
        end

        which_index = "long"
        which_cond = {engine._market_id, config.params.LONG}
        which_iter = {iterator='REQ'}
    end

    local taker_order = box.space.order:get(order_id)

    if start_condition() then
        for _, entry in box.space.orderbook.index[which_index]:pairs(which_cond, which_iter) do
            local fill_price = entry[d.entry_price]
            if end_condition(fill_price) then
                break
            end


            local maker_side = entry[d.entry_side]
            local maker_id = entry[d.entry_trader]
            local maker_order_id = entry[d.entry_id]

            -- we need to get order_id and check for liquidation
            local maker_is_liquidation = false
            local maker_order = box.space.order:get(maker_order_id)
            if maker_order ~= nil then
                if tostring(maker_order.reason) == "liquidation" then
                    maker_is_liquidation = true
                end
            end

            local fill_size
            if entry[d.entry_size] < left_size then -- PARTIAL match
                fill_size = entry[d.entry_size]


                left_size = left_size - fill_size

                e = _update_entry(entry[1], entry[d.entry_side], entry[d.entry_price], 0, sequence, entry[d.entry_trader])
                if e ~= nil then
                    return left_size, EngineError:new(e)
                end
            else -- FULL match
                fill_size = left_size

                local new_size = entry[d.entry_size] - fill_size
                e = _update_entry(entry[1], entry[d.entry_side], entry[d.entry_price], new_size, sequence, entry[d.entry_trader])
                if e ~= nil then
                    return left_size, e
                end

                left_size = ZERO
            end


            e = _fill(
                        maker_order.client_order_id,
                        taker_order.client_order_id,

                        fill_size,
                        fill_price,
                        side,       -- taker_side
                        trader_id,  -- taker_id
                        order_id,   -- taker_order_id
                        maker_side,
                        maker_id,
                        maker_order_id,
                        is_liquidation,
                        maker_is_liquidation,
                        sequence)

            if e ~= nil then
                return left_size, e
            end

            if left_size == 0 then
                break
            end
        end
    end

    if left_size ~= 0 then
        e = engine._create_entry(order_id, side, trader_id, price, left_size, sequence)
        if e ~= nil then
            return left_size, e
        end
    end

    e = _update_order(order_id, price, left_size)
    if e ~= nil then
        return left_size, e
    end

    return left_size, nil
end

local function _pre_create_order(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local order_mtype = engine._order_metatypes[order.order_type]
    if order_mtype ~= nil then
        return order_mtype.pre_create(order, position, market_data)
    end

    return ERR_WRONG_ORDER_TYPE
end

local function _pre_execute_order(order, position, market_data)
    checks('table|engine_order', '?table|engine_position', 'table|engine_market')

    local order_mtype = engine._order_metatypes[order.order_type]
    if order_mtype ~= nil then
        return order_mtype.pre_execute(order, position, market_data)
    end

    return ERR_WRONG_ORDER_TYPE
end

local function _post_execute_order(order, left_size, sequence)
    checks('table|engine_order', 'decimal', 'number')
    local err

    local order_mtype = engine._order_metatypes[order.order_type]
    if order_mtype == nil then
        return ERR_WRONG_ORDER_TYPE
    end

    err = order_mtype.post_execute(order, left_size, sequence)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    for _, profile in notif.changed_profiles_iterator() do
        local position = p.get_position(profile.profile_id, engine._market_id)

        for _, order_type in ipairs(position_dep_order_types) do
            for _, trader_order in o.iterator_by(profile.profile_id, config.params.ORDER_STATUS.PLACED, order_type) do
                if position == nil or position.size == 0 or position.side == trader_order.side then
                    local err = engine._cancel_order(trader_order.id)
                    if err ~= nil then
                        log.error(EngineError:new(err))
                        return err
                    end
                end
            end
        end
    end

    return nil
end

function engine._private_match(pm_counterparty, side, price, size, order_id, profile_id, market_data, is_liquidation, taker_client_order_id)
    checks('number', 'string', 'decimal', 'decimal', 'string', 'number', 'table|engine_market', 'boolean', "string")

    -- calc fill_price
    local cp_side = config.params.SHORT
    local fill_price = market_data.best_ask
    if side == config.params.SHORT then
        fill_price = market_data.best_bid
        cp_side = config.params.LONG
    end

    size = tick.round_to_nearest_tick(size, engine._min_order)

    local e
    local notion = fill_price * size

    -- UPDATE POSITION for maker ---
    e = _update_position(pm_counterparty, cp_side, fill_price, size)
    if e ~= nil then
        return EngineError:new(e)
    end

    -- UPDATE POSITION for taker ---
    e = _update_position(profile_id, side, fill_price, size)
    if e ~= nil then
        return EngineError:new(e)
    end

    local tm = time.now()
    local trade_id = trade.next_trade_id(engine._market_id)
    local maker_fill_id = trade.next_trade_id(engine._market_id)
    local taker_fill_id = trade.next_trade_id(engine._market_id)

    local makerTier = profile.which_tier(pm_counterparty)
    local takerTier = profile.which_tier(profile_id)

    local makerFee = notion * makerTier.maker_fee * -1
    local takerFee = notion * takerTier.taker_fee * -1


    local trade_item, err = archiver.insert(box.space.trade, {
        trade_id,
        engine._market_id,
        tm,
        fill_price,
        size,
        is_liquidation,
        side,
    })
    if err ~= nil then
        return EngineError:new(err)
    end

    local maker_fill, err = archiver.insert(box.space.fill, {
        maker_fill_id,
        pm_counterparty,
        engine._market_id,
        "pm",
        tm,

        trade_id,

        fill_price,
        size,
        cp_side,

        true,
        makerFee,
        is_liquidation,  -- we inherit is_liquidation from maker order (using reason field for that)

        ""
    })

    if err ~= nil then
        return err
    end

    local taker_fill, err = archiver.insert(box.space.fill, {
        taker_fill_id,
        profile_id,
        engine._market_id,
        order_id,
        tm,

        trade_id,

        fill_price,
        size,
        side,

        false,
        takerFee,
        is_liquidation,

        taker_client_order_id,
    })
    if err ~= nil then
        return err
    end

    err = _update_order(order_id, price, ZERO)
    if err ~= nil then
        return err
    end

    -- Create updates
    notif.add_private(engine._market_id, tostring(taker_fill_id), tm, profile_id, "fill", taker_fill)
    notif.add_private(engine._market_id, tostring(maker_fill_id), tm, pm_counterparty, "fill", maker_fill)

    -- GENERATE all candles for this trade on the fly
    candles.add_all_periods(fill_price, size, tm)

    -- PAY maker and taker FEE
    if makerFee ~= 0 then
        local balanceUpdate = balance.pay_fee(pm_counterparty, makerFee)
        if balanceUpdate['error'] ~= nil then
            err = "ERROR makerFee profile_pay_fee: " .. tostring(balanceUpdate['error'])
            return err
        end
    end

    if takerFee ~= 0 then
        local balanceUpdate = balance.pay_fee(profile_id, takerFee)
        if balanceUpdate['error'] ~= nil then
            err = "ERROR takerFee profile_pay_fee: " .. tostring(balanceUpdate['error'])
            return err
        end
    end
    


    -- UPDATE marketRT info
    local best_ask = market_data.best_ask
    if best_ask == nil or best_ask == ZERO then
        return "ZERO_ASK"
    end


    local best_bid = market_data.best_bid
    if best_bid == nil or best_bid == ZERO then
        return "ZERO_BID"
    end

    market.on_trade_update(engine._market_id,
                    best_bid,
                    best_ask,
                    fill_price,
                    notion,
                    size,
                    market_data.last_update_sequence)

    profile.update_volume(engine._market_id, profile_id, pm_counterparty, notion)

end


function engine._private_execute(is_liquidation, pm_counterparty, order, profile_data, position, market_data)
    checks('?boolean', 'number', 'table|engine_order', 'table', '?table|engine_position', 'table|engine_market')
    local err

    if is_liquidation == nil then
        is_liquidation = false
    end

    local cp_position_before = p.get_position(pm_counterparty, engine._market_id)
    local cp_profile_data = router.get_profile_data(pm_counterparty)
    if cp_profile_data == nil then
        local text = "NO_PROFILE_DATA for counterparty=" .. tostring(pm_counterparty)
        log.error(EngineError:new(text))
        return text
    end

    err = _pre_execute_order(order, position, market_data)
    if err ~= nil then
        if err == ERR_NO_CONDITION_MET then
            return nil
        end
        log.error(EngineError:new(err))
        return err
    end

    -- All or nothing by default, no partial fills possible
    err = engine._private_match(
        pm_counterparty,
        order.side,
        order.price,
        order.size,
        order.id,
        order.profile_id,
        market_data,
        is_liquidation,
        tostring(order.client_order_id)
    )
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    -- Post match for taker,
    err = risk.post_match(engine._market_id, profile_data, order.profile_id, position)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    -- Post match for maker (counterparty)
    err = risk.post_match(engine._market_id, cp_profile_data, pm_counterparty, cp_position_before)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end


    local _not_used_sequence = engine._current_sequence()
    err = _post_execute_order(order, ZERO, _not_used_sequence)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil

end 

local function _execute_order(order, is_liquidation, no_post_match, profile_data, position, sequence, market_data)
    checks('table|engine_order', 'boolean', 'boolean', 'table', '?table|engine_position', 'number', 'table|engine_market')
    local err

    err = _pre_execute_order(order, position, market_data)
    if err ~= nil then
        if err == ERR_NO_CONDITION_MET then
            return nil
        end
        log.error(EngineError:new(err))
        return err
    end

    local left_size
    left_size, err = engine._match_order(
        order.side,
        order.price,
        order.size,
        order.id,
        order.profile_id,
        is_liquidation,
        sequence
    )
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    if not no_post_match then
        err = risk.post_match(engine._market_id, profile_data, order.profile_id, position)
        if err ~= nil then
            log.error(EngineError:new(err))
            return err
        end
    end

    err = _post_execute_order(order, left_size, sequence)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    return nil
end

-- MAIN ENGINE METHODS ---------
function engine._handle_cancelall(order)
    checks('table|api_cancelall_order')
    local sequence = engine._next_sequence()

    local _, err = dml.atomic2(engine._rollback_with_sequence, function()
        local err

        err = profile.ensure_meta(order.profile_id, engine._market_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
        notif.add_profile(order.profile_id)

        for _, cond_order in o.iterator_by(order.profile_id, config.params.ORDER_STATUS.PLACED, nil) do
            err = engine._cancel_order(cond_order.id)
            if err ~= nil then
                log.error(EngineError:new(err))
                return nil, err
            end
        end

        for _, entry in ob.iterator_by_trader(order.profile_id) do
            err = _cancel_entry(entry, sequence)
            if err ~= nil then
                log.error(EngineError:new(err))
                return nil, err
            end
        end

        return nil, nil
    end)
    if err == nil then
        notif.notify(engine._market_id, sequence) -- notify through pub/sub
    end

    return err
end

function engine._handle_amend(amend, profile_data)
    checks('table|api_amend_order', 'table')

    local err, res
    res = market.get_market(engine._market_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return res.error
    end
    local market_data = res.res

    res = o.get_order_by_id(amend.order_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return nil, false, res.error
    end
    local order = res.res

    box.begin()
    err = profile.ensure_meta(order.profile_id, engine._market_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        box.rollback()
        return err
    end

    local order_mtype = engine._order_metatypes[order.order_type]
    if order_mtype == nil then
        log.error(EngineError:new("order handler not found: order_id=%s, order_type=%s", order.id, order.order_type))
        box.rollback()
        return ERR_WRONG_ORDER_TYPE
    end

    local sequence = engine._next_sequence()
    local entry = ob.get(order.id)
    local position_before = p.get_position(order.profile_id, engine._market_id)

    local need_execute = false
    order, need_execute, err = order_mtype.amend(amend, entry, order, position_before, sequence, market_data)
    if err ~= nil then
        log.error(EngineError:new(err))
        engine._rollback_with_sequence()
        return err
    end

    if need_execute then
        err = _execute_order(order, false, true, profile_data, position_before, sequence, market_data)
        if err ~= nil then
            log.error(EngineError:new(err))
            engine._rollback_with_sequence()
            return err
        end
    end

    err = order_mtype.post_amend(profile_data, order, position_before, market_data)
    if err ~= nil then
        log.error(EngineError:new(err))
        engine._rollback_with_sequence()
        return err
    end

    notif.add_profile(order.profile_id)

    box.commit()
    notif.notify(engine._market_id, sequence) -- notify through pub/sub

    return nil
end

function engine._handle_cancel(api_order)
    checks('table|api_cancel_order')

    local order
    local order_id = api_order.order_id
    local client_order_id = api_order.client_order_id

    -- if we cancel by client_order_id find first
    if order_id ~= nil and order_id ~= "" then
        local res = o.get_order_by_id(order_id)
        if res.error ~= nil then
            log.error(EngineError:new(res.error))
            return res.error
        end
        order = res.res
    else
        if client_order_id == nil or client_order_id == "" then
            local text = "order_id or client_order_id required"
            return text
        end

        local res = o.get_order_by_client_id(client_order_id)
        if res.error ~= nil then
            log.error(EngineError:new(res.error))
            local text = "client_order_id=" .. tostring(client_order_id) .. " NO OPEN orders FOUND"
            return text
        end
        order = res.res
    end

    local sequence = engine._next_sequence()

    local _, err = dml.atomic2(engine._rollback_with_sequence, function()
        local err

        err = profile.ensure_meta(order.profile_id, engine._market_id)
        if err ~= nil then
            log.error(EngineError:new(err))
            return nil, err
        end
        notif.add_profile(order.profile_id)

        if order.status == config.params.ORDER_STATUS.PLACED then
            if order.profile_id ~= api_order.profile_id then
                return nil, ERR_NOT_YOUR_ORDER
            end

            err = engine._cancel_order(order.id)
            if err ~= nil then
                log.error(EngineError:new(err))
                return nil, err
            end
        else
            local entry = ob.get(order.id)
            if entry == nil then
                return nil, ERR_ORDERBOOK_ENTRY_NOT_FOUND
            end
            if entry.trader_id ~= api_order.profile_id then
                return nil, ERR_NOT_YOUR_ORDER
            end

            err = _cancel_entry(entry, sequence)
            if err ~= nil then
                log.error(EngineError:new(err))
                return nil, err
            end
        end

        return nil, nil
    end)
    if err == nil then
        notif.notify(engine._market_id, sequence)
    end

    return err
end

function engine._handle_create(order, profile_data, matching_meta)
    checks('table|api_create_order', 'table', '?table')

    local err, res
    local ob_order = ob.get(order.order_id)
    if ob_order ~= nil then
        err = "exist order_id=" .. tostring(order.order_id)
        log.error(EngineError:new(err))
        return err
    end

    res = market.get_market(engine._market_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return res.error
    end
    local market_data = res.res

    local position_before = p.get_position(order.profile_id, engine._market_id)
    local pre_create_err = _pre_create_order(order, position_before, market_data)
    if  pre_create_err ~= nil then
        log.error(EngineError:new('%s: slog=%s', tostring(pre_create_err), json.encode(order)))
        --no return, create order anyway and reject it with reason on error
    end

    box.begin()
    err = profile.ensure_meta(order.profile_id, engine._market_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        box.rollback()
        return err
    end

    -- create order anyway and reject it with reason on error
    local new_order
    new_order, err = engine._create_order(
        order.order_id,
        order.profile_id,
        order.order_type,
        order.price,
        order.size,
        order.size,
        order.side,
        order.client_order_id,
        order.trigger_price,
        order.size_percent,
        order.time_in_force,
        order.is_liquidation
    )
    if err ~= nil then
        log.error(EngineError:new(err))
        box.rollback()
        return err
    end

    --TODO: needs refactoring to do it more consistent and organic
    -- because order must be stored anyway in engine (with reject reason)
    if  pre_create_err ~= nil then
        local rej_err = engine._reject_order(new_order.id, tostring(pre_create_err))
        if rej_err ~= nil then
            log.error(EngineError:new("reject order id=%s error: %s", new_order.id, EngineError:new(rej_err)))
        end

        box.commit()
        notif.notify_account(engine._market_id)
        return pre_create_err
    end

    notif.add_profile(new_order.profile_id)

    local svp_sequence = engine._current_sequence()
    local svp = box.savepoint()

    local pm_counterparty, cp_err = router.which_counterparty(order.is_liquidation, new_order, market_data, matching_meta)

    if cp_err ~= nil then
        log.info("PM_ERROR 5: cp_err=%s profile_id=%s pm_counterparty=%s", tostring(cp_err), tostring(order.profile_id), tostring(pm_counterparty))
    end

    if pm_counterparty ~= nil then
        err = profile.ensure_meta(pm_counterparty, engine._market_id)
        if err ~= nil then
            -- Something wrong - just move forward to public ob
            log.error(EngineError:new(err))
        else
            notif.add_profile(pm_counterparty)
            -- _private match doesn't inc the sequence
            err = engine._private_execute(order.is_liquidation, pm_counterparty, new_order, profile_data, position_before, market_data)
            if err ~= nil then
                log.error(EngineError:new(err))
                
                -- Something wrong - just revert and move forward to public orderbook
                engine._rollback_with_sequence(svp, svp_sequence)
            else
                box.commit()

                -- we don't send trade to public only to the user
                notif.notify_account(engine._market_id)
                return nil
            end
        end
    end


    if util.is_value_in(order.order_type, custom_execute_types) then
        local order_mtype = engine._order_metatypes[order.order_type]
        if order_mtype == nil or order_mtype.execute == nil then
            local err = EngineError:new(ERR_INTEGRITY_ERROR)
            log.error({
                message = err:backtrace(),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return err
        end

        local svp_sequence = engine._current_sequence()

        local notifications
        notifications, err = order_mtype.execute(new_order, order.is_liquidation, order.is_liquidation, profile_data, position_before)
        if err ~= nil then
            log.error(EngineError:new(err))

            engine._rollback_with_sequence(svp, svp_sequence)
            local rej_err = engine._reject_order(new_order.id, tostring(err))
            if rej_err ~= nil then
                log.error("reject order id=%s error: %s", order.id, EngineError:new(rej_err))
            end

            box.commit()
            notif.notify_account(engine._market_id)
            return err
        end

        box.commit()
        if notifications ~= nil then
            notif.notify_from_table(notifications, engine._market_id)
        end

        notif.notify_account(engine._market_id)
    else
        local sequence = engine._next_sequence()

        err = _execute_order(new_order, order.is_liquidation, order.is_liquidation, profile_data, position_before, sequence, market_data)
        if err ~= nil then
            log.error('execute order: %s', err)
            engine._rollback_with_sequence(svp)

            local rej_err = engine._reject_order(new_order.id, tostring(err))
            if rej_err ~= nil then
                log.error("reject order id=%s error: %s", new_order.id, EngineError:new(rej_err))
            end
            box.commit()
            notif.notify_account(engine._market_id)

            return err
        end

        box.commit()
        notif.notify(engine._market_id, sequence) -- notify through pub/sub
    end

    return nil
end

function engine._handle_execute(order_data, profile_data)
    checks('table|api_execute_order', 'table')
    local err, res

    res = market.get_market(engine._market_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return res.error
    end
    local market_data = res.res

    err = risk.check_market(market_data)
    if err ~= nil then
        log.error(EngineError:new(err))
        return err
    end

    res = o.get_order_by_id(order_data.order_id)
    if res.error ~= nil then
        log.error(EngineError:new(res.error))
        return res.error
    end
    local order = res.res
    if order.status ~= config.params.ORDER_STATUS.PLACED then
        return ERR_WRONG_ORDER_STATUS
    end

    local position = p.get_position(order.profile_id, engine._market_id)
    local sequence = engine._next_sequence()

    box.begin()
    notif.add_profile(order.profile_id)

    local is_liquidation = false
    err = _execute_order(order, is_liquidation, is_liquidation, profile_data, position, sequence, market_data)
    if err ~= nil then
        log.error(EngineError:new(err))
        engine._rollback_with_sequence()

        local rej_err = engine._reject_order(order.id, tostring(err))
        if rej_err ~= nil then
            log.error("reject order id=%s error: %s", order.id, EngineError:new(rej_err))
        end

        box.commit()
        notif.notify_account(engine._market_id)

        return err
    end

    box.commit()
    notif.notify(engine._market_id, sequence) -- notify through pub/sub

    return nil
end

function engine._handle_liquidate(task_data)
    checks('table|api_task')

    local liq_action = task_data.order
    local order_id = task_data.order_id
    if order_id == "" then
        return "BROKEN order_id"
    end

    -- INTEGRITY CHECK
    local market_id = liq_action[d.liq_action_market_id]
    if market_id ~= engine._market_id then
        local text = "WRONG_MARKET=" .. tostring(market_id) .. "for engine market_id=" .. tostring(engine._market_id)
        return text
    end

    local liquidate_kind = liq_action[d.liq_action_kind]

    notif.clear()   -- clear notifications

    local trader_id = liq_action[d.liq_action_trader_id]

    local res
    if liquidate_kind == config.params.LIQUIDATE_KIND.APLACESELLORDERS then
        local price = tick.round_to_nearest_tick(liq_action[d.liq_action_price], engine._min_tick)
        local size = tick.round_to_nearest_tick(liq_action[d.liq_action_size], engine._min_order)

        if price <= 0 or size <= 0 then
            local text = "BROKEN price=" .. tostring(price) .. " OR SIZE = " .. tostring(size)
            log.error(text)
            return text
        end

        local side = config.params.SHORT
        local positions = box.space.position.index.profile_id:select(trader_id, {iterator = "EQ"})

        if positions == nil or #positions == 0 then
            return "NO_POSITONS"
        end

        if positions[1][5] == side then
            side = config.params.LONG
        end

        local order, _ = action_creator.pack_create(
            liq_action[d.liq_action_trader_id],
            market_id,
            order_id,
            true,
            side,
            config.params.ORDER_TYPE.LIMIT,
            size,
            price,
            nil,
            nil,
            nil,
            config.params.TIME_IN_FORCE.GTC
        )

        res = engine._handle_create(order, {})
    elseif liquidate_kind == config.params.LIQUIDATE_KIND.AINSTAKEOVER then
        local price = liq_action[d.liq_action_price]
        local size = tick.round_to_nearest_tick(liq_action[d.liq_action_size], engine._min_order)

        if price <= 0 or size <= 0 then
            local text = "BROKEN price=" .. tostring(price) .. " OR SIZE = " .. tostring(size)
            log.error(text)
            return text
        end

        local insurance_id = profile.get_insurance_id()
        if insurance_id == nil then
            return "NO_INSURANCE_FOR_MARKET_" .. tostring(engine._market_id)
        end

        local positions = box.space.position.index.profile_id:select(trader_id, {iterator = "EQ"})

        if positions == nil or #positions == 0 then
            return "NO_POSITONS"
        end

        local trader_side = config.params.LONG
        local insurance_side = positions[1][5]
        if insurance_side == config.params.LONG then
            trader_side = config.params.SHORT
        end


        box.begin()
            notif.add_profile(trader_id)
            notif.add_profile(insurance_id)

            res = _fill_wf3(
                size,
                price,
                insurance_side,
                insurance_id,
                trader_side,
                trader_id,
                "wf3",
                true,
                true)
            if res ~= nil then
                box.rollback()
                return res
            end
        box.commit()

    elseif liquidate_kind == config.params.LIQUIDATE_KIND.AINSCLAWBACK then

        -- SKIP CLAWBACK order if market is active
        local cur_market = box.space.market:get(tostring(engine._market_id))
        if cur_market.status == config.params.MARKET_STATUS.ACTIVE then
            notif.clear()
            return nil
        end

        local price = liq_action[d.liq_action_price]
        local size =  liq_action[d.liq_action_size]

        if price <= 0 or size <= 0 then
            local text = "BROKEN price=" .. tostring(price) .. " OR SIZE = " .. tostring(size)
            log.error(text)
            return text
        end

        local insurance_id = profile.get_insurance_id()
        if insurance_id == nil then
            return "NO_INSURANCE_FOR_MARKET_" .. tostring(engine._market_id)
        end

        local positions = box.space.position.index.profile_id:select(trader_id, {iterator = "EQ"})

        if positions == nil or #positions == 0 then
            return "NO_POSITONS"
        end

        local trader_side = config.params.LONG
        local insurance_side = positions[1][5]
        if insurance_side == config.params.LONG then
            trader_side = config.params.SHORT
        end


        box.begin()
            notif.add_profile(trader_id)
            notif.add_profile(insurance_id)

            res = _fill_wf3(
                size,
                price,
                insurance_side,
                insurance_id,
                trader_side,
                trader_id,
                "clawback",
                false,
                false)
            if res ~= nil then
                box.rollback()
                return res
            end
        box.commit()

    else
        local text = "UNKNOWN_LIQUIDATE_KIND=" .. tostring(liquidate_kind)
        return text
    end

    _broadcast_changes()

    notif.clear()   -- clear notifications

    return res
end

function engine.handle_task(task)
    checks('table|api_task')

    local task_data = task[d.task_data]
    local which_action = task_data.action

    -- IF IT's liquidation, handle differenlty
    -- NO risk checks
    -- Private matching for some kinds
    if which_action == config.params.ORDER_ACTION.LIQUIDATE then
        return engine._handle_liquidate(task_data)
    end


    local order = task_data.order
    if order == nil then
        log.error("no order")
        return "no order"
    end

    if order.market_id ~= engine._market_id then
        local text = "wrong order.market_id=" .. tostring(order.market_id) .. " engine market_id=" .. tostring(engine._market_id)
        log.error(EngineError:new(text))
        return text
    end

    local res
    local ids = router.get_all_counterparties()
    table.insert(ids, order.profile_id)

    local profile_data, err = cm_get.handle_get_cache_and_meta(router, ids, engine._market_id, order.profile_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        return tostring(err)
    end
    
    if which_action == nil or which_action == "" then
        log.error(EngineError:new("NO_ACTION"))
        return "NO_ACTION"
    end

    notif.clear()   -- clear notifications

    if which_action == config.params.ORDER_ACTION.CREATE then
        res = engine._handle_create(order, profile_data, task_data.matching_meta)
    elseif which_action == config.params.ORDER_ACTION.CANCEL then
        res = engine._handle_cancel(order)
    elseif which_action == config.params.ORDER_ACTION.AMEND then
        res = engine._handle_amend(order, profile_data)
    elseif which_action == config.params.ORDER_ACTION.EXECUTE then
        res = engine._handle_execute(order, profile_data)
    elseif which_action == config.params.ORDER_ACTION.CANCELALL then
        res = engine._handle_cancelall(order)

        if res == "NO_ORDERS" then
            -- JUST DO nothing
            notif.clear()

            return nil
        end
    else
        local text = "UNKNOWN action " .. tostring(which_action)
        log.error(EngineError:new(text))
        return text
    end


    if res ~= nil then
        if order ~= nil then
            notif.send_profile_notification_with_order(order.profile_id, config.params.NOTIF_TYPE.NOTIF_ERROR, "order execution error", tostring(res), order)
        else
            notif.send_profile_notification(order.profile_id, config.params.NOTIF_TYPE.NOTIF_ERROR, "order execution error", tostring(res))
        end
    else
        notif.send_profile_notification(order.profile_id, config.params.NOTIF_TYPE.NOTIF_SUCCESS, "order placed success", "")
    end

    _broadcast_changes()

    notif.clear() -- clear all


    return res
end

function engine.get_orderbook_data(market_id)
    local update = {
        market_id,    -- market_id
        {},           -- bids
        {},           -- asks
        0,     -- sequence
        time.now()    -- timestamp
    }

    local sequence
    box.begin()
        sequence = box.sequence.orderbook_sequence:current()

        for _, bid in box.space.bids_to_size:pairs() do
            table.insert(update[2], bid)
        end

        for _, ask in box.space.asks_to_size:pairs() do
            table.insert(update[3], ask)
        end
    box.commit()
    update[4] = sequence

    return {res = update, error = nil}
end

-- ONLY FOR admin usage
-- Received original_fill as input (has profile_id, and taker_profile_id)
-- Will generate reverted trade
function engine.handle_revert(raw_fill)
    checks('table')

    local tup = revert.bind(raw_fill)
    if tup == nil then
        return {res = nil, error = ERR_REVERT_CANT_BIND}
    end

    local market_id = tup.market_id
    if market_id ~= engine._market_id then
        local text = "wrong original_fill.market_id=" .. tostring(market_id) .. " engine market_id=" .. tostring(engine._market_id)
        return {res = nil, error = text}
    end

    local maker_id = tup.profile_id
    local taker_id = tup.taker_profile_id

   local  err = profile.ensure_meta(maker_id, engine._market_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        return {res = nil, error = ERR_INTEGRITY_ERROR}
    end

    err = profile.ensure_meta(taker_id, engine._market_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        return {res = nil, error = ERR_INTEGRITY_ERROR}
    end

    local maker_data, taker_data

    maker_data, err = cm_get.handle_get_cache_and_meta(router, {maker_id}, engine._market_id, maker_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        return tostring(err)
    end

    taker_data, err = cm_get.handle_get_cache_and_meta(router, {taker_id}, engine._market_id, taker_id)
    if err ~= nil then
        log.error(EngineError:new(err))
        return tostring(err)
    end

    box.begin()

    local fill_price = tup.price
    local fill_size = tup.size
    local notion = fill_price * fill_size


    local taker_side = tup.side
    local maker_side = config.params.LONG
    if taker_side == config.params.LONG then
        maker_side = config.params.SHORT
    end

    local maker_position_before = p.get_position(maker_id, engine._market_id)
    err = _update_position(maker_id, maker_side, fill_price, fill_size)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end

    err = risk.post_match(engine._market_id, maker_data, maker_id, maker_position_before)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end

    local taker_position_before = p.get_position(taker_id, engine._market_id)
    err = _update_position(taker_id, taker_side, fill_price, fill_size)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end

    err = risk.post_match(engine._market_id, taker_data, taker_id, taker_position_before)
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end


    local tm = time.now()
    local trade_id = "revert"
    local maker_fill_id = trade.next_trade_id(engine._market_id)
    local taker_fill_id = trade.next_trade_id(engine._market_id)

    local maker_fill, err = archiver.insert(box.space.fill, {
        maker_fill_id,
        maker_id,
        engine._market_id,
        trade_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        maker_side,

        true,
        ZERO,
        false,

        "", -- client_order_id is empty
    })
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end

    local taker_fill, err = archiver.insert(box.space.fill, {
        taker_fill_id,
        taker_id,
        engine._market_id,
        trade_id,
        tm,

        trade_id,

        fill_price,
        fill_size,
        taker_side,

        false,
        ZERO,
        false,

        "", -- client_order_id is empty
    })
    if err ~= nil then
        box.rollback()
        return {res = nil, error = tostring(err)}
    end

    box.commit()

    local fills = {
        maker_fill = maker_fill,
        taker_fill = taker_fill
    }

    return {res = fills, error = nil}
end

local orphan_loop
local first_time = true
local task_id = nil
function engine._start_loop()
    engine._loop =
        fiber.new(
            function(instance)
                fiber.name("engine-main-loop")
                log.info("***** _start_loop for market_id=%s", instance._market_id)

                -- remove dead fiber from fibers info
                if orphan_loop then
                    orphan_loop:join()
                end

                while true do
                    fiber.testcancel()

                    -- TODO: wait for role
                    local res = rpc.wait_for_role("gateway")
                    if res.error ~= nil then
                        log.error(res.error)
                        fiber.sleep(1)
                    else
                        if first_time == true then
                            log.info("********** Success connect to gateway - waiting for tasks %s", instance._market_id)
                            first_time = false
                        end

                        local s, v = pcall(function()
                            res = rpc.callrw_gateway("next_task", {{market_id=instance._market_id, queue_type=config.sys.QUEUE_TYPE.MARKET, task_id=task_id}})
                        end)

                        if s == false then
                            log.error({
                                message = string.format('ENGINE_SYSTEM_ERROR: rpc next_task: %s', v),
                                [ALERT_TAG] = ALERT_CRIT,
                            })
                            fiber.sleep(1)
                        else

                            if res == nil then
                                log.error({
                                    message = string.format('UNKNOWN_ERROR market_id=%s', instance._market_id),
                                    [ALERT_TAG] = ALERT_CRIT,
                                })
                            elseif res.res == nil then
                                --log.info("no task")
                            elseif res["error"] ~= nil then
                                log.error({
                                    message = string.format('BROKEN_TASK: market_id=%s error=%s', instance._market_id, res.error),
                                    [ALERT_TAG] = ALERT_CRIT,
                                })
                                fiber.sleep(1)
                            else
                                local task = res["res"]
                                task_id = task[d.task_id]

                                local status, val = xpcall(function()
                                    return instance.handle_task(task)
                                end, function(err)
                                    -- if err is box.error object then debug.traceback won't return traceback
                                    return debug.traceback(tostring(err))
                                end)

                                if status == false then
                                    log.error({
                                        message = string.format('ENGINE_SYSTEM_ERROR: %s. task=%s', val, util.tostring(res)),
                                        [ALERT_TAG] = ALERT_CRIT,
                                    })
                                    -- fiber transaction can be in aborted state here, new fiber is needed with new tx
                                    fiber.testcancel()
                                    orphan_loop = fiber.self()
                                    engine._start_loop()
                                    return
                                else
                                    if val ~= nil then
                                        log.error('ENGINE_APP_ERROR: %s', val)
                                    end
                                end
                                local c = metrics.counter('rabbitx_exec_order_counter')
                                c:inc(1)
                            end
                        end
                       -- fiber.yield()
                    end
                end
            end,
        engine
    )

    engine._loop:set_joinable(true)
end

--[[
type FundingPayment struct {
	MarketId      string
	ProfileId     uint
	FundingAmount *tdecimal.Decimal
}
--]]
function engine.pay_funding(market_id, funding_payments, last_update_time, total_long, total_short)
    last_update_time = tonumber(last_update_time)
    checks('string', 'table', 'number', 'decimal', 'decimal')

    local changed_ids = {}

    local insideTx = box.is_in_txn()
    if insideTx == false then
        box.begin()
    end

    local count = 0
    for _, f_payment in ipairs(funding_payments) do
        local p_id = f_payment[2]
        local f_amount = f_payment[3]

        local balanceUpdate = balance.pay_funding(p_id, f_amount)
        if balanceUpdate['error'] ~= nil then
            local text = "ERROR profile_pay_funding: " .. tostring(balanceUpdate['error'])
            log.error(text)
            box.rollback()
            return {res = nil, error=text}
        end

        table.insert(changed_ids, p_id)

        -- WE don't do yield here: WAL log warn is ok
        -- CUZ we need to guarantee stability of the funding
        -- count = util.safe_yield(count, 0)
    end


    local status, res = pcall(function() return box.space.funding_meta:update(market_id, {
        {'=' , 'last_update', last_update_time},
        {'=' , 'total_long', total_long},
        {'=', 'total_short', total_short}
    }) end)

    if status == false then
        local text = "ERROR update funding_meta for market_id=" .. tostring(market_id) .. " err = %s" .. tostring(res)
        log.error(text)
        box.rollback()
        return {res = nil, error=text}
    end

    if insideTx == false then
        box.commit()
    end

    if #changed_ids > 0 then
        periodics.update_profiles(changed_ids)
    end
    changed_ids = nil

    return {res = nil, error = nil}
end

function engine.init(market_id, min_tick, min_order)
    checks('string', 'decimal', 'decimal')
    engine._market_id = market_id
    engine._min_tick = min_tick
    engine._min_order = min_order
end

function engine.start(market_id, min_tick, min_order)
    checks('string', 'decimal', 'decimal')
    engine.init(market_id, min_tick, min_order)

    if box.info.ro == true then
        return EngineError:new("engine.start on read-only instance")
    end

    if engine._loop ~= nil then
        return EngineError:new("engine.loop already started")
    end

    engine._start_loop()
    periodics.start(engine._market_id, engine._min_tick, engine._min_order)

    return nil
end


function engine.stop()
    periodics.killall()

    if engine._loop ~= nil and engine._loop:status() == "running" then
        engine._loop:cancel()

        local status, res = engine._loop:join()
        if status == false then
            return EngineError:new("loop join error=%s", res)
        end

        engine._loop = nil
    end

    return nil
end


function engine._test_set_rpc(override_rpc)
    rpc = override_rpc
end

function engine._test_update_position(trader_id, entry_side, fill_price, fill_size)
    return _update_position(trader_id, entry_side, fill_price, fill_size)
end

function engine._test_set_time(override_time)
    time = override_time
end

function engine._test_set_post_match(override_match)
    risk.post_match = override_match
end


return engine
