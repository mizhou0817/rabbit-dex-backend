local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local ddl = require('app.ddl')
local profile = require('app.engine.profile')
local setters = require('app.engine.setters')
local config = require('app.config')

local z = decimal.new(0)
require('app.config.constants')

local g = t.group('tier')

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

g.before_each(function(cg)
    archiver.init_sequencer("profile")
    profile.init_spaces()

end)

g.after_each(function(cg)
    box.space.tier:drop()
    box.space.special_tier:drop()
    box.space.profile_to_tier:drop()
end)

g.test_tier_flow = function(cg)

    local profile_id = 12

    -- return 0 tier if no special
    local which_tier = profile.which_tier(profile_id)
    t.assert_is(which_tier.tier, 0)
    t.assert_is(which_tier.taker_fee, decimal.new("0.0007"))

    -- return exact tier
    profile_id = 13
    local res

    res = setters.update_profiles_to_tiers({{profile_id, 0, 7}})
    t.assert_is(res["error"], nil)

    which_tier = profile.which_tier(profile_id)
    t.assert_is(which_tier.tier, 7)

    -- add special tier
    local special_tier_id = 22
    local specila_tier_title = "special_tier"
    local res = profile.add_special_tier(special_tier_id, specila_tier_title, decimal.new(0.1), decimal.new(0.2))
    t.assert_is(res["error"], nil)

    -- change tier to special
    res = setters.update_profiles_to_tiers({{profile_id, special_tier_id, 0}})
    t.assert_is(res["error"], nil)

    which_tier = profile.which_tier(profile_id)
    t.assert_is(which_tier.tier, special_tier_id)
    t.assert_is(which_tier.taker_fee, decimal.new(0.2))
end


