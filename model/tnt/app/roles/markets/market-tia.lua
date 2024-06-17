local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.tia
local MARKET = config.markets["TIA-USD"]

return market.new(ROLE_NAME, MARKET)
