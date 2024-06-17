local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.matic
local MARKET = config.markets["MATIC-USD"]

return market.new(ROLE_NAME, MARKET)
