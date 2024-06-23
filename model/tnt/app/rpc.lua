local crpc = require('cartridge.rpc')
local checks = require('checks')
local log = require('log')

local config = require('app.config')
local errors = require('app.lib.errors')

local err = errors.new_class("RPC")

-- domain specific rpc calls to cluster instanses using cartridge.rpc --

local rpc = {}

local function _rpc_call(role_name, fn_name, args, opts)
    local res, e = crpc.call(role_name, fn_name, args, opts)
    if e ~= nil then
        return {res=nil, error=e}
    end

    if res.error ~= nil then
        return {res=nil, error=res.error}    
    end

    return {res = res.res, error=nil}
end


function rpc.callrw_engine(market_id, fn_name, args)
    checks('string', 'string', '?table')

    local role_name = config.sys.ID_TO_ROLES[market_id]
    if role_name == nil then
        log.warn(config.sys.ID_TO_ROLES)
        local text = "ROLE_NOT_FOUND_FOR market_id= " .. tostring(market_id)
        return {res=nil, error=text}
    end

    return _rpc_call(role_name, fn_name, args, {leader_only=true, prefer_local=false})
end


function rpc.callro_engine(market_id, fn_name, args)
    checks('string', 'string', '?table')

    local role_name = config.sys.ID_TO_ROLES[market_id]
    if role_name == nil then
        log.warn(config.sys.ID_TO_ROLES)
        local text = "ROLE_NOT_FOUND_FOR market_id= " .. tostring(market_id)
        return {res=nil, error=text}
    end

    return _rpc_call(role_name, fn_name, args, {leader_only=false, prefer_local=false})
end


function rpc.callrw_gateway(fn_name, args)
    checks('string', '?table')

    return _rpc_call("gateway", fn_name, args, {leader_only=true, prefer_local=false})
end

function rpc.callrw_get_next_task(args)
    local conn, e = crpc.get_connection("gateway", {leader_only=true, prefer_local=false})
    if conn == nil then
        log.error(e)
        return {res = nil, error = text}
    end

    local res = conn:call("next_task", {args})
    return res
end

function rpc.callro_profile(fn_name, args)
    checks('string', '?table')

    return _rpc_call("profile", fn_name, args, {leader_only=false, prefer_local=false})
end

function rpc.callrw_profile(fn_name, args)
    checks('string', '?table')

    return _rpc_call("profile", fn_name, args, {leader_only=true, prefer_local=false})
end

function rpc.watcher_engine_subscribe(market_id, key, func)
    checks('string', 'string', "?")

    local role_name = config.sys.ID_TO_ROLES[market_id]
    if role_name == nil then
        log.warn(config.sys.ID_TO_ROLES)
        local text = "ROLE_NOT_FOUND_FOR market_id= " .. tostring(market_id)
        return {res=nil, error=text}
    end

    local conn, e = crpc.get_connection(role_name, {leader_only=true, prefer_local=false})
    if conn == nil then
        log.error(e)
        local text = "watcher_engine_subscribe market_id= " .. tostring(market_id) .. " error=" .. tostring(e)
        return {res = nil, error = text}
    end

    local handler = conn:watch(key, func)

    return {res = {conn=conn, handler=handler}, error = nil}
end

function rpc.callrw_pubsub_publish(channel, json_data, ttl, size, meta_ttl)
    return _rpc_call("pubsub", "publish_data", {channel, json_data, ttl, size, meta_ttl}, {leader_only=true, prefer_local=false})
end

function rpc.wait_for_role(rolename)
    local conn, e = crpc.get_connection(rolename, {leader_only=true, prefer_local=false})

    return {res = conn, error = e}
end

function rpc.is_role_present(rolename)
    local uris = crpc.get_candidates(rolename, {healthy_only = false})
    return #uris > 0
end


function rpc.test_set_mock_callro_profile(_profile_func)
    rpc.callro_profile = _profile_func
end

return rpc