local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.rlb
local MARKET = config.markets["RLB-USD"]

return market.new(ROLE_NAME, MARKET)
