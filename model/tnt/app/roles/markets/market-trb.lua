local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.trb
local MARKET = config.markets["TRB-USD"]

return market.new(ROLE_NAME, MARKET)
