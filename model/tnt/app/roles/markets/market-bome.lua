local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.bome
local MARKET = config.markets["BOME-USD"]

return market.new(ROLE_NAME, MARKET)
