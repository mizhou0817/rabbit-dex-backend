local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.rndr
local MARKET = config.markets["RNDR-USD"]

return market.new(ROLE_NAME, MARKET)
