local log = require('log')

local config = require('app.config')
local errors = require('app.lib.errors')

require("app.config.constants")

local err = errors.new_class("aggregate_error")

local ag = {}

function ag.init_spaces()
    local bids_to_size = box.schema.space.create('bids_to_size', {if_not_exists = true})
    bids_to_size:format({
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
    })
    bids_to_size:create_index('primary', {
        unique = true,
        parts = {{field = 'price'}},
        if_not_exists = true })


    local asks_to_size = box.schema.space.create('asks_to_size', {if_not_exists = true})
    asks_to_size:format({
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
    })
    asks_to_size:create_index('primary', {
        unique = true,
        parts = {{field = 'price'}},
        if_not_exists = true })


    local trader_order_to_notional = box.schema.space.create('trader_order_to_notional', {if_not_exists = true})
    trader_order_to_notional:format({
        {name = 'trader_id', type = 'unsigned'},
        {name = 'notional', type = 'decimal'},
    })
    trader_order_to_notional:create_index('primary', {
        unique = true,
        parts = {{field = 'trader_id'}},
        if_not_exists = true })

end

function ag.get_size(price, side)
    local size = ZERO
    
    local which_space = "bids_to_size"
    if side == config.params.SHORT then
        which_space = "asks_to_size"
    end

    local res = box.space[which_space]:get(price)
    if res ~= nil then
        size = res.size
    end

    return size
end

function ag.add_price_level(price, size, side)
    local new_item = {
        price,
        size
    }

    local which_space = "bids_to_size"
    if side == config.params.SHORT then
        which_space = "asks_to_size"
    end

    local status, res = pcall(function() return 
        box.space[which_space]:upsert(new_item, {
            {'+', "size", size}
        })
    end)
    if status == false then
        log.error(err:new("can't add_price_level error=%s", res))
        return res
    end

    local created = box.space[which_space]:get(price)
    if created.size == 0 then
        box.space[which_space]:delete(price)
    end

    return nil
end


function ag.get_order_total_notional(trader_id)
    local notional = ZERO
    
    local res = box.space.trader_order_to_notional:get(trader_id)
    if res ~= nil then
        notional = res.notional
    end

    return notional
end


function ag.add_order_notional(trader_id, notional)
    local new_item = {
        trader_id,
        notional
    }

    local status, res = pcall(function() return 
        box.space.trader_order_to_notional:upsert(new_item, {
            {'+', "notional", notional}
        })
    end)
    if status == false then
        log.error(err:new("can't add_order_notional error=%s", res))
        return res
    end

    local created = box.space.trader_order_to_notional:get(trader_id)
    if created.notional == 0 then
        box.space.trader_order_to_notional:delete(trader_id)
    end

    return nil
end

return ag