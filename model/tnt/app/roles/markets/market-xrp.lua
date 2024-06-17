local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.xrp
local MARKET = config.markets["XRP-USD"]

return market.new(ROLE_NAME, MARKET)
