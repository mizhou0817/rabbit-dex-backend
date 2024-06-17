local checks = require('checks')
local log = require('log')

local config = require('app.config')
local m = require('app.engine.market')
local pm = require('app.engine.profile')
local errors = require('app.lib.errors')

require("app.config.constants")

local ExtendedError = errors.new_class("enricher")

return {
    enrich_position = function(position)
        checks("table|engine_position")

        local res = m.get_market(position.market_id)
        if res.error ~= nil then
            log.error(ExtendedError:new(res.error))
            return position
        end
        local market = res.res

        res = pm.get_meta_for_profile(position.profile_id)
        if res.error ~= nil then
            log.error(ExtendedError:new(res.error))
            return position
        end
        local meta = res.res

        local sign = 1
        if position.side == config.params.SHORT then
            sign = -1
        end

        position.fair_price = market.fair_price
        position.unrealized_pnl = position.size * (market.fair_price - position.entry_price) * sign
        position.notional = position.size * market.fair_price
        position.margin = meta.initial_margin * position.size * market.fair_price
        local den = (1 + sign * -1 * market.forced_margin)
        if den ~= 0 then
            position.liquidation_price = (1 + sign * -1 * meta.initial_margin) * position.entry_price / den
        end

        return position
    end,
}
