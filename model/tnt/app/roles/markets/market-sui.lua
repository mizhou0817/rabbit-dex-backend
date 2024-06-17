local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.sui
local MARKET = config.markets["SUI-USD"]

return market.new(ROLE_NAME, MARKET)
