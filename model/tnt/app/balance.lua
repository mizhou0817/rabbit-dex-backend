local checks = require('checks')
local clock = require('clock')
local d = require('app.data')
local decimal = require('decimal')
local fun = require('fun')
local json = require("json")
local log = require('log')
local uuid = require('uuid')

local archiver = require('app.archiver')
local config = require('app.config')
local dml = require('app.dml')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')
local wdm = require('app.wdm')

require("app.config.constants")
require('app.errcodes')

local BalanceError = errors.new_class("BALANCE")
local VaultError = errors.new_class("VAULT")

local SETTLEMENT_STATUS_ID = 0
local SETTLEMENT_STATUS_DUMMY_CONTRACT = ""
local balance = {
    _shard_num = 0
}

local function _do_notif_profile(profile_id, ops_table)
    local channel = "account@" .. tostring(profile_id)
    local update = {
        id = profile_id,
        balance_operations = ops_table
    }

    local json_update = json.encode({ data = update })
    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)

    update = nil

    return nil
end

local function multiple_notif_profile(profile_id, multiple_balance_ops)
    return _do_notif_profile(profile_id, multiple_balance_ops)
end

local function notif_profile(profile_id, balance_op)
    return _do_notif_profile(profile_id, { balance_op })
end


local function create_lock_space(space_name)
    local lock_space = box.schema.space.create(
        space_name,
        { if_not_exists = true }
    )
    lock_space:format({
        { name = 'id',        type = 'unsigned' },
        { name = 'locked',    type = 'boolean' },
        { name = 'timestamp', type = 'number' }
    })

    lock_space:create_index(
        'primary',
        {
            unique = true,
            parts = { { field = 'id' } },
            if_not_exists = true
        }
    )
end

function balance.init_artifacts()
    local withdraw_white_list = box.schema.space.create('withdraw_white_list', { if_not_exists = true })
    withdraw_white_list:format({
        { name = 'profile_id', type = 'unsigned' }
    })
    withdraw_white_list:create_index('primary', {
        unique = true,
        parts = { { field = 'profile_id' } },
        if_not_exists = true
    })
end

function balance.init_spaces(shard_num)
    balance.init_artifacts()

    -- apply drop migrations
    local g_space = box.space['global_settlement_status']
    if g_space == nil then
        local ddl = require('app.ddl')
        require('migrations.common.eid_balance_migrations').drop_migration(ddl)
    end

    balance._shard_num = shard_num
    box.schema.sequence.create('withdrawal_id_sequence', { start = 0, min = 0, if_not_exists = true })
    box.schema.sequence.create('unstake_id_sequence', { start = 0, min = 0, if_not_exists = true })

    local exchange_wallets = box.schema.space.create('exchange_wallets', { if_not_exists = true })
    exchange_wallets:format({
        { name = 'wallet_id',    type = 'unsigned' },
        { name = 'balance',      type = 'decimal' },
        { name = 'last_updated', type = 'number' }
    })
    exchange_wallets:create_index('primary', {
        unique = true,
        parts = { { field = 'wallet_id' } },
        if_not_exists = true
    })

    local balance_sum = box.schema.space.create('balance_sum', { if_not_exists = true })
    balance_sum:format({
        { name = 'profile_id',   type = 'unsigned' },
        { name = 'balance',      type = 'decimal' },
        { name = 'last_updated', type = 'number' }
    })

    balance_sum:create_index('primary', {
        unique = true,
        parts = { { field = 'profile_id' } },
        if_not_exists = true
    })

    local yield, err = archiver.create('yield', { if_not_exists = true }, {
        { name = 'yield_id',         type = 'string' },
        { name = 'amount',           type = 'decimal' },
        { name = 'paid',             type = 'decimal' },
        { name = 'recipients',       type = 'unsigned' },
        { name = 'tx_hash',          type = 'string' },
        { name = 'timestamp',        type = 'number' },
        { name = 'exchange_id',      type = 'string' },
        { name = 'chain_id',         type = 'unsigned' },
        { name = 'exchange_address', type = 'string' },
    }, {
        unique = true,
        parts = { { field = 'yield_id' } },
        if_not_exists = true
    })
    if err ~= nil then
        log.error(BalanceError:new(err))
        error(err)
    end

    local balance_operations, err = archiver.create('balance_operations', { if_not_exists = true }, {
        { name = 'id',               type = 'string' }, -- this is the unique ID for withdraw
        { name = 'status',           type = 'string' },
        { name = 'reason',           type = 'string' },
        { name = 'txhash',           type = 'string' }, -- it's just some meta, for detect transaction onchain
        { name = 'profile_id',       type = 'unsigned' },
        -- if profile_id is known then wallet comes from that, and the
        -- wallet in balance_operations can be an empty string, but in case
        -- of a deposit straight to the L1 contract (not via front end) no
        -- profile_id is known initially (and there may not be one if the
        -- depositor has not onboarded) so wallet is used
        { name = 'wallet',           type = 'string' },
        { name = 'ops_type',         type = 'string' },
        { name = 'ops_id2',          type = 'string' }, -- this is the deposit/withdrawal id to never approve the same deposit or withdrawal twice
        { name = 'amount',           type = 'decimal' },
        { name = 'timestamp',        type = 'number' },
        { name = 'due_block',        type = 'unsigned' },
        { name = 'exchange_id',      type = 'string' },
        { name = 'chain_id',         type = 'unsigned' },
        { name = 'contract_address', type = 'string' },
    }, {
        unique = true,
        parts = { { field = 'id' } },
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(BalanceError:new(err))
        error(err)
    end

    -- Initially ops_id2 can be set equal to ops_id. For deposits it will eventually
    -- record the deposit id 'd_123', etc., but the deposit id is not known initially
    -- as it is assigned by the L1 contract when it starts processing the deposit
    balance_operations:create_index('ops_id2', {
        parts = { { field = 'ops_id2' } },
        unique = true,
        if_not_exists = true
    })

    balance_operations:create_index('profile_id_type_status_amount',
        {
            parts = { { field = 'profile_id' }, { field = 'ops_type' }, { field = 'status' }, { field = 'amount' } },
            unique = false,
            if_not_exists = true
        })

    balance_operations:create_index('txhash', {
        parts = { { field = 'txhash' } },
        unique = false,
        if_not_exists = true
    })

    balance_operations:create_index('reason', {
        parts = { { field = 'reason' } },
        unique = false,
        if_not_exists = true
    })

    -- if it has the due_block field, add the type_due_block index
    local format = balance_operations:format()
    for _, field in ipairs(format) do
        if field.name == 'exchange_id' then
            balance_operations:create_index('type_contract_due_block',
                {
                    parts = {
                        { field = 'ops_type' },
                        { field = 'contract_address' },
                        { field = 'due_block' } },
                    unique = false,
                    if_not_exists = true
                })

            balance_operations:create_index('exchange_id',
                {
                    parts = { { field = 'exchange_id' } },
                    unique = false,
                    if_not_exists = true
                })

            balance_operations:create_index('chain_id_contract_address',
                {
                    parts = { { field = 'chain_id' }, { field = 'contract_address' } },
                    unique = false,
                    if_not_exists = true
                })

            balance_operations:create_index('exchange_id_wallet_type_status',
                {
                    parts = { { field = 'exchange_id' }, { field = 'wallet' }, { field = 'ops_type' }, { field = 'status' } },
                    unique = false,
                    if_not_exists = true
                })

            balance_operations:create_index('type_status_exchange_id_chain_id', {
                parts = { { field = 'ops_type' }, { field = 'status' }, { field = "exchange_id" }, { field = "chain_id" } },
                unique = false,
                if_not_exists = true
            })
        end
    end

    local global_settlement_status = box.schema.space.create('global_settlement_status', { if_not_exists = true })
    global_settlement_status:format({
        { name = 'id',                    type = 'unsigned' },
        { name = 'last_processed_block',  type = 'string' },
        { name = 'withdrawals_suspended', type = 'boolean' },
    })

    global_settlement_status:create_index('primary', {
        unique = true,
        parts = { { field = 'id' } },
        if_not_exists = true
    })

    local vaults
    vaults, err = archiver.create('vaults', { if_not_exists = true }, {
        { name = 'vault_profile_id',     type = 'unsigned' },
        { name = 'manager_profile_id',   type = 'unsigned' },
        { name = 'treasurer_profile_id', type = 'unsigned' },
        { name = 'performance_fee',      type = 'decimal' },
        { name = 'status',               type = 'string' },
        { name = 'total_shares',         type = 'decimal' },
        { name = 'vault_name',           type = 'string' },
        { name = 'manager_name',         type = 'string' },
        { name = 'initialised_at',       type = 'number' },
    }, {
        unique = true,
        parts = { { field = 'vault_profile_id' } },
        if_not_exists = true,
    })

    local vault_holdings
    vault_holdings, err = archiver.create('vault_holdings', { if_not_exists = true }, {
        { name = 'vault_profile_id',  type = 'unsigned' },
        { name = 'staker_profile_id', type = 'unsigned' },
        { name = 'shares',            type = 'decimal' },
        { name = 'entry_nav',         type = 'decimal' },
        { name = 'entry_price',       type = 'decimal' },
    }, {
        unique = true,
        parts = { { field = 'vault_profile_id' }, { field = 'staker_profile_id' } },
        if_not_exists = true,
    })

    vault_holdings:create_index('vault', {
        parts = { { field = 'vault_profile_id' } },
        unique = false,
        if_not_exists = true
    })

    vault_holdings:create_index('staker', {
        parts = { { field = 'staker_profile_id' } },
        unique = false,
        if_not_exists = true
    })

    local max_withdraw_amount = box.schema.space.create('max_withdraw_amount', { if_not_exists = true })
    max_withdraw_amount:format({
        { name = 'key',    type = 'string' },
        { name = 'amount', type = 'unsigned' },
    })

    max_withdraw_amount:create_index('primary', {
        unique = true,
        parts = { { field = 'key' } },
        if_not_exists = true
    })



    local contract_map = box.schema.space.create('contract_map', { if_not_exists = true })
    contract_map:format({
        { name = 'contract_address', type = 'string' },
        { name = 'chain_id',         type = 'unsigned' },
        { name = 'exchange_id',      type = 'string' }
    })

    contract_map:create_index('primary', {
        unique = true,
        parts = { { field = 'contract_address' }, { field = 'chain_id' } },
        if_not_exists = true
    })

    contract_map:create_index('by_exchange_id', {
        unique = false,
        parts = { { field = 'exchange_id' } },
        if_not_exists = true
    })

    create_lock_space('withdraw_lock')
    create_lock_space('stake_lock')
    create_lock_space('unstake_lock')

    return true
end

function balance.max_withdraw_amount()
    local exist = box.space.max_withdraw_amount:get("default")
    if exist == nil then
        return { res = 0, error = nil }
    end

    return { res = exist.amount, error = nil }
end

function balance.add_to_contract_map(contract_address, chain_id, exchange_id)
    checks('string', 'number', 'string')

    local status, res = pcall(
        function()
            return box.space.contract_map:replace {
                contract_address,
                chain_id,
                exchange_id,
            }
        end)
    if status == false then
        return { res = nil, error = res }
    end

    return { res = res, error = nil }
end

-- Current version support 1 contract per exchange_id
function balance.contract_by_exchange_id(exchange_id)
    checks('string')

    return box.space.contract_map.index.by_exchange_id:min(exchange_id)
end

function balance.list_operations(profile_id, offset, limit)
    checks('number', 'number', 'number')

    local res = box.space.balance_operations.index.profile_id_type_status_amount:select({ profile_id },
        { iterator = 'EQ', offset = offset, limit = limit })

    return { res = res, error = nil }
end

function balance.list_operations_of_type_status(profile_id, op_type, op_status, offset, limit)
    checks('number', 'string', 'string', 'number', 'number')

    local status, res = pcall(
        function()
            return box.space.balance_operations.index.profile_id_type_status_amount:select(
                { profile_id, op_type, op_status },
                { iterator = 'EQ', offset = offset, limit = limit })
        end
    )
    if not status then
        return { res = nil, error = res }
    end
    return { res = res, error = nil }
end

function balance.init_vault(vault_profile_id, vault_name, manager_profile_id, manager_name, treasurer_profile_id,
                            performance_fee)
    checks('number', 'string', 'number', 'string', 'number', 'decimal')
    local tm = time.now()
    local _, err = archiver.upsert(
        box.space.vaults,
        {
            vault_profile_id,
            manager_profile_id,
            treasurer_profile_id,
            performance_fee,
            config.params.VAULT_STATUS.ACTIVE,
            decimal.new(0),
            vault_name,
            manager_name,
            tm,
        },
        {
            { '=', 'manager_profile_id',   manager_profile_id },
            { '=', 'treasurer_profile_id', treasurer_profile_id },
            { '=', 'performance_fee',      performance_fee },
            { '=', 'vault_name',           vault_name },
            { '=', 'manager_name',         manager_name },
        }
    )
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    return { res = nil, error = nil }
end

function balance.reactivate_vault(vault_profile_id)
    checks('number')
    local tm = time.now()
    local res, err = archiver.update(
        box.space.vaults, vault_profile_id,
        {
            { '=', 'status',         config.params.VAULT_STATUS.ACTIVE },
            { '=', 'initialised_at', tm }
        }
    )
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    return { res = nil, error = nil }
end

function balance.check_withdraw_allowed(profile_id)
    checks('number')

    local listed = box.space.withdraw_white_list:get(profile_id)
    if listed ~= nil then
        return { res = nil, error = nil }
    end

    box.begin()

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.PENDING)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist == nil then
        exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
            config.params.BALANCE_STATUS.CLAIMABLE)
        if err ~= nil then
            box.rollback()
            return { res = nil, error = err }
        end
    end
    if exist == nil then
        exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
            config.params.BALANCE_STATUS.CLAIMING)
        if err ~= nil then
            box.rollback()
            return { res = nil, error = err }
        end
    end
    if exist == nil then
        exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
            config.params.BALANCE_STATUS.PROCESSING)
        if err ~= nil then
            box.rollback()
            return { res = nil, error = err }
        end
    end

    if exist ~= nil then
        box.rollback()
        return { res = nil, error = "PENDING_WITHDRAW_EXIST" }
    end

    box.commit()

    return { res = nil, error = nil }
end

function balance.cancel_all_withdrawals(profile_id)
    checks('number')

    for _, ops in box.space.balance_operations.index.profile_id_type_status_amount:pairs({ profile_id }, { iterator = 'EQ' }) do
        if ops.profile_id ~= profile_id then
            break
        end

        balance.cancel_withdrawal(profile_id, ops.id)
    end

    return { res = nil, error = nil }
end

-- We allow to cancel only:
-- PENDING: means that transaction just created
-- CLAIMABLE: means that 6 hours passed, but it was never signed
function balance.cancel_withdrawal(profile_id, bops_id)
    checks('number', 'string')

    box.begin()

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.PENDING,
        bops_id)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist == nil then
        exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
            config.params.BALANCE_STATUS.CLAIMABLE,
            bops_id)
        if err ~= nil then
            box.rollback()
            return { res = nil, error = err }
        end
    end
    if exist == nil then
        box.rollback()
        return { res = nil, error = "NO_PENDING_WITHDRAWAL" }
    end

    local res
    res, err = archiver.update(box.space.balance_operations, exist.id, {
        { '=', 'status', config.params.BALANCE_STATUS.CANCELED } })
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if res == nil then
        box.rollback()
        return { res = nil, error = "WITHDRAWAL_NOT_FOUND" }
    end
    local b_ops = res:tomap({ names_only = true })

    err = balance.increase_balance_sum(profile_id, exist.amount)
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = tostring(err) }
    end

    box.commit()

    notif_profile(profile_id, b_ops)
    return { res = nil, error = nil }
end

function balance.claim_withdrawal(profile_id, bops_id)
    checks('number', 'string')

    box.begin()

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.CLAIMING,
        bops_id)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist ~= nil then
        box.commit()
        return { res = exist, error = nil }
    end

    exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.PROCESSING,
        bops_id)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist ~= nil then
        box.commit()
        return { res = exist, error = nil }
    end

    exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.CLAIMABLE,
        bops_id)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist == nil then
        box.rollback()
        return { res = nil, error = "NO_CLAIMABLE_WITHDRAWAL" }
    end

    local status, res
    res, err = archiver.update(box.space.balance_operations, exist.id, {
        { '=', 'status', config.params.BALANCE_STATUS.CLAIMING } })
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if res == nil then
        box.rollback()
        return { res = nil, error = "WITHDRAWAL_NOT_FOUND" }
    end

    box.commit()

    local b_ops = res:tomap({ names_only = true })
    notif_profile(profile_id, b_ops)

    return { res = res, error = nil }
end

function balance.find_bop(type, profile_id, status, bops_id)
    checks("string", "number", "string", "?string")

    local exist

    if bops_id ~= "" and bops_id ~= nil then
        --TODO: is that faster just to have secondary index id_status ?
        exist = box.space.balance_operations:get(bops_id)
        if exist ~= nil then
            if exist.status ~= status then
                return nil, nil
            end
        end
    else
        exist = box.space.balance_operations.index.profile_id_type_status_amount:min {
            profile_id,
            type,
            status,
        }
    end

    if exist ~= nil and exist.profile_id ~= profile_id then
        log.error({
            message = string.format("%s: check_withdraw_allowed: profile_id=%d exist.profile_id=%d", ERR_INTEGRITY_ERROR,
                profile_id, exist.profile_id),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, ERR_INTEGRITY_ERROR
    end
    return exist, nil
end

function balance.acquire_withdraw_lock(profile_id)
    checks('number')
    return balance.acquire_lock(box.space.withdraw_lock, profile_id)
end

function balance.release_withdraw_lock(profile_id)
    checks('number')
    return balance.release_lock(box.space.withdraw_lock, profile_id)
end

function balance.acquire_stake_lock(profile_id)
    checks('number')
    return balance.acquire_lock(box.space.stake_lock, profile_id)
end

function balance.release_stake_lock(profile_id)
    checks('number')
    return balance.release_lock(box.space.stake_lock, profile_id)
end

function balance.acquire_unstake_lock(profile_id)
    checks('number')
    return balance.acquire_lock(box.space.unstake_lock, profile_id)
end

function balance.release_unstake_lock(profile_id)
    checks('number')
    return balance.release_lock(box.space.unstake_lock, profile_id)
end

function balance.acquire_lock(lock_space, profile_id)
    local tm = time.now()

    box.begin()

    local lock = lock_space:get(profile_id)
    if lock ~= nil then
        if lock.locked == true then
            local text = "concurrent lock attempt for profile_id=" .. tostring(profile_id)
            box.rollback()
            log.error(text)
            return { res = lock, error = "ALREADY_LOCKED" }
        end
    end


    local is_locked = true
    local new_lock = {
        profile_id,
        is_locked,
        tm
    }

    local status, res = pcall(
        function()
            return lock_space:upsert(
                new_lock,
                {
                    { '=', "locked",    is_locked },
                    { '=', 'timestamp', tm }
                })
        end
    )

    if status == false then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = res }
    end

    -- UPSERT always return nil, but we want to return newly updated lock
    res = lock_space:get(profile_id)

    box.commit()

    return { res = res, error = nil }
end

function balance.release_lock(lock_space, profile_id)
    local tm = time.now()

    box.begin()

    local is_locked = false
    local new_lock = {
        profile_id,
        is_locked,
        tm
    }

    local status, res = pcall(
        function()
            return lock_space:upsert(
                new_lock,
                {
                    { '=', "locked",    is_locked },
                    { '=', 'timestamp', tm }
                })
        end
    )


    if status == false then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = res }
    end

    -- UPSERT always return nil, but we want to return newly updated lock
    res = lock_space:get(profile_id)

    box.commit()

    return { res = res, error = nil }
end

function balance.create_deposit(profile_id, wallet, amount, txhash, exchange_id, chain_id)
    checks('number', 'string', 'decimal', 'string', "string", "number")

    box.begin()

    local _id = uuid.str()

    local res, err = archiver.insert(box.space.balance_operations, {
        _id,
        config.params.BALANCE_STATUS.PENDING,
        "",
        txhash,
        profile_id,
        wallet,
        config.params.BALANCE_TYPE.DEPOSIT,
        _id,
        amount,
        time.now(),
        0,

        exchange_id,
        chain_id,
        "",
    })

    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    return { res = res, error = err }
end

function balance.create_stake(staker_profile_id, vault_wallet, amount, txhash, exchange_id, chain_id)
    checks('number', 'string', 'decimal', 'string', "string", "number")

    box.begin()

    local _id = uuid.str()

    local res, err = archiver.insert(box.space.balance_operations, {
        _id,
        config.params.BALANCE_STATUS.PENDING,
        "",
        txhash,
        staker_profile_id,
        vault_wallet,
        config.params.BALANCE_TYPE.STAKE,
        _id,
        amount,
        time.now(),
        0,
        exchange_id,
        chain_id,
        "",
    })

    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    notif_profile(res.profile_id, res:tomap({ names_only = true }))

    return { res = res, error = err }
end

function balance.create_unstake(staker_profile_id, vault_profile_id, vault_wallet, shares, exchange_id, chain_id)
    checks('number', 'number', 'string', 'decimal', 'string', 'number')

    local c_map = balance.contract_by_exchange_id(exchange_id)
    if c_map == nil then
        local text = "ERR_NO_CONTRACT_MAP exchange_id=" .. tostring(exchange_id)
        log.error(BalanceError:new(text))
        return { res = nil, error = ERR_NO_CONTRACT_MAP }
    end

    box.begin()

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.UNSTAKE_SHARES, staker_profile_id, config.params.BALANCE_STATUS.REQUESTED)
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end
    if exist ~= nil then
        box.rollback()
        log.error({
            message = "UNSTAKE_ALREADY_REQUESTED",
            [ALERT_TAG] = ALERT_LOW,
        })
        return { res = nil, error = "UNSTAKE_ALREADY_REQUESTED" }
    end

    local unstake_id = box.sequence.unstake_id_sequence:next()

    local vus, err = archiver.insert(box.space.balance_operations, {
        "vus_u_" .. unstake_id,
        config.params.BALANCE_STATUS.REQUESTED,
        "",
        "",
        vault_profile_id,
        vault_wallet,
        config.params.BALANCE_TYPE.VAULT_UNSTAKE_SHARES,
        "vus_u_" .. unstake_id,
        shares,
        time.now(),
        0,
        exchange_id,
        chain_id,
        c_map.contract_address,
    })

    if err ~= nil then
        box.rollback()
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end

    local unstake
    unstake, err = archiver.insert(box.space.balance_operations, {
        "u_" .. unstake_id,
        config.params.BALANCE_STATUS.REQUESTED,
        "",
        "",
        staker_profile_id,
        vault_wallet,
        config.params.BALANCE_TYPE.UNSTAKE_SHARES,
        "u_" .. unstake_id,
        shares,
        time.now(),
        0,
        exchange_id,
        c_map.chain_id,
        c_map.contract_address,
    })

    if err ~= nil then
        box.rollback()
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end

    local vlt = box.space.vaults:get(vault_profile_id)
    if vlt == nil then
        box.rollback()
        local msg = string.format(
            "VAULT_NOT_FOUND wallet=%s",
            vault_wallet
        )
        log.error({
            message = msg,
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end
    if vlt.status ~= config.params.VAULT_STATUS.ACTIVE then
        box.rollback()
        log.error({
            message = string.format(
                "%s: create_unstake: vault_profile_id=%s, staker_profile_id=%s",
                ERR_VAULT_NOT_ACTIVE,
                tostring(vault_profile_id),
                tostring(staker_profile_id)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = ERR_VAULT_NOT_ACTIVE }
    end
    if shares > vlt.total_shares then
        box.rollback()
        log.error({
            message = string.format(
                "%s: create_unstake: vault_profile_id=%s, staker_profile_id=%s, unstake shares=%s, total_shares=%s",
                ERR_INSUFFICIENT_SHARES,
                tostring(vault_profile_id),
                tostring(staker_profile_id),
                tostring(shares),
                tostring(vlt.total_shares)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = ERR_INSUFFICIENT_SHARES }
    end

    local exist = box.space.vault_holdings:get(
        { vault_profile_id, staker_profile_id }
    )

    if exist == nil then
        box.rollback()
        log.error({
            message = string.format(
                "%s: create_unstake: vault_profile_id=%s, staker_profile_id=%s",
                ERR_NOT_HOLDING_SHARES,
                tostring(vault_profile_id),
                tostring(staker_profile_id)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = ERR_NOT_HOLDING_SHARES }
    end
    if shares > exist.shares then
        box.rollback()
        log.error({
            message = string.format(
                "%s: create_unstake: vault_profile_id=%s, staker_profile_id=%s, unstake shares=%s, holding shares=%s",
                ERR_INSUFFICIENT_HOLDING,
                tostring(vault_profile_id),
                tostring(staker_profile_id),
                tostring(shares),
                tostring(exist.shares)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = ERR_INSUFFICIENT_HOLDING }
    end

    box.commit()

    notif_profile(vus.profile_id, vus:tomap({ names_only = true }))
    notif_profile(unstake.profile_id, unstake:tomap({ names_only = true }))

    return { res = unstake, error = err }
end

function balance.cancel_unstake(profile_id, bops_id)
    checks('number', 'string')

    box.begin()

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.UNSTAKE_SHARES, profile_id,
        config.params.BALANCE_STATUS.REQUESTED, bops_id)
    if err ~= nil then
        box.rollback()
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end
    if exist == nil then
        box.rollback()
        log.error({
            message = "NO_REQUESTED_UNSTAKE",
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = "NO_REQUESTED_UNSTAKE" }
    end

    local res, bres
    res, err = archiver.update(box.space.balance_operations, exist.id, {
        { '=', 'status', config.params.BALANCE_STATUS.CANCELED } })
    if err ~= nil then
        box.rollback()
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end
    if res == nil then
        box.rollback()
        log.error({
            message = "UNSTAKE_NOT_FOUND",
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = "UNSTAKE_NOT_FOUND" }
    end
    local balancing_id = "vus_" .. exist.id
    bres, err = archiver.update(box.space.balance_operations, balancing_id, {
        { '=', 'status', config.params.BALANCE_STATUS.CANCELED } })
    if err ~= nil then
        box.rollback()
        log.error({
            message = tostring(err),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = err }
    end
    if bres == nil then
        box.rollback()
        log.error({
            message = "BALANCING_UNSTAKE_NOT_FOUND",
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = "BALANCING_UNSTAKE_NOT_FOUND" }
    end

    box.commit()

    notif_profile(res.profile_id, res:tomap({ names_only = true }))
    notif_profile(bres.profile_id, bres:tomap({ names_only = true }))

    return { res = nil, error = nil }
end

function balance.resolve_all_unknown(profile_id, wallet, exchange_id)
    checks('number', 'string', 'string')

    local unknown_deposits = box.space.balance_operations.index.exchange_id_wallet_type_status:select({ exchange_id,
        wallet,
        config.params.BALANCE_TYPE.DEPOSIT,
        config.params.BALANCE_STATUS.UNKNOWN })

    local count = 0
    for _, unknown_deposit in pairs(unknown_deposits) do
        local success, err = pcall(
            balance.resolve_unknown, profile_id, unknown_deposit.id)
        if not success then
            log.error({
                message = tostring(err),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return { res = nil, error = tostring(err) }
        end
        count = util.safe_yield(count, 1000)
    end
    return { res = nil, error = nil }
end

function balance.create_referral_payout(id, profile_id, amount)
    checks('string', 'number', 'decimal')

    if amount <= 0 then
        return { res = nil, error = ERR_REFERRAL_PAYOUT_AMOUNT_NOT_POSITIVE }
    end

    box.begin()

    local exists = box.space.balance_operations:get(id)
    if exists ~= nil then
        box.rollback()
        return { res = nil, error = ERR_REFERRAL_PAYOUT_ID_DUPLICATE }
    end

    local res, err = archiver.insert(box.space.balance_operations, {
        id,
        config.params.BALANCE_STATUS.PENDING,
        "",
        "",
        profile_id,
        "",
        config.params.BALANCE_TYPE.REFERRAL_PAYOUT,
        id,
        amount,
        time.now(),
        0,

        "",
        0,
        "",
    })

    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    return { res = res, error = err }
end

function balance.increase_balance_sum(profile_id, amount)
    checks("number", "decimal")

    if amount == ZERO then
        return nil
    end

    local tm = time.now()
    local new_sum = {
        profile_id,
        amount,
        tm
    }

    local balance_before = decimal.new("0")
    local exist = box.space.balance_sum:get(profile_id)
    if exist ~= nil then
        balance_before = exist.balance
    end

    local status, res = pcall(
        function()
            return box.space.balance_sum:upsert(
                new_sum,
                {
                    { '+', "balance",      amount },
                    { '=', 'last_updated', tm }
                })
        end
    )

    if status == false then
        return BalanceError:new(res)
    end


    exist = box.space.balance_sum:get(profile_id)
    if exist == nil then
        return BalanceError:new("increase_balance_sum cant create for profile_id=%s",
        tostring(profile_id))
    end

    if exist.balance == balance_before then
        return BalanceError:new("increase_balance_sum amount not changed for profile_id=%s before=%s after=%s",
            tostring(profile_id),
            tostring(balance_before),
            tostring(exist.amount))
    end

    return nil

end

function balance.decrease_balance_sum(profile_id, amount)
    return balance.increase_balance_sum(profile_id, -amount)
end

-- process a deposit from a wallet linked to a profile id
function balance.resolve_unknown(profile_id, balance_ops_id)
    checks('number', 'string')

    local tm = time.now()
    box.begin()

    local exist = box.space.balance_operations:get(balance_ops_id)
    if exist == nil then
        local text = "RESOLVE_UNKNOWN no ops with id=" .. tostring(balance_ops_id)
        box.rollback()
        return { res = nil, error = text }
    elseif exist.status ~= config.params.BALANCE_STATUS.UNKNOWN then
        local text = "RESOLVE_UNKNOWN already resolved id =" .. tostring(balance_ops_id)
        box.rollback()
        return { res = nil, error = text }
    end

    local success = config.params.BALANCE_STATUS.SUCCESS

    local res, err = archiver.update(box.space.balance_operations, exist.id, {
        { '=', 'status',     success },
        { '=', 'profile_id', profile_id }
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = err }
    end

    local b_ops = res:tomap({ names_only = true })

    err = balance.increase_balance_sum(profile_id, exist.amount)
    if err ~= nil then
        box.rollback()
        log.error(err)
        return { res = nil, error = tostring(err) }
    end

    box.commit()

    notif_profile(profile_id, b_ops)
    return { res = nil, error = nil }
end

-- process a deposit from a wallet linked to a profile id
function balance.process_deposit(profile_id, wallet, deposit_id, amount, txhash, isPoolDeposit, exchange_id, chain_id,
                                 contract_address)
    checks('number', 'string', 'string', 'decimal', 'string', 'boolean', 'string', 'number', 'string')

    if amount <= ZERO then
        return { res = nil, error = "NEGATIVE_OR_ZERO_DEPOSIT_AMOUNT" }
    end

    local tm = time.now()

    box.begin()

    local exist = box.space.balance_operations.index.ops_id2:get(deposit_id)
    if exist ~= nil then
        local text = "DUPLICATE ID update attempt for ops_id2=" .. deposit_id
        box.rollback()
        return { res = nil, error = text }
    end

    -- For normal deposits there is only one deposit per tx and
    -- the front end creates the balance_op by calling
    -- create_deposit before the deposit_id is known (it's
    -- created later by the rabbit contract when it processes
    -- the deposit). If the deposit_id has not yet been set
    -- in ops_id2 then the balance op won't have been found by the
    -- search above, but we can find it by tx.
    --
    -- With pool deposits there can be multiple deposits in one tx
    -- so you can't find the balance op by tx, but the deposit_id
    -- is known at the time of balance op creation, so if the
    -- balance op existed we would have found it in the previous
    -- search by ops_id2.
    if not isPoolDeposit then
        exist = box.space.balance_operations.index.txhash:min { txhash }
    end

    -- INTEGRITY CHECKS in case of bugs: BUT should never happens
    if exist ~= nil and exist.profile_id ~= profile_id then
        log.error({
            message = string.format("INTEGRITY_ERROR_ID: process_deposit: profile_id=%d exist.profile_id=%d", profile_id,
                exist.profile_id),
            [ALERT_TAG] = ALERT_CRIT,
        })
        box.rollback()
        return { res = nil, error = "INTEGRITY_ERROR_ID" }
    end

    if exist ~= nil and exist.ops_type ~= config.params.BALANCE_TYPE.DEPOSIT then
        log.error({
            message = string.format("INTEGRITY_ERROR_TYPE: process_deposit: type different %s", exist.ops_type),
            [ALERT_TAG] = ALERT_CRIT,
        })
        box.rollback()
        return { res = nil, error = "INTEGRITY_ERROR_TYPE" }
    end

    -- Check its status, normally it should be pending.
    --
    -- If it is canceled that means the back end checked it and
    -- found no transaction with this txhash, so assumed the user
    -- had either canceled or replaced it, since we now have an event
    -- from the blockchain we can mark it as successful
    --
    -- If it is neither pending nor canceled that's an error

    if exist ~= nil and exist.status ~= config.params.BALANCE_STATUS.PENDING and exist.status ~= config.params.BALANCE_STATUS.CANCELED then
        log.error({
            message = string.format("INTEGRITY_ERROR_STATUS: process_deposit: wrong status", exist.status),
            [ALERT_TAG] = ALERT_CRIT,
        })
        box.rollback()
        return { res = nil, error = "INTEGRITY_ERROR_STATUS" }
    end

    local success = config.params.BALANCE_STATUS.SUCCESS

    local upserted_id = deposit_id
    if exist ~= nil then
        upserted_id = exist.id
    end
    local _, err = archiver.upsert(box.space.balance_operations, {
        upserted_id,
        success,
        "",
        txhash,
        profile_id,
        wallet,
        config.params.BALANCE_TYPE.DEPOSIT,
        deposit_id,
        amount,
        tm,
        0,

        exchange_id,
        chain_id,
        contract_address,
    }, {
        { '=', 'status',           success },
        { '=', 'wallet',           wallet },
        { '=', 'ops_id2',          deposit_id },
        { '=', 'amount',           amount },
        { '=', 'txhash',           txhash },
        { '=', 'exchange_id',      exchange_id },
        { '=', 'chain_id',         chain_id },
        { '=', 'contract_address', contract_address },
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    local res
    res, err = dml.get(box.space.balance_operations, upserted_id)
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end
    local b_ops = res:tomap({ names_only = true })

    err = balance.increase_balance_sum(profile_id, amount)
    if err ~= nil then
        box.rollback()
        log.error(err)
        return { res = nil, error = tostring(err) }
    end

    box.commit()

    notif_profile(profile_id, b_ops)
    return { res = nil, error = nil }
end

-- process a deposit from a wallet that is not linked to any profile id
function balance.process_deposit_unknown(wallet, deposit_id, amount, tx, exchange_id, chain_id, contract_address)
    checks('string', 'string', 'decimal', 'string', 'string', 'number', 'string')

    if amount <= 0 then
        return { res = nil, error = "NEGATIVE_OR_ZERO_DEPOSIT_AMOUNT" }
    end

    local exist = box.space.balance_operations.index.ops_id2:get(deposit_id)
    if exist ~= nil then
        local text = "DUPLICATE_ATTEMPT update attempt for ops_id2=" .. deposit_id
        return { res = nil, error = text }
    end

    local res, err = archiver.insert(box.space.balance_operations, {
        deposit_id,
        config.params.BALANCE_STATUS.UNKNOWN,
        "",
        tx,
        ANONYMOUS_ID_FOR_UNKNOWN,
        wallet,
        config.params.BALANCE_TYPE.DEPOSIT,
        deposit_id,
        amount,
        time.now(),
        0,
        exchange_id,
        chain_id,
        contract_address,
    })

    if err ~= nil then
        return { res = nil, error = err }
    end

    return { res = nil, error = nil }
end

function balance.process_stake(staker_profile_id, vault_profile_id, vault_wallet, stake_id, amount, current_nav, tx, from_balance, exchange_id)
    checks('number', 'number', 'string', 'string', 'decimal', 'decimal', 'string', 'boolean', 'string')

    local c_map = balance.contract_by_exchange_id(exchange_id)
    if c_map == nil then
        local text = "ERR_NO_CONTRACT_MAP exchange_id=" .. tostring(exchange_id)
        log.error(BalanceError:new(text))
        return { res = nil, error = ERR_NO_CONTRACT_MAP }
    end
    local _

    if amount <= ZERO then
        log.error({
            message = string.format(
                "%s: process_stake: staker_profile_id=%s, vault_wallet=%s, amount=%s",
                ERR_WRONG_STAKE_AMOUNT,
                tostring(staker_profile_id),
                tostring(vault_wallet),
                tostring(amount)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return { res = nil, error = ERR_WRONG_STAKE_AMOUNT }
    end

    local tm = time.now()

    box.begin()

    local stake_type, exist
    if from_balance then
        stake_type = config.params.BALANCE_TYPE.STAKE_FROM_BALANCE
        stake_id = uuid.str()
    else
        stake_type = config.params.BALANCE_TYPE.STAKE
        exist = box.space.balance_operations.index.ops_id2:get(stake_id)
        if exist ~= nil then
            log.error({
                message = string.format(
                    "%s: process_stake: update attempt for stake ops_id2=%s",
                    ERR_DUPLICATE_STAKE_ID,
                    tostring(stake_id)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_DUPLICATE_STAKE_ID }
        end

        exist = box.space.balance_operations.index.txhash:min { tx }

        -- INTEGRITY CHECKS in case of bugs: BUT should never happens
        if exist ~= nil and exist.profile_id ~= staker_profile_id then
            log.error({
                message = string.format(
                    "%s: process_stake: staker_profile_id=%s exist.profile_id=%s",
                    ERR_INTEGRITY_ERROR,
                    tostring(staker_profile_id),
                    tostring(exist.profile_id)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = "ERR_INTEGRITY_ERROR" }
        end

        if exist ~= nil and exist.ops_type ~= config.params.BALANCE_TYPE.STAKE then
            log.error({
                message = string.format("INTEGRITY_ERROR_TYPE: process_stake: type different %s", exist.ops_type),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = "INTEGRITY_ERROR_TYPE" }
        end

        -- Check its status, normally it should be pending.
        --
        -- If it is canceled that means the back end checked it and
        -- found no transaction with this txhash, so assumed the user
        -- had either canceled or replaced it, since we now have an event
        -- from the blockchain we can mark it as successful
        --
        -- If it is neither pending nor canceled that's an error

        if exist ~= nil and exist.status ~= config.params.BALANCE_STATUS.PENDING and exist.status ~= config.params.BALANCE_STATUS.CANCELED then
            log.error({
                message = string.format(
                    "%s: process_stake - wrong status %s",
                    ERR_INTEGRITY_ERROR,
                    exist.status),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_INTEGRITY_ERROR }
        end
    end

    local success = config.params.BALANCE_STATUS.SUCCESS

    local upserted_id = stake_id
    if exist ~= nil then
        upserted_id = exist.id
    end
    local res, err = archiver.upsert(box.space.balance_operations, {
        upserted_id,
        success,
        "",
        tx,
        staker_profile_id,
        vault_wallet,
        stake_type,
        stake_id,
        amount,
        tm,
        0,
        exchange_id,
        c_map.chain_id,
        c_map.contract_address,
    }, {
        { '=', 'status',           success },
        { '=', 'wallet',           vault_wallet },
        { '=', 'ops_id2',          stake_id },
        { '=', 'amount',           amount },
        { '=', 'exchange_id',      exchange_id },
        { '=', 'chain_id',         c_map.chain_id },
        { '=', 'contract_address', c_map.contract_address },
    })
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end

    local vstake_bop
    local balancing_op_id = "vs_" .. upserted_id
    local balancing_op_id2 = "vs_" .. stake_id
    vstake_bop, err = archiver.insert(box.space.balance_operations, {
        balancing_op_id,
        config.params.BALANCE_STATUS.SUCCESS,
        "",
        tx,
        vault_profile_id,
        vault_wallet,
        config.params.BALANCE_TYPE.VAULT_STAKE,
        balancing_op_id2,
        amount,
        tm,
        0,

        exchange_id,
        c_map.chain_id,
        c_map.contract_address,
    })
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end


    local stake_bop
    stake_bop, err = dml.get(box.space.balance_operations, upserted_id)
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end

    if from_balance then
        local staker_balance = box.space.balance_sum:get(staker_profile_id)
        if staker_balance == nil then
            box.rollback()
            log.error(VaultError:new(
                "STAKER_BALANCE_NOT_FOUND for staker_profile_id=%s",
                tostring(staker_profile_id)
            ))
            return { res = nil, error = "STAKER_BALANCE_NOT_FOUND" }
        end
        -- the caller has already checked that the withdrawable balance is sufficient
        -- so this check on balance should be unnecessary, but it does no harm
        if staker_balance.balance < amount then
            box.rollback()
            log.error(VaultError:new(
                "INSUFFICIENT_STAKER_BALANCE for staker_profile_id=%s, stake amount=%s, balance=%s",
                tostring(staker_profile_id),
                tostring(amount),
                tostring(staker_balance.balance)
            ))
            return { res = nil, error = "INSUFFICIENT_STAKER_BALANCE" }
        end
        err = balance.decrease_balance_sum(staker_profile_id, amount)
        if err ~= nil then
            box.rollback()
            log.error(string.format("STAKE_FROM_BALANCE_FAILED %s", tostring(err)))
            return { res = nil, error = tostring(err) }
        end
    end

    err = balance.increase_balance_sum(vault_profile_id, amount)
    if err ~= nil then
        box.rollback()
        log.error(err)
        return { res = nil, error = tostring(err) }
    end

    local vlt = box.space.vaults:get(vault_profile_id)
    if vlt == nil then
        box.rollback()
        err = VaultError:new(
            "VAULT_NOT_FOUND for vault_profile_id=%s",
            tostring(vault_profile_id)
        )
        log.error(err)
        return { res = nil, error = err }
    end
    if vlt.status ~= config.params.VAULT_STATUS.ACTIVE then
        box.rollback()
        log.error(VaultError:new(
            "VAULT_NOT_ACTIVE for vault_profile_id=%s",
            tostring(vault_profile_id)
        ))
        return { res = nil, error = "VAULT_NOT_ACTIVE" }
    end
    local total_shares = vlt.total_shares

    exist = box.space.vault_holdings:get(
        { vault_profile_id, staker_profile_id }
    )

    local prev_shares, prev_nav, prev_price
    if exist ~= nil then
        prev_shares = exist.shares
        prev_nav = exist.entry_nav
        prev_price = exist.entry_price
    else
        prev_shares = ZERO
        prev_nav = ZERO
        prev_price = ONE
    end

    local new_shares, updated_nav, updated_price
    if total_shares == ZERO then
        new_shares = amount
        updated_nav = amount
        updated_price = ONE
    elseif current_nav <= ONE then
        log.error(VaultError:new(
            "VAULT_NAV_TOO_LOW for vault_profile_id=%s, current_nav=%s",
            tostring(vault_profile_id),
            tostring(current_nav)
        ))
        box.rollback()
        return { res = nil, error = "WRONG_NAV" }
    else
        new_shares = (amount * total_shares) / current_nav;
        updated_nav = current_nav + amount;
        updated_price = current_nav / total_shares
    end

    total_shares = total_shares + new_shares
    _, err = archiver.update(
        box.space.vaults,
        vault_profile_id,
        { { '=', 'total_shares', total_shares } }
    )
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end

    local holding_shares, holding_nav, holding_price

    if prev_shares <= ZERO then
        holding_shares = new_shares
        holding_nav = updated_nav
        holding_price = updated_price
    else
        holding_shares = prev_shares + new_shares
        holding_nav = (prev_nav *
            prev_shares +
            updated_nav *
            new_shares) / holding_shares
        holding_price = (prev_price *
            prev_shares +
            updated_price *
            new_shares) / holding_shares
    end

    _, err = archiver.upsert(box.space.vault_holdings, {
        vault_profile_id,
        staker_profile_id,
        holding_shares,
        holding_nav,
        holding_price
    }, {
        { '=', 'shares',      holding_shares },
        { '=', 'entry_nav',   holding_nav },
        { '=', 'entry_price', holding_price }
    })
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end

    local shares_bop
    local stake_shares_op_id = "ss_" .. upserted_id
    local stake_shares_op_id2 = "ss_" .. stake_id
    shares_bop, err = archiver.insert(box.space.balance_operations, {
        stake_shares_op_id,
        config.params.BALANCE_STATUS.SUCCESS,
        "",
        tx,
        staker_profile_id,
        vault_wallet,
        config.params.BALANCE_TYPE.STAKE_SHARES,
        stake_shares_op_id2,
        new_shares,
        tm,
        0,

        exchange_id,
        c_map.chain_id,
        c_map.contract_address,
    })
    if err ~= nil then
        box.rollback()
        log.error(VaultError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    multiple_notif_profile(staker_profile_id, {
        stake_bop:tomap({ names_only = true }),
        shares_bop:tomap({ names_only = true }),
    })

    notif_profile(vstake_bop.profile_id, vstake_bop:tomap({ names_only = true }))

    return { res = stake_bop, error = nil }
end

function balance.create_withdrawal(profile_id, wallet, amount, exchange_id)
    checks("number", "string", "decimal", "string")
    local c_map = balance.contract_by_exchange_id(exchange_id)
    if c_map == nil then
        local text = "ERR_NO_CONTRACT_MAP exchange_id=" .. tostring(exchange_id)
        log.error(BalanceError:new(text))
        return { res = nil, error = ERR_NO_CONTRACT_MAP }
    end

    box.begin()

    local withdrawal_id = box.sequence.withdrawal_id_sequence:next()
    if withdrawal_id == 0 then
        local timestamp = math.floor(10 * clock.time())
        local offset = balance._shard_num * OFFSET_MULTIPLIER
        withdrawal_id = timestamp + offset
        box.sequence.withdrawal_id_sequence:set(withdrawal_id)
    end

    local id2 = 'w_' .. tostring(withdrawal_id)
    local tm = time.now()

    local res, err = archiver.insert(box.space.balance_operations, {
        id2,
        config.params.BALANCE_STATUS.PENDING,
        "",
        "",
        profile_id,
        wallet,
        config.params.BALANCE_TYPE.WITHDRAWAL,
        id2,
        amount,
        tm,
        0,

        exchange_id,
        c_map.chain_id,
        c_map.contract_address,
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = err }
    end
    local w_ops = res

    err = balance.decrease_balance_sum(profile_id, amount)
    if err ~= nil then
        box.rollback()
        log.error(string.format("CREATE_WITHDRAWAL_FAILED %s", tostring(err)))
        return { res = nil, error = tostring(err) }
    end

    -- we roll the volume
    wdm.roll_volume(amount)

    box.commit()

    return { res = w_ops, error = nil }
end

-- todo: refactoring is required because this func is a copy-paste of <create_withdrawal>
function balance.create_withdraw_fee(profile_id, wallet, amount, txhash)
    box.begin()

    local withdrawal_id = box.sequence.withdrawal_id_sequence:next()
    if withdrawal_id == 0 then
        local timestamp = math.floor(10 * clock.time())
        local offset = balance._shard_num * OFFSET_MULTIPLIER
        withdrawal_id = timestamp + offset
        box.sequence.withdrawal_id_sequence:set(withdrawal_id)
    end

    local id2 = 'w_' .. tostring(withdrawal_id)
    local tm = time.now()

    --TODO: Andrei did that
    local res, err = archiver.insert(box.space.balance_operations, {
        id2,
        config.params.BALANCE_STATUS.PENDING,
        "",
        txhash,
        profile_id,
        wallet,
        config.params.BALANCE_TYPE.WITHDRAWAL,
        id2,
        amount,
        tm,
        0,

        "",
        0,
        "",
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = err }
    end
    local w_ops = res

    local new_sum = {
        profile_id,
        ZERO, -- amount
        tm
    }
    local status

    status, res = pcall(
        function()
            return box.space.balance_sum:upsert(
                new_sum,
                {
                    { '=', "balance",      ZERO },
                    { '=', 'last_updated', tm }
                })
        end
    )

    if status == false then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = res }
    end

    box.commit()

    return { res = w_ops, error = nil }
end

function balance.pending_deposit_canceled(id)
    checks('string')
    -- check operation is a deposit or stake in pending state
    local exist = box.space.balance_operations:get(id)
    if exist == nil then
        return { res = nil, error = "OP_NOT_FOUND" }
    end
    if exist.ops_type ~= config.params.BALANCE_TYPE.DEPOSIT and exist.ops_type ~= config.params.BALANCE_TYPE.STAKE then
        return { res = nil, error = "OP_IS_NOT_DEPOSIT_OR_STAKE" }
    end
    if exist.status ~= config.params.BALANCE_STATUS.PENDING then
        return { res = nil, error = "OP_NOT_PENDING" }
    end
    -- mark it canceled
    local updated_op, err = archiver.update(box.space.balance_operations, id, {
        { '=', 'status', config.params.BALANCE_STATUS.CANCELED } })
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end
    if updated_op == nil then
        return { res = nil, error = "OP_NOT_FOUND" }
    end
    notif_profile(updated_op.profile_id, updated_op:tomap({ names_only = true }))
    return { res = nil, error = nil }
end

function balance.get_pending_deposits(exchange_id, chain_id)
    checks("string", "number")
    return balance.get_balance_ops_in_state(config.params.BALANCE_TYPE.DEPOSIT, config.params.BALANCE_STATUS.PENDING,
        exchange_id,
        chain_id)
end

function balance.get_pending_stakes(exchange_id, chain_id)
    checks("string", "number")

    return balance.get_balance_ops_in_state(config.params.BALANCE_TYPE.STAKE, config.params.BALANCE_STATUS.PENDING,
        exchange_id,
        chain_id)
end

function balance.get_pending_withdrawals(exchange_id, chain_id)
    checks("string", "number")

    return balance.get_balance_ops_in_state(config.params.BALANCE_TYPE.WITHDRAWAL, config.params.BALANCE_STATUS.PENDING,
        exchange_id, chain_id)
end

function balance.get_all_pending_withdrawals()
    return balance.get_balance_ops_in_state(config.params.BALANCE_TYPE.WITHDRAWAL, config.params.BALANCE_STATUS.PENDING)
end

function balance.get_balance_ops_in_state(op_type, state, exchange_id, chain_id)
    checks("string", "string", "?string", "?number")

    local cond = {
        op_type,
        state,
    }

    if exchange_id ~= nil and chain_id ~= 0 then
        cond = {
            op_type,
            state,
            exchange_id,
            chain_id,
        }
    end

    local res = {}
    local count = 0
    for _, balance_op in box.space.balance_operations.index.type_status_exchange_id_chain_id:pairs(cond,
        { iterator = "EQ" }
    )
    do
        table.insert(res, balance_op)
        count = util.safe_yield(count, 1000)
    end

    return { res = res, error = nil }
end

function balance.update_pending_withdrawals(current_block, future_block, for_contract)
    checks('number', 'number', 'string')

    local count = 0
    for _, balance_op in box.space.balance_operations.index.type_contract_due_block:pairs({ config.params.BALANCE_TYPE.WITHDRAWAL, for_contract, 0 }, { iterator = 'EQ' }) do
        if balance_op.contract_address ~= for_contract then
            break
        end

        local due_block = future_block
        local delay = wdm.get_delay(balance_op.exchange_id, balance_op.profile_id, balance_op.amount)
        if delay ~= nil then
            due_block = current_block + delay
        end
        
        local _, err = archiver.update(box.space.balance_operations, balance_op.id, {
            { '=', 'due_block', due_block } })
        if err ~= nil then
            return { res = nil, error = err }
        end

        count = util.safe_yield(count, 1000)
    end

    count = 0
    local claimable = config.params.BALANCE_STATUS.CLAIMABLE
    local pending = config.params.BALANCE_STATUS.PENDING
    for _, balance_op in box.space.balance_operations.index.type_contract_due_block:pairs({ config.params.BALANCE_TYPE.WITHDRAWAL, for_contract, current_block }, { iterator = 'LE' }) do
        if balance_op.ops_type ~= config.params.BALANCE_TYPE.WITHDRAWAL then
            break
        end

        if balance_op.contract_address ~= for_contract then
            break
        end

        if balance_op.status == pending then
            local err
            balance_op, err = archiver.update(box.space.balance_operations, balance_op.id, {
                { '=', 'status',    claimable },
                { '=', 'due_block', BLOCK_NUM_NEVER }
            })
            if err ~= nil then
                return { res = nil, error = err }
            end
            notif_profile(balance_op.profile_id, balance_op:tomap({ names_only = true }))
        else
            local _, err = archiver.update(box.space.balance_operations, balance_op.id, {
                { '=', 'due_block', BLOCK_NUM_NEVER } })
            if err ~= nil then
                return { res = nil, error = err }
            end
        end

        count = util.safe_yield(count, 1000)
    end
    return { res = nil, error = nil }
end

function balance.get_vault_manager_profile_id(vault_profile_id)
    checks('unsigned')

    local vault = box.space.vaults:get(vault_profile_id)
    if vault == nil then
        return {
            res = nil,
            error = string.format(
                "VAULT_NOT_FOUND %d",
                vault_profile_id
            )
        }
    end
    return { res = vault.manager_profile_id, error = nil }
end

function balance.get_vault_info(vault_profile_id)
    checks('number')

    local vault = box.space.vaults:get(vault_profile_id)
    if vault == nil then
        return {
            res = nil,
            error = string.format(
                "VAULT_NOT_FOUND %d",
                vault_profile_id
            )
        }
    end
    return { res = vault, error = nil }
end

function balance.get_holding_info(vault_profile_id, staker_profile_id)
    checks('unsigned', 'unsigned')

    local holding = box.space.vault_holdings:get(vault_profile_id, staker_profile_id)
    if holding == nil then
        return {
            res = {
                vault_profile_id = vault_profile_id,
                staker_profile_id = staker_profile_id,
                shares = decimal.new(0),
                entry_nav = decimal.new(0)
            },
            error = nil
        }
    end
    return { res = holding, error = nil }
end

function balance.process_unstakes(vault_profile_id, from_id, to_id, current_nav, withdrawable_balance, performance_fee, treasurer_profile_id, total_shares, exchange_id)
    checks('number', 'number', 'number', 'decimal', 'decimal', 'decimal', 'number', 'decimal', 'string')

    local remaining_withdrawable = withdrawable_balance
    local status, err
    local holding_shares, entry_price, current_price
    local performance_charge
    local unstake_id, unstake_shares, staker_profile_id, vault_wallet
    local exist, unstake_value, vault_balance
    local tm, staker_value, new_sum
    local bop_id
    local balancing_id
    local chain_id, contract_address
    local _

    local affected_pids = {}

    local count = 0
    box.begin()

    local vault_profile = box.space.profile:get(vault_profile_id)
    if not vault_profile then
        log.error({
            message = string.format(
                "%s: process_unstakes, vault not found: vault_profile_id=%s",
                ERR_UNSTAKE_VAULT_NOT_FOUND,
                tostring(vault_profile_id)
            ),
            [ALERT_TAG] = ALERT_CRIT,
        })
        box.rollback()
        return { res = nil, error = err }
    end
    vault_wallet = vault_profile.wallet

    for _, unstake_op in box.space.balance_operations.index.exchange_id_wallet_type_status:pairs(
        {
            exchange_id,
            vault_wallet,
            config.params.BALANCE_TYPE.UNSTAKE_SHARES,
            config.params.BALANCE_STATUS.REQUESTED
        },
        { iterator = 'EQ' }
    ) do
        local prefix, id_str = unstake_op.ops_id2:match("^(.)_(%d+)$")
        local id_num = tonumber(id_str)
        if prefix ~= 'u' or id_num == nil or id_num < from_id or id_num > to_id then
            goto continue
        end

        if unstake_op.amount <= ZERO then
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, amount=%s",
                    ERR_WRONG_UNSTAKE_AMOUNT,
                    tostring(unstake_op.ops_id2),
                    tostring(unstake_op.amount)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            unstake_op, err = archiver.update(box.space.balance_operations, unstake_op.id, {
                { '=', 'status', config.params.BALANCE_STATUS.CANCELED }
            })
            if err ~= nil then
                log.error({
                    message = string.format(
                        "%s: process_unstake: unstake_id=%s",
                        ERR_UNSTAKE_UPDATE_FAILED,
                        tostring(unstake_op.id)
                    ),
                    [ALERT_TAG] = ALERT_CRIT,
                })
                box.rollback()
                return { res = nil, error = err }
            end
            goto continue
        end

        unstake_id = unstake_op.id
        unstake_shares = unstake_op.amount
        staker_profile_id = unstake_op.profile_id
        vault_wallet = unstake_op.wallet
        exchange_id = unstake_op.exchange_id
        chain_id = unstake_op.chain_id
        contract_address = unstake_op.contract_address

        if total_shares < unstake_shares then
            log.error({
                message = string.format(
                    "%s: process_unstake: vault_profile_id=%s, staker_profile_id=%s, unstake shares=%s, total_shares=%s",
                    ERR_INSUFFICIENT_SHARES,
                    tostring(vault_profile_id),
                    tostring(staker_profile_id),
                    tostring(unstake_shares),
                    tostring(total_shares)
                ),
                [ALERT_TAG] = ALERT_CRIT
            })
            goto continue
        end

        exist = box.space.vault_holdings:get(
            { vault_profile_id, staker_profile_id }
        )

        if exist ~= nil then
            holding_shares = exist.shares
            entry_price = exist.entry_price
        else
            holding_shares = ZERO
            entry_price = ONE
        end

        if holding_shares < unstake_shares then
            log.error({
                message = string.format(
                    "%s: process_unstake: vault_profile_id=%s, staker_profile_id=%s, unstake shares=%s, holding shares=%s",
                    ERR_INSUFFICIENT_HOLDING,
                    tostring(vault_profile_id),
                    tostring(staker_profile_id),
                    tostring(unstake_shares),
                    tostring(holding_shares)
                ),
                [ALERT_TAG] = ALERT_CRIT
            })
            goto continue
        end

        unstake_value = (current_nav * unstake_shares) / total_shares;
        vault_balance = box.space.balance_sum:get(vault_profile_id)

        if remaining_withdrawable < unstake_value then
            box.rollback()
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, unstake_value=%s",
                    ERR_VAULT_WITHDRAWABLE_BALANCE_INSUFFICIENT,
                    tostring(unstake_id),
                    tostring(unstake_value)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return { res = nil, error = ERR_VAULT_WITHDRAWABLE_BALANCE_INSUFFICIENT }
        end
        remaining_withdrawable = remaining_withdrawable - unstake_value

        if vault_balance.balance < unstake_value then
            box.rollback()
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, unstake_value=%s, vault_balance=%s",
                    ERR_VAULT_BALANCE_INSUFFICIENT,
                    tostring(unstake_id),
                    tostring(unstake_value),
                    tostring(vault_balance.balance)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return { res = nil, error = ERR_VAULT_BALANCE_INSUFFICIENT }
        end

        -- can't `goto continue` after this; have to either update all of
        -- balances, share holding and balance_op status, or roll back
        _, err = archiver.update(
            box.space.vaults,
            vault_profile_id,
            {
                { '=', 'total_shares', total_shares - unstake_shares }
            }
        )
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, err=%s",
                    ERR_SHARES_UPDATE_FAILED,
                    tostring(unstake_op.id),
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_SHARES_UPDATE_FAILED }
        end

        _, err = archiver.update(
            box.space.vault_holdings,
            { vault_profile_id, staker_profile_id, },
            {
                { '=', 'shares', holding_shares - unstake_shares }
            }
        )
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, err=%s",
                    ERR_HOLDING_UPDATE_FAILED,
                    tostring(unstake_op.id),
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_HOLDING_UPDATE_FAILED }
        end

        tm = time.now()
        current_price = current_nav / total_shares
        if current_price > entry_price then
            performance_charge =
                (performance_fee * unstake_value * (current_price - entry_price)) / current_price
        else
            performance_charge = ZERO
        end

        err = balance.decrease_balance_sum(vault_profile_id, unstake_value)
        if err ~= nil then
            box.rollback()
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, unstake_value=%s, err=%s",
                    ERR_VAULT_BALANCE_UPDATE_FAILED,
                    tostring(unstake_id),
                    tostring(unstake_value),
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return { res = nil, error = ERR_VAULT_BALANCE_UPDATE_FAILED }
        end

        staker_value = unstake_value - performance_charge
        err = balance.increase_balance_sum(staker_profile_id, staker_value)
        if err ~= nil then
            box.rollback()
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, unstake_value=%s, err=%s",
                    ERR_STAKER_BALANCE_UPDATE_FAILED,
                    tostring(unstake_id),
                    tostring(staker_value),
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return { res = nil, error = ERR_STAKER_BALANCE_UPDATE_FAILED }
        end

        if performance_charge > ZERO then
            err = balance.increase_balance_sum(treasurer_profile_id, performance_charge)
            if err ~= nil then
                box.rollback()
                log.error({
                    message = string.format(
                        "%s: process_unstake: unstake_id=%s, performance_charge=%s, treasurer profile=%s, err=%s",
                        ERR_TREASURER_BALANCE_UPDATE_FAILED,
                        tostring(unstake_id),
                        tostring(performance_charge),
                        tostring(treasurer_profile_id),
                        tostring(err)
                    ),
                    [ALERT_TAG] = ALERT_CRIT,
                })
                return { res = nil, error = ERR_TREASURER_BALANCE_UPDATE_FAILED }
            end
        end

        bop_id = "uv_" .. unstake_id
        _, err = archiver.insert(box.space.balance_operations, {
            bop_id,
            config.params.BALANCE_STATUS.SUCCESS,
            "",
            "",
            staker_profile_id,
            vault_wallet,
            config.params.BALANCE_TYPE.UNSTAKE_VALUE,
            bop_id,
            staker_value,
            tm,
            0,
            exchange_id,
            chain_id,
            contract_address,
        })
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: balancing unstake_id=%s, err=%s",
                    ERR_BALANCING_UNSTAKE_UPDATE_FAILED,
                    bop_id,
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = err }
        end

        table.insert(affected_pids, staker_profile_id)

        bop_id = "uf_" .. unstake_id
        _, err = archiver.insert(box.space.balance_operations, {
            bop_id,
            config.params.BALANCE_STATUS.SUCCESS,
            "",
            "",
            treasurer_profile_id,
            vault_wallet,
            config.params.BALANCE_TYPE.UNSTAKE_FEE,
            bop_id,
            performance_charge,
            tm,
            0,
            exchange_id,
            chain_id,
            contract_address,
        })
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: balancing unstake_id=%s, err=%s",
                    ERR_BALANCING_UNSTAKE_UPDATE_FAILED,
                    bop_id,
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = err }
        end

        bop_id = "vuv_" .. unstake_id
        _, err = archiver.insert(box.space.balance_operations, {
            bop_id,
            config.params.BALANCE_STATUS.SUCCESS,
            "",
            "",
            vault_profile_id,
            vault_wallet,
            config.params.BALANCE_TYPE.VAULT_UNSTAKE_VALUE,
            bop_id,
            unstake_value,
            tm,
            0,
            exchange_id,
            chain_id,
            contract_address,
        })
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: balancing unstake_id=%s, err=%s",
                    ERR_BALANCING_UNSTAKE_UPDATE_FAILED,
                    bop_id,
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = err }
        end

        balancing_id = "vus_" .. unstake_id
        unstake_op, err = archiver.update(box.space.balance_operations, balancing_id, {
            { '=', 'status', config.params.BALANCE_STATUS.SUCCESS }
        })
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: balancing unstake_id=%s, err=%s",
                    ERR_BALANCING_UNSTAKE_UPDATE_FAILED,
                    balancing_id,
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_UNSTAKE_UPDATE_FAILED }
        end

        unstake_op, err = archiver.update(box.space.balance_operations, unstake_id, {
            { '=', 'status', config.params.BALANCE_STATUS.SUCCESS }
        })
        if err ~= nil then
            log.error({
                message = string.format(
                    "%s: process_unstake: unstake_id=%s, err=%s",
                    ERR_UNSTAKE_UPDATE_FAILED,
                    tostring(unstake_id),
                    tostring(err)
                ),
                [ALERT_TAG] = ALERT_CRIT,
            })
            box.rollback()
            return { res = nil, error = ERR_UNSTAKE_UPDATE_FAILED }
        end

        ::continue::
        count = util.safe_yield(count, 1000)
    end
    box.commit()

    return { res = affected_pids, error = nil }
end

function balance.get_withdrawals_suspended()
    local exist = box.space.global_settlement_status:get(SETTLEMENT_STATUS_ID)
    if exist == nil then
        return { res = false, error = nil }
    end

    return { res = exist.withdrawals_suspended, error = nil }
end

function balance.suspend_withdrawals()
    local new_item = {
        SETTLEMENT_STATUS_ID,
        "",
        true
    }

    local status, res = pcall(
        function()
            return box.space.global_settlement_status:upsert(
                new_item,
                {
                    { '=', 'withdrawals_suspended', true }
                })
        end
    )

    if status == false then
        log.error(BalanceError:new(res))
        return { res = nil, error = res }
    end
    return { res = nil, error = nil }
end

function balance.processing_withdrawal(profile_id, txhash, bops_id)
    checks('number', 'string', 'string')

    local exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
        config.params.BALANCE_STATUS.CLAIMING,
        bops_id)
    if err ~= nil then
        box.rollback()
        return { res = nil, error = err }
    end
    if exist == nil then
        exist, err = balance.find_bop(config.params.BALANCE_TYPE.WITHDRAWAL, profile_id,
            config.params.BALANCE_STATUS.PROCESSING,
            bops_id)
        if err ~= nil then
            box.rollback()
            return { res = nil, error = err }
        end
    end
    if exist ~= nil and exist.profile_id ~= profile_id then
        log.error({
            message = string.format("%s: processing_withdrawal: profile_id=%d exist.profile_id=%d", ERR_INTEGRITY_ERROR,
                profile_id, exist.profile_id),
            [ALERT_TAG] = ALERT_CRIT,
        })
        return nil, ERR_INTEGRITY_ERROR
    end
    if exist == nil then
        log.error(BalanceError:new("NO_CLAIMING_WITHDRAWAL for profile_id=%d", profile_id))
        return { res = nil, error = "NO_CLAIMING_WITHDRAWAL" }
    end
    local processing = config.params.BALANCE_STATUS.PROCESSING
    if exist.status ~= processing or exist.txhash ~= txhash then
        local id = exist.id
        local balance_op, err = archiver.update(box.space.balance_operations, id, {
            { '=', 'txhash', txhash },
            { '=', 'status', processing },
        })
        if err ~= nil then
            log.error(BalanceError:new("BALANCEOP_UPDATE_FAILED for id=%s, txhash=%s, err=%s", id, txhash, err))
            return { res = nil, error = err }
        elseif balance_op == nil then
            log.error(BalanceError:new("NEVER_HAPPENED no withdrawal with id=%s", id))
            return { res = nil, error = "NEVER_HAPPENED" }
        else
            notif_profile(balance_op.profile_id, balance_op:tomap({ names_only = true }))
        end
    end
    return { res = nil, error = nil }
end

function balance.get_last_processed_block_number(for_contract, chain_id, event_type)
    checks('string', 'number', 'string')

    local exist = box.space.processed_blocks:get({ for_contract, chain_id, event_type })
    if exist == nil then
        return { res = nil, error = nil }
    end
    return { res = exist.last_processed_block, error = nil }
end

function balance.set_last_processed_block_number(block_number, for_contract, chain_id, event_type)
    checks('string', 'string', 'number', 'string')

    local new_item = {
        for_contract,
        chain_id,
        event_type,
        block_number,
    }

    local status, res = pcall(
        function()
            return box.space.processed_blocks:upsert(
                new_item,
                {
                    { '=', 'contract_address',     for_contract },
                    { '=', 'chain_id',             chain_id },
                    { '=', 'event_type',           event_type },
                    { '=', 'last_processed_block', block_number }
                })
        end
    )

    if status == false then
        log.error(BalanceError:new(res))
        return { res = nil, error = res }
    end
    return { res = nil, error = nil }
end

function balance.completed_withdrawals(withdrawal_infos)
    checks('table')

    local success = config.params.BALANCE_STATUS.SUCCESS

    local count = 0
    for _, info in ipairs(withdrawal_infos) do
        local id = info[d.withdrawal_tx_info_id]
        local txhash = info[d.withdrawal_tx_info_hash]
        local exist = box.space.balance_operations:get(id)
        if exist == nil then
            log.error(BalanceError:new("NEVER_HAPPENED no withdrawal with id=%s", id))
        elseif exist.status ~= success or exist.txhash ~= txhash then
            local balance_op, err = archiver.update(box.space.balance_operations, id, {
                { '=', 'txhash', txhash },
                { '=', 'status', success },
            })
            if err ~= nil then
                log.error(BalanceError:new("BALANCEOP_UPDATE_FAILED for id=%s, txhash=%s, err=%s", id, txhash, err))
                return { res = nil, error = err }
            elseif balance_op == nil then
                log.error(BalanceError:new("NEVER_HAPPENED no withdrawal with id=%s", id))
            else
                notif_profile(balance_op.profile_id, balance_op:tomap({ names_only = true }))
            end
        end

        count = util.safe_yield(count, 1000)
    end
    return { res = nil, error = nil }
end

local function _update_balance(ops_id, txhash, profile_id, ops_type, amount, exchange_add, tm)
    checks('?string', "?string", 'number', 'string', 'decimal', 'decimal', '?number')

    local insideTx = false
    if box.is_in_txn() then
        insideTx = true
    end

    local id, id2, wallet

    -- CAN'T update balance TWICE
    if ops_id ~= "" then
        local exist = box.space.balance_operations:get(ops_id)
        if exist ~= nil then
            if exist.status == config.params.BALANCE_STATUS.SUCCESS then
                local err = string.format("ops_id=%s already success", ops_id)
                log.error(err)
                return { res = nil, error = err }
            end
        end

        id = ops_id
        id2 = exist.ops_id2
        wallet = exist.wallet
    else
        id = uuid.str()
        id2 = id
        wallet = ""
    end

    if insideTx == false then
        box.begin()
    end

    if tm == nil then
        tm = time.now()
    end

    local err, res_ops
    err = balance.increase_balance_sum(profile_id, amount)
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = tostring(err) }
    end

    res_ops, err = archiver.replace(box.space.balance_operations, {
        id,
        config.params.BALANCE_STATUS.SUCCESS,
        "",
        txhash,
        profile_id,
        wallet,
        ops_type,
        id2,
        amount,
        tm,
        0,

        "",
        0,
        "",
    })
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    if exchange_add ~= 0 then
        local which_wallet = {}

        if ops_type == config.params.BALANCE_TYPE.FEE or ops_type == config.params.BALANCE_TYPE.REFERRAL_PAYOUT then
            -- WE PAY FEE or REFERRAL_PAYOUT to separate wallet
            which_wallet = {
                FEE_WALLET_ID,
                exchange_add,
                tm
            }
        else
            -- BUT all other payments ADD/SUB from exchange
            which_wallet = {
                EXCHANGE_WALLET_ID,
                exchange_add,
                tm
            }
        end

        -- MAKE OPERATION HERE or revert
        local status, res = pcall(
            function()
                local res = box.space.exchange_wallets:upsert(which_wallet, {
                    { '+', 'balance',      exchange_add },
                    { '=', 'last_updated', tm } })

                -- for referral payout check that we don't dip negative.
                if ops_type == config.params.BALANCE_TYPE.REFERRAL_PAYOUT then
                    local b = box.space.exchange_wallets:get(FEE_WALLET_ID)
                    if b.balance < decimal.new(0) then
                        error({ err = ERR_REFERRAL_PAYOUT_NEGATIVE_FEE_WALLET })
                    end
                end

                return res
            end)
        if status == false then
            box.rollback()
            log.error(BalanceError:new(res.err))
            return { res = nil, error = res.err }
        end
    end

    if insideTx == false then
        box.commit()
    end

    return { res = res_ops, error = nil }
end

function balance.get_exchange_wallets_data()
    local res = {}

    local exchange = box.space.exchange_wallets:get(EXCHANGE_WALLET_ID)
    if exchange ~= nil then
        table.insert(res, exchange:totable())
    end

    local fee_wallet = box.space.exchange_wallets:get(FEE_WALLET_ID)
    if fee_wallet ~= nil then
        table.insert(res, fee_wallet:totable())
    end

    return { res = res, error = nil }
end

function balance.get_balance(profile_id)
    local exist = box.space.balance_sum:get(profile_id)
    if exist == nil then
        return ZERO
    end

    return exist.balance
end

function balance.pay_realized_pnl(profile_id, amount)
    checks('number', 'decimal')

    local exchange_add = amount * -1
    return _update_balance(
        "", -- ops_id
        "", -- txhash
        profile_id,
        config.params.BALANCE_TYPE.PNL,
        amount,
        exchange_add)
end

function balance.pay_fee(profile_id, amount)
    checks('number', 'decimal')

    local exchange_add = amount * -1
    return _update_balance(
        "", -- ops_id
        "", -- txhash
        profile_id,
        config.params.BALANCE_TYPE.FEE,
        amount,
        exchange_add)
end

function balance.pay_funding(profile_id, amount)
    checks('number', 'decimal')

    local exchange_add = amount * -1
    return _update_balance(
        "", -- ops_id
        "", -- txhash
        profile_id,
        config.params.BALANCE_TYPE.FUNDING,
        amount,
        exchange_add)
end

function balance.process_referral_payout()
    local balance_ops = balance.get_balance_ops_in_state(
        config.params.BALANCE_TYPE.REFERRAL_PAYOUT,
        config.params.BALANCE_STATUS.PENDING
    )

    balance_ops = balance_ops.res
    local profile_ids = {}
    for _, balance_op in pairs(balance_ops) do
        if balance_op.status ~= config.params.BALANCE_STATUS.PENDING then
            local err = string.format("balance_op.status expected to be = %s but got = %s",
                config.params.BALANCE_STATUS.PENDING,
                tostring(balance_op.status))
            log.error(err)
            return { res = nil, error = err }
        end

        if balance_op.id == nil then
            local err = "balance_op.id expected to not be empty"
            log.error(err)
            return { res = nil, error = err }
        end

        local res = _update_balance(
            balance_op.id, -- ops_id
            "",            -- txhash
            balance_op.profile_id,
            config.params.BALANCE_TYPE.REFERRAL_PAYOUT,
            balance_op.amount,
            balance_op.amount * -1,
            tonumber(balance_op.timestamp)
        )

        if res['error'] ~= nil then
            local err = res['error']
            log.error(BalanceError:new(err))
            return { res = nil, error = err }
        end

        table.insert(profile_ids, balance_op.profile_id)
    end

    return { res = profile_ids, error = nil }
end

-- TODO: THIS PART should be removed later, it used only for testing
function balance.deposit_credit(profile_id, amount)
    checks('number', 'decimal')

    if amount <= 0 then
        return { res = nil, error = "NEGATIVE_OR_ZERO_DEPOSIT_AMOUNT" }
    end

    local id = uuid.str()
    local tm = time.now()

    box.begin()

    local res, err

    err = balance.increase_balance_sum(profile_id, amount)
    if err ~= nil then
        log.error(err)
        return { res = nil, error = tostring(err) }
    end

    res, err = archiver.insert(box.space.balance_operations, {
        id,
        config.params.BALANCE_STATUS.SUCCESS,
        "",
        "",
        profile_id,
        "",
        config.params.BALANCE_TYPE.CREDIT,
        id,
        amount,
        tm,
        0,

        "",
        0,
        "",
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    return { res = res, error = nil }
end

function balance.withdraw_credit(profile_id, amount)
    checks('number', 'decimal')

    if amount <= 0 then
        return { res = nil, error = "NEGATIVE_OR_ZERO_WITHDRAW_AMOUNT" }
    end

    local tm = time.now()
    local id = uuid.str()

    box.begin()

    local res, err

    err = balance.decrease_balance_sum(profile_id, amount)
    if err ~= nil then
        log.error(string.format("WITHDRAW_CREDIT_FAILED %s", tostring(err)))
        return { res = nil, error = tostring(err) }
    end

    res, err = archiver.insert(box.space.balance_operations, {
        id,
        config.params.BALANCE_STATUS.SUCCESS,
        "",
        "",
        profile_id,
        "",
        config.params.BALANCE_TYPE.WITHDRAW_CREDIT,
        id,
        amount,
        tm,
        0,

        "",
        0,
        "",
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end

    box.commit()

    return { res = res, error = nil }
end

-- internal api only

-- returns fee amount or error
function balance.withdraw_fee(max_fee, ops_id)
    checks('decimal', 'string')

    local op, err = dml.get(box.space.balance_operations.index.ops_id2, ops_id)
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end
    if op ~= nil then
        return { res = op:tomap().amount, error = nil }
    end

    local res, err = dml.get(box.space.exchange_wallets, FEE_WALLET_ID)
    if err ~= nil then
        log.error(BalanceError:new(err))
        return { res = nil, error = err }
    end
    if res == nil then
        return { res = nil, error = 'FEE_WALLET_NOT_FOUND' }
    end
    local wallet = res:tomap({ names_only = true })

    local id = 'w_' .. box.sequence.withdrawal_id_sequence:next()
    local tm = time.now()

    local pay_amount = decimal.abs(wallet.balance)
    if pay_amount > max_fee then
        pay_amount = max_fee
    end
    if wallet.balance < 0 then
        pay_amount = -pay_amount
    end

    local _, err = dml.atomic(function()
        local _, err

        _, err = dml.update(box.space.exchange_wallets, wallet.wallet_id, {
            { '-', 'balance',      pay_amount },
            { '=', 'last_updated', tm },
        })
        if err ~= nil then
            log.error(BalanceError:new(err))
            return nil, err
        end

        -- TODO: do we nede due_block here?
        _, err = archiver.insert(box.space.balance_operations, {
            id,
            config.params.BALANCE_STATUS.SUCCESS,
            "", -- no reason
            ops_id,
            ANONYMOUS_ID,
            "", -- no wallet
            config.params.BALANCE_TYPE.WITHDRAW_FEE,
            ops_id,
            pay_amount,
            tm,
            0,

            "",
            0,
            "",
        })
        if err ~= nil then
            log.error(BalanceError:new(err))
            return nil, err
        end

        return nil, nil
    end)
    if err ~= nil then
        return { res = nil, error = err }
    end

    return { res = pay_amount, error = nil }
end

function balance.create_rebalance_ops(wallet, profile_id, amount, exchange_id, chain_id, to_contract)
    checks('string', 'number', 'decimal', 'string', 'number', 'string')
    box.begin()

    local withdrawal_id = box.sequence.withdrawal_id_sequence:next()
    if withdrawal_id == 0 then
        local timestamp = math.floor(10 * clock.time())
        local offset = balance._shard_num * OFFSET_MULTIPLIER
        withdrawal_id = timestamp + offset
        box.sequence.withdrawal_id_sequence:set(withdrawal_id)
    end

    local id2 = 'w_' .. tostring(withdrawal_id)
    local tm = time.now()

    local res, err = archiver.insert(box.space.balance_operations, {
        id2,
        config.params.BALANCE_STATUS.CLAIMABLE,
        "rebalance",
        "",
        profile_id,
        wallet,
        config.params.BALANCE_TYPE.WITHDRAWAL,
        id2,
        amount,
        tm,
        BLOCK_NUM_NEVER,

        exchange_id,
        chain_id,
        to_contract,
    })
    if err ~= nil then
        box.rollback()
        log.error(BalanceError:new(res))
        return { res = nil, error = err }
    end

    box.commit()

    return res
end

--[[
    BELOW only for debug purpose to understand the status of real blockchain
    NEVER should be used in prod
--]]

function balance.get_settlement_state()
    local res = {}

    local exist = box.space.global_settlement_status:get(SETTLEMENT_STATUS_ID)
    if exist ~= nil then
        return { res = exist.withdrawals_suspended, error = nil }
    end

    return { res = false, error = nil }
end

function balance.get_processing_ops()
    local res = fun.filter(
        function(item)
            return
                (item.ops_type == config.params.BALANCE_TYPE.WITHDRAWAL or item.ops_type == config.params.BALANCE_TYPE.DEPOSIT) and
                item.status ~= config.params.BALANCE_STATUS.SUCCESS
        end,
        box.space.balance_operations:pairs()):totable()

    return { res = res, error = nil }
end

function balance.delete_unknown_ops()
    local total = 0

    for _, ops in box.space.balance_operations.index.type_status_exchange_id_chain_id:pairs(
        {
            config.params.BALANCE_TYPE.DEPOSIT,
            config.params.BALANCE_STATUS.UNKNOWN,
        },
        { iterator = "EQ" }
    )
    do
        if ops.status == config.params.BALANCE_STATUS.UNKNOWN then
            box.space.balance_operations:delete(ops.id)
            total = total + 1
        else
            log.error({
                message = string.format("%s: DELETE status = %s", ERR_INTEGRITY_ERROR, ops.status),
                [ALERT_TAG] = ALERT_CRIT,
            })
        end
    end

    return { res = total, error = nil }
end

function balance.test_replace_ops(ops)
    local res = box.space.balance_operations:replace {
        ops[1],
        ops[2],
        ops[3],
        ops[4],
        ops[5],
        ops[6],
        ops[7],
        ops[8],
        ops[9],
        ops[10],

        "fake",
        0,

        "",
        0,
        "",
    }

    return { res = res, error = nil }
end

function balance.test_clear_spaces()
    box.space.withdraw_lock:drop()
    box.space.processed_blocks:drop()
    box.space.balance_operations:drop()
    box.space.balance_sum:drop()
    box.space.global_settlement_status:drop()
    box.space.exchange_wallets:drop()
    box.sequence.withdrawal_id_sequence:drop()
    box.sequence.unstake_id_sequence:drop()
end

return balance
