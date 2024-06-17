local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.ena
local MARKET = config.markets["ENA-USD"]

return market.new(ROLE_NAME, MARKET)
