local checks = require('checks')
local decimal = require("decimal")

local ddl = require('app.ddl')
local dml = require('app.dml')
local util = require('app.util')

local YIELD_LIMIT = 1000

local T = {
    tier_format = {
        {name = 'tier', type = 'unsigned'},
        {name = 'title', type = 'string'},
        {name = 'maker_fee', type = 'decimal'},
        {name = 'taker_fee', type = 'decimal'},
        {name = 'min_volume', type = 'decimal'},
        {name = 'min_assets', type = 'decimal'},
    },
    special_tier_format = {
        {name = 'tier', type = 'unsigned'},
        {name = 'title', type = 'string'},
        {name = 'maker_fee', type = 'decimal'},
        {name = 'taker_fee', type = 'decimal'},
    },
    profile_to_tier_format = {
        {name = 'profile_id', type = 'unsigned'},
        {name = 'special_tier_id', type = 'unsigned'},
        {name = 'tier_id', type = 'unsigned'},
    },
}

-- Create initial tiers for market if not exist
local function init_tiers()
    local tiers = {
        {0, "VIP 0 (Shrimp)", decimal.new(0), decimal.new("0.0007"), decimal.new(0), decimal.new(0)},
        {1, "VIP 1 (Herring)", decimal.new(0), decimal.new("0.0005"), decimal.new("1000000"), decimal.new("100000")},
        {2, "VIP 2 (Trout)", decimal.new(0), decimal.new("0.00045"), decimal.new("5000000"), decimal.new("500000")},
        {3, "VIP 3 (Tuna)", decimal.new(0), decimal.new("0.0004"), decimal.new("10000000"), decimal.new("1000000")},
        {4, "MM I / VIP 4 (Swordfish)", decimal.new("-0.0001"), decimal.new("0.00037"), decimal.new("50000000"), decimal.new("5000000")},
        {5, "MM II / VIP 5 (Shark)", decimal.new("-0.0001"), decimal.new("0.00035"), decimal.new("250000000"), decimal.new("20000000")},
        {6, "MM III / VIP 6 (Whale)", decimal.new("-0.0002"), decimal.new("0.00031"), decimal.new("1500000000"), decimal.new("100000000")},
        {7, "Special", decimal.new("-0.0002"), decimal.new("0.0002"), decimal.new("20000000000000"), decimal.new("1000000000000")}
    }

    -- NO need replace if exist
    local count = box.space.tier:count()
    if count > 1 then
        return
    end

    for _, tier in pairs(tiers) do
        box.space.tier:replace(tier)
    end
end

function T.init_spaces()
    ddl.create_space('tier', {if_not_exists = true}, T.tier_format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'tier'}},
            if_not_exists = true,
        },
    })

    ddl.create_space('special_tier', {if_not_exists = true}, T.special_tier_format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'tier'}},
            if_not_exists = true,
        },
    })

    ddl.create_space('profile_to_tier', {if_not_exists = true}, T.profile_to_tier_format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'profile_id'}},
            if_not_exists = true,
        },
    })

    init_tiers()
end

function T.get_tiers()
    local tiers, count = {}, 0

    for _, t in box.space.tier:pairs() do
        table.insert(tiers, t)
        count = util.safe_yield(count, YIELD_LIMIT)
    end

    return tiers
end

function T.update_profiles_to_tiers(profiles_tiers)
    checks('table')

    local count = 0
    return dml.atomic(function()
        for _, pt in pairs(profiles_tiers) do
            local profile_id, special_tier_id, tier_id = unpack(pt)

            local tuple = {
                profile_id,
                special_tier_id,
                tier_id,
            }

            if special_tier_id ~= 0 then
                box.space.profile_to_tier:upsert(tuple, {
                    {'=', 'special_tier_id', special_tier_id},
                    {'=', 'tier_id', 0},
                })
            elseif tier_id ~= 0 then
                box.space.profile_to_tier:upsert(tuple, {
                    {'=', 'tier_id', tier_id},
                    {'=', 'special_tier_id', 0},
                })
            else
                box.space.profile_to_tier:delete(profile_id)
            end

            count = util.safe_yield(count, YIELD_LIMIT)
        end
    end)
end

return T
