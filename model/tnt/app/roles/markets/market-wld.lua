local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.wld
local MARKET = config.markets["WLD-USD"]

return market.new(ROLE_NAME, MARKET)
