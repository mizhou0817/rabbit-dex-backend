local stats = require('app.stats')
local deadman = require('app.api.deadman')

-- NEED THIS so queue could init itself
local tqueue = require('app.tqueue')

local function init_spaces(opts)
    deadman.init_spaces()

    box.schema.sequence.create('order_sequence',{start=1000, min=1000, if_not_exists = true})

    local profile_meta_cache = box.schema.space.create('profile_meta_cache', {temporary=true, if_not_exists = true})
    profile_meta_cache:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'market_id', type = 'string'},
        {name = 'meta', type = '*'},
    })
    profile_meta_cache:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true })

    profile_meta_cache:create_index('profile_id', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true })    

    local used_client_order_id = box.schema.space.create('used_client_order_id', {temporary=false, if_not_exists = true})
    used_client_order_id:format({
        {name = 'profile_id', type = 'unsigned'},
        {name = 'client_order_id', type = 'string'},
        {name = 'order_id', type = 'string'},
        {name = 'market_id', type = 'string'},      -- we use market_id for invalidate later
    })

    used_client_order_id:create_index('primary', {
        unique = true,
        parts = {{field = 'profile_id'}, {field = 'client_order_id'}},
        if_not_exists = true })

    used_client_order_id:create_index('profile_market', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true })    

    used_client_order_id:create_index('profile_order_id', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'order_id'}},
        if_not_exists = true })
end



return {
    public = require('app.api.public'),
    internal = require('app.api.internal'),
    stats = stats,
    init_spaces = init_spaces
}