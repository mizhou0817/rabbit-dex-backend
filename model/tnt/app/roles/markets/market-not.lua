local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES._not
local MARKET = config.markets["NOT-USD"]

return market.new(ROLE_NAME, MARKET)
