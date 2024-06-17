local log = require('log')
local decimal = require('decimal')

local archiver = require('app.archiver')
local errors = require('app.lib.errors')
local util = require('app.util')

local TradeError = errors.new_class("TRADE")

local GET_BATCH_LIMIT = 1000

local trade = {}

function trade.init_spaces()
    box.schema.sequence.create('trade_sequence',{start=0, min=0, if_not_exists = true})

    local fill, err = archiver.create('fill', {if_not_exists = true}, {
        {name = 'id', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'market_id', type = 'string'},
        {name = 'order_id', type = 'string'},
        {name = 'timestamp', type = 'number'},

        {name = 'trade_id', type = 'string'},
                
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
        {name = 'side', type = 'string'},
        {name = 'is_maker', type = 'boolean'},

        {name = 'fee', type = 'decimal'},
        {name = 'liquidation', type = 'boolean'},

        {name = 'client_order_id', type = 'string'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(TradeError:new(err))
        error(err)
    end

    fill:create_index('order_id', {
        unique = false,
        parts = {{field = 'order_id'}},
        if_not_exists = true })

    fill:create_index('market_id', {
        unique = false,
        parts = {{field = 'market_id'}},
        if_not_exists = true })

    fill:create_index('profile_id', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })

    fill:create_index('profile_id_timestamp', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'timestamp'}},
        if_not_exists = true })
    
    
    local trade = archiver.create('trade', {if_not_exists = true}, {
        {name = 'id', type = 'string'},
        {name = 'market_id', type = 'string'},        
        {name = 'timestamp', type = 'number'},
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
        {name = 'liquidation', type = 'boolean'},
        {name = 'taker_side', type = 'string'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(TradeError:new(err))
        error(err)
    end

    trade:create_index('market_id', {
        unique = false,
        parts = {{field = 'market_id'}},
        if_not_exists = true })        

    trade:create_index('timestamp', {
        unique = false,
        parts = {{field = 'timestamp'}},
        if_not_exists = true })  
    
end

function trade.next_trade_id(market_id)
    local n =  box.sequence.trade_sequence:next()
    local id = tostring(market_id) .. "-" .. tostring(n)

    return id
end

function trade.get_trade_data(limit)
        local s = box.space['trade']
        local idx = s.index['timestamp']

        local res = idx:pairs(nil, {iterator=box.index.REQ}):take_n(limit):totable()
        return {res = res, error = nil}
end

function trade.total_volume(profile_id, start_timestamp, end_timestamp)
    local total = decimal.new(0)
    local last_timestamp_found = start_timestamp

    local count = 0

    -- GT iterrator on this index will select values where profile_id or timestamp > then ones in the key
    -- BUT they will be sorted in asc order - because it's tree index
    -- That's why we need to check in the loop both profile_id and timestamp 
    for _, fill in box.space.fill.index.profile_id_timestamp:pairs({profile_id, start_timestamp}, {iterator='GT'}) do
        if fill.profile_id ~= profile_id or
            fill.timestamp > end_timestamp then
                break
        end

        total = total + fill.price * fill.size
        last_timestamp_found = fill.timestamp

        count = util.safe_yield(count, GET_BATCH_LIMIT)
    end

    return {res = {total, last_timestamp_found}, error = nil}
end


return trade