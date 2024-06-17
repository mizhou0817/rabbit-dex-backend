local log      = require('log')
require('app.lib.table')
require('app.config.constants')

local MAX_CANDLES_PER_REQ = 1000
local _periods = {1, 5, 15, 30, 60, 240, 1440} -- minutes
local candles = {}

function candles.init_spaces()

    for _, period in pairs(_periods) do
        local space_name = "candles" .. tostring(period)

        local candles = box.schema.space.create(space_name, {if_not_exists = true})

        candles:format({
            {name = 'time', type = 'number'},
            {name = 'low', type = 'decimal'},
            {name = 'high', type = 'decimal'},
            {name = 'open', type = 'decimal'},
            {name = 'close', type = 'decimal'},
            {name = 'volume', type = 'decimal'}
        })

        candles:create_index('primary', {
            unique = true,
            parts = {{field = 'time'}},
            if_not_exists = true })
    end
end

function candles.add_all_periods(price, size, timestamp)
    timestamp = tonumber(timestamp)
    local timestamp_seconds = math.floor(timestamp / 1e6)

    local volume = price * size
    for _, _period in pairs(_periods) do
        local space_name = "candles" .. tostring(_period)

        local period_sec = _period * 60 
        local aligned_timestamp = math.floor(timestamp_seconds / period_sec) * period_sec


        local exist = box.space[space_name]:get(aligned_timestamp)
        if exist == nil then
            box.space[space_name]:replace({
                aligned_timestamp,
                price, -- low
                price, -- high
                price, -- open
                price, -- close
                volume
            })
        else
            local new_low = price < exist.low and price or exist.low
            local new_high = price > exist.high and price or exist.high
            local new_volume = exist.volume + volume
            box.space[space_name]:replace({
                aligned_timestamp,
                new_low, -- low
                new_high, -- high
                exist.open, -- open
                price, -- close
                new_volume
            })
        end
    end

    return {res = nil, error = nil}
end    

function candles.get_candles(period, time_from, time_to)
    local period_sec = period * 60
    local arr = {}
    local space_name = "candles" .. tostring(period)

    if box.space[space_name] == nil then
        return {res = arr, error = "NO_CANDLES_FOR_PERIOD"}
    end

    --[[
    local num_of_periods = math.ceil((time_to - time_from) / period_sec)
    if num_of_periods <= 0 or num_of_periods > MAX_CANDLES_PER_REQ then
        return {res = arr, error = "EXCEED_PERIOD"}
    end
    --]]

    local time_from_aligned = math.floor(time_from / period_sec) * period_sec
    local time_to_aligned = math.floor(time_to / period_sec) * period_sec
    local next_filled_timestamp = time_to_aligned

    local gap_value
    local last_candle
    local max_exceed = false
    local count = 0
    for _, candle in box.space[space_name]:pairs(time_to_aligned, {iterator="LE"}) do
        if candle.time < time_from_aligned then
            break
        end
    
        while (next_filled_timestamp ~= candle.time)
        do 
            gap_value = candle:update({
                {'=', 'volume', ZERO}, 
                {'=', 'time', next_filled_timestamp},
                {'=', 'low', candle.close},
                {'=', 'high', candle.close},
                {'=', 'open', candle.close}
            })
            
            table.insert(arr, gap_value:tomap({names_only=true}))    
            
            next_filled_timestamp = next_filled_timestamp - period_sec
            
            count = count + 1
            if count > MAX_CANDLES_PER_REQ then
                max_exceed = true
                break
            end    
        end
        count = count + 1
        if max_exceed == true or count > MAX_CANDLES_PER_REQ then
            break
        end

        table.insert(arr, candle:tomap({names_only=true}))
        next_filled_timestamp = candle.time - period_sec
        last_candle = candle 
    end

    if count == 0 then -- we didn't find any values: just fill the full range with nearest values
        gap_value = nil 
        for _, first in box.space[space_name]:pairs(time_to_aligned, {iterator="LT"}) do
            gap_value = first
            break
        end
        if gap_value == nil then
            return {res = arr, error = nil}
        end

        while (next_filled_timestamp >= time_from_aligned)
        do
            gap_value = gap_value:update({
                {'=', 'volume', ZERO}, 
                {'=', 'time', next_filled_timestamp},
                {'=', 'low', gap_value.close},
                {'=', 'high', gap_value.close},
                {'=', 'open', gap_value.close}
            })
            
            table.insert(arr, gap_value:tomap({names_only=true}))    
            
            next_filled_timestamp = next_filled_timestamp - period_sec
            
            count = count + 1
            if count > MAX_CANDLES_PER_REQ then
                break
            end    
        end
    elseif (count < MAX_CANDLES_PER_REQ and next_filled_timestamp >= time_from_aligned) then
        -- we didn't exceed count but don't have candles anymore for this range
        -- need to fill the gap
        gap_value = nil 
        for _, first in box.space[space_name]:pairs(next_filled_timestamp, {iterator="LT"}) do
            gap_value = first
            break
        end
        if gap_value == nil then
            return {res = arr, error = nil}
        end

        while (next_filled_timestamp >= time_from_aligned)
        do 
            gap_value = last_candle:update({
                {'=', 'volume', ZERO}, 
                {'=', 'time', next_filled_timestamp},
                {'=', 'low', gap_value.close},
                {'=', 'high', gap_value.close},
                {'=', 'open', gap_value.close}
            })
            
            table.insert(arr, gap_value:tomap({names_only=true}))    
            next_filled_timestamp = next_filled_timestamp - period_sec
            
            count = count + 1
            if count > MAX_CANDLES_PER_REQ then
                break
            end    
        end

    end

    return {res = arr, error = nil}
end    


return candles