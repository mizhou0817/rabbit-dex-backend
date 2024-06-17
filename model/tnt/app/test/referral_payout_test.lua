local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local archiver = require('app.archiver')
local engine_profile = require('app.engine.profile')
local engine_periodics = require('app.engine.periodics')
local profile = require('app.profile')
local log = require('log')
local balance = require('app.balance')
local uuid = require('uuid')
local config = require('app.config')
local market = require('app.engine.market')
local position = require('app.engine.position')
local aggs = require('app.engine.aggregate')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')
require('app.errcodes')

local g = t.group('referral')
local work_dir = fio.tempdir()
local shard = config.params.MARKETS.BTCUSDT

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    engine_periodics._market_id = shard
    archiver.init_sequencer('profile')
    engine_profile.init_spaces()
    profile.init_spaces({})
    balance.init_spaces(shard)
    market.init_spaces(config.markets[shard])
    position.init_spaces()
    aggs.init_spaces()

    -- add 100 balance to the FEE_WALLET
    local default_balance = decimal.new(100)
    box.space.exchange_wallets:upsert(
        {FEE_WALLET_ID, default_balance, 11111},
        {{ '+', 'balance', default_balance }}
    )
end)

g.after_each(function(cg)
    box.space.profile:drop()
    box.sequence.PID:drop()
    box.space.balance_operations:drop()
    box.space.balance_sum:drop()
    box.space.exchange_wallets:drop()
end)

function get_balance(profile_id)
    local sum = box.space.balance_sum:get(profile_id)
    if sum == nil then
        return 0
    end
    return sum.balance
end

g.test_simple_flow = function(cg)
    -- check the update function got called.
    local called_num = 0
    local called_event = nil
    local called_params = nil

    box.broadcast = function(event, params)
        print(event, params)
        called_num = called_num + 1
        called_event = event
        called_params = params
        return nil
    end

    t.assert_equals(called_num, 0)
    profile.profile.create("trader", "active", "0xqwe", DEFAULT_EXCHANGE_ID)

    local p = box.space.profile:get(1)
    t.assert_is_not(p, nil)

    local bops_len = box.space.balance_operations:count()
    t.assert_equals(bops_len, 0)

    local profile_id = p.id
    local balance_sum = get_balance(profile_id)
    t.assert_is(balance_sum, 0)

    local amount = decimal.new(100)
    local id = uuid.str()
    local res = balance.create_referral_payout(id, profile_id, amount)
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)
    t.assert_equals(called_num, 0)

    local bops_len = box.space.balance_operations:count()
    local bops = box.space.balance_operations:get(id)
    t.assert_equals(bops_len, 1)
    t.assert_is_not(bops, nil)
    t.assert_is(bops.profile_id, profile_id)
    t.assert_is(bops.amount, amount)
    t.assert_is(bops.status, config.params.BALANCE_STATUS.PENDING)
    t.assert_equals(called_num, 0)
    
    -- still 0, since not processed yet.
    balance_sum = get_balance(profile_id)
    t.assert_is(balance_sum, 0)

    local res = engine_profile.process_referral_payout('BTC-USD')
    t.assert_is(res['error'], nil)
    t.assert_is_not(res['res'], nil)

    bops = box.space.balance_operations:get(id)
    t.assert_is_not(bops, nil)
    t.assert_is(bops.amount, amount)
    t.assert_is(bops.status, config.params.BALANCE_STATUS.SUCCESS)

    -- now amount should be greater than 0 since it is now processed
    balance_sum = get_balance(profile_id)
    t.assert_is(balance_sum, amount)

    -- check update_profiles called broadcast
    local expected_event = config.sys.EVENTS.PROFILE_UPDATE
    local expected_params = {market_id=shard, profiles={profile_id}}
    t.assert_equals(called_num, 1)
    t.assert_equals(called_event, expected_event)
    t.assert_equals(expected_params, called_params)
end

g.test_create_referral_payout_negative_amount = function(cg)
    local res = balance.create_referral_payout('foobar1', 555, decimal.new(0))
    t.assert_is(res['res'], nil)
    t.assert_is_not(res['error'], nil)
    t.assert_equals(res['error'], ERR_REFERRAL_PAYOUT_AMOUNT_NOT_POSITIVE)

    res = balance.create_referral_payout('foobar1', 555, decimal.new(-1))
    t.assert_is(res['res'], nil)
    t.assert_is_not(res['error'], nil)
    t.assert_equals(res['error'], ERR_REFERRAL_PAYOUT_AMOUNT_NOT_POSITIVE)
end

g.test_create_referral_payout_duplicate = function(cg)
    local id = 'foobar1'
    local res = balance.create_referral_payout(id, 555, decimal.new(100))
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    res = balance.create_referral_payout(id, 555, decimal.new(100))
    t.assert_is(res['res'], nil)
    t.assert_is_not(res['error'], nil)
    t.assert_equals(res['error'], ERR_REFERRAL_PAYOUT_ID_DUPLICATE)
end

g.test_create_referral_payout_dip = function(cg)
    local id = 'foobar1'
    local res = balance.create_referral_payout(id, 555, decimal.new(101))
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    res = engine_profile.process_referral_payout(shard)
    t.assert_is(res['res'], nil)
    t.assert_equals(res['error'], ERR_REFERRAL_PAYOUT_NEGATIVE_FEE_WALLET)
end

g.test_process_multiple_times = function(cg)
    local id = 'foobar1'
    local res = balance.create_referral_payout(id, 555, decimal.new(20))
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    res = engine_profile.process_referral_payout(shard)
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    res = engine_profile.process_referral_payout(shard)
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    res = engine_profile.process_referral_payout(shard)
    t.assert_is_not(res['res'], nil)
    t.assert_is(res['error'], nil)

    local balance_sum = get_balance(555)
    t.assert_is(balance_sum, decimal.new(20))

    local b = box.space.exchange_wallets:get(FEE_WALLET_ID)
    t.assert_equals(b.balance, decimal.new(80))
end