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
local vault_profile_id
local vault_wallet = "0xe1a243b4e5b6"
local manager_profile_id = 13
local treasurer_profile_id = 14
local performance_fee = decimal.new(0.2)

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

g.before_each(function(cg)
    box.space.balance_sum:truncate()
    box.space.balance_operations:truncate()
    box.space.profile:truncate()
    box.space.profile_cache:truncate()
end)

g.test_vault_liquidation = function(cg)
    local trader1_profile = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x111", DEFAULT_EXCHANGE_ID).res
    local trader2_profile = profile.profile.create(config.params.PROFILE_TYPE.TRADER, "active", "0x222", DEFAULT_EXCHANGE_ID).res
    local holding_shares = decimal.new(123)
    local holding_nav = decimal.new(456)
    local holding_price = decimal.new(8)
    local err
    _, err = archiver.upsert(box.space.vault_holdings, {
        vault_profile_id,
        trader1_profile.id,
        holding_shares,
        holding_nav,
        holding_price
    }, {
        { '=', 'shares',      holding_shares },
        { '=', 'entry_nav',   holding_nav },
        { '=', 'entry_price', holding_price }
    })
    t.assert_is(err, nil)
    local t1_holding = box.space.vault_holdings:get({vault_profile_id, trader1_profile.id})
    t.assert_equals(t1_holding.shares, holding_shares)
    _, err = archiver.upsert(box.space.vault_holdings, {
        vault_profile_id,
        trader2_profile.id,
        holding_shares,
        holding_nav,
        holding_price
    }, {
        { '=', 'shares',      holding_shares },
        { '=', 'entry_nav',   holding_nav },
        { '=', 'entry_price', holding_price }
    })
    t.assert_is(err, nil)
    local t2_holding = box.space.vault_holdings:get({vault_profile_id, trader2_profile.id})
    t.assert_equals(t2_holding.shares, holding_shares)
    local vaults = {}
    vaults[1] = vault_profile_id
    profile.profile.liquidated_vaults(vaults)
    local vault_info = box.space.vaults:get(vault_profile_id)
    t.assert(vault_info.status == config.params.VAULT_STATUS.SUSPENDED)
    t1_holding = box.space.vault_holdings:get({vault_profile_id, trader1_profile.id})
    t.assert_equals(t1_holding.shares, ZERO)
    t.assert_equals(t1_holding.entry_nav, ZERO)
    t.assert_equals(t1_holding.entry_price, ONE)
    t2_holding = box.space.vault_holdings:get({vault_profile_id, trader2_profile.id})
    t.assert_equals(t2_holding.shares, ZERO)
    t.assert_equals(t2_holding.entry_nav, ZERO)
    t.assert_equals(t2_holding.entry_price, ONE)
end