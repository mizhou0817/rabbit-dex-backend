return {
    up = function()
        local ddl = require('app.ddl')

        local sp = box.space['used_client_order_id']
        if sp == nil then
            error('space `used_client_order_id` not found')
        end


        local order_id = 'order_id'
        if ddl.has_column(sp:format(), order_id) then
            return
        end

        -- STEP1: drop old tables
        box.space['used_client_order_id']:drop()

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
}
