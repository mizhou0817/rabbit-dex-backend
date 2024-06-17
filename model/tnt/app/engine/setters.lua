local checks = require('checks')

local tier = require('app.engine.tier')

local setters = {}

function setters.update_profiles_to_tiers(profiles_tiers)
    checks('table')

    local res, err = tier.update_profiles_to_tiers(profiles_tiers)
    return {res = res, error = err}
end

return setters
