local checks = require('checks')
local fio = require('fio')
local log = require('log')

local archiver = require('app.archiver')
local engine = require('app.engine')
local mt = require('app.engine.maintenance')
local setters = require('app.engine.setters')

local matching = engine.engine
local market = engine.market
local profile = engine.profile
local position = engine.position
local order = engine.order
local trade = engine.trade
local candles = engine.candles
local fortest = engine.fortest
local balance = engine.balance

local function stop()
    rawset(_G, 'engine', nil)
    matching.stop()

    return true
end

local function init(opts, market_config)
    checks('table', 'table')

    matching.init(market_config.id, market_config.min_tick, market_config.min_order)
    if opts.is_master then
        log.info("...starting market_id=%s", market_config.id)

        archiver.init_sequencer(market_config.id)
        engine.init_spaces(market_config)
        
        box.schema.func.create('engine', {if_not_exists = true})
        box.schema.func.create('market', {if_not_exists = true})
        box.schema.func.create('periodics', {if_not_exists = true})
        box.schema.func.create('profile', {if_not_exists = true})
        box.schema.func.create('trade', {if_not_exists = true})
        box.schema.func.create('candles', {if_not_exists = true})
        box.schema.func.create('fortest', {if_not_exists = true})
        box.schema.func.create('balance', {if_not_exists = true})
        box.schema.func.create('position', {if_not_exists = true})
        box.schema.func.create('order', {if_not_exists = true})
        box.schema.func.create('mt', {if_not_exists = true})

        -- market.update_index_price(MARKET.id, TEST_FAIR_PRICE)

        require('app.util').migrator_upgrade(fio.pathjoin('migrations', 'engine'))

        matching.start(market_config.id, market_config.min_tick, market_config.min_order)

        --TODO: it doesn't work here cuz role is not initialized yet. 
        -- Need to find the way how to call this func on instance stop
        -- mt.cancel_all_listed()
    end

    rawset(_G, 'order', order)
    rawset(_G, 'engine', matching)
    rawset(_G, 'market', market)
    rawset(_G, 'periodics', engine.periodics)
    rawset(_G, 'profile', profile)
    rawset(_G, 'trade', trade)
    rawset(_G, 'candles', candles)
    rawset(_G, 'fortest', fortest)
    rawset(_G, 'balance', balance)
    rawset(_G, 'position', position)
    rawset(_G, 'archiver', archiver)
    rawset(_G, 'mt', mt)
    rawset(_G, 'setters', setters)

    return true
end

local function new(role_name, market_config)
    checks('string', 'table')

    return {
        role_name = role_name,
        dependencies = {
            'app.roles.pubsub',
        },
        init = function(opts)
            checks('?table')
            return init(opts, market_config)
        end,
        stop = stop,
        utils = {
            engine = matching,
            market = market,
            periodics = engine.periodics,
            profile = profile,
            position = position,
            order = order,
            trade = trade,
            candles = candles,
            fortest = fortest,
            balance = balance,
            mt = mt
        },

        get_market = market.get_market,
        update_index_price = market.update_index_price,
        get_profile_meta = profile.get_profile_meta,
        get_market_data = market.get_market_data,
        get_orderbook_data = matching.get_orderbook_data,
        get_positions = position.get_positions,
        get_extended_position = engine.extended.get_extended_position,
        get_orders = order.get_orders,
        get_trade_data = trade.get_trade_data,
        get_order_by_id = order.get_order_by_id,
        get_exchange_wallets_data = balance.get_exchange_wallets_data,
        change_status = market.change_status,
        list_balance_ops = balance.list_operations,
        withdraw_fee = balance.withdraw_fee,
        total_volume = trade.total_volume,
        check_coid_for_reuse = order.check_coid_for_reuse,

        handle_revert = matching.handle_revert,
        mt_add_profile_id = mt.add_profile_id,
        mt_remove_profile_id = mt.remove_profile_id,
        mt_cancel_all_listed = mt.cancel_all_listed
    }
end

return {
    new = new,
}
