local fio = require('fio')

local api = require('app.api')
local archiver = require('app.archiver')
local equeue = require('app.enginequeue')
local config = require('app.config')

local function mode()
    return {res = config.sys.MODE,  error=nil}
end


local function stop()
    return true
end

local function init(opts) -- luacheck: no unused args
    if opts.is_master then
        archiver.init_sequencer("gateway")

        api.init_spaces()
        equeue.init_spaces()
    end
    return true
end

local function validate_config(conf_new, conf_old) 
    return true
end

local function apply_config(conf, opts)
    if opts.is_master then
        box.schema.func.create('public', {if_not_exists = true})
        box.schema.func.create('internal', {if_not_exists = true})
        box.schema.func.create('next_task', {if_not_exists = true})
        box.schema.func.create('mode', {if_not_exists = true})

        require('app.util').migrator_upgrade(fio.pathjoin('migrations', 'api'))
    end

    rawset(_G, 'public', api.public)
    rawset(_G, 'internal', api.internal)
    rawset(_G, 'next_task', api.internal.next_task)
    rawset(_G, 'mode', mode)

    return true
end

return {
    role_name = 'gateway',
    init = init,
    stop = stop,
    validate_config = validate_config,
    apply_config = apply_config,
    utils = {
        public   = api.public,
        internal = api.internal,
        mode = mode
    },

    -- FUNCTIONS available for RPC calls --
    next_task = api.internal.next_task,
    --ack_task = api.internal.ack_task, --TODO: unused, remove it if it's true
    new_order = api.public.new_order,
    cancel_order = api.public.cancel_order,
    amend_order = api.public.amend_order,
    cancel_all = api.public.cancel_all
}