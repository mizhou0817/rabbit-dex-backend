local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.doge
local MARKET = config.markets["DOGE-USD"]

return market.new(ROLE_NAME, MARKET)
