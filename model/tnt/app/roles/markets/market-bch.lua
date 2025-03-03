local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.bch
local MARKET = config.markets["BCH-USD"]

return market.new(ROLE_NAME, MARKET)
