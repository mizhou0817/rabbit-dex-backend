local balance = require('app.balance')
local ag = require('app.engine.aggregate')
local candles = require('app.engine.candles')
local extended = require('app.engine.extended_data')
local fortest = require('app.engine.fortest')
local market = require('app.engine.market')
local notif = require('app.engine.notif')
local order = require('app.engine.order')
local orderbook = require('app.engine.orderbook')
local position = require('app.engine.position')
local profile = require('app.engine.profile')
local trade = require('app.engine.trade')
local stats = require('app.stats')
local mt = require('app.engine.maintenance')
local router = require('app.engine.router')

local function init_spaces(market_data)
    -- INIT 3rd party modules --
    market.init_spaces(market_data)
    router.init_spaces()

    ag.init_spaces()
    trade.init_spaces()
    balance.init_spaces(0)
    notif.init_spaces()
    order.init_spaces()
    position.init_spaces()
    profile.init_spaces()
    candles.init_spaces()
    fortest.init_spaces()
    orderbook.init_spaces(market_data)

    mt.init_spaces()
    return true
end

local engine = require('app.engine.engine')
local methods = require('app.engine.methods')
local periodics = require('app.engine.periodics')

return {
    engine = engine,
    market = market,
    methods = methods,
    periodics = periodics,
    profile = profile,
    position = position,
    order = order,
    orderbook = orderbook,
    trade = trade,
    fortest = fortest,
    balance = balance,
    candles = candles,
    stats = stats,
    extended = extended,
    init_spaces = init_spaces
}
