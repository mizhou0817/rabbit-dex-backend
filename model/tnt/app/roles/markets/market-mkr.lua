local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.mkr
local MARKET = config.markets["MKR-USD"]

return market.new(ROLE_NAME, MARKET)
