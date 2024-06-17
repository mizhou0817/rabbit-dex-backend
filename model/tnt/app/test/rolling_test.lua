local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')

local rolling = require('app.rolling')
require('app.config.constants')

local g = t.group('rolling')
g.before_each(function(cg)
    rolling.init_spaces()
end)

local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.test_rolling_min_max = function(cg)

    local timestamp = 1


    local title = "24h_last_trade_price" 
    local item_id = "btc" 
    local period_sec = 3600   -- we aggregate it per hour 
    local max_periods = 24 -- for 24 hours  
    local is_replace = true 

    local inc_timestamp = 1800 -- 30 min or 1800 second 
    local number_of_days = 2
    local number_of_steps = number_of_days * 24 * (3600 / inc_timestamp)

    local initial_trade_price = 1

    local expected_min = decimal.new(number_of_steps - 24 * (3600 / inc_timestamp) + 1)
    local expected_max = decimal.new(number_of_steps)

    local last_trade_price = decimal.new(initial_trade_price)
    for i = 1, number_of_steps do
        rolling.impl_update_roll_value(
            title,
            item_id,
            last_trade_price,
            period_sec,
            max_periods,
            is_replace,
            timestamp
        )
        timestamp = timestamp + inc_timestamp
        last_trade_price = last_trade_price + 1
    end

    local p_min, p_max = rolling.get_period_min_max(title, item_id, decimal.new(0))
    -- log.info("expected_max=%s period_max=%s",expected_max, p_max)
    -- log.info("expected_min=%s period_min=%s",expected_min, p_min)

    t.assert_equals(expected_min, p_min)
    t.assert_equals(expected_max, p_max)
end
