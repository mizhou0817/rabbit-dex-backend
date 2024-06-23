local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local archiver = require('app.archiver')
local balance = require('app.balance')
local profile = require('app.profile')
local config = require('app.config')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')
local time = require('app.lib.time')

require('app.config.constants')

function pubsub_publish(channel, json_data)
end

local g = t.group('stake')
local vault_profile_id
local manager_profile_id = 13
local treasurer_profile_id = 14
local vault_wallet = "0xe1a243b4e5b6"
local performance_fee = decimal.new(0.2)
local stake_id_num = 0
local work_dir = fio.tempdir()

t.before_suite(function()
    box.cfg {
        work_dir = work_dir,
    }
    archiver.init_sequencer("balance")
    profile.init_spaces({})
    balance.init_spaces(0)
    balance.add_to_contract_map("", 0, DEFAULT_EXCHANGE_ID)
    local res = profile.profile.create("vault", "active", vault_wallet, DEFAULT_EXCHANGE_ID)
    t.assert_is(res["error"], nil)
    local vault_profile = res["res"]
    vault_profile_id = vault_profile.id
    balance.init_vault(vault_profile_id, "vault_name", manager_profile_id, "manager_name", treasurer_profile_id, performance_fee)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.test_stake = function(cg)
    local amount = decimal.new(13)
    local current_nav = decimal.new(100)
    local staker_profile_id = 11
    local txhash = random_txhash()
    local from_balance = false
    local stake_id = next_stake_id()
    create_and_process_stake(amount, staker_profile_id, current_nav, txhash, from_balance, stake_id)
end

g.test_stake_from_balance = function(cg)
    local amount = decimal.new(10)
    local current_nav = decimal.new(100)
    local staker_profile_id = 11
    local stake_txhash = random_txhash()
    local from_balance = true
    local stake_id = next_stake_id()
    local deposit_id = "d_1"
    local staker_wallet = "0x123"
    local deposit_txhash = "0x321d"
    create_and_process_deposit(amount, staker_profile_id, staker_wallet, deposit_txhash, deposit_id)
    create_and_process_stake(amount, staker_profile_id, current_nav, stake_txhash,
        from_balance, stake_id)
end

g.test_stake_unstake = function(cg)
    local amount = decimal.new(13)
    local entry_nav = decimal.new(150)
    local staker_profile_id = 13457347822
    local txhash = random_txhash()
    local from_balance = false
    local stake_id = next_stake_id()
    local shares_before = get_shares(vault_profile_id, staker_profile_id)
    log.info("shares_before %s", tostring(shares_before))
    log.info("entry_nav %s", tostring(entry_nav))
    create_and_process_stake(amount, staker_profile_id, entry_nav, txhash, from_balance, stake_id)
    entry_nav = entry_nav + amount
    local shares_after = get_shares(vault_profile_id, staker_profile_id)
    log.info("shares_after %s", tostring(shares_after))
    local entry_nav_after = get_entry_nav(vault_profile_id, staker_profile_id)
    log.info("entry_nav_after %s", tostring(entry_nav_after))
    local shares_to_unstake = shares_after / 2
    local withdrawable_balance = decimal.new(150)
    local current_nav = decimal.new(200)
    set_balance(vault_profile_id, current_nav)
    local vault_balance_before = get_balance(vault_profile_id)
    local staker_balance_before = get_balance(staker_profile_id)
    local treasurer_balance_before = get_balance(treasurer_profile_id)
    create_and_process_unstake(shares_to_unstake, staker_profile_id, current_nav, withdrawable_balance)
    local vault_balance_after = get_balance(vault_profile_id)
    local staker_balance_after = get_balance(staker_profile_id)
    local treasurer_balance_after = get_balance(treasurer_profile_id)
    local vault_unstake_value = vault_balance_before - vault_balance_after
    local staker_unstake_value = staker_balance_after - staker_balance_before
    local treasurer_fee = treasurer_balance_after - treasurer_balance_before
    local total_unstake_value = staker_unstake_value + treasurer_fee
    assert_close_to(vault_unstake_value, total_unstake_value)
    local performance_gain = (total_unstake_value * (current_nav - entry_nav)) / current_nav
    local expected_treasurer_fee = performance_gain * performance_fee
    assert_close_to(treasurer_fee, expected_treasurer_fee)
end

g.test_stake_unstake_no_fee = function(cg)
    local amount = decimal.new(12)
    local current_nav = decimal.new(350)
    local staker_profile_id = 121
    local txhash = random_txhash()
    local from_balance = false
    local stake_id = next_stake_id()
    create_and_process_stake(amount, staker_profile_id, current_nav, txhash, from_balance, stake_id)
    local shares_after = get_shares(vault_profile_id, staker_profile_id)
    local shares_to_unstake = shares_after
    local withdrawable_balance = decimal.new(150)
    current_nav = decimal.new(200) --unstake at lower nav than stake
    set_balance(vault_profile_id, current_nav)
    local treasurer_balance_before = get_balance(treasurer_profile_id)
    create_and_process_unstake(shares_to_unstake, staker_profile_id, current_nav, withdrawable_balance)
    local treasurer_balance_after = get_balance(treasurer_profile_id)
    assert_close_to(treasurer_balance_after, treasurer_balance_before)
end

function next_stake_id()
    stake_id_num = stake_id_num + 1
    return "s_" .. tostring(stake_id_num)
end

function random_txhash()
    local r = math.random(0, 2 ^ 53 - 1)
    return string.format("0x%x", r)
end

function get_balance(profile_id)
    local b_sum = box.space.balance_sum:get(profile_id)
    if b_sum == nil then
        return 0
    end
    return b_sum.balance
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
end

function get_shares(vault_id, staker_id)
    local shares = box.space.vault_holdings:get({ vault_id, staker_id })
    if shares == nil then
        return 0
    end
    return shares.shares
end

function get_entry_nav(vault_id, staker_id)
    local shares = box.space.vault_holdings:get({ vault_id, staker_id })
    if shares == nil then
        return 0
    end
    return shares.entry_nav
end

function get_entry_price(vault_id, staker_id)
    local shares = box.space.vault_holdings:get({ vault_id, staker_id })
    if shares == nil then
        return 0
    end
    return shares.entry_price
end

function get_total_shares(vault_id)
    local vault = box.space.vaults:get(vault_id)
    if vault == nil then
        return 0
    end
    return vault.total_shares
end

function assert_close_to(actual, expected)
    local tolerance = 1e-12
    local diff = actual - expected
    if (diff > tolerance) or (diff < -tolerance) then
        t.fail(string.format(
            "Expected %s to be close to %s", tostring(actual), tostring(expected)))
    end
end

function create_and_process_deposit(amount, profile_id, wallet, txhash, deposit_id)
    local balance_before = get_balance(profile_id)
    local res = balance.create_deposit(profile_id, wallet, amount, txhash, "", 0)
    t.assert_is(res["error"], nil)
    local exist = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is_not(exist, nil)
    local id = exist.id
    res = balance.process_deposit(profile_id, wallet, deposit_id, amount, txhash, false, DEFAULT_EXCHANGE_ID, 0, "")
    t.assert_is(res["error"], nil)
    local deposit = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is(deposit.id, id)
    t.assert_is(deposit.ops_id2, deposit_id)
    local balance_after = get_balance(profile_id)
    assert_close_to(balance_after, balance_before + amount)
end

function create_and_process_stake(amount, staker_profile_id, current_nav, txhash, from_balance, stake_id)
    local balance_before = get_balance(staker_profile_id)
    local created_id
    if (not from_balance) then
        local res = balance.create_stake(staker_profile_id, vault_wallet, amount, txhash, "", 0)
        t.assert_is(res["error"], nil)
        created_id = res["res"].id
        local exist = box.space.balance_operations.index.txhash:min { txhash }
        t.assert_is_not(exist, nil)
        t.assert_is(exist.status, config.params.BALANCE_STATUS.PENDING)
    end
    balance_before = get_balance(staker_profile_id)
    local shares_before = get_shares(vault_profile_id, staker_profile_id)
    local total_shares_before = get_total_shares(vault_profile_id)
    local res = balance.process_stake(staker_profile_id, vault_profile_id, vault_wallet, stake_id, amount, current_nav, txhash,
        from_balance, DEFAULT_EXCHANGE_ID)
    t.assert_is(res["error"], nil)
    local balance_after = get_balance(staker_profile_id)
    if not from_balance then
        local stake = box.space.balance_operations.index.txhash:min { txhash }
        t.assert_is(stake.id, created_id)
        t.assert_is(stake.ops_id2, stake_id)
        t.assert_is(stake.status, config.params.BALANCE_STATUS.SUCCESS)
    else
        t.assert_is(res["res"].status, config.params.BALANCE_STATUS.SUCCESS)
    end
    if from_balance then
        assert_close_to(balance_after, balance_before - amount)
    else
        assert_close_to(balance_after, balance_before)
    end
    local shares_after = get_shares(vault_profile_id, staker_profile_id)
    local expected_new_shares
    if total_shares_before == 0 then
        expected_new_shares = amount
    else
        expected_new_shares = amount * total_shares_before / current_nav
    end
    local expected_shares_after = shares_before + expected_new_shares
    assert_close_to(shares_after, expected_shares_after)
    local total_shares_after = get_total_shares(vault_profile_id)
    assert_close_to(total_shares_after, total_shares_before + expected_new_shares)
end

function create_and_process_unstake(shares, staker_profile_id, current_nav, withdrawable_balance)
    local staker_balance_before = get_balance(staker_profile_id)
    local vault_balance_before = get_balance(vault_profile_id)
    local treasurer_balance_before = get_balance(treasurer_profile_id)
    local shares_before = get_shares(vault_profile_id, staker_profile_id)
    local entry_nav = get_entry_nav(vault_profile_id, staker_profile_id)
    local entry_price = get_entry_price(vault_profile_id, staker_profile_id)
    local total_shares_before = get_total_shares(vault_profile_id)
    local unstake_res = balance.create_unstake(staker_profile_id, vault_profile_id, vault_wallet, shares, DEFAULT_EXCHANGE_ID, 0)
    t.assert_is(unstake_res["error"], nil)
    t.assert_is_not(unstake_res["res"], nil)
    local unstake_id = unstake_res["res"].id
    t.assert_is_not(unstake_id, nil)
    local prefix, id_str = unstake_id:match("^(.)_(%d+)$")
    t.assert_is(prefix, "u")
    local id_num = tonumber(id_str)
    t.assert_is_not(id_num, nil)
    local res = balance.process_unstakes(vault_profile_id, id_num, id_num, current_nav, withdrawable_balance,
        performance_fee, treasurer_profile_id, total_shares_before, DEFAULT_EXCHANGE_ID)
    t.assert_is(res["error"], nil)
    t.assert_is(res["res"][1], staker_profile_id)
    local unstake = box.space.balance_operations:get(unstake_id)
    t.assert_is(unstake.status, config.params.BALANCE_STATUS.SUCCESS)
    local shares_after = get_shares(vault_profile_id, staker_profile_id)
    assert_close_to(shares_after, shares_before - shares)
    local total_shares_after = get_total_shares(vault_profile_id)
    assert_close_to(total_shares_after, total_shares_before - shares)
    local staker_balance_after = get_balance(staker_profile_id)
    local vault_balance_after = get_balance(vault_profile_id)
    local treasurer_balance_after = get_balance(treasurer_profile_id)
    local balance_sum_before = staker_balance_before + vault_balance_before + treasurer_balance_before
    local balance_sum_after = staker_balance_after + vault_balance_after + treasurer_balance_after
    assert_close_to(balance_sum_after, balance_sum_before)
    local unstake_value = current_nav * shares / total_shares_before
    local unstake_price = current_nav / total_shares_before
    local performance_charge
    if unstake_price > entry_price then
        performance_charge =
            (performance_fee * unstake_value * (unstake_price - entry_price)) / unstake_price
    else
        performance_charge = ZERO
    end
    assert_close_to(staker_balance_after, staker_balance_before + unstake_value - performance_charge)
    assert_close_to(treasurer_balance_after, treasurer_balance_before + performance_charge)
    assert_close_to(vault_balance_after, vault_balance_before - unstake_value)
end

function onboard_vault(wallet)
    local res = profile.profile.create("vault", "active", wallet, DEFAULT_EXCHANGE_ID)
    t.assert_is(res["error"], nil)
    local vault_profile = res["res"]
    vault_profile_id = vault_profile.id
    balance.init_vault(vault_profile_id, manager_profile_id, treasurer_profile_id, performance_fee)
end

function debugInfo(balance_op, desc)
    log.info("<%s>.id %s", desc, balance_op.id)
    log.info("<%s>.status %s", desc, balance_op.status)
    log.info("<%s>.reason %s", desc, balance_op.reason)
    log.info("<%s>.txhash %s", desc, balance_op.txhash)
    log.info("<%s>.profile_id %s", desc, tostring(balance_op.profile_id))
    log.info("<%s>.wallet %s", desc, balance_op.wallet)
    log.info("<%s>.ops_type %s", desc, balance_op.ops_type)
    log.info("<%s>.ops_id2 %s", desc, balance_op.ops_id2)
    log.info("<%s>.amount %s", desc, tostring(balance_op.amount))
    log.info("<%s>.timestamp %s", desc, tostring(balance_op.timestamp))
    log.info("<%s>.due_block %s", desc, tostring(balance_op.due_block))
end
