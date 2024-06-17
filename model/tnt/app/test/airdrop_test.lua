local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local airdrop = require('app.profile.airdrop')
local profile = require('app.profile')
local config = require('app.config')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

local z = decimal.new(0)
require('app.config.constants')

local g = t.group('airdrop')

local work_dir = fio.tempdir()

local VOLUME = 100
local TM = 100
local LFT = 11

local mock_rpc = {call={}}
function mock_rpc.callro_engine(market_id, fn, ops)
    return {res = {decimal.new(VOLUME), LFT}, error = nil}
end

local mock_time = {}
function mock_time.now()
    return TM
end

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }

    airdrop.test_set_rpc(mock_rpc)
    airdrop.test_set_time(mock_time)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    archiver.init_sequencer("profile")
    profile.init_spaces({})

    mock_rpc.call = {}
end)

g.after_each(function(cg)
    box.space.airdrop:drop()
    box.space.profile_airdrop:drop()
    box.space.airdrop_claim_ops:drop()
    box.space.profile:drop()
    box.sequence.airdrop_claim_ops_id_sequence:drop()
    box.sequence.PID:drop()
end)

g.test_creation_flow = function(cg)

    -- CREATE 1 profile
    profile.profile.create("trader", "active", "0xqwe", DEFAULT_EXCHANGE_ID)

    local p = box.space.profile:get(1)
    t.assert_is_not(p, nil)

    -- CREATE AIRDROP
    local res = airdrop.create_airdrop("airdrop", 1, 10)
    t.assert_is(res["error"], nil)
    
    -- can't create twice
    res = airdrop.create_airdrop("airdrop", 1, 10)
    t.assert_is(res["error"], "AIRDROP_EXIST")

    -- wrong timestamp 
    res = airdrop.create_airdrop("airdrop1", 1, 1)
    t.assert_is(res["error"], "TIMESTAMP_END_LE_START")

    -- WRONG airdrop
    res = airdrop.set_profile_total(p.id, "airdropXXX", decimal.new(1000), decimal.new(1))
    t.assert_is(res["error"], "AIRDROP_NOT_EXIST")

    -- NO_PROFILE
    res = airdrop.set_profile_total(100000000, "airdrop", decimal.new(1000), decimal.new(1))
    t.assert_is(res["error"], "PROFILE_NOT_EXIST")

    -- WRONG numbers
    res = airdrop.set_profile_total(p.id, "airdrop", decimal.new(0), decimal.new(1))
    t.assert_is(res["error"], "CLAIMABLE_MORE_THAN_TOTAL")
    
    -- INIT PROFILE
    res = airdrop.set_profile_total(p.id, "airdrop", decimal.new(1000), decimal.new(1))
    t.assert_is(res["error"], nil)

    -- CAN'T init twice
    res = airdrop.set_profile_total(p.id, "airdrop", decimal.new(1000), decimal.new(1))
    t.assert_is(res["error"], "ALREADY_INIT")

end


g.test_airdrop_flow = function(cg)

    -- CREATE 1 profile
    profile.profile.create("trader", "active", "0x1", DEFAULT_EXCHANGE_ID)
    profile.profile.create("trader", "active", "0x2", DEFAULT_EXCHANGE_ID)

    local p1 = box.space.profile:get(1)
    t.assert_is_not(p1, nil)

    local p2 = box.space.profile:get(2)
    t.assert_is_not(p2, nil)

    -- CREATE AIRDROP-1
    local res = airdrop.create_airdrop("airdrop-1", 0, 200)
    t.assert_is(res["error"], nil)

    res = airdrop.create_airdrop("airdrop-2", 210, 400)
    t.assert_is(res["error"], nil)
        
    -- INIT PROFILE 1
    local p1_total = decimal.new(1000)
    local p1_claimable = decimal.new(10)
    res = airdrop.set_profile_total(p1.id, "airdrop-1", p1_total, p1_claimable)
    t.assert_is(res["error"], nil)
    res = airdrop.set_profile_total(p1.id, "airdrop-2", p1_total, p1_claimable)
    t.assert_is(res["error"], nil)

    -- INIT PROFILE 2
    local p2_total = decimal.new(1000)
    local p2_claimable = decimal.new(10)
    res = airdrop.set_profile_total(p2.id, "airdrop-1", p2_total, p2_claimable)
    t.assert_is(res["error"], nil)

    --[[
        Getters
    --]]
    res = airdrop.get_profile_airdrops(p1.id)
    local pas = res["res"]
    local pa1 = pas[1]
    local pa2 = pas[2]
    t.assert_is(#pas, 2)
    t.assert_is(pa1.profile_id, p1.id)
    t.assert_is(pa2.profile_id, p1.id)

    t.assert_is(pa1.total_rewards, p1_total)
    t.assert_is(pa1.claimable, p1_claimable)
    t.assert_is(pa1.claimed, ZERO)


    res = airdrop.get_profile_airdrops(p2.id)
    pas = res["res"]
    pa1 = pas[1]
    t.assert_is(#res["res"], 1)
    t.assert_is(pa1.profile_id, p2.id)

    res = airdrop.pending_claim(p1.id)
    t.assert_is(res["res"], nil)
    t.assert_is(res["error"], nil)

    res = airdrop.finish_claim(p1.id)
    t.assert_is(res["error"], "NO_PENDING_CLAIMS")


    -- CLAIM initial
    TM = 201
    res = airdrop.claim_all(p1.id, "airdrop-1")
    t.assert_is(res["error"], nil)
    local ops = res["res"]
    local ops_id = ops[1]
    t.assert_is(ops.airdrop_title, "airdrop-1")
    t.assert_is(ops.profile_id, p1.id)
    t.assert_is(ops.status, config.params.AIRDROP_CLAIM_STATUS.CLAIMING)
    t.assert_is(ops.amount, p1_claimable)


    -- CAN'T claim twice
    res = airdrop.claim_all(p1.id, "airdrop-1")
    t.assert_is(res["error"], "NOTHING_TO_CLAIM")

    -- CAN'T claim for other while one is pending
    TM = 401
    res = airdrop.claim_all(p1.id, "airdrop-2")
    t.assert_is(res["error"], "PENDING_CLAIM_EXIST")

    -- NOW we see it's pending
    res = airdrop.pending_claim(p1.id)
    ops = res["res"]
    t.assert_is(ops.id,ops_id)
    t.assert_is(ops.profile_id, p1.id)
    t.assert_is(res["error"], nil)



    -- FINISH pending
    res = airdrop.finish_claim(p1.id)
    ops = res["res"]
    t.assert_is(ops.id,ops_id)
    t.assert_is(ops.profile_id, p1.id)
    t.assert_is(ops.status, config.params.AIRDROP_CLAIM_STATUS.CLAIMED)
    t.assert_is(res["error"], nil)

    res = airdrop.finish_claim(p1.id)
    t.assert_is(res["error"], "NO_PENDING_CLAIMS")


    --[[
        The full sequence for profile1 
        during airdrop it has volume = 100
        then 300 for 2 periods
    --]]    
    TM = 100
    VOLUME = 125 -- VILL init volume for period to 100
    res = airdrop.update_profile_claimable(p1.id, "airdrop-1")
    t.assert_is(res["error"], "AIRDROP_IN_PROGRESS")

    TM = 201
    LFT = 210
    res = airdrop.update_profile_claimable(p1.id, "airdrop-1")
    t.assert_is(res["error"], nil)
    local pa = res["res"]
    t.assert_is(pa.claimable, p1_total - p1_claimable)
    
    -- Wait untill total will be accumulated
    VOLUME = 250
    TM = 300
    for i=1,100 do
        res = airdrop.update_profile_claimable(p1.id, "airdrop-1")
       -- t.assert_is(res["error"], nil)
        --pa = res["res"]
    end
    t.assert_is(pa.claimable, p1_total - p1_claimable)    

    res = airdrop.claim_all(p1.id, "airdrop-1")
    t.assert_is(res["error"], nil)
    local ops = res["res"]
    local ops_id = ops[1]
    t.assert_is(ops.airdrop_title, "airdrop-1")
    t.assert_is(ops.profile_id, p1.id)
    t.assert_is(ops.status, config.params.AIRDROP_CLAIM_STATUS.CLAIMING)
    t.assert_is(ops.amount, p1_total - p1_claimable)

    local _pa = box.space.profile_airdrop:get{p1.id, "airdrop-1"}
    t.assert_is(_pa.status, config.params.PROFILE_AIRDROP_STATUS.ACTIVE)
    t.assert_is(_pa.claimable, ZERO)
    t.assert_is(_pa.claimed, p1_total)
    
    res = airdrop.finish_claim(p1.id)
    ops = res["res"]
    t.assert_is(ops.id,ops_id)
    t.assert_is(ops.profile_id, p1.id)
    t.assert_is(ops.status, config.params.AIRDROP_CLAIM_STATUS.CLAIMED)
    t.assert_is(res["error"], nil)

    _pa = box.space.profile_airdrop:get{p1.id, "airdrop-1"}
    t.assert_is(_pa.status, config.params.PROFILE_AIRDROP_STATUS.FINISHED)
    t.assert_is(_pa.claimable, ZERO)
    t.assert_is(_pa.claimed, p1_total)

    res = airdrop.update_profile_claimable(p1.id, "airdrop-1")
    t.assert_is(res["error"], "FINISHED")


    -- DELETE all
    local c = box.space.airdrop:count()
    t.assert_is_not(c, 0)

    airdrop.delete_all_airdrops("airdrop-1", 1, 10, p1.id, decimal.new(100), decimal.new(10))
    
    c = box.space.airdrop:count()
    t.assert_is(c, 1)

end
