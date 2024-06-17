local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.ordi
local MARKET = config.markets["ORDI-USD"]

return market.new(ROLE_NAME, MARKET)
