local checks = require('checks')
local log = require('log')

local archiver = require('app.archiver')
local config = require('app.config')
local ddl = require('app.ddl')
local time = require('app.lib.time')
local util = require('app.util')

require('app.errcodes')
require('app.config.constants')

local function migrate_once()
    box.begin()
    for _, api_secret in box.space.api_secret.index.api_secret_tag:pairs({""}) do     
        local status, res = pcall(function() return box.space.api_secret:update({api_secret.key}, {
            {'=', 'status', TEMP_API_SECRET_STATUS}
        }) end)
        if status == false then
            box.rollback()
            return
        end
    end
    box.commit()
end

local function init_space()
    archiver.init_sequencer("auth")

    -- once migrate only
    local sp = box.space['api_jwt']
    if sp ~= nil then
        if not ddl.has_column(sp:format(), "created_at") then
            box.space.api_jwt:drop()
        end
    end

    -- for local storage on frontend
    local frontend_storage = box.schema.space.create('frontend_storage', {if_not_exists = true})

    frontend_storage:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'data', type = '*'},
    })

    frontend_storage:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })


    
    -- jwt which are used for MM api calls, and must be invalidated together
    local api_jwt = box.schema.space.create('api_jwt', {if_not_exists = true})

    api_jwt:format({
        {name = 'key', type = 'string'},
        {name = 'jwt', type = 'string'},
        {name = 'refresh_token', type = 'string'},
        {name = "created_at", type = 'number'}
    })

    api_jwt:create_index('key_jwt', {
        unique = true,
        parts = {{field = 'key'}, {field = 'jwt'}},
        if_not_exists = true })

    api_jwt:create_index('by_refresh_token', {
        unique = false,
        parts = {{field = 'refresh_token'}},
        if_not_exists = true })
    

    -- jwt which are used for MM api calls, and must be invalidated together
    local allowed_ips = box.schema.space.create('allowed_ips', {if_not_exists = true})

    allowed_ips:format({
        {name = 'key', type = 'string'},
        {name = 'ip', type = 'string'},
    })

    allowed_ips:create_index('api_key_ip', {
        unique = true,
        parts = {{field = 'key'}, {field = 'ip'}},
        if_not_exists = true })

    -- Create space for storing api secrets for market makers
    local api_secret = box.schema.space.create('api_secret', {if_not_exists = true})

    api_secret:format({
        {name = 'key', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'secret', type = 'string'},
        {name = 'tag', type = 'string'},
        {name = 'expiration', type = 'unsigned'},
        {name = 'status', type = 'string'} })

    api_secret:create_index('api_secret_key', {
        unique = true,
        parts = {{field = 'key'}},
        if_not_exists = true })

    api_secret:create_index('api_secret_profile', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })
    
    api_secret:create_index('api_secret_profile_status', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'status'}},
        if_not_exists = true })
    
    api_secret:create_index('api_secret_tag', {
        unique = false,
        parts = {{field = 'tag'}},
        if_not_exists = true })
    

    -- Create space for storing frontend secrets
    local frontend_secret = box.schema.space.create('frontend_secret', {if_not_exists = true})

    frontend_secret:format({
        {name = 'jwt', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'random_secret', type = 'string'},
        {name = 'refresh_token', type = 'string'},
        {name = 'status', type = 'string'} })

    frontend_secret:create_index('frontend_secret_jwt', {
        unique = true,
        parts = {{field = 'jwt'}},
        if_not_exists = true })

    frontend_secret:create_index('frontend_secret_refresh_token', {
        unique = true,
        parts = {{field = 'refresh_token'}},
        if_not_exists = true })

    frontend_secret:create_index('frontend_secret_profile', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })

    --TODO: delete on the next release
    migrate_once()
end

local function drop_spaces()
    box.space.api_jwt:drop()
    box.space.allowed_ips:drop()
    box.space.api_secret:drop()
    box.space.frontend_secret:drop()
end

local function write_frontend_storage(profile_id, data)
    checks('number', '?')

    local res = box.space.frontend_storage:replace{profile_id, data}

    return {res = res, error = ""}
end

local function read_frontend_storage(profile_id)
    checks('number')

    local res = box.space.frontend_storage:get{profile_id}

    if res ~= nil then
        res = res.data
    end

    return {res = res, error = ""}
end


local function api_update_api_secret_expire(key, new_expired_at)
    checks('string', 'number')

    local exist = box.space.api_secret:get(key)
    if not exist then
        return {error = ERR_NO_SUCH_KEY}
    end

    local res = box.space.api_secret:update({key}, {
        {'=', 'expiration', new_expired_at}
    })
    if not res then
        return {error = ERR_API_SECRET_UPDATE}
    end

    return {error = ""}
end

local function api_list_secrets(profile_id)
    checks("number")

    local res = {}
    local count = 0
    for _, secret in box.space.api_secret.index.api_secret_profile:pairs(
        {
             profile_id, 
         },
         {iterator = "EQ"})do

        local tokens = box.space.api_jwt:select({secret.key})
            
        -- if no jwt then skip it
        if tokens ~= nil and #tokens > 0 then
            local token = tokens[1]
            local item = {
                api_secret=secret,
                allowed_ip_list={},
                jwt_private=token.jwt,
                refresh_token=token.refresh_token,
                jwt_public=config.params.JWT_PUBLIC,
                created_at=token.created_at,
            }
        
            for _, ip in box.space.allowed_ips:pairs(
                {
                    secret.key, 
                },
                {iterator = "EQ"}) do
                
                table.insert(item.allowed_ip_list, ip.ip)
            end

            table.insert(res, item)
        end
         -- should never happened
         count = util.safe_yield(count, 1000)
     end
 

    return {secrets = res, error = ""}
end

local function api_refresh_secret(profile_id, refresh_token, new_jwt, new_refresh, new_expired_at)
    checks('number', 'string', 'string', 'string', 'number')

    local jwt = box.space.api_jwt.index.by_refresh_token:select({refresh_token})
    if not jwt or #jwt <= 0 then
        return {secret = nil, error = ERR_REFRESH_TOKEN_NOT_FOUND}
    end

    local key = jwt[1].key
    local old_jwt = jwt[1].jwt

    local api_secret = box.space.api_secret:get(key)
    if not api_secret then
        return {secret = nil, error = ERR_NO_API_KEY_FOR_REFRESH_TOKEN}    
    end

    if api_secret.profile_id ~= profile_id then
        return {secret = nil, error = ERR_NOT_YOUR_REFRESH_TOKEN}        
    end

    local new_secret = {    
        api_secret=api_secret:update({{'=', 'expiration', new_expired_at}}),
        allowed_ip_list={},
        jwt_private=new_jwt,
        refresh_token=new_refresh,
        jwt_public=config.params.JWT_PUBLIC

    }
    
    local tm = time.now()
    box.begin()
        box.space.api_jwt:delete({key, old_jwt})
        local res = box.space.api_jwt:insert({
            key,
            new_jwt,
            new_refresh,
            tm
        })
        if not res then
            box.rollback()
            return {secret = nil, error = ERR_API_JWT_UPDATE_ERROR}
        end

        res = box.space.api_secret:update({key}, {
            {'=', 'expiration', new_expired_at}
        })
        if not res then
            box.rollback()
            return {secret = nil, error = ERR_API_SECRET_UPDATE_ERROR}
        end
    box.commit()


    return {secret = new_secret, error = ""}
end

local function api_secret_validate(key, ip, now)
    checks('string', 'string', 'number')

    local exist = box.space.api_secret:get(key)
    if not exist then
        return {is_valid = false, error = ERR_NO_SUCH_KEY}
    end

    if exist.expiration < now then
        return {is_valid = false, error = ERR_KEY_EXPIRED}
    end

    if box.space.allowed_ips:count({key}) > 0 then
        local allowed = box.space.allowed_ips:get({key, ip})
        if not allowed then
            return {is_valid = false, error = ERR_NOT_ALLOWED_FOR_IP}
        end
    end

    return {is_valid = true, error = ""}
end


local function api_delete_secret_key(profile_id, key)
    checks('number', 'string')

    local status, res

    local exist = box.space.api_secret:get(key)
    if not exist then
        return {api_secret = nil, error = ERR_NO_SUCH_KEY}
    end

    if exist.profile_id ~= profile_id then
        return {api_secret = nil, error = ERR_NOT_YOUR_KEY}
    end

    local delete_double_key = function(space_name, filter, iter)
        for _, item in box.space[space_name]:pairs(filter, iter) do
            box.space[space_name]:delete({item[1], item[2]})    
        end    
    end

    box.begin()
        box.space.api_secret:delete(key)
        
        delete_double_key("api_jwt", {key}, {iterator='EQ'})
        delete_double_key("allowed_ips", {key}, {iterator='EQ'})

    box.commit()

    return {api_secret = nil, error = ""}
end

-- Functions for api_secret
local function api_secret_pair_create(api_secret, jwt, refresh_token, ips)
    checks('table', 'string', 'string', '?table')

    local status, res1, res2, res3

    local key = api_secret[1]

    local total = box.space.api_secret.index.api_secret_profile:count(api_secret[2])

    if total >= config.params.MAX_SECRETS_PER_ACCOUNT then
        return {api_secret = api_secret, error = ERR_MAX_SECRETS_EXCEED}
    end

    local tm = time.now()
    box.begin()

    status, res1 = pcall(function() return box.space.api_secret:insert(api_secret) end)
    if status == false then
        box.rollback()
        return {api_secret = api_secret, error = res1}
    end

    status, res2 = pcall(function() return box.space.api_jwt:insert({key, jwt, refresh_token, tm}) end)
    if status == false then
        box.rollback()
        return {api_secret = api_secret, error = res2}
    end

    if ips ~= nil and #ips > 0 then
        for _, ip in pairs(ips) do
            status, res3 = pcall(function() return box.space.allowed_ips:replace({key, ip}) end)
            if status == false then
                box.rollback()
                return {api_secret = api_secret, error = res3}
            end        
        end
    end

    box.commit()

    return {api_secret = api_secret,  error = ""}
end

local function remove_session_secrets(profile_id)
    checks("number")
    
    box.begin()
    local exist = box.space.frontend_secret.index.frontend_secret_profile:min(profile_id)
    if exist ~= nil then 
        box.space.frontend_secret:delete(exist.jwt)
    end
    box.commit()

    return {error = ""}
end

local function api_secret_create(api_secret)
    checks('table')

    local status, res = pcall(function() return box.space.api_secret:insert(api_secret) end)
    if status == false then
        return {api_secret = api_secret, error = res}
    end

    return {api_secret = res, error = ""}
end

local function api_secret_by_key(api_secret_key)
    checks('string')

    local api_secret = box.space.api_secret.index.api_secret_key:get(api_secret_key)
    if api_secret == nil then
        return {api_secret = nil, error = ERR_NO_SUCH_KEY}
    end

    return {api_secret = api_secret, error = ""}
end

--TODO: temp function while we are not fully migrated to the new type of secrets
-- status is used for backward compability, to split old api_secrets from the new ones
-- we will need fully moved to new secrets.
local function get_or_refresh_api_secret_by_profile_id(profile_id, new_api_secret)
    checks('number', 'table')

    local res = nil 
    local status

    local unixts = time.now_sec()

    -- Find the current MM api secret, if it's expired delete, and generate the new one
    box.begin()
    for _, api_secret in box.space.api_secret.index.api_secret_profile_status:pairs(
        {profile_id, TEMP_API_SECRET_STATUS}, 
        {iterator = "EQ"}) do 
        if api_secret.expiration < unixts then
            box.space.api_secret:delete(api_secret.key)
        else
            res = api_secret
            break
        end
    end

    if res == nil then
        status, res = pcall(function() return box.space.api_secret:insert(new_api_secret) end)
        if status == false then
            box.rollback()
            return {api_secret = nil, error = res}
        end
    end

    box.commit()

    return {api_secrets = {res}, error = ""}
end

-- TODO: add offset/limit
local function api_secret_by_profile_id(profile_id)
    checks('number')

    local status, res = pcall(function() return box.space.api_secret.index.api_secret_profile:select(profile_id) end)
    if status == false then
        return {api_secrets = nil, error = res}
    end

    return {api_secrets = res, error = ""}
end

-- Functions for frontend_secret

local function frontend_secret_create(frontend_secret, old_jwt)
    checks('table', '?string')

    if old_jwt then
        box.space.frontend_secret:delete(old_jwt)
    end

    local status, res = pcall(function() return box.space.frontend_secret:replace(frontend_secret) end)
    if status == false then
        return {frontend_secret = frontend_secret, error = res}
    end

    return {frontend_secret = res, error = ""}
end

local function frontend_secret_by_jwt(jwt)
    checks('string')

    local frontend_secret = box.space.frontend_secret.index.frontend_secret_jwt:get(jwt)
    if frontend_secret == nil then
        return {frontend_secret = nil, error = ERR_NO_SUCH_KEY}
    end

    return {frontend_secret = frontend_secret, error = ""}
end

local function frontend_secret_by_refresh_token(refresh_token)
    checks('string')

    local frontend_secret = box.space.frontend_secret.index.frontend_secret_refresh_token:get(refresh_token)
    if frontend_secret == nil then
        return {frontend_secret = nil, error = ERR_NO_SUCH_KEY}
    end

    return {frontend_secret = frontend_secret, error = ""}
end

local function frontend_secret_update(frontend_secret)
    checks('table')

    local current = box.space.frontend_secret:get(frontend_secret[1])
    if current == nil then
        return {frontend_secret = frontend_secret, error = ERR_NO_SUCH_KEY}
    end

    local status, res = pcall(function() return box.space.frontend_secret:replace(frontend_secret) end)
    if status == false then
        return {frontend_secret = frontend_secret, error = res}
    end

    return {frontend_secret = res, error = ""}
end

local function _test_set_time(override_time)
    time = override_time
end


local function init(opts)
    opts = opts or {}

    if opts.is_master then
        init_space()

        box.schema.func.create('api_secret_create', {if_not_exists = true})
        box.schema.func.create('api_secret_by_key', {if_not_exists = true})
        box.schema.func.create('api_secret_by_profile_id', {if_not_exists = true})
        box.schema.func.create('frontend_secret_create', {if_not_exists = true})
        box.schema.func.create('frontend_secret_by_jwt', {if_not_exists = true})
        box.schema.func.create('frontend_secret_by_refresh_token', {if_not_exists = true})
        box.schema.func.create('frontend_secret_update', {if_not_exists = true})
        box.schema.func.create('api_secret_pair_create', {if_not_exists = true})
        box.schema.func.create('api_delete_secret_key', {if_not_exists = true})
        box.schema.func.create('api_secret_validate', {if_not_exists = true})
        box.schema.func.create('api_list_secrets', {if_not_exists = true})
        box.schema.func.create('api_refresh_secret', {if_not_exists = true})
        box.schema.func.create('api_update_api_secret_expire', {if_not_exists = true})
        box.schema.func.create('remove_session_secrets', {if_not_exists = true})
        box.schema.func.create('write_frontend_storage', {if_not_exists = true})
        box.schema.func.create('read_frontend_storage', {if_not_exists = true})
        box.schema.func.create('get_or_refresh_api_secret_by_profile_id', {if_not_exists = true})

    end

    rawset(_G, 'api_secret_create', api_secret_create)
    rawset(_G, 'api_secret_by_key', api_secret_by_key)
    rawset(_G, 'api_secret_by_profile_id', api_secret_by_profile_id)
    rawset(_G, 'frontend_secret_create', frontend_secret_create)
    rawset(_G, 'frontend_secret_by_jwt', frontend_secret_by_jwt)
    rawset(_G, 'frontend_secret_by_refresh_token', frontend_secret_by_refresh_token)
    rawset(_G, 'frontend_secret_update', frontend_secret_update)
    rawset(_G, 'api_secret_pair_create', api_secret_pair_create)
    rawset(_G, 'api_delete_secret_key', api_delete_secret_key)
    rawset(_G, 'api_secret_validate', api_secret_validate)
    rawset(_G, 'api_list_secrets', api_list_secrets)
    rawset(_G, 'api_refresh_secret', api_refresh_secret)
    rawset(_G, 'api_update_api_secret_expire', api_update_api_secret_expire)
    rawset(_G, 'remove_session_secrets', remove_session_secrets)
    rawset(_G, 'write_frontend_storage', write_frontend_storage)
    rawset(_G, 'read_frontend_storage', read_frontend_storage)
    rawset(_G, 'get_or_refresh_api_secret_by_profile_id', get_or_refresh_api_secret_by_profile_id)

end

return {
    role_name = 'auth',
    init = init,
    utils = {
        api_secret_create = api_secret_create,
        api_secret_by_key = api_secret_by_key,
        api_secret_by_profile_id = api_secret_by_profile_id,
        frontend_secret_create = frontend_secret_create,
        frontend_secret_by_jwt = frontend_secret_by_jwt,
        frontend_secret_by_refresh_token = frontend_secret_by_refresh_token,
        frontend_secret_update = frontend_secret_update,
        api_secret_pair_create = api_secret_pair_create,
        api_delete_secret_key = api_delete_secret_key,
        api_secret_validate = api_secret_validate,
        api_list_secrets = api_list_secrets,
        api_refresh_secret = api_refresh_secret,
        drop_spaces = drop_spaces,
        init_space = init_space,
        api_update_api_secret_expire = api_update_api_secret_expire,
        remove_session_secrets = remove_session_secrets,
        write_frontend_storage = write_frontend_storage,
        read_frontend_storage = read_frontend_storage,
        get_or_refresh_api_secret_by_profile_id = get_or_refresh_api_secret_by_profile_id,
        _test_set_time = _test_set_time,
    }
}
