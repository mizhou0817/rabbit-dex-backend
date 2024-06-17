local checks = require('checks')
local decimal = require('decimal')
local log = require('log')

local archiver = require('app.archiver')
local errors = require('app.lib.errors')
local tick = require("app.lib.tick")
local time = require('app.lib.time')
local rolling = require('app.rolling')
local tuple = require('app.tuple')
local tick = require('app.lib.tick')

require("app.config.constants")

local EngineError = errors.new_class("ENGINE_ERROR")

local M = {
    format = {
        {name = 'id', type = 'string'},
        {name = 'status', type = 'string'},

        {name = 'min_initial_margin', type = 'decimal'},
        {name = 'forced_margin', type = 'decimal'},
        {name = 'liquidation_margin', type = 'decimal'},
        {name = 'min_tick', type = 'decimal'},
        {name = 'min_order', type = 'decimal'},

        {name = 'best_bid', type = 'decimal'},
        {name = 'best_ask', type = 'decimal'},
        {name = 'market_price', type = 'decimal'},
        {name = 'index_price', type = 'decimal'},
        {name = 'last_trade_price', type = 'decimal'},
        {name = 'fair_price', type = 'decimal'},
        {name = 'instant_funding_rate', type = 'decimal'},
        {name = 'last_funding_rate_basis', type = 'decimal'},

        {name = 'last_update_time', type = 'number'},
        {name = 'last_update_sequence', type = 'number'},
        {name = 'average_daily_volume_q', type = 'decimal'},
        {name = 'last_funding_update_time', type = 'number'},

        {name = 'icon_url', type = 'string'},
        {name = 'market_title', type = 'string'},
    },
    strict_type = 'engine_market',
}

local function new_market()
    return {
        --TODO: add market records creation
    }
end

function M.init_spaces(market_config)
    rolling.init_spaces()

    -- CONSTANTS that rare change and status
    local market, err = archiver.create('market', {if_not_exists = true}, M.format, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        error(err)
    end

    -- LAST STEP: create market here
    local exist = box.space.market:get(market_config.id)
    if exist ~= nil then
        log.warn("*** market exist %s", market_config.id)

        -- UPDATE dynamic params only
        local n, err = archiver.update(box.space.market, market_config.id, {
            {'=' , 'min_initial_margin', market_config.min_initial_margin},
            {'=' , 'forced_margin', market_config.forced_margin},
            {'=' , 'liquidation_margin', market_config.liquidation_margin}
        })
        if err ~= nil then
            log.error(EngineError:new("**** can't update params error=%s", err))
            return false
        end

        return true
    end

    --[[
        params.MARKETS.BTCUSDT,
        params.MARKET_STATUS.ACTIVE,

        default_min_initial_margin,
        default_forced_margin,
        default_liquidation_margin,

        1.0,        -- min_tick
        0.0001,     -- min_order
        default_adv_constant
    --]]
    local z = decimal.new(0)
    local res, err = archiver.insert(box.space.market, {
        market_config.id,
        market_config.status,

        market_config.min_initial_margin,
        market_config.forced_margin,
        market_config.liquidation_margin,
        market_config.min_tick,
        market_config.min_order,

        z,z,z,z,z,z,z,z,

        0, 0, z, 0,

        "", "",     -- icon_url, market_title
    })
    if err ~= nil then
        log.error(EngineError:new("**** can't create market error=%s", err))
        return false
    end

    -- CREATE FUNDING_META
    local funding_meta = box.schema.space.create("funding_meta", {if_not_exists = true})
    -- CONSTANTS that rare change and status
    funding_meta:format({
        {name = 'id', type = 'string'},
        {name = 'last_update', type = 'number'},
        {name = 'total_long', type = 'decimal'},
        {name = 'total_short', type = 'decimal'}
    })

    funding_meta:create_index('primary', {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true })

    local status
    status, res = pcall(function() return box.space.funding_meta:insert{
        market_config.id,
        0,
        z,
        z
    } end)
    if status == false then
        log.error(EngineError:new("**** can't create funding_meta error=%s", res))
        return false
    end

    log.warn("*** SUCCESS CREATE MARKET %s", market_config.id)
    return true
end

function M.bind(market)
    return tuple.new(market, M.format, M.strict_type)
end

-- tuple behaviour
function M.tomap(self, opts)
    return self:tomap(opts)
end

function M.get_market(market_id)
    local market = box.space.market:get(market_id)
    if market == nil then
        local text = "MARKET_NOT_FOUND " .. tostring(market_id)
        log.error(EngineError:new(text))
        return {res=nil, error=text}
    end

    return {res = M.bind(market), error = nil}
end

function M.on_trade_update(market_id, best_bid, best_ask, last_trade_price, last_trade_volume, last_trade_size, sequence)
    local market = box.space.market:get(market_id)
    local new_market_price = market.market_price

    if best_bid ~= 0 and best_ask ~= 0 then
        new_market_price = (best_bid + best_ask) / 2
    elseif best_bid ~= 0 then
        new_market_price = best_bid
    elseif best_ask ~= 0 then
        new_market_price = best_ask
    end

    local e = M.update_roll_value("24h_trade_q", 
            market_id, 
            last_trade_size, 
            3600,   -- we aggregate it per hour 
            24, -- for 24 hours  
            false) -- we sum the value
    if e ~= nil then
        log.warn("on_trade_update update_roll_value 24h_trade_q error=%s for market_id=%s", e, market_id)
    end
    local average_daily_volume_q = M.get_roll_sum("24h_trade_q", market_id)

    local last_update = time.now()
    local _, err = archiver.update(box.space.market, market_id, {
        {'=' , 'last_update_time', last_update},
        {'=' , 'last_update_sequence', sequence},
        {'=', 'best_bid', best_bid},
        {'=', 'best_ask', best_ask},
        {'=', 'market_price', new_market_price},
        {'=', 'last_trade_price', last_trade_price},
        {'=', 'average_daily_volume_q', average_daily_volume_q},
    })
    if err ~= nil then
        log.warn("on_trade_update: update market space: error=%s for market_id=%s", err, market_id)
    end
end

function M.bid_ask_update(market_id, best_bid, best_ask, sequence)
    local market = box.space.market:get(market_id)
    local new_market_price = market.market_price

    if best_bid ~= 0 and best_ask ~= 0 then
        new_market_price = (best_bid + best_ask) / 2
    elseif best_bid ~= 0 then
        new_market_price = best_bid
    elseif best_ask ~= 0 then
        new_market_price = best_ask
    end    
    local last_update = time.now()

    if market.best_bid ~= best_bid or 
        market.best_ask ~= best_ask or 
        market.market_price ~= new_market_price then

        local _, err = archiver.update(box.space.market, market_id, {
            {'=' , 'last_update_time', last_update},
            {'=' , 'last_update_sequence', sequence},
            {'=', 'best_bid', best_bid},
            {'=', 'best_ask', best_ask},
            {'=', 'market_price', new_market_price},
        })
        if err ~= nil then
            log.warn("bid_ask_update: update market space: error=%s for market_id=%s", err, market_id)
        end
    end
end

function M.update_funding(market_id, instant_funding_rate, new_funding_rate, new_funding_update_time, need_update_funding)
    checks("string", "decimal", "decimal", "number", "boolean")

    local market = box.space.market:get(market_id)
    if market == nil then
        return "market_not_found"
    end

    local last_funding_rate = market.last_funding_rate_basis
    local last_funding_update_time = market.last_funding_update_time
    if need_update_funding == true then
        last_funding_rate = new_funding_rate
        last_funding_update_time = new_funding_update_time
    end

    local _, err = archiver.update(box.space.market, market_id, {
        {'=', 'instant_funding_rate', instant_funding_rate},
        {'=', 'last_funding_rate_basis', last_funding_rate},
        {'=', 'last_funding_update_time', last_funding_update_time},
    })
    if err ~= nil then
        log.error(EngineError:new("can't update funding error=%s", err))
        return err
    end

    return nil
end

function M.update_fair_price(market_id, fair_price)
    checks("string", "decimal")

    if fair_price <= 0 then
        return "zero_or_negative_fair_price"
    end

    local market = box.space.market:get(market_id)
    if market == nil then
        return "market_not_found"
    end

    fair_price = tick.round_to_nearest_tick(fair_price, market.min_tick)

    local _, err = archiver.update(box.space.market, market_id, {
        {'=', 'fair_price', fair_price},
    })
    if err ~= nil then
        log.error(EngineError:new("can't update fair_price error=%s", res))
        return err
    end

    return nil
end

function M.update_index_price(market_id, index_price)
    checks("string", "decimal")

    if index_price <= 0 then
        return {res=nil, error="zero_or_negative_index_price"}
    end

    local market = box.space.market:get(market_id)
    if market == nil then
        return "market_not_found"
    end

    index_price = tick.round_to_nearest_tick(index_price, market.min_tick)
    if index_price <= 0 then
        return {res=nil, error="zero_or_negative_index_price_after_tick_rounding"}
    end

    local update_ops = {{'=', 'index_price', index_price}}

    local market = box.space.market:get(market_id)
    if market.fair_price == 0 then
        update_ops = {{'=', 'index_price', index_price}, {'=', 'fair_price', index_price}}
    end

    local _, err = archiver.update(box.space.market, market_id, update_ops)
    if err ~= nil then
        log.error(EngineError:new("can't update index_price error=%s", err))
        return {res=nil,  error=err}
    end

    return {res=nil, error=nil}
end

function M.update_roll_value(title, market_id, new_value, period_sec, max_values, is_replace)
    return rolling.update_roll_value(title, market_id, new_value, period_sec, max_values, is_replace)
end

function M.get_roll_avg(title, market_id)
    return rolling.get_roll_avg(title, market_id)
end

function M.reset_roll_value(title, market_id)
    return rolling.reset_roll_value(title, market_id)
end

function M.get_roll_sum(title, market_id)
    return rolling.get_roll_sum(title, market_id)
end

function M.diff_roll_value(title, market_id)
    return rolling.diff_roll_value(title, market_id)
end

function M.get_period_min_max(title, market_id, current)
    return rolling.get_period_min_max(title, market_id, current)
end

function M.get_mid_price(market_id)
    checks("string")

    local market = box.space.market:get(market_id)
    if market == nil then
        return {res = nil, error = ERR_MARKET_NOT_FOUND}
    end

    if market.best_bid <= 0 then
        return {res = nil, error = ERR_BEST_BID_ZERO}
    end

    if market.best_ask <= 0 then
        return {res = nil, error = ERR_BEST_ASK_ZERO}
    end

    local mid_price = tick.calc_mid_price(market.best_ask, market.min_tick)

    if not tick.is_mid_price_valid(mid_price, market.best_ask, market.best_bid) then
        return {res = nil, error = ERR_MID_PRICE_NOT_POSSIBLE}    
    end
    
    return {res = mid_price, error = nil}
end


function M.get_market_data(market_id)
    checks("string")

    local market = box.space.market:get(market_id)
    if market == nil then
        return {res = nil, error = ERR_MARKET_NOT_FOUND}
    end

    local res = market:tomap({names_only=true})

    return {res = res, error = nil}
end

function M.get_funding_meta(market_id)
    checks("string")

    local meta = box.space.funding_meta:get(market_id)
    if meta == nil then
        return {res = nil, error = "FUNDING_META_NOT_FOUND"}
    end

    local res = meta:tomap({names_only=true})

    return {res = res, error = nil}

end

function M.change_status(market_id, new_status)
    checks("string", "string")

    local _, err = archiver.update(box.space.market, market_id, {
        {'=' , 'status', new_status},
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        return {res = false, error = err}
    end

    return {res = true, error = nil}
end

function M.update_icon_url(market_id, new_url)
    checks("string", "string")

    local res, err = archiver.update(box.space.market, market_id, {
        {'=' , 'icon_url', new_url},
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

function M.update_market_title(market_id, new_title)
    checks("string", "string")

    local res, err = archiver.update(box.space.market, market_id, {
        {'=' , 'market_title', new_title},
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

return M
