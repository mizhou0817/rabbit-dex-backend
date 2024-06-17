local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.inj
local MARKET = config.markets["INJ-USD"]

return market.new(ROLE_NAME, MARKET)
