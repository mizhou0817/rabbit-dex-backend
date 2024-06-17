local checks = require('checks')
local log = require('log')

local balance = require('app.balance')
local config = require('app.config')
local ag = require("app.engine.aggregate")
local pos = require('app.engine.position')
local errors = require('app.lib.errors')
local time = require('app.lib.time')

require("app.config.constants")

local ExtendedError = errors.new_class("enricher-meta")

return {
    update_profile_meta = function(market, profile_id, profile_meta)
        checks("cdata", "number", "cdata")

                -- totals
        local cum_unrealized_pnl = ZERO
        local total_position_margin = ZERO
        local total_notional = ZERO

        -- Calc metrics and update ALL positionRT fir this profile
        for _, position in box.space.position.index.profile_id:pairs(profile_id, {iterator = 'EQ'}) do
            local sign = 1
            if position.side == config.params.SHORT then
                sign = -1
            end

            local unrealized_pnl_fair = position.size * (market.fair_price - position.entry_price) * sign
            local notional_fair = position.size * market.fair_price
            local position_margin = profile_meta.initial_margin * position.size * market.fair_price

            local den = (1 + sign * -1 * market.forced_margin)
            local liquidation_price = ZERO

            if den ~= 0 then
                liquidation_price = (1 + sign * -1 * profile_meta.initial_margin) * position.entry_price / den
            end

            local err = pos.replace({
                position.id,
                position.market_id,
                position.profile_id,
                position.size,
                position.side,
                position.entry_price,

                unrealized_pnl_fair,
                notional_fair,
                position_margin,
                liquidation_price,
                market.fair_price
            })

            if err ~= nil then
                log.error('update_profile_meta: error=%s', ExtendedError:new(err))
            end

            cum_unrealized_pnl = cum_unrealized_pnl + unrealized_pnl_fair
            total_position_margin = total_position_margin + position_margin
            total_notional = total_notional + notional_fair
        end

        local total_order_margin = profile_meta.initial_margin * ag.get_order_total_notional(profile_id)

        local profile_balance = balance.get_balance(profile_id)

        box.space.profile_meta:replace({
            profile_id,  -- profile_id
            profile_meta.market_id,
            config.params.PROFILE_STATUS.ACTIVE,

            cum_unrealized_pnl,
            total_notional,
            total_position_margin,
            total_order_margin,

            profile_meta.initial_margin,
            profile_meta.market_leverage,
            profile_balance,
            profile_meta.cum_trading_volume,
            time.now(),
        })

        return nil
    end,
}
