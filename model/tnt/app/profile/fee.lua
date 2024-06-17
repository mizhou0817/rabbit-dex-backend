local checks = require('checks')
local decimal = require('decimal')
local log = require('log')
local uuid = require('uuid')

local balance = require('app.balance')
local config = require('app.config')
local dml = require('app.dml')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')

require("app.config.constants")

local function init_spaces()
    local withdraw_fee = box.schema.space.create('withdraw_fee', {
        if_not_exists = true,
        format = {
            {name = 'ops_id', type = 'string'},
            {name = 'created_at', type = 'number'},
        },
    })

    withdraw_fee:create_index('primary', {
        unique = true,
        parts = {{field = 'ops_id'}},
        if_not_exists = true,
    })
end

local function withdraw_fee(profile_id, wallet, max_fee, opts)
    checks('number', 'string', 'decimal', '?table')
    if opts == nil then
        opts = {retries_count = 3}
    end
    local ops_id
    local tm = time.now()

    local tuples, err = dml.select(box.space.withdraw_fee)
    if err ~= nil then
        return nil, err
    end
    if #tuples > 0 then
        ops_id = tuples[1][1]
    else
        ops_id = uuid.str()
        local _, err = dml.insert(box.space.withdraw_fee, {ops_id, tm})
        if err ~= nil then
            return nil, err
        end
    end

    local total_fee = decimal.new(0)

    for _, market in pairs(config.markets) do
        local market_fee, err = util.retry(opts.retries_count, function()
            local res = rpc.callrw_engine(market.id, 'withdraw_fee', {max_fee, ops_id})
            if res['error'] ~= nil then
                error(res['error'])
            end
            return res['res']
        end)
        if err ~= nil then
            return nil, err
        end

        total_fee = total_fee + market_fee
    end

    if total_fee > 0 then
        local txhash= 'feewithdrawal_' .. ops_id
        local res = balance.create_withdraw_fee(profile_id, wallet, total_fee, txhash)
        if res['error'] ~= nil then
            return nil, res['error']
        end
    end

    box.space.withdraw_fee:truncate()
    return total_fee, nil
end

return {
    init_spaces = init_spaces,
    withdraw_fee = withdraw_fee,
}
