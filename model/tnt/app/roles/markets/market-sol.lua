local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.sol
local MARKET = config.markets["SOL-USD"]

return market.new(ROLE_NAME, MARKET)
