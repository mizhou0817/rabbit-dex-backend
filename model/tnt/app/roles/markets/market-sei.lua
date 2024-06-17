local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.sei
local MARKET = config.markets["SEI-USD"]

return market.new(ROLE_NAME, MARKET)
