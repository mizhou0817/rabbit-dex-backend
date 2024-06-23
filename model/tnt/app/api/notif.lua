local checks = require('checks')
local json = require('json')
local rpc = require('app.rpc')
local log = require('log')
local time = require('app.lib.time')

local notif = {}

function notif.notify_order(profile_id, task_order, status)
    checks("number", "table|cdata", "string")

    task_order["status"] = status
    task_order["id"] = task_order["order_id"]
    task_order["order_id"] = nil
    task_order["timestamp"] = time.now()

    local channel = "account@" .. tostring(profile_id)
    local update = {
        id = profile_id,
        orders = {task_order}
    }
    local json_update = json.encode(update)

    pubsub_publish(channel, json_update)
    update = nil

    return nil

end

function notif._test_set_rpc(override_rpc)
    rpc = override_rpc
end


return notif
