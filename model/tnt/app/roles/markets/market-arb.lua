local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.arb
local MARKET = config.markets["ARB-USD"]

return market.new(ROLE_NAME, MARKET)
