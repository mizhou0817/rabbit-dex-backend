local decimal = require('decimal')
local fio = require('fio')
local t = require('luatest')
local log = require('log')
local time = require('app.lib.time')
local archiver = require('app.archiver')
local ddl = require('app.ddl')
local config = require('app.config')
local auth = require('app.roles.auth')

require('app.config.constants')
require('app.errcodes')

local g = t.group('auth.api_secret_refresh')

local mock_time = {}
local NOW_TIME = 10
local NOW_SEC = 10
function mock_time.now()
    return NOW_TIME
end

function mock_time.now_sec()
    return NOW_SEC
end


local cur_dir = fio.cwd()
local work_dir = fio.tempdir()
t.before_suite(function()
    box.cfg{
        work_dir = work_dir,
    }

    auth.utils._test_set_time(mock_time)
end)

t.after_suite(function()
    fio.rmtree(work_dir)
end)

g.before_each(function(cg)
    auth.utils.init_space()
end)

g.after_each(function(cg)
    auth.utils.drop_spaces()
end)

g.test_auth_flow = function(cg)

    local profile_id = 1
    local secrets = {
        {
            "1",
            profile_id,
            "secret1",
            "tag1",
            123,
            "status",
        },
        {
            "2",
            profile_id,
            "secret2",
            "tag2",
            10,
            TEMP_API_SECRET_STATUS
        },
        {
            "3",
            profile_id,
            "secret3",
            "tag3",
            20,
            TEMP_API_SECRET_STATUS
        },
        {
            "4",
            profile_id,
            "secret4",
            "tag4",
            30,
            TEMP_API_SECRET_STATUS      
        },
    }

    for _, s in ipairs(secrets) do
        local res = auth.utils.api_secret_create(s)
        t.assert_is(res["error"], "")
    end

    local count = 2
    for _, s in box.space.api_secret.index.api_secret_profile_status:pairs({profile_id, TEMP_API_SECRET_STATUS}) do
        t.assert_is(s.key, tostring(count))
        count = count + 1
    end

    -- One expired, need to return 3
    NOW_SEC = 11
    local res = auth.utils.get_or_refresh_api_secret_by_profile_id(profile_id, {})
    local s = res['api_secrets'][1]
    local e = res['error']
    t.assert_is(e, "")
    t.assert_is(s[1], "3")

    -- all expired need to create the new one
    NOW_SEC = 33
    local new_secret = {
        "5",
        profile_id,
        "secret4",
        "tag4",
        300,
        TEMP_API_SECRET_STATUS      
    }
    res = auth.utils.get_or_refresh_api_secret_by_profile_id(profile_id, new_secret)
    s = res['api_secrets'][1]
    e = res['error']
    t.assert_is(e, "")
    t.assert_is(s[1], "5")

    for _, s in box.space.api_secret.index.api_secret_profile_status:pairs({profile_id, TEMP_API_SECRET_STATUS}) do
        t.assert_is(s.key, "5")
    end

    

end

