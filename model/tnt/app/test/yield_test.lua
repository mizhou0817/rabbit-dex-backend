local decimal = require('decimal')
local fio = require('fio')
local math = require('math')
local t = require('luatest')
local log = require('log')
local archiver = require('app.archiver')
local balance = require('app.balance')
local profile = require('app.profile')
local config = require('app.config')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')
local time = require('app.lib.time')

local profile_cache = profile.cache

require('app.config.constants')

local g = t.group('stake')
local work_dir = fio.tempdir()
local ONE_MILLIONTH = decimal.new(0.000001)

local function enable_strict_mode()
    setmetatable(_G, {
        __newindex = function(table, key, value)
            local where = debug.getinfo(2, "S").what
            if where ~= "main" and where ~= "C" then
                error("attempt to write to undeclared variable " .. tostring(key), 2)
            end
            rawset(table, key, value)
        end,
        __index = function(table, key)
            error("attempt to read undeclared variable " .. tostring(key), 2)
        end,
    })
end

t.before_suite(function()
    box.cfg {
        work_dir = work_dir,
    }
    math.randomseed(os.time())
    archiver.init_sequencer("balance")
    profile.init_spaces({})
    balance.init_spaces(0)

    balance.add_to_contract_map("", 0, DEFAULT_EXCHANGE_ID)
    enable_strict_mode()
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    box.space.balance_sum:truncate()
    box.space.balance_operations:truncate()
    box.space.profile:truncate()
    box.space.profile_cache:truncate()
end)

g.test_payout = function(cg)
    local total_balance = decimal.new(0)
    local traders = {}
    for i = 1, 20 do
        local res = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x" .. tostring(i), DEFAULT_EXCHANGE_ID)
        t.assert_is(res["error"], nil)
        local trader_profile = res["res"]
        local exponent = math.random(1, 10)
        local profile_balance = decimal.new(math.random()) * decimal.new(10)^exponent
        set_balance(trader_profile.id, profile_balance)
        local trader = {
            profile_id = trader_profile.id,
            balance = profile_balance
        }
        traders[i] = trader
        total_balance = total_balance + profile_balance
    end
    local yield = decimal.new(54321)
    profile_cache.process_yield_and_invalidate("y_1", yield, "0x123", DEFAULT_EXCHANGE_ID, 0, "0xabc")
    for i = 1, 20 do
        local trader = traders[i]
        local profile_balance = trader.balance
        local frac = (yield / total_balance) * decimal.new(0.9999999)
        local expected_yield = profile_balance * frac
        local expected_balance = profile_balance + expected_yield
        local actual_balance = get_balance(trader.profile_id)
        assert_close_to(actual_balance, expected_balance)
    end
end

g.test_split_payout = function(cg)
    -- create balances and total them for all traders
    local total_balance = decimal.new(0)
    local partial_balance
    local traders = {}
    for i = 1, 10 do
        local exponent = math.random(1, 10)
        local profile_balance = decimal.new(math.random()) * decimal.new(10)^exponent
        local trader = {
            balance = profile_balance
        }
        traders[i] = trader
        total_balance = total_balance + profile_balance
    end
    partial_balance = total_balance
    for i = 0, 9 do
        local profile_balance = traders[10 - i].balance
        local trader = {
            balance = profile_balance
        }
        traders[11 + i] = trader
        total_balance = total_balance + profile_balance
    end
    -- create profiles and set balances for half the traders
    for i = 1, 10 do
        local res = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x" .. tostring(i), DEFAULT_EXCHANGE_ID)
        t.assert_is(res["error"], nil)
        local trader_profile = res["res"]
        local trader = traders[i]
        trader.profile_id = trader_profile.id
        set_balance(trader.profile_id, trader.balance)
    end
    -- process yield, it will be distributed amongst the first thousand traders
    local yield = decimal.new(123456)
    profile_cache.process_yield_and_invalidate("y_1", yield, "0x123", DEFAULT_EXCHANGE_ID, 0, "0xabc")
    -- check the results
    local halfway_total_yield = decimal.new(0)
    local frac = (yield / partial_balance) * ROUND_DOWN_FACTOR
    for i = 1, 10 do
        local trader = traders[i]
        local profile_balance = trader.balance
        local expected_yield = profile_balance * frac
        local expected_balance = profile_balance + expected_yield
        local actual_balance = get_balance(trader.profile_id)
        t.assert_equals(actual_balance, expected_balance)
        trader.halfway_balance = actual_balance
        halfway_total_yield = halfway_total_yield + expected_yield
    end
    -- create profiles and set balances for the other thousand traders
    for i = 11, 20 do
        local res = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x" .. tostring(i), DEFAULT_EXCHANGE_ID)
        t.assert_is(res["error"], nil)
        local trader_profile = res["res"]
        local trader = traders[i]
        trader.profile_id = trader_profile.id
        set_balance(trader.profile_id, trader.balance)
    end
    -- process yield again, this time with roughly twice as much yield,
    -- the first thousand traders should keep their initial allocation and
    -- get no more this time,
    -- what they got before should be deducted from the yield left to distribute, 
    -- and the unallocated half of the yield should be distributed amongst the
    -- second thousand traders
    local yield2 = yield * decimal.new(2.1)
    profile_cache.process_yield_and_invalidate("y_1", yield2, "0x123", DEFAULT_EXCHANGE_ID, 0, "0xabc")
    local total_yield = halfway_total_yield
    local unallocated_yield = yield2 - halfway_total_yield
    local eligible_balance = total_balance - partial_balance
    frac = (unallocated_yield / eligible_balance) * ROUND_DOWN_FACTOR
    for i = 1, 20 do
        local trader = traders[i]
        local profile_balance = trader.balance
        local expected_balance
        if i <= 10 then
            expected_balance = trader.halfway_balance
        else
            local expected_yield = profile_balance * frac
            total_yield = total_yield + expected_yield
            expected_balance = profile_balance + expected_yield
        end
        local actual_balance = get_balance(trader.profile_id)
        assert_close_to(actual_balance, expected_balance)
    end
end

g.test_no_double_payout = function(cg)
    local total_balance = decimal.new(0)
    local traders = {}
    for i = 1, 20 do
        local res = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x" .. tostring(i), DEFAULT_EXCHANGE_ID)
        t.assert_is(res["error"], nil)
        local trader_profile = res["res"]
        local exponent = math.random(1, 10)
        local profile_balance = decimal.new(math.random()) * decimal.new(10)^exponent
        set_balance(trader_profile.id, profile_balance)
        local trader = {
            profile_id = trader_profile.id,
            balance = profile_balance
        }
        traders[i] = trader
        total_balance = total_balance + profile_balance
    end
    local yield = decimal.new(54321)
    profile_cache.process_yield_and_invalidate("y_1", yield, "0x123", DEFAULT_EXCHANGE_ID, 0, "0xabc")
    for i = 1, 20 do
        local trader = traders[i]
        local profile_balance = trader.balance
        local frac = (yield / total_balance) * decimal.new(0.9999999)
        local expected_yield = profile_balance * frac
        local expected_balance = profile_balance + expected_yield
        local actual_balance = get_balance(trader.profile_id)
        assert_close_to(actual_balance, expected_balance)
        trader.balance = actual_balance
    end
    --process the same yield again, shouldn't change anybody's balance
    profile_cache.process_yield_and_invalidate("y_1", yield, "0x123", DEFAULT_EXCHANGE_ID, 0, "0xabc")
    for i = 1, 20 do
        local trader = traders[i]
        local expected_balance = trader.balance
        local actual_balance = get_balance(trader.profile_id)
        assert_close_to(actual_balance, expected_balance)
    end
end

function get_balance(profile_id)
    local res = profile_cache.get_cache(profile_id)
    if res.res == nil then
        return 0
    end

    local p_cache = res.res
    return p_cache.balance
end

function set_balance(profile_id, new_balance)
    local tm = time.now()
    local new_sum = {
        profile_id,
        new_balance,
        tm
    }
    box.space.balance_sum:upsert(
        new_sum,
        {
            { '=', "balance",      new_balance },
            { '=', 'last_updated', tm }
        }
    )

    profile_cache.update(profile_id)
end

function assert_close_to(actual, expected)
    local diff = actual - expected
    if diff < 0 then
        diff = -diff
    end
    if (diff > TENTH_OF_A_CENT) then
        if expected < 0 then
            expected = -expected
        end
        local ratio_ok = false
        if expected > 1 then
            local ratio = diff / expected
            ratio_ok = ratio < ONE_MILLIONTH
        end
        if not ratio_ok then
            t.fail(string.format(
                "Expected %s to be close to %s, the difference is %s", tostring(actual), tostring(expected), tostring(diff)))
        end
    end
end
