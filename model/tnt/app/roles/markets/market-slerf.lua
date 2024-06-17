local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.slerf
local MARKET = config.markets["SLERF-USD"]

return market.new(ROLE_NAME, MARKET)
