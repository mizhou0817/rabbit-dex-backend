local checks = require('checks')
local log = require('log')

local config = require('app.config')
local en = require('app.engine.enricher')
local o = require('app.engine.order')
local p = require('app.engine.position')
local errors = require('app.lib.errors')
local tuple = require('app.tuple')

local ExtendedError = errors.new_class("extended-data")

local cp = config.params

local format = {}
for _, f in ipairs(p.format) do
    table.insert(format, f)
end
table.extend(format, {
    -- these fields are not described for entities stored in space,
    -- but need to add it here to serialize extednded_position correclty
    --TODO: later move it to entity format description
    {name = 'shard_id', type = 'string'},
    {name = 'archive_id', type = 'number'},

    {name = 'stop_loss', type = '*'},
    {name = 'take_profit', type = '*'},
})

local function bind(pos)
    checks('table|engine_position')
    return tuple.new(pos, format, 'extended_position')
end

local stop_loss_orders = {cp.ORDER_TYPE.STOP_LOSS, cp.ORDER_TYPE.STOP_LOSS_LIMIT}
local take_profit_orders = {cp.ORDER_TYPE.TAKE_PROFIT, cp.ORDER_TYPE.TAKE_PROFIT_LIMIT}

return {
    get_extended_position = function(profile_id, market_id)
        checks("number", "string")

        -- current SLTP orders logic awaits only one SL and one TP per profile
        local res

        local stop_loss = box.NULL
        for _, order_type in ipairs(stop_loss_orders) do
            res = o.get_orders(profile_id, cp.ORDER_STATUS.PLACED, order_type)
            if res.error ~= nil then
                log.error(ExtendedError:new(res.error))
                return {res = nil, error = res.error}
            end
            if #res.res > 0 then
                stop_loss = res.res[1]
                break
            end
        end

        local take_profit = box.NULL
        for _, order_type in ipairs(take_profit_orders) do
            res = o.get_orders(profile_id, cp.ORDER_STATUS.PLACED, order_type)
            if res.error ~= nil then
                log.error(ExtendedError:new(res.error))
                return {res = nil, error = res.error}
            end
            if #res.res > 0 then
                take_profit = res.res[1]
                break
            end
        end

        local position = p.get_position(profile_id, market_id)
        if position == nil then
            return {res = nil, error = nil}
        end

        position = en.enrich_position(position)

        position = bind(position)
        position.stop_loss = stop_loss
        position.take_profit = take_profit

        return {res = position, error = nil}
    end,
}
