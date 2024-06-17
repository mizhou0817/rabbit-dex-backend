local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.blur
local MARKET = config.markets["BLUR-USD"]

return market.new(ROLE_NAME, MARKET)
