local checks = require('checks')
local decimal = require('decimal')
local log = require('log')

local archiver = require('app.archiver')
local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local rpc = require('app.rpc')
local util = require('app.util')

require("app.config.constants")

local AirdropError = errors.new_class("AIRDROP")

local function init_spaces()

    local airdrop, err = archiver.create('airdrop', {if_not_exists = true}, {
        {name = 'title', type = 'string'},
        {name = 'start_timestamp', type = 'number'},
        {name = 'end_timestamp', type = 'number'},
    }, {
    {
        name = "primary",
        unique = true,
        parts = {{field = 'title'}},
        if_not_exists = true,
    },
    {
        name = "timestamp",
        unique = false,
        parts = {{field = 'start_timestamp'}, {field = 'end_timestamp'}},
        if_not_exists = true,
    }
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        error(err)
    end

    local profile_airdrop, err = archiver.create('profile_airdrop', {if_not_exists = true}, {
        {name = 'profile_id', type = 'unsigned'},
        {name = 'airdrop_title', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'total_volume_for_airdrop', type = 'decimal'},
        {name = 'total_volume_after_airdrop', type = 'decimal'},
        {name = 'total_rewards', type = 'decimal'},
        {name = 'claimable', type = 'decimal'},
        {name = 'claimed', type = 'decimal'},
        {name = 'last_fill_timestamp', type = '*'},
        {name = 'initial_rewards', type = 'decimal'},
    }, {
    {
        name = "primary",
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'airdrop_title'}},
        if_not_exists = true,
    },
    {
        name = "status",
        unique = false,
        parts = {{field = 'status'}},
        if_not_exists = true,
    }
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        error(err)
    end
    

    box.schema.sequence.create('airdrop_claim_ops_id_sequence',{start=1, min=1, if_not_exists = true})
    local airdrop_claim_ops, err = archiver.create('airdrop_claim_ops', {if_not_exists = true}, {
        {name = 'id', type = 'unsigned'},
        {name = 'airdrop_title', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'status', type = 'string'},
        {name = 'amount', type = 'decimal'},
        {name = 'timestamp', type = 'number'},
    }, {
    {
        name = "primary",
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    },
    {
        name = "status_profile",
        unique = false,
        parts = {{field = 'status'}, {field = 'profile_id'}},
        if_not_exists = true,
    }
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        error(err)
    end
end

local function create_airdrop(title, start_timestamp, end_timestamp)
    start_timestamp = tonumber(start_timestamp)
    end_timestamp = tonumber(end_timestamp)
    checks('string', 'number', 'number')

    if end_timestamp <= start_timestamp then
        return {res = nil, error = "TIMESTAMP_END_LE_START"}
    end

    local exist = box.space.airdrop:get(title)
    if exist ~= nil then
        return {res = exist, error = "AIRDROP_EXIST"}
    end

    local res, err = archiver.insert(box.space.airdrop, {
        title,
        start_timestamp,
        end_timestamp,
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

local function set_profile_total(profile_id, airdrop_title, total_rewards, claimable)
    checks('number', 'string', 'decimal', 'decimal')


    local profile = box.space.profile:get(profile_id)
    if profile == nil then
        return {res = nil, error = "PROFILE_NOT_EXIST"}
    end

    local airdrop = box.space.airdrop:get(airdrop_title)
    if airdrop == nil then
        return {res = nil, error = "AIRDROP_NOT_EXIST"}
    end

    if claimable > total_rewards then
        return {res = nil, error = "CLAIMABLE_MORE_THAN_TOTAL"}
    end

    local exist = box.space.profile_airdrop:get{profile_id, airdrop_title}
    if exist ~= nil then
        return {res = nil, error = "ALREADY_INIT"}
    end
    
    -- need to be a map for msgpack decoding
    local last_fill_timestamp = {}
    last_fill_timestamp["BTC-USD"] = 0
    local res, err = archiver.insert(box.space.profile_airdrop, {
        profile_id,
        airdrop_title,
        config.params.PROFILE_AIRDROP_STATUS.INIT,
        ZERO,
        ZERO,
        total_rewards,
        claimable,
        ZERO,
        last_fill_timestamp,
        total_rewards - claimable
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

local function calc_profile_total_volume(market_ids, profile_id, start_timestamp, end_timestamp)
    local total = ZERO
    local last_fill_timestamp = 0
    for _, market_id in pairs(market_ids) do    
        local data, err = util.retry(2, function()
            local res = rpc.callro_engine(market_id, 'total_volume', {profile_id, start_timestamp, end_timestamp})
            if res['error'] ~= nil then
                error(res['error'])
            end
            return res['res']
        end)

        if err ~= nil then
            log.error(AirdropError:new(err))
            return nil, nil, tostring(err)
        end

        total = total + data[1]
        last_fill_timestamp = data[2]
    end    

    return total, last_fill_timestamp, nil
end

local function update_profile_claimable(profile_id, airdrop_title)
    checks('number', 'string')

    local airdrop = box.space.airdrop:get(airdrop_title)
    if airdrop == nil then
        return {res = nil, error = "NO_AIRDROP_WITH_TITLE"}
    end

    if time.now() <= airdrop.end_timestamp then
        return {res = nil, error = "AIRDROP_IN_PROGRESS"}
    end

    local profile_airdrop = box.space.profile_airdrop:get{profile_id, airdrop_title}
    if profile_airdrop == nil then
        return {res = nil, error = "NO_DATA_FOR_PROFILE_WITH_AIRDROP_TITLE"}
    end

    if profile_airdrop.status == config.params.PROFILE_AIRDROP_STATUS.FINISHED then
        return {res = nil, error = "FINISHED"}    
    end

    local total_volume_for_airdrop = profile_airdrop.total_volume_for_airdrop
    if profile_airdrop.status == config.params.PROFILE_AIRDROP_STATUS.INIT then
        -- need to calc total for period
        local market_ids = {}
        for _, market_id in pairs(config.params.MARKETS) do
            table.insert(market_ids, market_id)
        end   

        local total, lft, err = calc_profile_total_volume(market_ids, profile_id, airdrop.start_timestamp, airdrop.end_timestamp)
        if err ~= nil then
            log.error(AirdropError:new(err))
            return {res=nil, error = tostring(err)}
        end

        total_volume_for_airdrop = total

        local res, err = archiver.update(box.space.profile_airdrop, {profile_id, airdrop_title}, {
            {'=', "status", config.params.PROFILE_AIRDROP_STATUS.ACTIVE},
            {'=', "total_volume_for_airdrop", total_volume_for_airdrop}
        })

        if err ~= nil then
            log.error(AirdropError:new(err))
            return {res = nil, error = err}
        end
    end

    if total_volume_for_airdrop == ZERO then   
        local res, err = archiver.update(box.space.profile_airdrop, {profile_id, airdrop_title}, {
            {'=', "status", config.params.PROFILE_AIRDROP_STATUS.FINISHED}
        })
        if err ~= nil then
            log.error(AirdropError:new(err))
            return {res = nil, error = err}
        end
    
        return {res = nil, error = "FINISHED"}
    end

    --[[
        We calc the trading volume for the user since last check
        * from_timestamp = airdrop.end_timestamp to now (if we never calced)
        * from_timestamp = last_fill_timestamp (if we checked)
    --]]
    local current_total_volume = ZERO
    local last_fill_timestamp = profile_airdrop.last_fill_timestamp
    for _, market in pairs(config.markets) do
        local from_timestamp = airdrop.end_timestamp
        if last_fill_timestamp[market.id] ~= nil and last_fill_timestamp[market.id] >= airdrop.end_timestamp then
            from_timestamp = last_fill_timestamp[market.id]
        end
        local to_timestamp = time.now()

        if to_timestamp > from_timestamp then
            local total, lft, err = calc_profile_total_volume({market.id}, profile_id, from_timestamp, to_timestamp)
            if err ~= nil then
                log.error(AirdropError:new(err))
                return {res=nil, error = tostring(err)}
            end

            current_total_volume = current_total_volume + total
            last_fill_timestamp[market.id] = to_timestamp
        end 
    end

    if current_total_volume == 0 then
        return {res = nil, error = "NO_FILLS_FOR_PERIOD"}
    end

 
    -- update claimable, finish if total achieved
    local new_unlocked_amount =  profile_airdrop.initial_rewards * (current_total_volume / total_volume_for_airdrop)
    local diff = profile_airdrop.total_rewards - (profile_airdrop.claimable + profile_airdrop.claimed)
    if diff < 0 then
        error("DIFF_INTEGRITY_ERROR on update claimable")
        return {res = exist, error = "DIFF_INTEGRITY_ERROR"}
    end

    if new_unlocked_amount < diff then
        diff = new_unlocked_amount
    end
    
    local res, err = archiver.update(box.space.profile_airdrop, {profile_id, airdrop_title}, {
        {'=', "status", config.params.PROFILE_AIRDROP_STATUS.ACTIVE},
        {'+', "claimable", diff},
        {'+', "total_volume_after_airdrop", current_total_volume},
        {'=', "last_fill_timestamp", last_fill_timestamp}
    })
    if err ~= nil then
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end

    return {res = res, error = nil}
end

local function update_all_profile_airdrops(profile_id)
    checks('number')

    -- silently update
    for _, profile_airdrop in box.space.profile_airdrop:pairs({profile_id}, {iterator = "EQ"}) do 
        local res, err = update_profile_claimable(profile_id, profile_airdrop.airdrop_title)
        if err ~= nil then
            return {res = nil, error = err}
        end
    end

    return {res = nil, error = nil}
end

local function pending_claim(profile_id)
    checks('number')

    local ops = box.space.airdrop_claim_ops.index.status_profile:select({config.params.AIRDROP_CLAIM_STATUS.CLAIMING, profile_id}, {iterator = "EQ"})
    if ops ~= nil and #ops > 0 then
        return {res = ops[1], error = nil}
    end

    return {res = nil, error = nil}
end

-- By design only 1 pending claim per time possible
-- No concurrency issues here. 
-- If several clients request pending_claim at this moment they will receive the ops
local function finish_claim(profile_id)
    checks('number')

    local alreadyInTx = box.is_in_txn()
    if not alreadyInTx then
        box.begin()
    end

    local ops = box.space.airdrop_claim_ops.index.status_profile:select({config.params.AIRDROP_CLAIM_STATUS.CLAIMING, profile_id}, {iterator = "EQ"})
    if ops == nil or #ops == 0 then
        box.rollback()
        return {res = nil, error = "NO_PENDING_CLAIMS"}
    end

    local the_ops = ops[1]
    local ops_id = the_ops[1]

    local res, err = archiver.update(box.space.airdrop_claim_ops, ops_id, {
        {'=', "status", config.params.AIRDROP_CLAIM_STATUS.CLAIMED},
    })
    if err ~= nil then
        box.rollback()
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end

    -- Can we finalize the profile_airdrop?
    local profile_airdrop = box.space.profile_airdrop:get{profile_id, the_ops.airdrop_title}
    if profile_airdrop == nil then
        box.rollback()
        return {res = nil, error = "INTEGRITY_ERROR: NO_DATA_FOR_PROFILE_WITH_AIRDROP_TITLE"}
    end

    if profile_airdrop.total_rewards == profile_airdrop.claimed then
        local res, err = archiver.update(box.space.profile_airdrop, {profile_id, the_ops.airdrop_title}, {
            {'=', "status", config.params.PROFILE_AIRDROP_STATUS.FINISHED},
        })
        if err ~= nil then
            box.rollback()
            log.error(AirdropError:new(err))
            return {res = nil, error = err}
        end
    end

    if not alreadyInTx then
        box.commit()
    end

    ops = box.space.airdrop_claim_ops:get(ops_id)

    return {res = ops, error = nil}
end


local function claim_all(profile_id, airdrop_title)
    checks('number', 'string')
    local airdrop = box.space.airdrop:get(airdrop_title)
    if airdrop == nil then
        return {res = nil, error = "NO_AIRDROP_WITH_TITLE"}
    end

    if time.now() <= airdrop.end_timestamp then
        return {res = nil, error = "AIRDROP_IN_PROGRESS"}
    end

    local profile_airdrop = box.space.profile_airdrop:get{profile_id, airdrop_title}
    if profile_airdrop == nil then
        return {res = nil, error = "NO_DATA_FOR_PROFILE_WITH_AIRDROP_TITLE"}
    end

    if profile_airdrop.claimable == 0 then
        return {res = nil, error = "NOTHING_TO_CLAIM"}
    end

    local res = pending_claim(profile_id)
    if res["res"] ~= nil then
        return {res = nil, error = "PENDING_CLAIM_EXIST"}
    end

    local tm = time.now()

    local alreadyInTx = box.is_in_txn()
    if not alreadyInTx then
        box.begin()
    end

    local ops_id = box.sequence.airdrop_claim_ops_id_sequence:next()
    local res, err = archiver.insert(box.space.airdrop_claim_ops, {
        ops_id,
        airdrop_title,
        profile_id,
        config.params.AIRDROP_CLAIM_STATUS.CLAIMING,
        profile_airdrop.claimable,
        tm,
    })
    if err ~= nil then
        box.rollback()
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end
    local ops = res

    res, err = archiver.update(box.space.profile_airdrop, {profile_id, airdrop_title}, {
        {'=', "claimable", ZERO},
        {'+', 'claimed', profile_airdrop.claimable}
    })

    if err ~= nil then
        box.rollback()
        log.error(AirdropError:new(err))
        return {res = nil, error = err}
    end

    -- INTEGRITY_CHECK to never allow claim more than total
    local pa = box.space.profile_airdrop:get{profile_id, airdrop_title}
    local diff = pa.total_rewards - (pa.claimable + pa.claimed)
    if diff < 0 then
        box.rollback()
        log.error(AirdropError:new("CLAIM_ALL_MORE_THAN_TOTAL"))
        return {res = nil, error = "CLAIM_ALL_MORE_THAN_TOTAL"}
    end

    if not alreadyInTx then
        box.commit()
    end

    return {res = ops, error = nil}
end

local function get_profile_airdrops(profile_id)
    checks('number')

    local airdrops = {}

    for _, profile_airdrop in box.space.profile_airdrop:pairs({profile_id}, {iterator = "EQ"}) do 
        table.insert(airdrops, profile_airdrop)
    end

    return {res = airdrops, error = nil}
end

local function delete_all_airdrops(title, from, to, profile_id, total_rewards, claimable)
    box.space.airdrop:truncate()
    box.space.profile_airdrop:truncate()
    box.space.airdrop_claim_ops:truncate()

    create_airdrop(title, from, to)
    set_profile_total(profile_id, title, total_rewards, claimable)

    return {res = nil, error = nil}
end

local function test_create_claim_ops(profile_id, airdrop_title, claimable)

    local tm = time.now()
    local ops_id = box.sequence.airdrop_claim_ops_id_sequence:next()
    local res, err = archiver.insert(box.space.airdrop_claim_ops, {
        ops_id,
        airdrop_title,
        profile_id,
        config.params.AIRDROP_CLAIM_STATUS.CLAIMING,
        decimal.new(claimable),
        tm,
    })

    return {res = res, error = err}
end

local function test_get_claim_ops(ops_id)

    local res = box.space.airdrop_claim_ops:get(ops_id)

    return {res = res, error = nil}
end



--TODO: refactor as notif:new(rpc), notif:notify_xxx() and remove this func
local function test_set_rpc(override_rpc)
    rpc = override_rpc
end

local function test_set_time(override_time)
    time = override_time
end


return {
    init_spaces = init_spaces,
    create_airdrop = create_airdrop,
    set_profile_total = set_profile_total,
    update_profile_claimable = update_profile_claimable,
    claim_all = claim_all,
    get_profile_airdrops = get_profile_airdrops,
    update_all_profile_airdrops = update_all_profile_airdrops,
    pending_claim = pending_claim,
    finish_claim = finish_claim,
    calc_profile_total_volume = calc_profile_total_volume,
    test_set_rpc = test_set_rpc,
    test_set_time = test_set_time,
    delete_all_airdrops = delete_all_airdrops,
    test_create_claim_ops = test_create_claim_ops,
    test_get_claim_ops = test_get_claim_ops
}
