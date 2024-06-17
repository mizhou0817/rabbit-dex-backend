local checks = require('checks')
local log = require('log')

local errors = require('app.lib.errors')
local time = require('app.lib.time')
local util = require('app.util')
local decimal = require('decimal')
local rolling = require('app.rolling')

require("app.config.constants")

local WdmError = errors.new_class("WdmError")

local wdm = {}
local item_title = "24h_wds"
local item_id = "total_wds"


-- Create initial tiers for market if not exist
local function _create_wd_tiers()
    local tiers = {
		{"rbx", decimal.new("0"), 1}, -- 15 sec
		{"bfx", decimal.new("0"), 1}, -- 2 sec
		{"rbx", decimal.new("10000"), 10}, -- 2 min  
		{"bfx", decimal.new("10000"), 100}, -- 3 min
		{"rbx", decimal.new("100000"), 100}, -- 25 min 
		{"bfx", decimal.new("100000"), 1000}, -- 33 min
		{"rbx", decimal.new("1000000"), 1800}, -- 6 hours
		{"bfx", decimal.new("1000000"), 10800}, -- 6 hours
	}

    for _, tier in pairs(tiers) do 
        box.space.wd_tier:replace(tier)
    end

    tiers = nil
end


function wdm.init_spaces()
    -- CREATE ENGINE RELATED DATA --
    local wd_tier = box.schema.space.create('wd_tier', {if_not_exists = true})

    wd_tier:format({
        {name = 'exchange_id', type = 'string'},
        {name = 'min_amount', type = 'decimal'},
        {name = 'delay', type = 'unsigned'},
    })

    wd_tier:create_index('primary', {
        unique = true,
        parts = {{field = 'exchange_id'}, {field = "min_amount"}},
        if_not_exists = true })
    

    -- SPECIAL tiers without volume
    -- CREATE ENGINE RELATED DATA --
    local wd_special_tier = box.schema.space.create('wd_special_tier', {if_not_exists = true})
    wd_special_tier:format({
        {name = 'exchange_id', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'from_amount', type = 'decimal'},
        {name = 'to_amount', type = 'decimal'},
        {name = 'delay', type = 'unsigned'},
    })

    wd_special_tier:create_index('primary', {
        unique = true,
        parts = {{field = 'exchange_id'}, {field = 'profile_id'}, {field = 'from_amount'}, {field = "to_amount"}},
        if_not_exists = true })
    
    if box.space.wd_tier:count() <= 0 then
        _create_wd_tiers()
    end

    rolling.init_spaces()
end

function wdm.get_delay(exchange_id, profile_id, amount)
    checks("string", "number", "decimal")
    
    local delay = nil

    --check special tier first: find the first range
    for _, next_tier in box.space.wd_special_tier:pairs({exchange_id, profile_id}) do
        if next_tier.exchange_id ~= exchange_id or next_tier.profile_id ~= profile_id then
            break
        end

        if  amount >= next_tier.from_amount and  amount <= next_tier.to_amount then
            delay = next_tier.delay
            break
        end
    end

    if delay ~= nil then
        return delay
    end

    for _, next_tier in box.space.wd_tier:pairs({exchange_id}) do
        if next_tier.exchange_id ~= exchange_id then
            break
        end

        if next_tier.min_amount > amount then
            break
        end
    
        delay = next_tier.delay
    end

    return delay
end

function wdm.replace_tier(exchange_id, min_amount, delay)
    checks("string", "decimal", "number")

    return box.space.wd_tier:replace{exchange_id, min_amount, delay}
end

function wdm.replace_special_tier(exchange_id, profile_id, from_amount, to_amount, delay)
    checks("string", "number", "decimal", "decimal", "number")

    return box.space.wd_special_tier:replace({exchange_id, profile_id, from_amount, to_amount, delay})
end


function wdm.roll_volume(amount)
    checks("decimal")

    
    local e = rolling.update_roll_value(item_title, 
    item_id,
    amount,
    3600,   -- we aggregate it per hour 
    24, -- for 24 hours  
    false) -- we sum the value
    if e ~= nil then
        log.warn("wdm.roll_volume 24h_withdrawls error=%s", e)
    end
end

function wdm.wds_per_24h()
    return rolling.get_roll_sum(item_title, item_id)
end


return wdm