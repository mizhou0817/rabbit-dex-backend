local config = require('app.config')
local market = require('app.roles.abstract_market')

local ROLE_NAME = config.sys.ROLES.meme
local MARKET = config.markets["MEME-USD"]

return market.new(ROLE_NAME, MARKET)
