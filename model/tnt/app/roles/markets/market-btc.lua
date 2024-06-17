local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.btc
local MARKET = config.markets["BTC-USD"]

return market.new(ROLE_NAME, MARKET)
