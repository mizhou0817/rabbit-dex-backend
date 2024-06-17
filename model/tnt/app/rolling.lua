local log = require('log')

local errors = require('app.lib.errors')
local time = require('app.lib.time')

require("app.config.constants")

local err = errors.new_class("ROLLING_ERR")

-- TODO: make universal rolling lib
local C = {}

local SPACE_NAME = "roll_values"
local MIN_MAX_SPACE_NAME = "mix_max_roll_values"

C.space_name = SPACE_NAME
C.min_max_space_name = MIN_MAX_SPACE_NAME

function C.init_spaces()

    -- ROLL VALUES: 60m_basis, 30m_basis, 24h_trading_volume, etc..
    local roll_values = box.schema.space.create(SPACE_NAME, {if_not_exists = true})

    roll_values:format({
        {name = 'title', type = 'string'},
        {name = 'item_id', type = 'string'},
        {name = 'period', type = 'number'},
        {name = 'timestamp', type = 'number'},
        {name = 'value', type = 'decimal'},
    })

    roll_values:create_index('primary', {
        unique = true,
        parts = {{field = 'title'}, {field = 'item_id'}, {field = 'period'}},
        if_not_exists = true })   

    local min_max_roll_values = box.schema.space.create(MIN_MAX_SPACE_NAME, {if_not_exists = true})

    min_max_roll_values:format({
        {name = 'title', type = 'string'},
        {name = 'item_id', type = 'string'},
        {name = 'period', type = 'number'},
        {name = 'min_value', type = 'decimal'},
        {name = 'max_value', type = 'decimal'},
    })

    min_max_roll_values:create_index('primary', {
        unique = true,
        parts = {{field = 'title'}, {field = 'item_id'}, {field = 'period'}},
        if_not_exists = true })   
    
end

function C.reset_roll_value(title, item_id)
    for _, item in box.space[SPACE_NAME]:pairs({title, item_id}, {iterator='EQ'}) do
        box.space[SPACE_NAME]:delete({title, item_id, item.period})    
    end
end


function C.impl_update_roll_value(title, item_id, new_value, period_sec, max_values, is_replace, timestamp)
    local period = math.ceil(timestamp / period_sec)

    local new_item = {
        title,
        item_id,
        period,
        timestamp,
        new_value
    }

    local ops = {'+', 'value', new_value}
    if is_replace == true then
        ops = {'=', 'value', new_value}
    end

    local status, res = pcall(function() return 
        box.space[SPACE_NAME]:upsert(new_item, {
            ops
        })
    end)   
    if status == false then
        log.error(err:new("can't roll value error=%s", res))
        return res
    end


    -- Update min max for period
    local min_value = new_value
    local max_value = new_value
    local current_min_max = box.space[MIN_MAX_SPACE_NAME]:get({title, item_id, period})
    if current_min_max ~= nil then
        if min_value > current_min_max.min_value then
            min_value = current_min_max.min_value
        end 
        
        if max_value < current_min_max.max_value then
            max_value = current_min_max.max_value
        end
    end

    box.space[MIN_MAX_SPACE_NAME]:replace({
        title,
        item_id,
        period,
        min_value,
        max_value
    })


    -- rotate rolling
    local count = box.space[SPACE_NAME]:count({title, item_id}, {iterator='EQ'})
    if count > max_values then
        local first = box.space[SPACE_NAME].index.primary:min({title, item_id})
        if first ~= nil then
            box.space[SPACE_NAME]:delete({title, item_id, first.period})
        end
    end

    -- rotate min max rolling
    count = box.space[MIN_MAX_SPACE_NAME]:count({title, item_id}, {iterator='EQ'})
    if count > max_values then
        local first = box.space[MIN_MAX_SPACE_NAME].index.primary:min({title, item_id})
        if first ~= nil then
            box.space[MIN_MAX_SPACE_NAME]:delete({title, item_id, first.period})
        end
    end

    return nil
end

function C.update_roll_value(title, item_id, new_value, period_sec, max_values, is_replace)
    
    --[[
        The time is in seconds: 
        1001 - X
        1002 - X1
        1003 - X2
        1004 - X3

        1001 - period
        1001.009
        1001.010
        1001.011

        1002.010
        1003.012
        let's say we need to "calc values" per minute.

        1001 / 60
        1002 / 60
    --]]
    local timestamp = math.ceil(time.now_sec())
    return C.impl_update_roll_value(title, item_id, new_value, period_sec, max_values, is_replace, timestamp)
end


function C.get_roll_avg(title, item_id)
    local avg = ZERO
    local counter = 0 
    for _, item in box.space[SPACE_NAME]:pairs({title, item_id}, {iterator='EQ'}) do
        counter = counter + 1
        avg = avg + item.value
    end

    if counter > 0 then
        avg = avg / counter
    end

    return avg
end


function C.get_roll_sum(title, item_id)
    local sum = ZERO
    for _, item in box.space[SPACE_NAME]:pairs({title, item_id}, {iterator='EQ'}) do
        sum = sum + item.value
    end

    return sum
end

-- return diff: (premium, basis)
function C.diff_roll_value(title, item_id)
    local count = box.space[SPACE_NAME]:count({title, item_id}, {iterator='EQ'})
    if count <= 1 then
        return ZERO, ZERO
    end

    local first = box.space[SPACE_NAME].index.primary:min({title, item_id})
    local last = box.space[SPACE_NAME].index.primary:max({title, item_id})
    if first == nil or last == nil then
        return ZERO, ZERO
    end

    local basis = ZERO
    local premium = last.value - first.value
    if first.value ~= ZERO then
        basis = premium / first.value
    end

    return premium, basis
end

function C.get_period_min_max(title, item_id, current)
    local count = box.space[MIN_MAX_SPACE_NAME]:count({title, item_id}, {iterator='EQ'})
    if count <= 0 then
        return current, current
    end

    local new_max = nil
    local new_min = nil

    for _, item in box.space[MIN_MAX_SPACE_NAME]:pairs({title, item_id}, {iterator='EQ'}) do
        if new_max == nil then
            new_max = item.max_value
        end

        if new_min == nil then
            new_min = item.min_value
        end

        if item.min_value < new_min then
            new_min = item.min_value
        end

        if item.max_value > new_max then
            new_max = item.max_value
        end
    end

    return new_min, new_max
end

function C.get_all_min_max_values(title, item_id)
    return box.space[MIN_MAX_SPACE_NAME]:select({title, item_id})
end


return C