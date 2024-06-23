local log = require('log')

local archiver = require('app.archiver')
local balance = require('app.balance')
local errors = require('app.lib.errors')
local airdrop = require('app.profile.airdrop')
local fee = require('app.profile.fee')

local integrity = require('app.profile.integrity')
local p = require('app.profile.profile')
local revert = require('app.revert')
local tuple = require('app.tuple')
local dynamic = require('app.profile.dynamic')
local wdm = require('app.wdm')

require("app.config.constants")

local ApiSpacesError = errors.new_class("API_SPACES")

local P = {
    profile_format = {
        {name = 'id', type = 'unsigned'},
        {name = 'profile_type', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'wallet', type = 'string'},
        {name = 'created_at', type = 'number'},
        {name = 'exchange_id', type = 'string'},
    },
    profile_strict_type = 'profile',
}

local function init_spaces(opts)
    wdm.init_spaces()

    integrity.init_spaces()
    dynamic.init_spaces()

    revert.init_spaces()

    fee.init_spaces()

    balance.init_spaces(1)
    airdrop.init_spaces()

    box.schema.sequence.create('PID', {start = 0, min = 0, if_not_exists = true})
    local profile, err = archiver.create('profile', { if_not_exists = true }, P.profile_format, {
        name = 'primary',
        sequence = 'PID',
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(ApiSpacesError:new(err))
        error(err)
    end

    profile:create_index('profile_type', {
        unique = false,
        parts = {{field = 'profile_type'}},
        if_not_exists = true })

    -- dfd
    -- if it has the due_block field, add the type_due_block index
    local format = profile:format()
    for _, field in ipairs(format) do
        if field.name == 'exchange_id' then
            profile:create_index('exchange_id',
            {
                parts = { {field = 'exchange_id'} },
                unique = false,
                if_not_exists = true
            })

            profile:create_index('exchange_id_wallet', {
                unique = true,
                parts = {{field = 'exchange_id'}, {field = 'wallet'}},
                if_not_exists = true
            })
        end
    end

    local profile_cache, err = archiver.create('profile_cache', {if_not_exists = true}, {
        {name = 'id', type = 'unsigned'},
        {name = 'profile_type', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'wallet', type = 'string'},

        {name = 'last_update', type = 'number'},
        {name = 'balance', type = 'decimal'},
        {name = 'account_equity', type = 'decimal'},
        {name = 'total_position_margin', type = 'decimal'},
        {name = 'total_order_margin', type = 'decimal'},
        {name = 'total_notional', type = 'decimal'},
        {name = 'account_margin', type = 'decimal'},
        {name = 'withdrawable_balance', type = 'decimal'},
        {name = 'cum_unrealized_pnl', type = 'decimal'},
        {name = 'health', type = 'decimal'},
        {name = 'account_leverage', type = 'decimal'},
        {name = 'cum_trading_volume', type = 'decimal'},
        {name = 'leverage', type = '*'},
        {name = 'last_liq_check', type = 'number'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(ApiSpacesError:new(err))
        error(err)
    end
    box.space.profile_cache:truncate()

    profile_cache:create_index('for_liquidation', {
        unique = false,
        parts = {{field = 'status'}, {field = "id"}},
        if_not_exists = true })


    local profile_meta = box.schema.space.create('profile_meta', {if_not_exists = true})
    profile_meta:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'market_id', type = 'string'},
        {name = 'status', type = 'string'},

        {name = 'cum_unrealized_pnl', type = 'decimal'},
        {name = 'total_notional', type = 'decimal'},
        {name = 'total_position_margin', type = 'decimal'},
        {name = 'total_order_margin', type = 'decimal'},
        {name = 'initial_margin', type = 'decimal'},
        {name = 'market_leverage', type = 'decimal'},
        {name = 'balance', type = 'decimal'},
        {name = 'cum_trading_volume', type = 'decimal'},
        {name = 'timestamp', type = 'number'},
    })
    box.space.profile_meta:truncate()

    profile_meta:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true })


    local exchange_total = box.schema.space.create('exchange_total', {if_not_exists = true})
    exchange_total:format({
        {name = 'id', type = 'number'},
        {name = 'trading_fee', type = 'decimal'},
        {name = 'total_balance', type = 'decimal'},
    })
    box.space.exchange_total:truncate()

    exchange_total:create_index('primary', {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true })

    local inv3_data, err = archiver.create('inv3_data', {temporary=true, if_not_exists = true}, {
        {name = 'id', type = 'number'},
        {name = 'valid', type = 'boolean'},
        {name = 'last_updated', type = 'number'},
        {name = 'margined_ae_sum', type = 'decimal'},
        {name = 'exchange_balance', type = 'decimal'},
        {name = 'insurance_balance', type = 'decimal'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(ApiSpacesError:new(err))
        error(err)
    end


    local vault_permissions = box.schema.space.create('vault_permissions', {if_not_exists = true})
    vault_permissions:format({
        {name = 'vault', type = 'string'},
        {name = 'wallet', type = 'string'},
        {name = 'role', type = 'number'},
    })

    vault_permissions:create_index('primary', {
        unique = true,
        parts = {{field = 'vault'}, {field = "wallet"}, {field = "role"}},
        if_not_exists = true })

    p.create_insurance(DEFAULT_EXCHANGE_ID)
end

local function bind(value)
    return tuple.new(value, P.profile_format, P.profile_strict_type)
end

-- tuple behaviour
local function tomap(self, opts)
    return self:tomap(opts)
end

return {
    init_spaces = init_spaces,
    bind = bind,
    tomap = tomap,
    profile = require('app.profile.profile'),
    getters = require('app.profile.getters'),
    setters = require('app.profile.setters'),
    cache = require('app.profile.cache'),
    periodics = require('app.profile.periodics'),
    fortest = require('app.profile.fortest'),
    balance = balance,
    fee = require('app.profile.fee'),
    airdrop = airdrop
}
