local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local bops_migration = require('migrations.common.eid_balance_migrations')
local profile_migration = require('migrations.profile.2023122019000001_profile_eid_field')
local config = require('app.config')
local ddl = require('app.ddl')


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

local g = t.group('ied_related_migration')
g.before_each(function(cg)
    -- Create old market and init
    archiver.init_sequencer("BTC-USD")

    -- CREATE old schema and add some records
    local old_balance_operations, err = archiver.create('balance_operations', { if_not_exists = true }, {
        { name = 'id',         type = 'string' }, -- this is the unique ID for withdraw
        { name = 'status',     type = 'string' },
        { name = 'reason',     type = 'string' },
        { name = 'txhash',     type = 'string' }, -- it's just some meta, for detect transaction onchain
        { name = 'profile_id', type = 'unsigned' },
        -- if profile_id is known then wallet comes from that, and the
        -- wallet in balance_operations can be an empty string, but in case
        -- of a deposit straight to the L1 contract (not via front end) no
        -- profile_id is known initially (and there may not be one if the
        -- depositor has not onboarded) so wallet is used
        { name = 'wallet',     type = 'string' },
        { name = 'ops_type',   type = 'string' },
        { name = 'ops_id2',    type = 'string' }, -- this is the deposit/withdrawal id to never approve the same deposit or withdrawal twice
        { name = 'amount',     type = 'decimal' },
        { name = 'timestamp',  type = 'number' },
        { name = 'due_block',  type = 'unsigned' },
    }, {
        unique = true,
        parts = { { field = 'id' } },
        if_not_exists = true,
    })
    t.assert_is_not(old_balance_operations, nil)

    -- Initially ops_id2 can be set equal to ops_id. For deposits it will eventually
    -- record the deposit id 'd_123', etc., but the deposit id is not known initially
    -- as it is assigned by the L1 contract when it starts processing the deposit
    old_balance_operations:create_index('ops_id2', {
        parts = { { field = 'ops_id2' } },
        unique = true,
        if_not_exists = true
    })

    old_balance_operations:create_index('type_status', {
        parts = { { field = 'ops_type' }, { field = 'status' } },
        unique = false,
        if_not_exists = true
    })

    old_balance_operations:create_index('wallet_type_status',
        {
            parts = { { field = 'wallet' }, { field = 'ops_type' }, { field = 'status' } },
            unique = false,
            if_not_exists = true
        })

    old_balance_operations:create_index('profile_id_type_status_amount',
        {
            parts = { { field = 'profile_id' }, { field = 'ops_type' }, { field = 'status' }, { field = 'amount' } },
            unique = false,
            if_not_exists = true
        })

    old_balance_operations:create_index('txhash', {
        parts = { { field = 'txhash' } },
        unique = false,
        if_not_exists = true
    })
    t.assert_is(err, nil)

    for _, val in pairs({1,2,3,4,5}) do 
        local res, err = archiver.insert(box.space.balance_operations, {
            tostring(val),
            config.params.BALANCE_STATUS.PENDING,
            "",
            "txhash",
            val,
            "0xsomewallet",
            config.params.BALANCE_TYPE.DEPOSIT,
            tostring(val),
            decimal.new(0),
            time.now(),
            0,
        })
    
    end


    -- CREATE old schema and add some records
    box.schema.sequence.create('PID', {start = 0, min = 0, if_not_exists = true})
    local profile, err = archiver.create('profile', { if_not_exists = true },
        {
            {name = 'id', type = 'unsigned'},
            {name = 'profile_type', type = 'string'},
            {name = 'status', type = 'string'},
            {name = 'wallet', type = 'string'},
            {name = 'created_at', type = 'number'},
        },
    {
        name = 'primary',
        sequence = 'PID',
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    t.assert_is(err, nil)

    
    profile:create_index('wallet', {
        unique = true,
        parts = {{field = 'wallet'}},
        if_not_exists = true })

    profile:create_index('profile_type', {
        unique = false,
        parts = {{field = 'profile_type'}},
        if_not_exists = true })

    for _, val in pairs({1,2,3,4,5}) do
        local p, e = archiver.insert(box.space.profile, {
            val,
            "trader",
            "active",
            tostring(val),
            time.now(),
        })
        t.assert_is(e, nil)
        t.assert_is_not(p, nil)
    end
end)

g.after_each(function(cg)
    box.space.balance_operations:drop()
    box.space.profile:drop()
   -- box.schema.sequence:drop()
end)

g.test_bops_migration = function(cg)

    local bops_new_format = {
        { name = 'id',         type = 'string' }, -- this is the unique ID for withdraw
        { name = 'status',     type = 'string' },
        { name = 'reason',     type = 'string' },
        { name = 'txhash',     type = 'string' }, -- it's just some meta, for detect transaction onchain
        { name = 'profile_id', type = 'unsigned' },
        { name = 'wallet',     type = 'string' },
        { name = 'ops_type',   type = 'string' },
        { name = 'ops_id2',    type = 'string' }, -- this is the deposit/withdrawal id to never approve the same deposit or withdrawal twice
        { name = 'amount',     type = 'decimal' },
        { name = 'timestamp',  type = 'number' },
        { name = 'due_block',  type = 'unsigned' },

        {is_nullable = false, name = "exchange_id", type = "string"},
        {is_nullable = false, name = "chain_id", type = "unsigned"},
        {is_nullable = false, name = "contract_address", type = "string"},
        {name = "shard_id", type = "string"},
        {name = "archive_id", type = "number"},
    }
    
    -- let's try to do it twice

    for _, step in pairs{1,2} do
        -- Migrate to new schema
        bops_migration.up()

        local sp = box.space['balance_operations']
        t.assert_is_not(sp, nil)
    
        -- Check that format migrated
        t.assert_equals(sp:format(), bops_new_format)
    
        -- Check that indexes created
        local expected = {
            primary=1,
            ops_id2=1,
            type_status_exchange_id_chain_id=1,
            wallet_type_status=1,
            profile_id_type_status_amount=1,
            txhash=1,
            archive_id=1,
            exchange_id=1,
            chain_id_contract_address=1,
            wallet_type_status=1,
            txhash=1,
            ops_id2=1,
            chain_id_contract_address=1,
            exchange_id_wallet_type_status=1,
            profile_id_type_status_amount=1,
            exchange_id=1,
            type_status=1,
            type_contract_due_block=1,
            primary=1,
            archive_id=1,
            exchange_id_wallet_type_status=1,    
        }
        
        local actual = {}
        for _, item in pairs(sp.index) do
            actual[item.name] = 1
        end
    
        t.assert_equals(actual, expected)
    
        -- check that vaues initialized correclty
        local some_item = sp:get("1")
        
        t.assert_is(some_item[12], 'rbx')
        t.assert_is(some_item[13], 1)
        t.assert_is(some_item[14], '0xrbxl1')
        t.assert_is(some_item[15], 'BTC-USD')
        t.assert_is(some_item[16], 1)    
    end
end


g.test_profile_migration = function(cg)

    local profile_new_format = {
        {name = 'id', type = 'unsigned'},
        {name = 'profile_type', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'wallet', type = 'string'},
        {name = 'created_at', type = 'number'},
        {is_nullable = false, name = "exchange_id", type = "string"},
        {name = "shard_id", type = "string"},
        {name = "archive_id", type = "number"},    
    }
    
    for _, step in pairs({1,2}) do
            -- Migrate to new schema
        profile_migration.up()

        local sp = box.space['profile']
        t.assert_is_not(sp, nil)

        -- Check that format migrated
        t.assert_equals(sp:format(), profile_new_format)

        -- Check that indexes created
        local expected = {
            primary=1,
            profile_type=1,
            archive_id=1,
            exchange_id=1,
            exchange_id_wallet=1,
        }


        local actual = {}


        for _, item in pairs(sp.index) do
            actual[item.name] = 1
            if item.name == "primary" then
                t.assert_is(item.sequence_fieldno, 1)
            end
        end

        t.assert_equals(actual, expected)

        -- check that vaues initialized correclty
        local some_item = sp:get(1)
        
        t.assert_is(some_item[6], 'rbx')
        t.assert_is(some_item[7], 'BTC-USD')
        t.assert_is(some_item[8], 16)

    end
end
