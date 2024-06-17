local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.mother
local MARKET = config.markets["MOTHER-USD"]

return market.new(ROLE_NAME, MARKET)
