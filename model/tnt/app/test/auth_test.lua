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

local g = t.group('auth')

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
    auth.utils.init_space()
end)

g.after_each(function(cg)
    auth.utils.drop_spaces()
end)

g.test_auth_flow = function(cg)
    local secret = {
        "0",
        1,
        "secret1",
        "tag1",
        123,
        "status"
    }

    local res = auth.utils.api_secret_create(secret)
    t.assert_is(res["error"], "")
    res = auth.utils.api_secret_by_key("0")
    t.assert_is(res["api_secret"][5], 123)

    res = auth.utils.api_update_api_secret_expire("0", 321)
    t.assert_is(res["error"], "")

    res = auth.utils.api_secret_by_key("0")
    t.assert_is(res["api_secret"][5], 321)

    secret[2] = 2
    secret[1] = "01"
    res = auth.utils.api_secret_create(secret)
    t.assert_is(res["error"], "")

    secret[2] = 1

    secret[1] = "1"
    res = auth.utils.api_secret_pair_create(secret, "jwt1", "refresh_token1")

    -- create 102 secrets
    for i = 2, 102, 1 do
        secret[1] = tostring(i)
        local jwt = "jwt" .. tostring(i)
        local refresh_token = "refresh_token" ..tostring(i)
        res = auth.utils.api_secret_pair_create(secret, jwt, refresh_token, {"1", "2", "3"})
        if i < 100 then
            t.assert_is(res["error"], "")     
        else
            t.assert_is(res["error"], ERR_MAX_SECRETS_EXCEED)
        end   
    end


    res = auth.utils.api_list_secrets(1)
    local secrets = res["secrets"]

    --[[
        Expected res
        {"refresh_token":"refresh_token",
        "jwt_private":"jwt",
        "api_secret":["3",1,"secret1","tag1",123,"status"],
        "jwt_public":"qq",
        "allowed_ip_list":["1","2","3"]}    
    --]]
    for i, s in ipairs(secrets) do
        t.assert_is(s["api_secret"][2], 1)
        t.assert_str_contains(s["refresh_token"], "refresh_token")
        t.assert_str_contains(s["jwt_private"], "jwt")
    end


    -- CHECK validation
    res = auth.utils.api_secret_validate("000", "1", 1)
    t.assert_is(res["error"], ERR_NO_SUCH_KEY)

    res = auth.utils.api_secret_validate("0", "1", 1000)
    t.assert_is(res["error"], ERR_KEY_EXPIRED)

    res = auth.utils.api_secret_validate("1", "111", 1)
    t.assert_is(res["error"], "")

    res = auth.utils.api_secret_validate("11", "111", 1)
    t.assert_is(res["error"], ERR_NOT_ALLOWED_FOR_IP)

    res = auth.utils.api_secret_validate("11", "3", 1)
    t.assert_is(res["error"], "")


    -- Check delition
    res = auth.utils.api_delete_secret_key(1, "11")
    t.assert_is(res["error"], "")

    --can't delete twice
    res = auth.utils.api_delete_secret_key(1, "11")
    t.assert_is(res["error"], ERR_NO_SUCH_KEY)

    -- check  that all artifacts were deleted
    t.assert_is(box.space.api_secret:get("11"), nil)
    t.assert_is(box.space.api_jwt:count({"11"}), 0)
    t.assert_is(box.space.allowed_ips:count({"11"}), 0)
    


    -- can't refresh non-existing
    res = auth.utils.api_refresh_secret(1, "refresh_token11", "new_jwt", "new_refresh", 33)
    t.assert_is(res["error"], ERR_REFRESH_TOKEN_NOT_FOUND)



    res = auth.utils.api_refresh_secret(1, "refresh_token66", "new_jwt", "new_refresh", 33)
    t.assert_is(res["error"], "")

    -- old refresh shouldn't exist anymore
    res = auth.utils.api_refresh_secret(1, "refresh_token66", "new_jwt", "new_refresh", 33)
    t.assert_is(res["error"], ERR_REFRESH_TOKEN_NOT_FOUND)

    local n_key = box.space.api_secret:get("66")
    local n_jwt = box.space.api_jwt:get({"66", "new_jwt"})
    local n_ips = box.space.allowed_ips:select({"66"})
    t.assert_is(n_key.expiration, 33)
    t.assert_is(n_jwt.key, "66")
    t.assert_is(n_jwt.jwt, "new_jwt")
    t.assert_is(n_jwt.refresh_token, "new_refresh")
    
    for i, ip in ipairs(n_ips) do
        t.assert_is(ip.key, "66")
        t.assert_is(ip.ip, tostring(i))

    end

    

end

