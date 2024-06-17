local state    = require('app.tqueue.abstract.state')
local str_type = require('app.tqueue.compat').str_type
local log      = require('log')

local tube = {}
local method = {}


-- validate space of queue
local function validate_space(space)
    -- check indexes
    local indexes = {'task_id', 'status', 'by_inside_outside_priority_timestamp'}
    for _, index in pairs(indexes) do
        if space.index[index] == nil then
            error(string.format('space "%s" does not have "%s" index',
                space.name, index))
        end
    end
end

-- create space
function tube.create_space(space_name, opts)
    local space_opts         = {}
    local if_not_exists      = opts.if_not_exists or false
    space_opts.temporary     = opts.temporary or false
    space_opts.engine        = opts.engine or 'memtx'
    space_opts.format = {
        {name = 'task_id', type = "number"},
        {name = 'status', type = str_type()},
        {name = 'data', type = '*'},
        {name = 'order_id', type = str_type()},
        {name = 'profile_id', type = str_type()},
        {name = 'order_type', type = str_type()},
        {name = 'i_priority', type = "number"},
        {name = 'o_priority', type = "number"},
        {name = 'timestamp', type = "number"},
    }

    -- id, status, utube, data
    local space = box.space[space_name]
    if if_not_exists and space then
        -- Validate the existing space.
        validate_space(box.space[space_name])
        return space
    end

    space = box.schema.create_space(space_name, space_opts)
    space:create_index('task_id', {
        type = 'tree',
        unique = true,
        parts = {{field = "task_id"}},
        if_not_exists = true
    })
    space:create_index('status', {
        type = 'tree',
        unique = false,
        parts = {{field = 'status'}, {field = 'task_id'}},
        if_not_exists = true
    })
    space:create_index('by_profile_id', {
        type = 'tree',
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'status'}, {field = "task_id"}},
        if_not_exists = true
    })
    space:create_index('by_order_id', {
        type = 'tree',
        unique = true,
        parts = {{field = 'order_id'}, {field = 'status'}, {field = "task_id"}},
        if_not_exists = true
    })
    space:create_index('by_profile_id_order_id', {
        type = 'tree',
        unique = true,
        parts = {{field = 'status'}, {field = 'profile_id'}, {field = "order_id"}},
        if_not_exists = true
    })

    space:create_index('by_inside_outside_priority_timestamp', {
        type = 'tree',
        unique = false,
        parts = {{field = 'status'}, {field = 'o_priority'}, {field = 'i_priority'}, {field = "timestamp"}},
        if_not_exists = true
    })



    return space
end

-- start tube on space
function tube.new(space, on_task_change)
    validate_space(space)

    on_task_change = on_task_change or (function() end)
    local self = setmetatable({
        space          = space,
        on_task_change = on_task_change,
    }, { __index = method })
    return self
end

-- normalize task: cleanup all internal fields
function method.normalize_task(self, task)
    return task -- and task:transform(3, 1)
end


-- put task in space
function method.put(self, data, opts)    
    if opts.i_priority == nil then
        error(string.format('opts.i_priority is nil for space=%s',
        self.space.name))
    elseif opts.o_priority == nil then
        error(string.format('opts.o_priority is nil for space=%s',
        self.space.name))
    elseif opts.timestamp == nil then
        error(string.format('opts.timestamp is nil for space=%s',
        self.space.name))
    end

    local max = self.space.index.task_id:max()
    local id = max and max[1] + 1 or 0
    local task = self.space:insert{
        id,
        state.READY,
        data,
        tostring(opts.order_id),
        tostring(opts.profile_id),
        tostring(opts.order_type),
        opts.i_priority,
        opts.o_priority,
        opts.timestamp,
    }
    self.on_task_change(task, 'put')
    return task
end

-- take task
function method.take(self)
    local task = self.space.index.by_inside_outside_priority_timestamp:min{state.READY}
    if task ~= nil and task[2] == state.READY then
        task = self.space:update(task[1], { { '=', 2, state.TAKEN } })
        self.on_task_change(task, 'take')
        return task
    end
end

-- touch task
function method.touch(self, id, ttr)
    error('pqueue queue does not support touch')
end

-- delete task
function method.delete(self, id)
    local task = self.space:get(id)
    self.space:delete(id)
    if task ~= nil then
        task = task:transform(2, 1, state.DONE)
    end
    return task
end

function method.delete_by_profile_id(self, profile_id)
    for s, task in self.space.index.by_profile_id:pairs({tostring(profile_id), state.READY}, { iterator = 'EQ' }) do
        if task[2] == state.READY then
            self:delete(task[1])
        end
    end
end

function method.find_task(self, profile_id, order_id)
    return self.space.index.by_profile_id_order_id:get({state.READY, tostring(profile_id), tostring(order_id)})
end


-- release task
function method.release(self, id, opts)
    local task = self.space:update(id, {{ '=', 2, state.READY }})
    if task ~= nil then
        self.on_task_change(task, 'release')
    end
    return task
end

-- peek task
function method.peek(self, id)
    return self.space:get{id}
end

-- get iterator to tasks in a certain state
function method.tasks_by_state(self, task_state)
    return self.space.index.status:pairs(task_state)
end

function method.truncate(self)
    self.space:truncate()
end


-- not implemented
function method.bury(self, id)
    error('pqueue queue does not support bury')
end

-- not implemented
function method.kick(self, count)
    error('pqueue queue does not support kick')
end

-- This driver has no background activity.
-- Implement dummy methods for the API requirement.
function method.start()
    return
end

function method.stop()
    return
end

return tube
