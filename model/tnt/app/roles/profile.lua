local confapplier = require('cartridge.confapplier')
local fio = require('fio')
local log = require('log')

local archiver = require('app.archiver')
local config = require('app.config')
local p = require('app.profile')
local tier = require('app.profile.tier')

local profile = p.profile
local getters = p.getters
local setters = p.setters
local balance = p.balance
local cache = p.cache
local periodics = p.periodics
local fortest = p.fortest
local airdrop = p.airdrop

local function stop()
    return true
end

local function init(opts) -- luacheck: no unused args
    if opts.is_master then
        archiver.init_sequencer("profile")

        p.init_spaces(opts)
        tier.init_spaces()
        box.schema.func.create('profile', {if_not_exists = true})
        box.schema.func.create('getters', {if_not_exists = true})
        box.schema.func.create('balance', {if_not_exists = true})
        box.schema.func.create('cache', {if_not_exists = true})
        box.schema.func.create('periodics', {if_not_exists = true})
        box.schema.func.create('get_cache_and_meta', {if_not_exists = true})
        box.schema.func.create('fortest', {if_not_exists = true})
        box.schema.func.create('airdrop', {if_not_exists = true})

        require('app.util').migrator_upgrade(fio.pathjoin('migrations', 'profile'))

        if config.sys.MODE ~= "sync" then
            local profile_conf = confapplier.get_readonly('profile')
            if profile_conf and profile_conf.validate_markets then
                local err = periodics.validate_markets()
                if err ~= nil then
                    error(err)
                end
            end

            periodics.start()
        else
            log.warn("*** SYNC MODE for profile role")
        end
    end
    rawset(_G, 'profile', profile)
    rawset(_G, 'getters', getters)
    rawset(_G, 'setters', setters)
    rawset(_G, 'balance', balance)
    rawset(_G, 'cache', cache)
    rawset(_G, 'periodics', periodics)
    rawset(_G, 'get_cache_and_meta', cache.get_cache_and_meta)
    rawset(_G, 'fortest', fortest)
    rawset(_G, 'archiver', archiver)
    rawset(_G, 'airdrop', airdrop)

    return true
end


return {
    role_name = 'profile',
    init = init,
    stop = stop,
    utils = {
        profile = profile,
        getters = getters,
        balance = balance,
        cache = cache,
        periodics = periodics,
        fortest = fortest,
        airdrop = airdrop
    },
    -- FUNCTIONS available for RPC calls --
    create = profile.create,
    get = profile.get,
    get_cache_and_meta = cache.get_cache_and_meta,
    ensure_cache = cache.ensure_cache,
}