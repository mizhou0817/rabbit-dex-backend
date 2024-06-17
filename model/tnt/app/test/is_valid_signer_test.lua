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

local g = t.group('is_valid_signer')
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

g.before_each(function(cg)
    archiver.init_sequencer('profile')
    profile.init_spaces({})
end)

g.after_each(function(cg)
    box.space.profile:drop()
    box.space.vault_permissions:drop()
end)


g.test_signer = function(cg)

    local p = profile.profile

    local vault = "0xVAULT"
    local wallet = "0xWALLET"
    local role = 1

    local is_valid = p.is_valid_signer(vault, wallet, role)
    t.assert_is(is_valid.res, false)

    local res = p.add_permission(vault, wallet, role)
    t.assert_is_not(res, nil)

    is_valid = p.is_valid_signer(vault, wallet, role)
    t.assert_is(is_valid.res, true)

    is_valid = p.is_valid_signer(vault, wallet, role + 1)
    t.assert_is(is_valid.res, false)
end

