local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')

-- TODO: rewrite equeue path, to not allow init abstract on require
-- local internal = require('app.api.internal')

require('app.config.constants')

local g = t.group('clear_coid_test')

local work_dir = fio.tempdir()

local global_res = {}

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    local used_client_order_id = box.schema.space.create('used_client_order_id', {temporary=false, if_not_exists = true})
    used_client_order_id:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'client_order_id', type = 'string'},
        {name = 'market_id', type = 'string'},      -- we use market_id for invalidate later
    })

    used_client_order_id:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'client_order_id'}},
        if_not_exists = true })

    used_client_order_id:create_index('profile_market', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true })
end)

g.after_each(function(cg)
    box.space.used_client_order_id:drop()
end)

local function clear_coid_table()
    for _, pt in box.space.used_client_order_id:pairs() do
        if pt.market_id ~= "" then
            local res = global_res
            if res["error"] == nil and res["res"] == pt.client_order_id then
                box.space.used_client_order_id:delete{pt.profile_id, pt.client_order_id}
            end
        end
    end
end


g.test_clear_coid = function(cg)
    -- INSERT a lot of used_client_order_id
    box.space.used_client_order_id:insert{1, "coid-1", "btc"}
    box.space.used_client_order_id:insert{1, "coid-2", "btc"}
    box.space.used_client_order_id:insert{1, "coid-3", "btc"}
    box.space.used_client_order_id:insert{1, "coid-4", "btc"}

    box.space.used_client_order_id:insert{2, "coid-1", "btc"}
    box.space.used_client_order_id:insert{2, "coid-2", "btc"}
    box.space.used_client_order_id:insert{2, "coid-3", "btc"}
    box.space.used_client_order_id:insert{2, "coid-4", "btc"}

    local initial_count = box.space.used_client_order_id:count()

    -- nothing to clear
    global_res = {res = nil, error = "OPEN_ONLY"}
    clear_coid_table()
    local c = box.space.used_client_order_id:count()
    t.assert_is(c, initial_count)

    -- 1 found
    global_res = {res = "coid-1", error = nil}
    for i = 1,2 do
        clear_coid_table()
    end
    
    c = box.space.used_client_order_id:count()
    t.assert_is(c, initial_count - 2)

end
