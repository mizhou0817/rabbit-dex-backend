local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.apt
local MARKET = config.markets["APT-USD"]

return market.new(ROLE_NAME, MARKET)
