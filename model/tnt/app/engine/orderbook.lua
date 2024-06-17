local checks = require('checks')
local log = require('log')

local archiver = require('app.archiver')
local profile = require('app.engine.profile')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local tuple = require('app.tuple')

local EngineError = errors.new_class("ENGINE_ERROR")

local OB = {
    format = {
        {name = 'order_id', type = 'string'},
        {name = 'timestamp', type = 'number'},

        {name = 'market_id', type = 'string'},
        {name = 'trader_id', type = 'unsigned'},

        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
        {name = 'side', type = 'string'},
        {name = 'reverse', type = 'number'},
    },
    strict_type = 'engine_ob_entry',
}

-- func should create tuple with exact described O.format
local function new_entry(
    order_id,
    timestamp,
    market_id,
    trader_id,
    price,
    size,
    side
)
    checks('string', 'number', 'string', 'number', 'decimal', 'decimal', 'string')

    return {
        order_id,
        timestamp,
        market_id,
        trader_id,
        price,
        size,
        side,
        -timestamp,
    }
end

function OB.bind(value)
    return tuple.new(value, OB.format, OB.strict_type)
end

-- tuple behaviour
function OB.tomap(self, opts)
    return self:tomap(opts)
end

function OB.init_spaces(market_data)
    checks('table')

    if box.sequence.orderbook_sequence == nil then
        box.schema.sequence.create('orderbook_sequence', {start = 0, min = 0, if_not_exists = true})
        -- sequence need to be started
        box.sequence.orderbook_sequence:next()
    end
    OB.sequence = box.sequence.orderbook_sequence

    local orderbook, err = archiver.create('orderbook', {if_not_exists = true}, OB.format, {
        unique = true,
        parts = {{field = 'order_id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(EngineError:new(err))
        error(err)
    end

    orderbook:create_index('market_id', {
        unique = false,
        parts = {{field = 'market_id'}},
        if_not_exists = true })

    orderbook:create_index('by_trader_id', {
        unique = false,
        parts = {{field = 'trader_id'}},
        if_not_exists = true })

    orderbook:create_index('by_price', {
        unique = false,
        parts = {{field = 'market_id'}, {field = 'price'}},
        if_not_exists = true })

    orderbook:create_index('long', {
        unique = false,
        parts = {{field = 'market_id'}, {field = 'side'}, {field = 'price'}, {field = 'reverse'}},
        if_not_exists = true })

    orderbook:create_index('short', {
        unique = false,
        parts = {{field = 'market_id'}, {field = 'side'}, {field = 'price'}, {field = 'timestamp'}},
        if_not_exists = true })        


    profile.set_insurance_id(market_data.id)
    return true
end

function OB.create(
    order_id,
    market_id,
    trader_id,
    price,
    size,
    side
)
    checks('string', 'string', 'number', 'decimal', 'decimal', 'string')

    local entry = OB.bind(
        new_entry(
            order_id,
            time.now(),
            market_id,
            trader_id,
            price,
            size,
            side
        )
    )
    local _, err = archiver.insert(box.space.orderbook, entry)
    if err ~= nil then
        return nil, err
    end

    return entry, nil
end

function OB.delete(order_id)
    local res = box.space.orderbook:delete(order_id)
    if res ~= nil then
        return OB.bind(res)
    end

    return nil
end

function OB.update_size(order_id, new_size)
    checks('string', 'decimal')

    local res, err = archiver.update(box.space.orderbook, order_id, {
        {'=', 'size', new_size},
    })
    if res == nil then
        return ERR_ORDERBOOK_ENTRY_NOT_FOUND
    end

    return OB.bind(res)
end

function OB.get(order_id)
    local res = box.space.orderbook:get(order_id)
    if res ~= nil then
        return OB.bind(res)
    end

    return nil
end

function OB.get_by_price(market_id, price)
    local res = box.space.orderbook.index.by_price:min{market_id, price}
    if res ~= nil then
        return OB.bind(res)
    end

    return nil
end


function OB.iterator_by_trader(profile_id)
    checks('number')
    local iter, iter_param, iter_state = box.space.orderbook.index.by_trader_id:pairs(profile_id, {iterator = box.index.EQ})

    return function(param, state)
        local state, value = iter(param, state)
        return state, value and OB.bind(value)
    end,
    iter_param,
    iter_state
end

return OB
