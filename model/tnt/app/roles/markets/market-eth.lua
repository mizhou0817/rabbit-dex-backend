local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.eth
local MARKET = config.markets["ETH-USD"]

return market.new(ROLE_NAME, MARKET)
