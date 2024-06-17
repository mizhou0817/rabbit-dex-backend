local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local ddl = require('app.ddl')
local config = require('app.config')
local revert = require('app.revert')
local profile = require('app.profile')
local p = require('app.profile.profile')
local engine = require('app.engine')
local matching = require('app.engine.engine')
local risk = require('app.engine.risk')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')
require('app.errcodes')

local g = t.group('revert')

local mock_rpc = {call=nil}
function mock_rpc.callrw_engine(market_id, fn, params)
    t.assert_is_not(revert.bind(params[1]), nil)
    
    local res = matching.handle_revert(params[1])
    t.assert_is(res["error"], nil)

    return res
end

function mock_rpc.callrw_profile(fn, params)
    return {
        res = {},
        error = nil
    }
end

function mock_post_match(market_id, profile_data, profile_id, position_before)
    return nil
end

local cur_dir = fio.cwd()
local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    revert.init_spaces()
    archiver.init_sequencer("test")
  
    matching.init("BTC-USD", ONE, ONE)
    engine.init_spaces({
        id = 'BTC-USD',
        status = 'active',
        min_initial_margin = ONE,
        forced_margin = ONE,
        liquidation_margin = ONE,
        min_tick = ONE,
        min_order = ONE,
    })

    risk._test_set_post_match(mock_post_match)
    revert.test_set_rpc(mock_rpc)
    matching._test_set_rpc(mock_rpc)

    profile.init_spaces()

    for i = 0, 50, 1 do
        local w = "0x" .. tostring(i)
        p.create(config.params.PROFILE_TYPE.TRADER, config.params.PROFILE_STATUS.ACTIVE, w, DEFAULT_EXCHANGE_ID)
    end
end)

g.after_each(function(cg)
    revert.drop_spaces()
end)

g.test_upload_csv = function(cg)
    local valid_csv = fio.pathjoin(cur_dir, "app/test/data/valid.csv")
    local invalid_header = fio.pathjoin(cur_dir, "app/test/data/invalid_header.csv")
    local invalid_market = fio.pathjoin(cur_dir, "app/test/data/invalid_market.csv")
    local invalid_taker = fio.pathjoin(cur_dir, "app/test/data/invalid_taker.csv")

    local res = revert.upload_csv(valid_csv)
    t.assert_is(res, nil)
    t.assert_is(box.space.revert_list.index.status:count(config.params.REVERT_STATUS.TX_UPLOADED), 4)

    -- TODO: fix later, as never used in prod
    --[[
    res = revert.execute_revert_list()
    for i, tup in box.space.revert_list:pairs() do
        t.assert_is_not(tup.result, nil)
    end


    t.assert_is(box.space.revert_list.index.status:count(config.params.REVERT_STATUS.TX_UPLOADED), 0)
    t.assert_is(box.space.revert_list.index.status:count(config.params.REVERT_STATUS.TX_PROCESSED), 4)


    res = revert.upload_csv(invalid_header)
    t.assert_is(res, ERR_REVERT_INVALID_HEADER)

    res = revert.upload_csv(invalid_market)
    t.assert_is(res, ERR_REVERT_CANT_BIND)

    res = revert.upload_csv(invalid_taker)
    t.assert_is(res, ERR_REVERT_CANT_BIND)
    --]]
end

