
return {    
    up = function()
        local archiver = require('app.archiver')
        local ddl = require('app.ddl')

        local sp = box.space['profile_airdrop']
        if sp == nil then
            error('space `profile_airdrop` not found')
        end


        local initial_rewards = 'initial_rewards'
        if ddl.has_column(sp:format(), initial_rewards) then
            return
        end

        -- STEP1: drop old tables
        box.space['profile_airdrop']:drop()
        if box.space['airdrop'] ~= nil then
            box.space['airdrop']:drop()            
        end
        
        if box.space['airdrop_claim_ops'] ~= nil then
            box.space['airdrop_claim_ops']:drop()
        end

        -- STEP2: create from scratch
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
            error(err)
        end
    end
}
