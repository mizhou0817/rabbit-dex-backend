local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.bonk
local MARKET = config.markets["BONK1000-USD"]

return market.new(ROLE_NAME, MARKET)
