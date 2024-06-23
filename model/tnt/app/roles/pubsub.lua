local checks = require('checks')
local fun = require('fun')
local log = require('log')

local logger = require('rbx.logger')
local pubsub = require('rbx.pubsub')

require('app.config.constants')

local role_name = 'pubsub'
local app

--TODO: move to config
local centrifugo_uri = 'http://centrifugo:10000';

local function stop()
    if app then
        pubsub.stop(app)
        app = nil
        rawset(_G, 'pubsub_publish', nil)
    end

    return true
end

local function init(opts) -- luacheck: no unused args
    local err

    logger.init()

    if not app then
        local opts = {address = centrifugo_uri}
        app, err = pubsub.start(opts)
        if err then
            log.error(err:unpack())
            error(err:unpack())
        end
        rawset(_G, 'pubsub_publish', function(channel, data)
            checks('string', 'string')
            local err = pubsub.publish(app, channel, data)
            if err then
                log.error(fun.chain(err:unpack(), {[ALERT_TAG] = ALERT_CRIT}):tomap())
            end
        end)
    end

    return true
end

return {
    role_name = role_name,
    init = init,
    stop = stop,
}
