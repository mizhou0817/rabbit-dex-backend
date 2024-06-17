
return {    
    up = function()
        local archiver = require('app.archiver')
        local ddl = require('app.ddl')

        local new_format = {
            {name = 'id', type = 'string'},
            {name = 'status', type = 'string'},
        
            {name = 'min_initial_margin', type = 'decimal'},
            {name = 'forced_margin', type = 'decimal'},
            {name = 'liquidation_margin', type = 'decimal'},
            {name = 'min_tick', type = 'decimal'},
            {name = 'min_order', type = 'decimal'},
        
            {name = 'best_bid', type = 'decimal'},
            {name = 'best_ask', type = 'decimal'},
            {name = 'market_price', type = 'decimal'},
            {name = 'index_price', type = 'decimal'},
            {name = 'last_trade_price', type = 'decimal'},
            {name = 'fair_price', type = 'decimal'},
            {name = 'instant_funding_rate', type = 'decimal'},
            {name = 'last_funding_rate_basis', type = 'decimal'},
        
            {name = 'last_update_time', type = 'number'},
            {name = 'last_update_sequence', type = 'number'},
            {name = 'average_daily_volume_q', type = 'decimal'},
            {name = 'last_funding_update_time', type = 'number'},

            {name = 'icon_url', type = 'string'},
            {name = 'market_title', type = 'string'},    
        }
        
        local sp = box.space['market']
        if sp == nil then
            error('space `market` not found')
        end

        local cur_fmt, err = archiver.format(sp)
        if err ~= nil then
            error(err)
        end

        local migration_done = true
        if #cur_fmt.columns ~= #new_format then
            migration_done = false
        else
            -- We have formats of the same length, no need to check out of range error in the loop
            for key, value in pairs(cur_fmt.columns or {}) do
                if value.name ~= new_format[key].name then
                    migration_done = false
                    break
                end
            end    
        end

        if migration_done == true then
            return
        end

        local tmp_sp, err = archiver.create('market_tmp', 
            cur_fmt.options, 
            new_format,
            cur_fmt.indices
        )
        if err ~= nil then
            error(err)
        end

        for _, tuple in sp.index.primary:pairs(nil, {iterator = box.index.ALL}) do
            tmp_sp:insert{
                tuple.id,
                tuple.status,

                tuple.min_initial_margin,
                tuple.forced_margin,
                tuple.liquidation_margin,
                tuple.min_tick,
                tuple.min_order,

                tuple.best_bid,
                tuple.best_ask,
                tuple.market_price,
                tuple.index_price,
                tuple.last_trade_price,
                tuple.fair_price,
                tuple.instant_funding_rate,
                tuple.last_funding_rate_basis,

                tuple.last_update_time,
                tuple.last_update_sequence,
                tuple.average_daily_volume_q,
                tuple.last_funding_update_time,

                "", 
                "",

                tuple.shard_id,
                tuple.archive_id,
            }
        end


        sp:drop()
        tmp_sp:rename(sp.name)
    end
}
