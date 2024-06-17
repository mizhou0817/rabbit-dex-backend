local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.sats
local MARKET = config.markets["SATS1000000-USD"]

return market.new(ROLE_NAME, MARKET)
