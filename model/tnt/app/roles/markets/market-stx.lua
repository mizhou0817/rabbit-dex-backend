local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.stx
local MARKET = config.markets["STX-USD"]

return market.new(ROLE_NAME, MARKET)
