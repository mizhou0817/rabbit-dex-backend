local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.bnb
local MARKET = config.markets["BNB-USD"]

return market.new(ROLE_NAME, MARKET)
