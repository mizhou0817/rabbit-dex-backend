local checks = require('checks')
local decimal = require('decimal')
local fiber = require('fiber')
local json = require("json")
local log = require('log')

local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')
local equeue = require('app.enginequeue')
local action = require('app.action')

require("app.config.constants")

local DeadmanError = errors.new_class("DEADMAN")

local DEFAULT_SLEEP = 5000
local DEFAULT_TITLE = "deadman"

local function reset_timer_tick(title, new_min)
    checks('string', 'number')

    local exist = box.space.deadman_interval:get(title)

    if exist == nil or exist.timeout > new_min then
        return box.space.deadman_interval:replace({title, new_min})
    end

    -- if not deadmans left then just set timeout to default, to not spend resources
    local left = box.space.deadman:count()
    if left == 0 then
        return box.space.deadman_interval:replace({title, DEFAULT_SLEEP})    
    end
end

local function get_timer_tick(title)
    checks('string')

    local exist = box.space.deadman_interval:get(title)

    if exist == nil then
        return DEFAULT_SLEEP / 1000
    end

    return exist.timeout / 1000
end

local function put_cancel_all(profile_ids)
    checks("table")

        for _, profile_id in pairs(profile_ids) do
            for _, market in pairs(config.markets) do
                local market_id = market.id
        
                local r1 = equeue.which_qname(market_id, config.sys.QUEUE_TYPE.MARKET)
                if r1["error"] == nil then
                    local qname = r1["res"]

                                    -- CAN'T SEND TWO cancellall orders --
                    local oid = "cancelall"
                    local r2 = equeue.find_task(qname, tostring(profile_id), oid)
                    if r2["res"] == nil then
                        local order, task_data = action.pack_cancelall(profile_id, market_id)
                        local r3 = equeue.put(qname, task_data, profile_id, oid, config.params.ORDER_ACTION.CANCEL)
                        if r3["error"] ~= nil then
                            log.error(DeadmanError:new('DEADMAN cancel all error %s: profile_id=%s market_id=%s', r3["error"], tostring(profile_id), tostring(market_id)))
                        end            
                    end
                end
            end        
        end
end

local function init_spaces()
    local deadman = box.schema.space.create('deadman', {if_not_exists = true})
    deadman:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'timeout', type = 'number'},
        {name = 'last_updated', type = 'number'},
        {name = 'status', type = 'string'},
    })
    deadman:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })
    deadman:create_index('by_timeout', {
        unique = false,
        parts = {{field = 'timeout'}},
        if_not_exists = true })
    deadman:create_index('by_status', {
        unique = false,
        parts = {{field = 'status'}},
        if_not_exists = true })
    
    
    local deadman_interval = box.schema.space.create('deadman_interval', {if_not_exists = true})
    deadman_interval:format({
        {name = 'fiber_title', type = 'string'},
        {name = 'timeout', type = 'number'},
    })
    deadman_interval:create_index('primary', {
        unique = true,
        parts = {{field = 'fiber_title'}},
        if_not_exists = true })

    reset_timer_tick(DEFAULT_TITLE, DEFAULT_SLEEP)

    fiber.create(function()
        while true do
            local sleep_interval = get_timer_tick(DEFAULT_TITLE)

            -- go through all and
            local ids_to_cancel = {}
            local tm = time.now_milli()
            
            box.begin()
                for _, item in box.space.deadman.index.by_status:pairs(DEADMAN_ACTIVE, {iterator=box.index.EQ}) do
                    if tm > item.last_updated + item.timeout then
                        table.insert(ids_to_cancel, item.profile_id)
                        box.space.deadman:update(item.profile_id, {{'=', 'status', DEADMAN_CANCELED}})
                    end
                end

                if #ids_to_cancel > 0 then
                    put_cancel_all(ids_to_cancel)
                end
            box.commit()
            
            fiber.sleep(sleep_interval)
        end
    end)
end

local function deadman_replace(profile_id, timeout)
    checks("number", "number")

    reset_timer_tick(DEFAULT_TITLE, timeout)

    return box.space.deadman:replace({
        profile_id, 
        timeout,
        time.now_milli(),
        DEADMAN_ACTIVE,
    })
end

local function deadman_touch(profile_id)
    checks("number")

    if box.space.deadman:get(profile_id) == nil then
        return nil
    end

    return box.space.deadman:update(profile_id, {
        {'=', 'status', DEADMAN_ACTIVE},
        {'=', 'last_updated', time.now_milli()}
    })
end


local function deadman_delete(profile_id)
    checks("number")

    local res = box.space.deadman:delete(profile_id)

    local new_min = box.space.deadman.index.by_timeout:min() or DEFAULT_SLEEP

    reset_timer_tick(DEFAULT_TITLE, new_min)

    return res
end

local function deadman_get(profile_id)
    checks("number")

    return box.space.deadman:get(profile_id)
end


return {
    init_spaces = init_spaces,
    replace = deadman_replace,
    delete = deadman_delete,
    list = deadman_get,
    touch = deadman_touch,
}
