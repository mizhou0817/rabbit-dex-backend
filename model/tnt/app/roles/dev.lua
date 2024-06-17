local cartridge = require('cartridge')
local checks = require('checks')

local function init(opts) -- luacheck: no unused args
    if opts.is_master then
        box.schema.user.grant('guest',
            'read,write,execute,create,session,alter,drop,usage',
            'universe',
            nil, { if_not_exists = true }
        )
    end

    return true
end

return {
    role_name = 'dev',
    init = init,
    dependencies = {'cartridge.roles.vshard-router'},
}

