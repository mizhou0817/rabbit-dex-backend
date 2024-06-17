local log = require('log')
local checks = require('checks')

local archiver = require('app.archiver')
local errors = require('app.lib.errors')
local tuple = require('app.tuple')

require("app.config.constants")

local PositionError = errors.new_class("POSITION")

local P = {
    format = {
        {name = 'id', type = 'string'},
        {name = 'market_id', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'size', type = 'decimal'},
        {name = 'side', type = 'string'},
        {name = 'entry_price', type = 'decimal'},

        {name = 'unrealized_pnl', type = 'decimal'},
        {name = 'notional', type = 'decimal'},
        {name = 'margin', type = 'decimal'},
        {name = 'liquidation_price', type = 'decimal'},
        {name = 'fair_price', type = 'decimal'},
    },
    strict_type = 'engine_position',
}

-- func should create tuple with exact described O.format
local function new_position(
    pos_id,
    market_id,
    trader_id,
    size,
    side,
    entry_price
)
    checks('string', 'string', 'number', 'decimal', 'string', 'decimal')

    return {
        pos_id,
        market_id,
        trader_id,
        size,
        side,
        entry_price,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
        ZERO,
     }
end

function P.bind(value)
    return tuple.new(value, P.format, P.strict_type)
end

-- tuple behaviour
function P.tomap(self, opts)
    return self:tomap(opts)
end

function P.init_spaces()
    local position, err = archiver.create('position', {if_not_exists = true}, P.format, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(PositionError:new(err))
        error(err)
    end

    position:create_index('market_id', {
        unique = false,
        parts = {{field = 'market_id'}},
        if_not_exists = true })


    position:create_index('profile_id', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })

    position:create_index('pos_by_market_profile', {parts = {{field = 'market_id'}, {field = 'profile_id'}},
        unique = true,
        if_not_exists = true })
end

function P.create(market_id, trader_id, size, side, entry_price)
    local pos_id = "pos-" .. tostring(market_id) .. "-tr-" .. tostring(trader_id)

    local pos = P.bind(
        new_position(
            pos_id,
            market_id,
            trader_id,
            size,
            side,
            entry_price
        )
    )
    local _, err = archiver.replace(box.space.position, pos)
    if err ~= nil then
        log.error(PositionError:new(err))
        return nil, err
    end

    return pos, nil
end

function P.update(position_id, new_side, new_size, new_price)
    checks('string', 'string', 'decimal', 'decimal')

    if new_size ~= 0 then
        local res, err = archiver.update(box.space.position, position_id, {
            {'=', 'size', new_size},
            {'=', 'side', new_side},
            {'=', 'entry_price', new_price},
        })
        if err ~= nil then
            log.error(PositionError:new(err))
            return nil, err
        end

        return P.bind(res), nil
    end

    local status, res = pcall(function() return box.space.position:delete(position_id) end)
    if status == false then
        log.error(PositionError:new(res))
        return nil, res
    end

    res = res:update({{"=", "size", ZERO}})
    return P.bind(res), nil
end

function P.replace(pos)
    checks('table|engine_position')

    local _, err = archiver.replace(box.space.position, pos)

    return err
end

function P.get_positions(profile_id)
    checks("number")

    local res = {}
    for _, row in box.space.position.index.profile_id:pairs({profile_id}, {iterator=box.index.EQ}) do
        table.insert(res, P.bind(row))
    end

    return {res = res, error = nil}
end

function P.get_position(profile_id, market_id)
    checks("number", "string")

    local position = box.space.position.index.pos_by_market_profile:get({market_id, profile_id})
    if position == nil then
        return nil
    end

    return P.bind(position)
end

-- TODO: implement offset and limit
function P.get_all_active_positions(market_id, offset, limit)
    checks("string", "number", "number")
    local res = {}

    for _, position in box.space.position.index.market_id:pairs(market_id, {iterator = "EQ"}) do
        table.insert(res, P.bind(position))
    end

    return {res = res, error = nil}
end

function P.get_winning_positions(side)
    checks('string')

    local res = {}

    for _, position in box.space.position:pairs() do
        if position.unrealized_pnl > 0 and position.side == side then
            table.insert(res, P.bind(position))
        end
    end

    return {res = res, error = nil}
end

function P.iterator_by_market_profile(market_id, profile_id)
    checks('string', '?number')

    return box.space.position.index.pos_by_market_profile:pairs(
        {market_id, profile_id},
        {iterator = box.index.EQ}
    )
end

return P
