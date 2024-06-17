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
local wdm = require('app.wdm')

local z = decimal.new(0)
require('app.config.constants')

local g = t.group('balance')

local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg {
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    archiver.init_sequencer("balance")

    -- truncate tiers to not break prev tests we have sep test for it
 
    profile.init_spaces({})
    balance.init_spaces(0)

    balance.add_to_contract_map("", 0, DEFAULT_EXCHANGE_ID)

    box.space.wd_tier:truncate()
end)

g.after_each(function(cg)
    balance.test_clear_spaces()
end)

g.test_deposit_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local txhash = "0x321"
    local deposit_id = "d_1"
    create_and_process_deposit(amount, profile_id, wallet, txhash, deposit_id)
end

g.test_withdrawal_before_deposit_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local withdrawal = create_withdrawal(profile_id, wallet, amount)
end

g.test_deposit_unknown_then_onboarded_flow = function(cg)
    -- Create 3 deposits for the first unknown wallet and
    -- 1 for the second then onboard the first one.
    -- The balance of the onboarded wallet should increase by
    -- the sum of its 3 deposits, and their status should all be
    -- success. The fourth deposit should still be unknown.
    local wallet = "0x123"
    local otherwallet = "0x234"
    local amount1 = decimal.new(10)
    local amount2 = decimal.new(150)
    local amount3 = decimal.new(50)
    local amount4 = decimal.new(250)
    local txhash1 = "0x321"  -- Same tx for deposits 1 and 2, with the deposit pool
    local txhash2 = "0x321"  -- there can be multiple deposits in the same tx.
    local txhash3 = "0x4321" -- Multiple deposits for different wallets can also
    local txhash4 = "0x4321" -- be in the same tx.
    local deposit_id1 = "d_1"
    local deposit_id2 = "d_2"
    local deposit_id3 = "d_3"
    local deposit_id4 = "d_4"
    create_and_process_deposits_unknown_then_onboard(wallet, otherwallet,
        amount1, amount2, amount3, amount4,
        txhash1, txhash2, txhash3, txhash4,
        deposit_id1, deposit_id2, deposit_id3, deposit_id4)
end

g.test_cancel_deposit_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local txhash = "0x321"
    local balance_before = get_balance(profile_id)
    local res = balance.create_deposit(profile_id, wallet, amount, txhash, "", 0)
    assert_success(res)
    local exist = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is_not(exist, nil)
    local id = exist.id
    res = balance.pending_deposit_canceled(id)
    assert_success(res)
    local deposit = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is(deposit.id, id)
    t.assert_is(deposit.ops_id2, id)
    t.assert_is(deposit.status, config.params.BALANCE_STATUS.CANCELED)
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
end

g.test_cant_cancel_succesful_deposit = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local txhash = "0x321"
    local deposit_id = "d_1"
    local balance_before = get_balance(profile_id)
    local res = balance.create_deposit(profile_id, wallet, amount, txhash, "", 0)
    assert_success(res)
    local exist = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is_not(exist, nil)
    local id = exist.id
    res = balance.process_deposit(profile_id, wallet, deposit_id, amount, txhash, false, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    res = balance.pending_deposit_canceled(id)
    assert_failure(res)
    local deposit = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is(deposit.id, id)
    t.assert_is(deposit.ops_id2, deposit_id)
    t.assert_is(deposit.status, config.params.BALANCE_STATUS.SUCCESS)
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before + amount)
end

g.test_deposit_unknown_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local txhash = "0x321"
    local deposit_id = "d_1"

    --process deposit for unknown profile, should not change balance for any profile
    local balance_before = get_balance(profile_id)
    log.info("unknown amount in %s", tostring(amount))
    local res = balance.process_deposit_unknown(wallet, deposit_id, amount, txhash, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is(deposit.profile_id, ANONYMOUS_ID_FOR_UNKNOWN)
    t.assert_is(deposit.ops_id2, deposit_id)
    t.assert_is(deposit.status, config.params.BALANCE_STATUS.UNKNOWN)
    local id = deposit.id
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)

    --resolve the deposit to a known profile, should change balance for that profile
    res = balance.resolve_unknown(profile_id, deposit_id)
    deposit = box.space.balance_operations.index.txhash:min { txhash }
    -- debugInfo(deposit, "resolved")
    t.assert_is(deposit.profile_id, profile_id)
    t.assert_is(deposit.id, id)
    t.assert_is(deposit.ops_id2, deposit_id)
    t.assert_is(deposit.status, config.params.BALANCE_STATUS.SUCCESS)
    balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before + amount)
end

local function debugInfo(balance_op, desc)
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

g.test_claim_withdrawal_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local deposit_txhash = "0x321"
    local deposit_id = "d_1"
    create_and_process_deposit(amount, profile_id, wallet, deposit_txhash, deposit_id)
    local withdrawal = create_withdrawal(profile_id, wallet, amount)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PENDING)
    t.assert_is(withdrawal.due_block, 0)

    --update pending should set the due block
    balance.update_pending_withdrawals(1, 10, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PENDING)
    t.assert_is(withdrawal.due_block, 10)

    --update pending when due block reached should make it claimable
    balance.update_pending_withdrawals(10, 20, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMABLE)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)

    --update pending again should do nothing
    balance.update_pending_withdrawals(20, 30, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMABLE)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)

    --claim should change status to claiming and not change balance
    local balance_before = get_balance(profile_id)
    balance.claim_withdrawal(profile_id, withdrawal.id)
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMING)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)

    --processing should change status to processing and set txhash
    local txhash = "0x432"
    balance.processing_withdrawal(profile_id, txhash, withdrawal.id)
    balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PROCESSING)
    t.assert_is(withdrawal.txhash, txhash)

    --claim again should do nothing
    balance.claim_withdrawal(profile_id, withdrawal.id)
    balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PROCESSING)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)
    t.assert_is(withdrawal.txhash, txhash)

    --processing again should set txhash again
    local txhash2 = "0x543"
    balance.processing_withdrawal(profile_id, txhash2, withdrawal.id)
    balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PROCESSING)
    t.assert_is(withdrawal.txhash, txhash2)
end

g.test_cancel_withdrawal_flow = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local deposit_txhash = "0x321"
    local deposit_id = "d_1"
    create_and_process_deposit(amount, profile_id, wallet, deposit_txhash, deposit_id)
    local withdrawal = create_withdrawal(profile_id, wallet, amount)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PENDING)
    t.assert_is(withdrawal.due_block, 0)

    --update pending should set the due block
    balance.update_pending_withdrawals(1, 10, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.PENDING)
    t.assert_is(withdrawal.due_block, 10)

    --update pending when due block reached should make it claimable
    balance.update_pending_withdrawals(10, 20, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMABLE)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)

    --update pending again should do nothing
    balance.update_pending_withdrawals(20, 30, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMABLE)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)
    local amount = withdrawal.amount

    --cancel should change status to canceled and increase balance
    local balance_before = get_balance(profile_id)
    balance.cancel_withdrawal(profile_id, withdrawal.id)
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before + amount)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CANCELED)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)

    --claim should now fail
    balance_before = get_balance(profile_id)
    local res = balance.claim_withdrawal(profile_id, withdrawal.id)
    assert_failure(res)
    balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CANCELED)
    t.assert_is(withdrawal.due_block, BLOCK_NUM_NEVER)
end

g.test_check_withdraw_allowed = function(cg)
    local amount = decimal.new(10)
    local profile_id = 1
    local wallet = "0x123"
    local deposit_txhash = "0x321"
    local withdrawal_txhash = "0x432"
    local deposit_id = "d_1"
    create_and_process_deposit(amount, profile_id, wallet, deposit_txhash, deposit_id)

    --withdrawal should be allowed initially
    local res = balance.check_withdraw_allowed(profile_id)
    assert_success(res)

    --after creating a withdrawal further withdrawals should not be allowed
    local withdrawal = create_withdrawal(profile_id, wallet, amount)
    res = balance.check_withdraw_allowed(profile_id)
    assert_failure(res)

    --update pending should set the due block
    balance.update_pending_withdrawals(1, 10, "")
    --update pending when due block reached should make it claimable
    balance.update_pending_withdrawals(10, 20, "")
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMABLE)

    --with a claimable withdrawal further withdrawals should not be allowed
    res = balance.check_withdraw_allowed(profile_id)
    assert_failure(res)

    --claim should change status to claiming
    local res = balance.claim_withdrawal(profile_id, withdrawal.id)
    assert_success(res)
    withdrawal = box.space.balance_operations:get(withdrawal.id)
    t.assert_is_not(withdrawal, nil)
    t.assert_is(withdrawal.status, config.params.BALANCE_STATUS.CLAIMING)

    --with a claiming withdrawal further withdrawals should not be allowed
    res = balance.check_withdraw_allowed(profile_id)
    assert_failure(res)
end

-- test for type_due_block index where you have a lot of balance_ops of
-- different type with 0 due_block and iter through them
-- balance.update_pending_withdrawals relies on this index
g.test_type_due_block_index = function(cg)
    -- insert some tuples, of different types and due_blocks
    for i = 0, 100 do
        insert_tuple(i)
    end

    -- see if we can find all tuples with type withdrawal and due_block = 0
    local eq0_seen = {}
    for i = 0, 100 do
        eq0_seen["w_" .. tostring(i)] = false
    end
    for _, v in box.space.balance_operations.index.type_contract_due_block:pairs({ config.params.BALANCE_TYPE.WITHDRAWAL, "", 0 }, { iterator = 'EQ' }) do
        if v.ops_type ~= config.params.BALANCE_TYPE.WITHDRAWAL then
            break
        end
        t.assert_not(eq0_seen[v.id], "duplicate id")
        eq0_seen[v.id] = true
    end
    for i = 0, 100 do
        t.assert_is(eq0_seen["w_" .. tostring(i)], i % 9 == 3)
    end

    -- see if we can find all tuples with type withdrawal and due_block <= 11
    local le11_seen = {}
    for i = 0, 100 do
        le11_seen["w_" .. tostring(i)] = false
    end
    for _, v in box.space.balance_operations.index.type_contract_due_block:pairs({ config.params.BALANCE_TYPE.WITHDRAWAL, "", 11 }, { iterator = 'LE' }) do
        if v.ops_type ~= config.params.BALANCE_TYPE.WITHDRAWAL then
            break
        end
        t.assert_not(le11_seen[v.id], "duplicate id")
        le11_seen[v.id] = true
    end
    for i = 0, 100 do
        t.assert_is(le11_seen["w_" .. tostring(i)], i ~= 11 and (i % 9 == 2 or i % 9 == 3))
    end

    -- see if we can correctly update due blocks of withdrawals
    -- due_block 0 => 10
    -- due_block 1..10 => BLOCK_NUM_NEVER (10^15)
    balance.update_pending_withdrawals(10, 100, "")
    for i = 0, 100 do
        local id = "w_" .. tostring(i)
        local b_op = box.space.balance_operations:get(id)
        if i == 11 then
            t.assert_is(b_op.due_block, 20)
        elseif i % 9 == 2 then
            t.assert_is(b_op.due_block, BLOCK_NUM_NEVER)
        elseif i % 9 == 3 then
            t.assert_is(b_op.due_block, 100)
        else
            t.assert_is(b_op.due_block, 0)
        end
    end

    -- see if we can find all tuples with type withdrawal and due_block <= 100
    local le100_seen = {}
    for i = 0, 100 do
        le100_seen["w_" .. tostring(i)] = false
    end
    for _, v in box.space.balance_operations.index.type_contract_due_block:pairs({ config.params.BALANCE_TYPE.WITHDRAWAL, "", 100 }, { iterator = 'LE' }) do
        if v.ops_type ~= config.params.BALANCE_TYPE.WITHDRAWAL then
            break
        end
        t.assert_not(le100_seen[v.id], "duplicate id")
        le100_seen[v.id] = true
    end
    for i = 0, 100 do
        t.assert_is(le100_seen["w_" .. tostring(i)], i == 11 or i % 9 == 3)
    end
end

function insert_tuple(i)
    local type
    local due_block = 0
    local id = "w_" .. tostring(i)
    if i % 9 == 0 then
        type = config.params.BALANCE_TYPE.DEPOSIT
    elseif i % 9 == 1 then
        type = config.params.BALANCE_TYPE.CREDIT
    elseif i % 9 == 2 then
        type = config.params.BALANCE_TYPE.WITHDRAWAL
        if i == 11 then
            due_block = 20
        else
            due_block = 10
        end
    elseif i % 9 == 3 then
        type = config.params.BALANCE_TYPE.WITHDRAWAL
    elseif i % 9 == 4 then
        type = config.params.BALANCE_TYPE.WITHDRAW_CREDIT
    elseif i % 9 == 5 then
        type = config.params.BALANCE_TYPE.FUNDING
    elseif i % 9 == 6 then
        type = config.params.BALANCE_TYPE.PNL
    elseif i % 9 == 7 then
        type = config.params.BALANCE_TYPE.FEE
    elseif i % 9 == 8 then
        type = config.params.BALANCE_TYPE.WITHDRAW_FEE
    end
    local res, err = archiver.insert(box.space.balance_operations, {
        id,
        config.params.BALANCE_STATUS.PENDING,
        "",
        "",
        1,
        "0x123",
        type,
        id,
        decimal.new(10),
        123,
        due_block,

        DEFAULT_EXCHANGE_ID,
        0,
        "",
    })
end

g.test_last_processed = function(cg)
    local res = balance.set_last_processed_block_number("3", "", 0, "")
    assert_success(res)
    local res = balance.get_last_processed_block_number("", 0, "")
    assert_success(res)
    t.assert_is(res["res"], "3")
end

function get_balance(profile_id)
    local b_sum = box.space.balance_sum:get(profile_id)
    if b_sum == nil then
        return 0
    end
    return b_sum.balance
end

function create_and_process_deposit(amount, profile_id, wallet, txhash, deposit_id)
    local balance_before = get_balance(profile_id)
    local res = balance.create_deposit(profile_id, wallet, amount, txhash, "", 0)
    assert_success(res)
    local exist = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is_not(exist, nil)
    local id = exist.id
    res = balance.process_deposit(profile_id, wallet, deposit_id, amount, txhash, false, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit = box.space.balance_operations.index.txhash:min { txhash }
    t.assert_is(deposit.id, id)
    t.assert_is(deposit.ops_id2, deposit_id)
    local balance_after = get_balance(profile_id)
    t.assert_is(balance_after, balance_before + amount)
end

function create_and_process_deposits_unknown_then_onboard(
    wallet, otherwallet,
    amount1, amount2, amount3, amount4,
    txhash1, txhash2, txhash3, txhash4,
    deposit_id1, deposit_id2, deposit_id3, deposit_id4)
    -- first process all the deposits
    local res = balance.process_deposit_unknown(wallet, deposit_id1, amount1, txhash1, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit1 = box.space.balance_operations:get(deposit_id1)
    t.assert_is(deposit1.ops_id2, deposit_id1)
    t.assert_is(deposit1.status, config.params.BALANCE_STATUS.UNKNOWN)
    res = balance.process_deposit_unknown(wallet, deposit_id2, amount2, txhash2, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit2 = box.space.balance_operations:get(deposit_id2)
    t.assert_is(deposit2.ops_id2, deposit_id2)
    t.assert_is(deposit2.status, config.params.BALANCE_STATUS.UNKNOWN)
    res = balance.process_deposit_unknown(wallet, deposit_id3, amount3, txhash3, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit3 = box.space.balance_operations:get(deposit_id3)
    t.assert_is(deposit3.ops_id2, deposit_id3)
    t.assert_is(deposit3.status, config.params.BALANCE_STATUS.UNKNOWN)
    res = balance.process_deposit_unknown(otherwallet, deposit_id4, amount4, txhash4, DEFAULT_EXCHANGE_ID, 0, "")
    assert_success(res)
    local deposit4 = box.space.balance_operations:get(deposit_id4)
    t.assert_is(deposit4.ops_id2, deposit_id4)
    t.assert_is(deposit4.status, config.params.BALANCE_STATUS.UNKNOWN)
    
    -- now onboard the first wallet
    log.info(box.space.balance_operations.index)
    res = profile.profile.create("trader", "active", wallet, DEFAULT_EXCHANGE_ID)
    assert_success(res)
    local new_profile = res.res
    local balance_after = get_balance(new_profile.id)
    t.assert_is(balance_after, amount1 + amount2 + amount3)
    local deposit1_after = box.space.balance_operations:get(deposit_id1)
    t.assert_is(deposit1_after.status, config.params.BALANCE_STATUS.SUCCESS)
    local deposit2_after = box.space.balance_operations:get(deposit_id2)
    t.assert_is(deposit2_after.status, config.params.BALANCE_STATUS.SUCCESS)
    local deposit3_after = box.space.balance_operations:get(deposit_id3)
    t.assert_is(deposit3_after.status, config.params.BALANCE_STATUS.SUCCESS)
    local deposit4_after = box.space.balance_operations:get(deposit_id4)
    t.assert_is(deposit4_after.status, config.params.BALANCE_STATUS.UNKNOWN)
    -- now onboard the other wallet
    res = profile.profile.create("trader", "active", otherwallet, DEFAULT_EXCHANGE_ID)
    assert_success(res)
    local other_profile = res.res
    local other_balance_after = get_balance(other_profile.id)
    t.assert_is(other_balance_after, amount4)
    -- first profile should not have changed balance
    t.assert_is(balance_after, amount1 + amount2 + amount3)
    local deposit4_final = box.space.balance_operations:get(deposit_id4)
    t.assert_is(deposit4_final.status, config.params.BALANCE_STATUS.SUCCESS)
end

function create_withdrawal(profile_id, wallet, amount)
    local balance_before = get_balance(profile_id)
    local res = balance.create_withdrawal(profile_id, wallet, amount, DEFAULT_EXCHANGE_ID)
    assert_success(res)
    local withdrawal = res["res"]
    t.assert_is_not(withdrawal, nil)
    local balance_after = get_balance(profile_id)
    local expected_balance = balance_before - amount
    local message = string.format("The balance after withdrawal was %s but should be the balance before, which was %s, minus the amount %s, which is %s", tostring(balance_after), tostring(balance_before), tostring(amount), tostring(expected_balance))
    t.assert_is(balance_after, expected_balance, message)
    return withdrawal
end

function assert_success(res)
    t.assert_is(res["error"], nil)
end

function assert_failure(res)
    t.assert_is_not(res["error"], nil)
end
