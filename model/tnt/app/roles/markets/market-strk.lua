local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.strk
local MARKET = config.markets["STRK-USD"]

return market.new(ROLE_NAME, MARKET)
