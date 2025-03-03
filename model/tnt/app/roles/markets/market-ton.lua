local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.ton
local MARKET = config.markets["TON-USD"]

return market.new(ROLE_NAME, MARKET)
