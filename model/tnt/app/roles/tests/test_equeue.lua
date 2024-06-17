local cartridge = require('cartridge')
local checks = require('checks')
local equeue = require('app.enginequeue')

local function init(opts) -- luacheck: no unused args
    if opts.is_master then
        equeue.init_spaces()

        box.schema.func.create('equeue', {if_not_exists = true})
    end

    rawset(_G, 'equeue', equeue)
    return true
end

return {
    role_name = 'test_equeue',
    init = init,
    utils = {
        equeue = equeue
    }
}

