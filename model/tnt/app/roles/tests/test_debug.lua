local json = require("json")
local log = require('log')

local config = require('app.config')
local time = require('app.lib.time')
local rpc = require('app.rpc')

local D = {}

function D.push_orderbook(market_id,
    sequence,
    bid,
    bid_size,
    ask,
    ask_size
)
    local market_id = config.markets.test_market[1]
    local channel = "orderbook:" .. tostring(market_id)
    local update = {
        market_id=market_id,
        timestamp=time.now(),
        sequence=sequence,
        bids={},
        asks={}
    }

    log.info("publish to channel %s", channel)

    table.insert(update.bids, {bid, bid_size})  
    table.insert(update.asks, {ask, ask_size})

    local json_update = json.encode({data=update})
    
    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)
    update = nil

    return {res = "success", error = nil}
end

function D.push_profile(profile_id)
    local channel = "account@" .. tostring(profile_id)
    local update = {
        id=profile_id,
        last_update=time.now()
    }

    log.info("publish to channel %s", channel)

    local json_update = json.encode({data=update})
    rpc.callrw_pubsub_publish(channel, json_update, 0, 0, 0)

    update = nil

    return {res = "success", error = nil}
end


local function stop()
    return true
end

local function init(opts) -- luacheck: no unused args
    if opts.is_master then

        box.schema.func.create('test_debug', {if_not_exists = true})
    end

    rawset(_G, 'test_debug', D)
    return true
end

return {
    role_name = 'test_debug',
    init = init,
    stop = stop,
    utils = {
        test_debug = D
    }
}

