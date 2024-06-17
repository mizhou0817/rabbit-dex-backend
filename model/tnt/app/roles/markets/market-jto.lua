local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.jto
local MARKET = config.markets["JTO-USD"]

return market.new(ROLE_NAME, MARKET)
