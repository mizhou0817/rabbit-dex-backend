local fiber = require('fiber')
local utils = require('metrics.utils')

local collectors_list = {}

local function update_fibers_metrics()
    local fibers_info = fiber.info({backtrace = false})

    for _, f in pairs(fibers_info) do
        collectors_list.fiber_cpu_time = utils.set_gauge('fiber1_cpu_time', 'Fiber cpu time',
            f.time, {name = f.name}, nil, {default = true})
        collectors_list.fiber_csw = utils.set_gauge('fiber1_csw', 'Fiber csw',
            f.csw, {name = f.name}, nil, {default = true})
        collectors_list.fiber_memalloc = utils.set_gauge('fiber1_memalloc', 'Fiber memalloc',
            f.memory.total, {name = f.name}, nil, {default = true})
        collectors_list.fiber_memused = utils.set_gauge('fiber1_memused', 'Fiber memused',
            f.memory.used, {name = f.name}, nil, {default = true})
    end
end

return {
    update = update_fibers_metrics,
    list = collectors_list,
}
