local checks = require('checks')
local decimal = require('decimal')
local log = require('log')
local uuid = require('uuid')

local config = require('app.config')
local time = require('app.lib.time')

require("app.config.constants")

local DEFAULT_LOCK_ID = 1

local function init_spaces()

    -- it's temp = true, so on replica, it will not work, and it's ok
    local init_lock = box.schema.space.create('init_lock', {temporary=true, if_not_exists = true})
    init_lock:format({
        {name = 'id', type = 'unsigned'},
        {name = 'is_valid', type = 'boolean'},  -- false or non-exist by default. We need explicitly signal that it's done
    })

    init_lock:create_index('primary', {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true })

end

local function is_valid()
    local lock = box.space.init_lock:get(DEFAULT_LOCK_ID)
    if not lock or not lock.is_valid then
        return false
    end

    return true
end

local function make_valid()
    local res = box.space.init_lock:replace({DEFAULT_LOCK_ID, true})

    if res == nil then
        log.error({
            message = 'INTEGRITY_CHANGED_ERROR: make_valid replace returned nil',
            [ALERT_TAG] = ALERT_CRIT,
        })
        return
    end

    log.info("INTEGRITY_CHANGED: to valid")
end

local function make_invalid()
    local res = box.space.init_lock:replace({DEFAULT_LOCK_ID, false})

    if res == nil then
        log.error({
            message = 'INTEGRITY_CHANGED_ERROR: make_invalid replace returned nil',
            [ALERT_TAG] = ALERT_CRIT,
        })
        return
    end

    log.error({
        message = 'INTEGRITY_CHANGED: to invalid',
        [ALERT_TAG] = ALERT_CRIT,
    })
end


return {
    init_spaces = init_spaces,
    is_valid = is_valid,
    make_valid = make_valid,
    make_invalid = make_invalid,
}
