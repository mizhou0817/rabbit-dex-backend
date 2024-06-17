-- Engine queue based on modified tqueue - refactor of official queue package

local checks = require('checks')

local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local tqueue = require('app.tqueue')

local err = errors.new_class("EQUEUE")

--[[
    engine queue
    implements queue with priorities based on multi table structure from high to low:

    (profile_id)
    (profile_id, order_type)
    (order_type)
--]] 

local equeue = {}

equeue.lowest_priority = tonumber(9e10)
equeue.highest_priority = -1

function equeue.find_task(qname, profile_id, order_id)
    checks('string', 'string', 'string')
    
    local status, res = pcall(
       function()
          return tqueue.tube[qname]:find_task(profile_id, order_id)
       end
    )
    if status == false then
       return {res=nil, error=res}
    end

    return {res=res, error=nil}
end

-- delete all READY tasks by utube name
function equeue.delete_by_profile_id(qname, profile_id)
    checks('string', 'string')
    
    local status, res = pcall(
       function()
          return tqueue.tube[qname]:delete_by_profile_id(profile_id)
       end
    )
    if status == false then
       return {error=res}
    end

    return {error=nil}
end


function equeue.stats(qname)
    checks('string')
 
    if tqueue.tube[qname] == nil then
       return {res=nil, error=err:new("NOT_FOUND")}
    end
 
    local stats = tqueue.statistics(qname)
    return {res=stats, error=nil}
end
 

function equeue.delete(qname, task_id)
    checks('string', 'number')
    local status, res = pcall(
       function()
          return tqueue.tube[qname]:delete(task_id)
       end
    )
    if status == false then
       return {error=res}
    end

    return {error=nil}
end
 

function equeue.ack(qname, task_id)
    checks('string', 'number')
    local status, res = pcall(
       function()
          return tqueue.tube[qname]:ack(task_id)
       end
    )
    if status == false then
        return {error=res}
    end

    return {error=nil}
end
  

function equeue.take(qname, timeout)
    checks('string', '?number')
    local status, res = pcall(
       function()
          return tqueue.tube[qname]:take(timeout)
       end
    )
    if status == false then
       return {res=nil, error=res}
    end
    
    return {res=res, error=nil}
end

-- return i_priority, o_priority
function equeue.decide_priority(profile_id, order_type)
    checks('number', 'string')

    -- map of priorities
    -- from high to low
    local priorities = {
        {
            space=box.space.priority_by_profile_id,
            key={profile_id}
        },
        {
            space=box.space.priority_by_profile_id_order_type,
            key={profile_id, order_type}
        },
        {
            space=box.space.priority_by_order_type,
            key={order_type}
        },
    }

    for _, p in ipairs(priorities) do
        local i_priority, o_priority = equeue.impl_decide_priority(p.space, p.key)
        if i_priority ~= equeue.lowest_priority then
            return i_priority, o_priority
        end
    end


    -- no priority just add to the end
    return equeue.lowest_priority, equeue.lowest_priority
end

function equeue.impl_decide_priority(space, key)
    local p = space:get(key)
    if p == nil then
        return equeue.lowest_priority, equeue.lowest_priority
    end

    if p.unlimited == true then
        return p.i_priority, p.o_priority
    end

    local new_current = p.current - 1
    if new_current + p.wait == 0 then
        new_current = p.wait
    end

    space:update(key, {{'=', 'current', new_current}})

    if p.current > 0 then
        return p.i_priority, p.o_priority
    end

    return equeue.lowest_priority, equeue.lowest_priority
end

function equeue.put(qname, data, profile_id, order_id, order_type)
    checks('string', 'table', 'number', 'string', 'string')

    local i_priority, o_priority = equeue.decide_priority(profile_id, order_type)
    local timestamp = time.now()

    return equeue._put_with_priority(qname, data, tostring(profile_id), order_id, order_type, i_priority, o_priority, timestamp)
end

function equeue.put_after(qname, data, profile_id, order_id, order_type, task)
    checks('string', 'table', 'number', 'string', 'string', '?tuple')

    if not task then
        return equeue.put(qname, data, profile_id, order_id, order_type)
    end

    local i_priority = task.i_priority + 1
    local o_priority = task.o_priority

    local timestamp = time.now()

    return equeue._put_with_priority(qname, data, tostring(profile_id), order_id, order_type, i_priority, o_priority, timestamp)
end


function equeue.put_with_highest_priority(qname, data, profile_id, order_id, order_type)
    checks('string', 'table', 'string', 'string', 'string')

    local timestamp = time.now()

    return equeue._put_with_priority(qname, data, profile_id, order_id, order_type, equeue.highest_priority, equeue.highest_priority, timestamp)
end

function equeue.put_with_lowest_priority(qname, data, profile_id, order_id, order_type)
    checks('string', 'table', 'string', 'string', 'string')

    local timestamp = time.now()

    return equeue._put_with_priority(qname, data, profile_id, order_id, order_type, equeue.lowest_priority, equeue.lowest_priority, timestamp)
end


function equeue._put_with_priority(qname, data, profile_id, order_id, order_type, i_priority, o_priority, timestamp)
    checks('string', 'table', 'string', 'string', 'string', 'number', 'number', 'number')

    local status, res = pcall(
       function()
          return tqueue.tube[qname]:put(data, {
            order_id = order_id,
            profile_id = profile_id,
            order_type = order_type,
            i_priority = i_priority,
            o_priority = o_priority,
            timestamp = timestamp,
        })
       end
    )
    if status == false then
       return {res=nil, error=res}
    end

    return {res=res, error=nil}
end


function equeue.get_or_create(qname)
    checks('string')
    local status, res = pcall(
       function()
          return tqueue.create_tube(qname, 'pqueue', {if_not_exists = true })
       end
    )
    if status == false then
        return {error=res}
    end

    return {error=nil}
end


function equeue.which_qname(market_id, queue_type)
    checks("string", "string")

    local queue_params = {}

    if queue_type == config.sys.QUEUE_TYPE.MARKET then
        queue_params = config.sys.ID_TO_MARKET_QUEUE[market_id]
    elseif queue_type == config.sys.QUEUE_TYPE.LIQUIDATION then
        queue_params = config.sys.ID_TO_LIQUIDATION_QUEUE[market_id]
    else
        return {res=nil, error="unknow queue_type"}
    end

    if queue_params == nil then
        return {res=nil, error="queue not found"}    
    end

    local e = equeue.get_or_create(queue_params.name)
    if e["error"] ~= nil then
        return {res=nil, error=e["error"]}
    end

    return {res=queue_params.name, error=nil}
end

function equeue.white_list(profile_id)
    box.space.white_list:replace({profile_id})

    return box.space.white_list:select()
end

function equeue.check_queue_limit(qname)
    checks("string")

    if tqueue.tube[qname] == nil then
        return {error=nil}
    end
  
    local stats = tqueue.statistics(qname)
    local total_tasks = 0
    if stats ~= nil then
        if stats["tasks"] ~= nil then
            total_tasks = stats["tasks"]["total"]
        end
    end

    if total_tasks == nil or total_tasks == 0 then
        return {error=nil}
    end

    if total_tasks >= config.sys.QUEUE_LIMITS.TOTAL_SIZE then
        return {error="QUEUE_RATE_LIMIT"}
    end

    return {error=nil}
end


function equeue.check_limit(profile_id)
    local now = time.now()

    local white_listted = box.space.white_list:get(profile_id)
    if white_listted ~= nil then
        return {error=nil}    
    end

    local stats = box.space.equeue_stats:get(profile_id)
    if stats == nil then
        return {error=nil}
    end

    if stats.count >= config.sys.QUEUE_LIMITS.LIMIT then
        if now - stats.last_update <= config.sys.QUEUE_LIMITS.PERIOD then
            return {error="RATE_LIMIT"}
        end
    end

    return {error=nil}
end

function equeue.inc_count(profile_id)
    local stats = {
        profile_id,
        time.now(),
        1
    }

    local exist = box.space.equeue_stats:get(profile_id)
    if exist ~= nil then
        stats = exist:totable()
        stats[3] = stats[3] + 1
        
        if stats[3] > config.sys.QUEUE_LIMITS.LIMIT then
            stats[2] = time.now()
            stats[3] = 1
        end
    end

    box.space.equeue_stats:replace(stats)

    return {error=nil}
end

function equeue.add_priority_by_profile_id(profile_id, current, wait, unlimited, i_priority, o_priority)
    checks('number', 'number', 'number', 'boolean', 'number', 'number')

    return box.space.priority_by_profile_id:insert{
        profile_id,
        current,
        wait,
        unlimited,
        i_priority,
        o_priority
    }
end

function equeue.add_priority_by_profile_id_order_type(profile_id, order_type, current, wait, unlimited, i_priority, o_priority)
    checks('number', 'string', 'number', 'number', 'boolean', 'number', 'number')

    return box.space.priority_by_profile_id_order_type:insert{
        profile_id,
        order_type,
        current,
        wait,
        unlimited,
        i_priority,
        o_priority
    }
end

function equeue.add_priority_by_order_type(order_type, current, wait, unlimited, i_priority, o_priority)
    checks('string', 'number', 'number', 'boolean', 'number', 'number')

    return box.space.priority_by_order_type:insert{
        order_type,
        current,
        wait,
        unlimited,
        i_priority,
        o_priority
    }
end


function equeue.init_spaces()
    local equeue_stats = box.schema.space.create('equeue_stats', {if_not_exists = true})
    equeue_stats:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'last_update', type = 'number'},
        {name = 'count', type = 'number'},
    })

    equeue_stats:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })

    local white_list = box.schema.space.create('white_list', {if_not_exists = true})
    white_list:format({
        {name = 'profile_id', type = 'unsigned'}
    })
    white_list:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })

    

    -- CREATE tables for detecting priorities
    local priority_by_profile_id = box.schema.space.create('priority_by_profile_id', {if_not_exists = true})
    priority_by_profile_id:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'current', type = 'number'},
        {name = 'wait', type = 'number'},
        {name = 'unlimited', type = 'boolean'},
        {name = 'i_priority', type = 'number'},
        {name = 'o_priority', type = 'number'},
    })

    priority_by_profile_id:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })


    local priority_by_profile_id_order_type = box.schema.space.create('priority_by_profile_id_order_type', {if_not_exists = true})
    priority_by_profile_id_order_type:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'order_type', type = 'string'},
        {name = 'current', type = 'number'},
        {name = 'wait', type = 'number'},
        {name = 'unlimited', type = 'boolean'},
        {name = 'i_priority', type = 'number'},
        {name = 'o_priority', type = 'number'},
    })

    priority_by_profile_id_order_type:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'order_type'}},
        if_not_exists = true })

    local priority_by_order_type = box.schema.space.create('priority_by_order_type', {if_not_exists = true})
    priority_by_order_type:format({
        {name = 'order_type', type = 'string'},
        {name = 'current', type = 'number'},
        {name = 'wait', type = 'number'},
        {name = 'unlimited', type = 'boolean'},
        {name = 'i_priority', type = 'number'},
        {name = 'o_priority', type = 'number'},
    })

    priority_by_order_type:create_index('primary', {
        unique = true,
        parts = {{field = 'order_type'}},
        if_not_exists = true })
end

return equeue

