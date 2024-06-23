local fio = require('fio')
local t = require('luatest')

local a = require('app.archiver')
local profile = require('app.profile')
local notif = require('app.profile.notif')
local ddl = require('app.ddl')
local m = require('migrations.common.eid_balance_migrations')

require('app.config.constants')

local g = t.group('profile.notif')

local work_dir = fio.tempdir()

local mock_rpc = {call=nil}
function pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    mock_rpc.call = {channel, json_data}
end

t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
    notif._test_set_rpc(mock_rpc)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    t.assert_is_not(a.init_sequencer('shard'), nil)
    profile.init_spaces()

    local leverage = {}
    leverage["BTC-USD"] = ONE

    a.replace(box.space.profile_cache, {
        1, "trader", "active", "0xabcdef", 123456,
        ONE, ONE,
        ONE, ONE, ONE, ONE, ONE, ONE, ONE, ONE, ONE,
        leverage,
        234567,
    })

    mock_rpc.call = nil
    cg.params = {}
end)

g.after_each(function(cg)
    box.space.profile_cache:drop()
    box.sequence.shard_archive_id_sequencer:drop()
end)

g.test_all_profiles = function(cg)
    local expected = {"account@1", '{"data":{"account_leverage":"1","shard_id":"shard","id":1,"profile_type":"trader","total_position_margin":"1","leverage":{"BTC-USD":"1"},"cum_unrealized_pnl":"1","total_notional":"1","status":"active","cum_trading_volume":"1","account_equity":"1","withdrawable_balance":"1","last_liq_check":234567,"balance":"1","wallet":"0xabcdef","last_update":123456,"health":"1","account_margin":"1","archive_id":3,"total_order_margin":"1"}}'}
    notif.notify_profiles({})
    t.assert_equals(mock_rpc.call, expected)
end

g.test_one_profile = function(cg)
    local expected = {"account@1", '{"data":{"account_leverage":"1","shard_id":"shard","id":1,"profile_type":"trader","total_position_margin":"1","leverage":{"BTC-USD":"1"},"cum_unrealized_pnl":"1","total_notional":"1","status":"active","cum_trading_volume":"1","account_equity":"1","withdrawable_balance":"1","last_liq_check":234567,"balance":"1","wallet":"0xabcdef","last_update":123456,"health":"1","account_margin":"1","archive_id":1,"total_order_margin":"1"}}'}
    local profiles = {1}
    notif.notify_profiles(profiles)
    t.assert_equals(mock_rpc.call, expected)
end
