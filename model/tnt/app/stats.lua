
local time = require('app.lib.time')

local stats = {}

function stats.init_spaces()
    local d_stats = box.schema.space.create('d_stats', {temporary = true, if_not_exists = true})
    d_stats:format({
        {name = 'timestamp', type = 'number'},
        {name = 'metric_name', type = 'string'},
        {name = 'count', type = 'number'},
        {name = 'instance_name', type = 'string'}
    })
    d_stats:create_index('primary', {
        unique = true,
        parts = {{field = 'timestamp'}, {field = 'metric_name'}},
        if_not_exists = true })    
end

function stats.show_all(total)
    local res = {}

    local i = 0
    for _, m in box.space.d_stats:pairs(nil, {iterator="REQ"}) do
        table.insert(res, m)
        i = i + 1
        if i > total then
            break
        end
    end

    return {res = res, error = nil}
end


return stats