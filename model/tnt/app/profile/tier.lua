local checks = require('checks')

local ddl = require('app.ddl')
local tier = require('app.tier')
local util = require('app.util')

local YIELD_LIMIT = 1000

local T = {
    profile_special_tier_format = {
        {name = 'profile_id', type = 'unsigned'},
        {name = 'special_tier_id', type = 'unsigned'},
    },
    affiliate_profile_tier_format = {
        {name = 'profile_id', type = 'unsigned'},
        {name = 'special_tier_id', type = 'unsigned'},
        {name = 'replace_tier_id', type = 'unsigned'},
    },

    get_tiers = tier.get_tiers,
}

function T.init_spaces()
    tier.init_spaces()

    ddl.create_space('profile_special_tier', {if_not_exists = true}, T.profile_special_tier_format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'profile_id'}},
            if_not_exists = true,
        },
    })

    ddl.create_space('affiliate_profile_tier', {if_not_exists = true}, T.affiliate_profile_tier_format, {
        {
            name = 'primary',
            unique = true,
            parts = {{field = 'profile_id'}},
            if_not_exists = true,
        },
    })
end

function T.get_profiles_special_tiers()
    local profiles_tiers, count = {}, 0

    for _, t in box.space.profile_special_tier:pairs() do
        table.insert(profiles_tiers, t)
        count = util.safe_yield(count, YIELD_LIMIT)
    end

    return profiles_tiers
end

function T.get_affiliate_profiles_tiers(profiles_ids)
    checks('?table')

    local profiles_tiers, count = {}, 0

    if util.is_nil(profiles_ids) or #profiles_ids == 0 then
        for _, t in box.space.affiliate_profile_tier:pairs() do
            table.insert(profiles_tiers, t)
            count = util.safe_yield(count, YIELD_LIMIT)
        end
    else
        for _, id in ipairs(profiles_ids) do
            local t = box.space.affiliate_profile_tier:get(id)
            if t then
                table.insert(profiles_tiers, t)
            end
            count = util.safe_yield(count, YIELD_LIMIT)
        end
    end

    return profiles_tiers
end

return T
