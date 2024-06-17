local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.yes
local MARKET = config.markets["YES-USD"]

return market.new(ROLE_NAME, MARKET)
