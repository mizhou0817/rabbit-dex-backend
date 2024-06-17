local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.taiko
local MARKET = config.markets["TAIKO-USD"]

return market.new(ROLE_NAME, MARKET)
