local checks = require('checks')
local log = require('log')

local archiver = require('app.archiver')
local config = require('app.config')
local errors = require('app.lib.errors')
local time = require('app.lib.time')
local tuple = require('app.tuple')
local util = require('app.util')

require('app.errcodes')

local OrderError = errors.new_class("ORDER")

local YIELD_THRESHOLD = 1000 -- tarantool recommended value

local stop_loss_order = {}

function stop_loss_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function stop_loss_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        -- clear unused
        amend.price = nil
        amend.size = nil
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
        amend.size_percent = nil
    end

    return nil
end

local take_profit_order = {}

function take_profit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function take_profit_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        -- clear unused
        amend.price = nil
        amend.size = nil
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
        amend.size_percent = nil
    end

    return nil
end

local stop_loss_limit_order = {}

function stop_loss_limit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function stop_loss_limit_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        -- clear unused
        amend.size = nil
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
        amend.size_percent = nil
    end

    return nil
end

local take_profit_limit_order = {}

function take_profit_limit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function take_profit_limit_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        -- clear unused
        amend.size = nil
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
        amend.size_percent = nil
    end

    return nil
end

local stop_market_order = {}

function stop_market_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function stop_market_order.amend(amend)
    checks('table')

    -- clear unused
    amend.size_percent = nil

    if amend.status == config.params.ORDER_STATUS.PLACED then
        -- clear unused
        amend.price = nil
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
    end

    return nil
end

local stop_limit_order = {}

function stop_limit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.PLACED

    return nil
end

function stop_limit_order.amend(amend)
    checks('table')

    -- clear unused
    amend.size_percent = nil

    if amend.status == config.params.ORDER_STATUS.PLACED then
    else
        amend.status = config.params.ORDER_STATUS.OPEN

        -- clear unused
        amend.trigger_price = nil
    end

    return nil
end

local market_order = {}

function market_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.OPEN

    return nil
end

function market_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        log.error(OrderError:new('bad order status: order_id=%s, order_type=%s, order_status=%s', order.id, order.order_type, order.status))
        return ERR_WRONG_ORDER_STATUS
    end

    -- clear unused
    amend.size_percent = nil
    amend.trigger_price = nil

    amend.status = config.params.ORDER_STATUS.OPEN
    return nil
end

local limit_order = {}

function limit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.OPEN

    return nil
end

function limit_order.amend(amend)
    checks('table')

    if amend.status == config.params.ORDER_STATUS.PLACED then
        log.error(OrderError:new('bad order status: order_id=%s, order_type=%s, order_status=%s', order.id, order.order_type, order.status))
        return ERR_WRONG_ORDER_STATUS
    end

    -- clear unused
    amend.size_percent = nil
    amend.trigger_price = nil

    amend.status = config.params.ORDER_STATUS.OPEN
    return nil
end


local ping_limit_order = {}

function ping_limit_order.create(order)
    checks('table|engine_order')
    order.status = config.params.ORDER_STATUS.OPEN

    return nil
end

function ping_limit_order.amend(amend)
    checks('table')

    return ERR_ACTION_NOT_AVAILABLE
end


local O = {
    format = {
        {name = 'id', type = 'string'},
        {name = 'profile_id', type = 'unsigned'},
        {name = 'market_id', type = 'string'},
        {name = 'order_type', type = 'string'},
        {name = 'status', type = 'string'},
        {name = 'price', type = 'decimal'},
        {name = 'size', type = 'decimal'},
        {name = 'initial_size', type = 'decimal'},
        {name = 'total_filled_size', type = 'decimal'},
        {name = 'side', type = 'string'},
        {name = 'timestamp', type = 'number'},
        {name = 'reason', type = 'string'},
        {name = 'client_order_id', type = 'string'},
        {name = 'trigger_price', type = 'decimal'},
        {name = 'size_percent', type = 'decimal'},
        {name = 'time_in_force', type = 'string'},
        {name = 'created_at', type = 'number'},
        {name = 'updated_at', type = 'number'},
    },
    strict_type = 'engine_order',
    _metatypes = {
        limit = limit_order,
        market = market_order,
        stop_loss = stop_loss_order,
        take_profit = take_profit_order,
        stop_loss_limit = stop_loss_limit_order,
        take_profit_limit = take_profit_limit_order,
        stop_limit = stop_limit_order,
        stop_market = stop_market_order,
        ping_limit = ping_limit_order,
    },
}

-- func should create tuple with exact described O.format
local function new_order(
    order_id,
    profile_id,
    market_id,
    order_type,
    status,
    price,
    size,
    initial_size,
    total_filled_size,
    side,
    timestamp,
    reason,
    client_order_id,
    trigger_price,
    size_percent,
    time_in_force,
    created_at,
    updated_at
)
    checks('string', 'number', 'string', 'string', 'string', 'decimal', 'decimal', 'decimal', 'decimal',
     'string', 'number', 'string', 'string', 'decimal', 'decimal', 'string', 'number', 'number')

    return {
        order_id,
        profile_id,
        market_id,
        order_type,
        status,
        price,
        size,
        initial_size,
        total_filled_size,
        side,
        timestamp,
        reason,
        client_order_id,
        trigger_price,
        size_percent,
        time_in_force,
        created_at,
        updated_at,
    }
end

function O.init_spaces()
    local order, err = archiver.create('order', {if_not_exists = true}, O.format, {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true,
    })
    if err ~= nil then
        log.error(OrderError:new(err))
        error(err)
    end

    order:create_index('profile_id', {
        unique = false,
        parts = {{field = 'profile_id'}},
        if_not_exists = true,
    })
    order:create_index('profile_market', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'market_id'}},
        if_not_exists = true,
    })
    order:create_index('status_type', {
        unique = false,
        parts = {{field = 'status'}, {field = 'order_type'}},
        if_not_exists = true,
    })
    order:create_index('profile_status_type', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'status'}, {field = 'order_type'}},
        if_not_exists = true,
    })
    order:create_index('client_order_id', {
        unique = false,
        parts = {{field = 'client_order_id'}},
        if_not_exists = true,
    })
    order:create_index('coid_status', {
        unique = false,
        parts = {{field = 'client_order_id'}, {field = 'status'}},
        if_not_exists = true,
    })
    order:create_index('profile_coid_status', {
        unique = false,
        parts = {{field = 'profile_id'}, {field = 'client_order_id'}, {field = "status"}},
        if_not_exists = true,
    })
end

function O.new(...)
    return O.bind(new_order(...))
end

function O.bind(value)
    return tuple.new(value, O.format, O.strict_type)
end

-- tuple behaviour
function O.tomap(self, opts)
    return self:tomap(opts)
end

function O.create(
    order_id,
    profile_id,
    market_id,
    order_type,
    price,
    size,
    initial_size,
    side,
    client_order_id,
    trigger_price,
    size_percent,
    time_in_force,
    is_liquidation
)
    checks('string', 'number', 'string', 'string', '?decimal', '?decimal', '?decimal',
        'string', '?string', '?decimal', '?decimal', 'string', 'boolean')

    local reason = is_liquidation and 'liquidation' or ''
    local timestamp = time.now()

    -- our current phylosophy to store not nil values only because of lua/tarantool nil specifics
    -- so fill nil values with default ones
    client_order_id = util.return_not_nil(client_order_id, '')
    price = util.return_not_nil(price, ZERO)
    size = util.return_not_nil(size, ZERO)
    initial_size = util.return_not_nil(initial_size, ZERO)
    trigger_price = util.return_not_nil(trigger_price, ZERO)
    size_percent = util.return_not_nil(size_percent, ZERO)
    time_in_force = util.return_not_nil(time_in_force, config.params.TIME_IN_FORCE.GTC)

    local total_filled_size = initial_size - size

    local order = O.bind(
        new_order(
            order_id,
            profile_id,
            market_id,
            order_type,
            config.params.ORDER_STATUS.UNKNOWN,
            price,
            size,
            initial_size,
            total_filled_size,
            side,
            timestamp,
            reason,
            client_order_id,
            trigger_price,
            size_percent,
            time_in_force,
            timestamp,
            timestamp
        )
    )

    local order_mtype = O._metatypes[order.order_type]
    if order_mtype == nil then
        log.error(OrderError:new("order handler not found: id=%s, type=%s", order.id, order.order_type))
        return nil, ERR_WRONG_ORDER_TYPE
    end

    local err = order_mtype.create(order)
    if err ~= nil then
        log.error(OrderError:new(err))
        return nil, err
    end

    local _, err = archiver.insert(box.space.order, order)
    if err ~= nil then
        log.error(OrderError:new(err))
        return nil, err
    end

    return order, nil
end

function O.update(order_id, price, size)
    local exist = box.space.order:get(order_id)
    if exist == nil then
        local err = "not found order_id=" .. tostring(order_id)
        return nil, err
    end

    local total_filled_size = exist.total_filled_size + (exist.size - size)

    local status = config.params.ORDER_STATUS.OPEN
    if size == 0 then
        status = config.params.ORDER_STATUS.CLOSED
    end

    local timestamp = time.now()

    local res, err = archiver.update(box.space.order, order_id, {
        {'=', 'status', status},
        {'=', 'size', size},
        {'=', 'price', price},
        {'=', 'total_filled_size', total_filled_size},
        {'=', 'price', price},
        {'=', 'updated_at', timestamp},
    })
    if err ~= nil then
        return nil, err
    end

    return O.bind(res), nil
end

function O.open(order_id, price, size)
    checks('string', '?decimal', '?decimal')
    local timestamp = time.now()

    local ops = {
        {'=', 'status', config.params.ORDER_STATUS.OPEN},
        {'=', 'updated_at', timestamp},
    }
    if price ~= nil then
        table.insert(ops, {'=', 'price', price})
    end
    if size ~= nil then
        table.insert(ops, {'=', 'size', size})
        table.insert(ops, {'=', 'initial_size', size})
    end

    local res, err = archiver.update(box.space.order, order_id, ops)
    if err ~= nil then
        return nil, err
    end

    return O.bind(res), nil
end

function O.cancel(order_id)
    checks('string')
    local timestamp = time.now()

    local res, err = archiver.update(box.space.order, order_id, {
        {'=', 'status', config.params.ORDER_STATUS.CANCELED},
        {'=', 'updated_at', timestamp},
    })
    if err ~= nil then
        return nil, err
    end

    return O.bind(res), nil
end

function O.reject(order_id, reason)
    checks('string', 'string')
    local timestamp = time.now()

    local res, err = archiver.update(box.space.order, order_id, {
        {'=', 'status', config.params.ORDER_STATUS.REJECTED},
        {'=', 'reason', reason},
        {'=', 'updated_at', timestamp},
    })
    if err ~= nil then
        return nil, err
    end

    return O.bind(res), nil
end

function O.amend(order_id, new_price, new_size, new_trigger_price, new_size_percent)
    checks('string', '?decimal', '?decimal', '?decimal', '?decimal')

    local order = box.space.order:get(order_id)
    if order == nil then
        log.error(OrderError:new("order not found: order_id=", order_id))
        return nil, ERR_ORDER_NOT_FOUND
    end

    local order_mtype = O._metatypes[order.order_type]
    if order_mtype == nil then
        log.error(OrderError:new("order handler not found: order_id=%s, order_type=%s", order.id, order.order_type))
        return nil, ERR_WRONG_ORDER_TYPE
    end

    local amend = { --TODO: add type
        status = order.status,
        price = new_price,
        size = new_size,
        trigger_price = new_trigger_price,
        size_percent = new_size_percent,
    }
    local err = order_mtype.amend(amend)
    if err ~= nil then
        log.error(OrderError:new(err))
        return nil, err
    end

    local timestamp = time.now()
    local ops = {
        {'=', 'status', amend.status},
        {'=', 'updated_at', timestamp},
    }
    if amend.price ~= nil then
        table.insert(ops, {'=', 'price', amend.price})
    end
    if amend.trigger_price ~= nil then
        table.insert(ops, {'=', 'trigger_price', amend.trigger_price})
    end
    if amend.size ~= nil then
        table.insert(ops, {'=', 'size', amend.size})
        table.insert(ops, {'=', 'initial_size', amend.size})
    end
    if amend.size_percent ~= nil then
        table.insert(ops, {'=', 'size_percent', amend.size_percent})
    end

    local res, err = archiver.update(box.space.order, order_id, ops)
    if err ~= nil then
        log.error(OrderError:new(err))
        return nil, err
    end
    if res == nil then
        log.error(OrderError:new('order not found: order_id=%s', order_id))
        return nil, ERR_ORDER_NOT_FOUND
    end

    return O.bind(res), nil
end

function O.get_orders(profile_id, statuses, order_type, limit)
    checks('number', 'string|table', '?string', '?number')
    limit = limit or 40 --FIXME: why such const value?

    if type(statuses) == 'string' then
        statuses = {statuses}
    end

    local idx = box.space.order.index.profile_status_type
    local key = {profile_id, 'status', order_type}

    local orders = {}
    for _, status in ipairs(statuses) do
        key[2] = status
        for _, row in idx:pairs(key, {iterator = box.index.EQ}):take_n(limit) do
            table.insert(orders, O.bind(row))
        end
        if limit == #orders then
            break
        end
        limit = limit - #orders
    end

    return {res = orders, error = nil}
end

function O.get_order_by_id(order_id)
    checks('string')

    local order = box.space.order:get(order_id)
    if order == nil then
        return {res = nil, error = ERR_ORDER_NOT_FOUND}
    end

    return {res = O.bind(order), error = nil}
end

function O.get_order_by_client_id(order_id, status)
    checks('string', '?string')
    status = status or config.params.ORDER_STATUS.OPEN

    local order = box.space.order.index.coid_status:min{order_id, status}
    if order == nil then
        return {res = nil, error = ERR_ORDER_NOT_FOUND}
    end

    return {res = O.bind(order), error = nil}
end

function O.get_order_by_oneof(order_id, client_order_id)
    checks('?string', '?string')

    if order_id == nil or order_id == '' then
        if client_order_id == nil or client_order_id == '' then
            return {res = nil, error = OrderError:new('%s: neither order_id nor client_order_id specified', ERR_ORDER_NOT_FOUND)}
        end

        return O.get_order_by_client_id(client_order_id)
    end

    return O.get_order_by_id(order_id)
end

function O.check_coid_for_reuse(profile_id, client_order_id)
    local res = client_order_id
    local err = nil

    -- go through all orders with 
    for _, po in box.space.order.index.profile_coid_status:pairs({profile_id, client_order_id, config.params.ORDER_STATUS.OPEN}, {iterator = "EQ"}) do
        res = nil
        err = "OPEN_FOUND"
        break
    end

    return {res = res, error = err}
end


function O.get_all_orders(profile_id, limit)
    checks('number', 'number')

    local idx = box.space.order.index.profile_id
    local orders = {}
    for _, row in idx:pairs(profile_id, {iterator = box.index.EQ}):take_n(limit) do
        table.insert(orders, O.bind(row))
    end

    return {res = orders, error = nil}
end

function O.iterator_by(profile_id, order_status, order_type)
    checks('?number', '?string', '?string')

    local idx = box.space.order
    local itr = box.index.ALL
    local key = {}

    if profile_id ~= nil then
        idx = idx.index.profile_status_type
        itr = box.index.EQ
        table.insert(key, profile_id)
    elseif order_status ~= nil then
        idx = idx.index.status_type
        itr = box.index.EQ
    elseif order_type ~= nil then
        return error('no index found in space `order` for order_type only')
    end

    if order_status ~= nil then
        table.insert(key, order_status)
    end
    if order_type ~= nil then
        table.insert(key, order_type)
    end

    local iter, iter_param, iter_state = idx:pairs(key, {iterator = itr})

    return function(param, state)
        local state, value = iter(param, state)
        return state, value and O.bind(value)
    end,
    iter_param,
    iter_state
end

function O.get_all_orders2(profile_id, order_status, order_type)
    checks('?number', '?string', '?string')

    local orders = {}
    local count = 0
    for _, row in O.iterator_by(profile_id, order_status, order_type) do
        table.insert(orders, O.bind(row))
        count = util.safe_yield(count, YIELD_THRESHOLD)
    end

    return {res = orders, error = nil}
end

function O.count_orders(profile_id, status, order_type)
    checks('number', 'string', 'string')
    return box.space.order.index.profile_status_type:count({profile_id, status, order_type}, {iterator = box.index.EQ})
end

local sltp_orders_statuses = {config.params.ORDER_STATUS.PLACED, config.params.ORDER_STATUS.OPEN}
local stop_loss_orders_types = {config.params.ORDER_TYPE.STOP_LOSS, config.params.ORDER_TYPE.STOP_LOSS_LIMIT}
function O.is_new_stop_loss_allowed(profile_id)
    checks('number')

    local cp = config.params
    local cnt = 0

    for _, order_status in ipairs(sltp_orders_statuses) do
        for _, order_type in ipairs(stop_loss_orders_types) do
            cnt = cnt + O.count_orders(profile_id, order_status, order_type)
            if cnt >= cp.MAX_POSITION_DEPENDENT_ORDERS then
                return false
            end
        end
    end

    return true
end

local take_profit_orders_types = {config.params.ORDER_TYPE.TAKE_PROFIT, config.params.ORDER_TYPE.TAKE_PROFIT_LIMIT}
function O.is_new_take_profit_allowed(profile_id)
    checks('number')

    local cp = config.params
    local cnt = 0

    for _, order_status in ipairs(sltp_orders_statuses) do
        for _, order_type in ipairs(take_profit_orders_types) do
            cnt = cnt + O.count_orders(profile_id, order_status, order_type)
            if cnt >= cp.MAX_POSITION_DEPENDENT_ORDERS then
                return false
            end
        end
    end

    return true
end

local stop_orders_types = {config.params.ORDER_TYPE.STOP_LIMIT, config.params.ORDER_TYPE.STOP_MARKET}
function O.is_new_stop_order_allowed(profile_id)
    checks('number')

    local cp = config.params
    local os = cp.ORDER_STATUS
    local cnt = 0

    for _, order_type in ipairs(stop_orders_types) do
        cnt = cnt + O.count_orders(profile_id, os.PLACED, order_type)
        if cnt >= config.params.MAX_CONDITIONAL_ORDERS then
            return false
        end
    end

    return true
end

-- for unit testing

function O._test_set_time(override_time)
    time = override_time
end

return O
