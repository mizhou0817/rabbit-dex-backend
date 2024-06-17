local market = require('app.engine.market')
local config = require('app.config')
local log      = require('log')
local o = require('app.engine.order')
local tick = require('app.lib.tick')
local decimal = require('decimal')
local time = require('app.lib.time')
local checks = require('checks')

require('app.lib.table')

local fortest = {}

local function _is_allowed()
    if config.sys.MODE == "sync" then
        return true
    end

    return false
end

function fortest.init_spaces()
    local ft = box.schema.space.create('ft', {if_not_exists = true})

    ft:format({
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'}
    })

    ft:create_index('primary', {
        unique = true,
        parts = {{field = 'price'}},
        if_not_exists = true })

    return true
end

function fortest.update_fair_price(market_id, fair_price)
    checks("string", "decimal")
    
    local res = market.update_fair_price(market_id, fair_price)
    if res ~= nil then
        return {res = nil, error = res}
    end

    return {res = fair_price, error = nil}
end    

function fortest.get_fills()
    local res = box.space.fill:select(nil, {iterator = 'ALL', fullscan=true})

    return {res = res, error = nil}
end

function fortest.get_trades()
    local res = box.space.trade:select(nil, {iterator = 'ALL', fullscan=true})

    return {res = res, error = nil}
end

function fortest.get_order_by_id(order_id)
    return o.get_order_by_id(order_id)
end

function fortest.create_order(price, size, n_tick)
    log.info("...in params")
    log.info(price)
    log.info(size)
    log.info(n_tick)

    local d_price = tick.round_to_nearest_tick(price, n_tick)
    local d_size = tick.round_to_nearest_tick(size, n_tick)

    log.info("is_decimal d_price, price")
    log.info(decimal.is_decimal(d_price))
    log.info(decimal.is_decimal(price))

    log.info("d_price == price")
    log.info(d_price == price)

    log.info(d_price)
    log.info(d_price - price)

    log.info(d_size)

    local i = tick.round_to_nearest_tick(price, n_tick)
    log.info(i)
    log.info(i - price)

    local tm = time.now()

    box.space.ft:insert({
        tm,
        d_price,
        d_size
    })

    local res = box.space.ft:get(tm)
    
    return {res = res, error = nil}
end

function fortest.test_round(d_num, d_tick)   
    checks("decimal", "decimal")
   
   
    local d_round = tick.round_to_nearest_tick(d_num, d_tick)
   
    log.info("****")
    log.info(d_round)

    return {res = d_round, error = nil}
end

function fortest.create_value(price, size)   
    checks("decimal", "decimal")
   
    log.info(price)
    log.info(size)

    price = tick.round_to_nearest_tick(price, 1)
    box.space.ft:insert({
        price,
        size
    })

    local one = box.space.ft:get(price)
    
    return {res = one, error = nil}

    --[[
    box.space.ft:insert({
        price+1,
        size
    })

    local res = {
        market_id = "123",
        bids = {}
    }
    for _, t in box.space.ft:pairs() do
        table.insert(res.bids, t)
    end

    return {res = res, error = nil}
    --]]
end

function fortest.pcall_test()   

    local r1, e1 = pcall(function() 
        return "qwe"
    end)

    log.info("res1:")
    log.info(r1)
    log.info(e1)

    local r2, e2 = pcall(function() 
        local e = 1 / gg
    end)

    log.info("res2:")
    log.info(r2)
    log.info(e2)

    return {res="11", error=nil}
end

function fortest.get_all_orders()
    return o.get_all_orders2()
end

return fortest
