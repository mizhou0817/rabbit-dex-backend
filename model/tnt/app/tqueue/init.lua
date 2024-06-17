local queue_state = require('app.tqueue.abstract.queue_state')
local queue = nil

-- load all core drivers
local core_drivers = {
    fifo        = require('app.tqueue.abstract.driver.fifo'),
    fifottl     = require('app.tqueue.abstract.driver.fifottl'),
    utubettl    = require('app.tqueue.abstract.driver.utubettl'),
    limfifottl  = require('app.tqueue.abstract.driver.limfifottl'),
    pqueue       = require('app.tqueue.abstract.driver.pqueue'),
}

local function remove_utube_type()
    if box.space["_queue"] == nil then
        return
    end

    local found = false
    for _, tube in box.space._queue:pairs() do
        if tube.tube_type == "utube" then
            if box.space[tube.space_name] ~= nil then
                box.space[tube.space_name]:drop()

                box.space._queue:delete(tube.tube_name)
                found = true
            end
        end        
    end

    -- if we dropped utube spaces, let's free the sessions
    if found == true then
        local sps = {"_queue_consumers", "_queue_taken_2", "_queue_session_ids", "_queue_shared_sessions"}
        
        for _, space_name in ipairs(sps) do
            if box.space[space_name] ~= nil then
                box.space[space_name]:truncate()
            end
        end
    end
end

local function register_driver(driver_name, tube_ctr)
    if type(tube_ctr.create_space) ~= 'function' or
        type(tube_ctr.new) ~= 'function' then
        error('tube control methods must contain functions "create_space"'
              .. ' and "new"')
    end
    if queue.driver[driver_name] then
        error(('overriding registered driver "%s"'):format(driver_name))
    end
    queue.driver[driver_name] = tube_ctr
end

local deferred_opts = {}

-- We cannot call queue.cfg() while tarantool is in read_only mode.
-- This method stores settings for later original queue.cfg() call.
local function deferred_cfg(opts)
    opts = opts or {}

    for k, v in pairs(opts) do
        deferred_opts[k] = v
    end
end

queue = setmetatable({
    driver = core_drivers,
    register_driver = register_driver,
    state = queue_state.show,
    cfg = deferred_cfg,
    remove_utube_type = remove_utube_type,
}, { __index = function()
        print(debug.traceback())
        error('Please configure box.cfg{} in read/write mode first')
    end
})

-- Used to store the original methods
local orig_cfg = nil
local orig_call = nil

local wrapper_impl

local function cfg_wrapper(...)
    box.cfg = orig_cfg
    return wrapper_impl(...)
end

local function cfg_call_wrapper(cfg, ...)
    local cfg_mt = getmetatable(box.cfg)
    cfg_mt.__call = orig_call
    return wrapper_impl(...)
end

local function wrap_box_cfg()
    if type(box.cfg) == 'function' then
        -- box.cfg before the first box.cfg call
        orig_cfg = box.cfg
        box.cfg = cfg_wrapper
    elseif type(box.cfg) == 'table' then
        -- box.cfg after the first box.cfg call
        local cfg_mt = getmetatable(box.cfg)
        orig_call = cfg_mt.__call
        cfg_mt.__call = cfg_call_wrapper
    else
        error('The box.cfg type is unexpected: ' .. type(box.cfg))
    end
end

function wrapper_impl(...)
    local result = { pcall(box.cfg,...) }
    if result[1] then
        table.remove(result, 1)
    else
        wrap_box_cfg()
        error(result[2])
    end

    if box.info.ro == false then
        remove_utube_type()

        local abstract = require 'app.tqueue.abstract'
        for name, val in pairs(abstract) do
            rawset(queue, name, val)
        end
        abstract.driver = queue.driver
        -- Now the "register_driver" method from abstract will be used.
        queue.register_driver = nil
        setmetatable(queue, getmetatable(abstract))
        queue.cfg(deferred_opts)
        queue.start()
    else
        -- Delay a start until the box will be configured
        -- with read_only = false
        wrap_box_cfg()
    end
    return unpack(result)
end

--- Implementation of the “lazy start” procedure.
-- The queue module is loaded immediately if the instance was
-- configured with read_only = false. Otherwise, a start is
-- delayed until the instance will be configured with read_only = false.
local function queue_init()
    if rawget(box, 'space') ~= nil and box.info.ro == false then
        -- because of "lazy start" we need to do this check here
        remove_utube_type()

        -- The box was configured with read_only = false
        queue = require('app.tqueue.abstract')
        queue.driver = core_drivers
        queue.start()
    else
        wrap_box_cfg()
    end
end

queue_init()

return queue
