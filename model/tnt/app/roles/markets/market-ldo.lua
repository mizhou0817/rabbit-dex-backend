local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.ldo
local MARKET = config.markets["LDO-USD"]

return market.new(ROLE_NAME, MARKET)
