local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.link
local MARKET = config.markets["LINK-USD"]

return market.new(ROLE_NAME, MARKET)
