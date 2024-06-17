local checks = require('checks')
local log = require('log')
local decimal = require('decimal')

local config = require('app.config')
local time = require('app.lib.time')
local util = require('app.util')
local csv = require('csv')
local fio = require('fio')
local balance = require('app.balance')
local tuple = require('app.tuple')

local p = require('app.profile.profile')
local rpc = require('app.rpc')

require("app.config.constants")
require('app.errcodes')

local R = {
    format = {
        { name='id', type='string' },
        { name='profile_id', type="number" },
        { name='market_id', type="string" },
        { name='price', type="decimal" },
        { name='size', type="decimal" },
        { name='side', type="string" },
        { name='taker_profile_id', type="number" },
    },
    strict_type = 'revert_fill_record',
}

local function init_spaces()
    local revert_list = box.schema.space.create('revert_list', {
        if_not_exists = true,
        format = {
            {name = 'fill_id', type = 'string'},
            {name = 'status', type = 'string'},
            {name = 'created_at', type = 'number'},
            {name = 'market_id', type = 'string'},
            {name = 'maker_fill', type = '*'},
            {name = 'result', type = '*'},
            {name = 'error', type = 'string'},
        },
    })

    -- we can't revert some TX twice, only one revert is possible
    revert_list:create_index('primary', {
        unique = true,
        parts = {{field = 'fill_id'}},
        if_not_exists = true,
    })

    revert_list:create_index('by_market', {
        unique = false,
        parts = {{field = 'market_id'}},
        if_not_exists = true,
    })


    revert_list:create_index('status', {
        unique = false,
        parts = {{field = 'status'}},
        if_not_exists = true,
    })


end

local function drop_spaces()
    box.space.revert_list:drop()
end

local function _test_set_rpc(override_rpc)
    rpc = override_rpc
end

function R._validate_row(row, i)
    checks("table|revert_fill_record", "?number")

    if row.profile_id == row.taker_profile_id then
        log.error("TAKER_AND_MAKER_SAME index = %s row = %s", tostring(i), table.concat(row, ","))
        return ERR_REVERT_TAKER_AND_MAKER_SAME
    end

    if not row.taker_profile_id then
        log.error("NO_TAKER index = %s row = %s", tostring(i), table.concat(row, ","))
        return ERR_REVERT_NO_TAKER
    end

    local side = row.side
    if side ~= config.params.LONG and side ~= config.params.SHORT then
        log.error("UNKNOWN_SIDE index = %s row = %s", tostring(i), table.concat(row, ","))
        return ERR_REVERT_UNKNOWN_SIDE
    end

    if not config.markets[row.market_id] then
        log.error("NO_SUCH_MARKET index = %s row = %s", tostring(i), table.concat(row, ","))
        return ERR_REVERT_NO_SUCH_MARKET
    end

    return nil
end

function R.bind(value)
    local res = tuple.new(value, R.format, R.strict_type)

    local is_valid = R._validate_row(res)
    if is_valid ~= nil then
        return nil
    end

    -- convert types. (TODO: move to tuple)
    res.profile_id = tonumber(res.profile_id)
    res.taker_profile_id = tonumber(res.taker_profile_id)
    res.price = decimal.new(res.price)
    res.size = decimal.new(res.size)

    return res
end


function R.raw_insert_revert_fill(raw, i)
    checks("table", "?number")

    local tup = R.bind(raw)
    if tup == nil then
        return ERR_REVERT_CANT_BIND
    end

    return R.insert_revert_fill(tup, i)
end

function R.insert_revert_fill(tup, i)
    checks("table|revert_fill_record", "?number")

    -- skip duplicated values
    local status, res = pcall(function() return box.space.revert_list:insert{
        tup.id,
        config.params.REVERT_STATUS.TX_UPLOADED,
        time.now(),
        tup.market_id,
        tup,
        {},
        ""
    } end)

    -- but log
    if status == false then
        log.error(res)
        return ERR_REVERT_ALREADY_EXIST
    end

    return nil
end

function R.execute_revert_list()
    if box.space.revert_list:count() == 0 then
        log.info("Nothing to execute")
        return nil
    end

    local total = 0
    for i, tup in box.space.revert_list.index.status:pairs(config.params.REVERT_STATUS.TX_UPLOADED, {iterator = "EQ"}) do
        
        local revert_fill = R.bind(tup.maker_fill)
        if revert_fill == nil then
            log.error({
                message = string.format("%s: can't convert to revert_fill", ERR_INTEGRITY_ERROR),
                [ALERT_TAG] = ALERT_CRIT,
            })
            return ERR_INTEGRITY_ERROR
        end

        -- pause accounts
        local res = p.update_status(revert_fill.profile_id, config.params.PROFILE_STATUS.BLOCKED)
        if res["error"] ~= nil then
            log.error("CAN'T block for maker_id = %s", tostring(revert_fill.maker_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end

        res = p.update_status(revert_fill.taker_profile_id, config.params.PROFILE_STATUS.BLOCKED)
        if res["error"] ~= nil then
            log.error("CAN'T block for taker_id = %s", tostring(revert_fill.taker_profile_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end

        -- cancel withdrawls
        res = balance.cancel_all_withdrawals(revert_fill.profile_id)
        if res["error"] ~= nil and res["error"] ~= "NO_PENDING_WITHDRAWAL" then
            log.error("CAN'T cancel for maker_id = %s", tostring(revert_fill.profile_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end

        res = balance.cancel_all_withdrawals(revert_fill.taker_profile_id)
        if res["error"] ~= nil and res["error"] ~= "NO_PENDING_WITHDRAWAL" then
            log.error("CAN'T cancel for taker_id = %s", tostring(revert_fill.taker_profile_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end

        -- Start reverting the trade
        box.space.revert_list:update(tup.fill_id, {{'=', "status", config.params.REVERT_STATUS.TX_PROCESSED}})

        local market_id = tup.market_id

        local res = rpc.callrw_engine(market_id, "handle_revert", {tup.maker_fill})
        if res["error"] ~= nil then
            box.space.revert_list:update(tup.fill_id, {
                {'=', "status", config.params.REVERT_STATUS.TX_FAILED},
                {'=', "error", tostring(res["error"])}
            })
            return tostring(res["error"])
        else
            box.space.revert_list:update(tup.fill_id, {
                {'=', "result", res["res"]},
                {'=', "error", ""}
            })
        end
        
        -- activate accounts back
        local res = p.update_status(revert_fill.profile_id, config.params.PROFILE_STATUS.ACTIVE)
        if res["error"] ~= nil then
            log.error("CAN'T activate for maker_id = %s", tostring(revert_fill.profile_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end

        res = p.update_status(revert_fill.taker_profile_id, config.params.PROFILE_STATUS.ACTIVE)
        if res["error"] ~= nil then
            log.error("CAN'T activate for taker_id = %s", tostring(revert_fill.taker_profile_id))
            log.error(tostring(res["error"]))
            return tostring(res["error"])
        end
        
        total = total + 1
    end

    log.info("EXECUTE_SUCCESS: total=%d", total)

    return nil
end


function R.upload_csv(csv_path)
    checks("string")
    local f, err = fio.open(csv_path, {'O_RDONLY'})
    
    if err ~= nil then
        log.error("upload_list file=%s open error = %s", tostring(csv_path), tostring(err))
        return err
    end

    local validate_header = function(header)
        for i, field in ipairs(R.format) do
            if header[i] ~= field.name then
                log.error("INVALID_HEADER")
                log.error("wrong index: %d in header: %s", i, table.concat(header, ","))        
                return false
            end
        end

        return true
    end




    local total = 0
    local skipped = 0
    local load_csv = function(readable, opts)
        opts = opts or {}
        for i, raw in csv.iterate(readable, opts) do
            if i == 1 and not validate_header(raw) then
                return ERR_REVERT_INVALID_HEADER
            elseif i > 1 then
                local e = R.raw_insert_revert_fill(raw, i)
                if e == ERR_REVERT_CANT_BIND then
                    return e
                elseif e == ERR_REVERT_ALREADY_EXIST then
                    skipped = skipped + 1
                else
                    total = total + 1
                end
            end
        end

        return nil
    end

    err = load_csv(f)
    if err ~= nil then
        log.error("load_csv error = %s", tostring(err))
        return err
    end

    log.info("LOADED_SUCCESS: path=%s loaded total=%d skipped=%d", tostring(csv_path), total, skipped)

    return nil
end


return {
    init_spaces = init_spaces,
    drop_spaces = drop_spaces,
    test_set_rpc = _test_set_rpc,
    upload_csv = R.upload_csv,
    execute_revert_list = R.execute_revert_list,
    bind = R.bind
}
