local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local migration = require('migrations.engine.2023031415000000_engine_market')
local ddl = require('app.ddl')

local z = decimal.new(0)
local num = decimal.new(111)

require('app.config.constants')
local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        listen = 4301,
        work_dir = work_dir,
    }
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

local g = t.group('market_migration')
g.before_each(function(cg)
    -- Create old market and init
    archiver.init_sequencer("BTC-USD")

    -- CONSTANTS that rare change and status
    local market, err = archiver.create('market', {if_not_exists = true}, {
        {name = 'id', type = 'string'},
        {name = 'status', type = 'string'},

        {name = 'min_initial_margin', type = 'decimal'},
        {name = 'forced_margin', type = 'decimal'},
        {name = 'liquidation_margin', type = 'decimal'},
        {name = 'min_tick', type = 'decimal'},
        {name = 'min_order', type = 'decimal'},

        {name = 'best_bid', type = 'decimal'},
        {name = 'best_ask', type = 'decimal'},
        {name = 'market_price', type = 'decimal'},
        {name = 'index_price', type = 'decimal'},
        {name = 'last_trade_price', type = 'decimal'},
        {name = 'fair_price', type = 'decimal'},
        {name = 'last_trade_price_24high', type = 'decimal'},
        {name = 'last_trade_price_24low', type = 'decimal'},
        {name = 'average_daily_volume', type = 'decimal'},
        {name = 'instant_funding_rate', type = 'decimal'},
        {name = 'instant_daily_volume', type = 'decimal'},
        {name = 'last_funding_rate_basis', type = 'decimal'},
        {name = 'last_trade_price_24h_change_premium', type = 'decimal'},
        {name = 'last_trade_price_24h_change_basis', type = 'decimal'},
        {name = 'average_daily_volume_change_premium', type = 'decimal'},
        {name = 'average_daily_volume_change_basis', type = 'decimal'},

        {name = 'last_update_time', type = 'number'},
        {name = 'last_update_sequence', type = 'number'},
        {name = 'average_daily_volume_q', type = 'decimal'},
        {name = 'last_funding_update_time', type = 'number'},
    }, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    t.assert_is(err, nil)

    local res, err = archiver.insert(box.space.market, {
        "BTC-USD",
        "active",

        num,
        num,
        num,
        num,
        num,

        z,z,z,z,z,z,z,z,z,z,z,z,z,z,z,z, 0,0, z, 0  -- best_bid .. last_update_sequence
    })
    t.assert_is(err, nil)
end)

g.after_each(function(cg)
    box.space.market:drop()
end)

g.test_market_migration = function(cg)
    local new_format = {
        {name = 'id', type = 'string'},
        {name = 'status', type = 'string'},
    
        {name = 'min_initial_margin', type = 'decimal'},
        {name = 'forced_margin', type = 'decimal'},
        {name = 'liquidation_margin', type = 'decimal'},
        {name = 'min_tick', type = 'decimal'},
        {name = 'min_order', type = 'decimal'},
    
        {name = 'best_bid', type = 'decimal'},
        {name = 'best_ask', type = 'decimal'},
        {name = 'market_price', type = 'decimal'},
        {name = 'index_price', type = 'decimal'},
        {name = 'last_trade_price', type = 'decimal'},
        {name = 'fair_price', type = 'decimal'},
        {name = 'instant_funding_rate', type = 'decimal'},
        {name = 'last_funding_rate_basis', type = 'decimal'},
    
        {name = 'last_update_time', type = 'number'},
        {name = 'last_update_sequence', type = 'number'},
        {name = 'average_daily_volume_q', type = 'decimal'},
        {name = 'last_funding_update_time', type = 'number'},

        {name = 'icon_url', type = 'string'},
        {name = 'market_title', type = 'string'},

        {name = 'shard_id', type = 'string'},
        {name = 'archive_id', type = 'number'},
    }
    
    -- Migrate to new schema
    migration.up()

    local sp = box.space['market']
    t.assert_is_not(sp, nil)

    -- Ceck that format migrated
    t.assert_equals(sp:format(), new_format)

    -- Check random tuple for copied correctly
    local val = box.space.market:get("BTC-USD")
    t.assert_equals(val.liquidation_margin, num)

    t.assert_is(val.icon_url, "")
    t.assert_is(val.market_title, "")
    t.assert_is(val.last_trade_price_24h_change_premium, nil)

end
