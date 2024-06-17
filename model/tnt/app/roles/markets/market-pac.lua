local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.pac
local MARKET = config.markets["PAC-USD"]

return market.new(ROLE_NAME, MARKET)