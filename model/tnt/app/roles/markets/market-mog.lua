local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.mog
local MARKET = config.markets["MOG1000-USD"]

return market.new(ROLE_NAME, MARKET)
