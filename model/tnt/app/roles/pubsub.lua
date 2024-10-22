local clock = require "clock"
local fiber = require "fiber"
local log = require "log"
local ffi = require "ffi"
local json = require "json".new()
local argparse = require('cartridge.argparse')

json.cfg {encode_invalid_as_nil = true}

--==================================================================================
-- Embedded indexpiration with some adjustments to stop a worker on non-leader node.
-- Original: https://github.com/moonlibs/indexpiration
--==================================================================================

local M = {}

local function table_clear(t)
    if type(t) ~= "table" then
        error("bad argument #1 to 'clear' (table expected, got " .. (t ~= nil and type(t) or "no value") .. ")", 2)
    end
    local count = #t
    for i = 0, count do
        t[i] = nil
    end
    return
end

local function typeeq(src, ref)
    if ref == "str" then
        return src == "STR" or src == "str" or src == "string"
    elseif ref == "num" then
        return src == "NUM" or src == "num" or src == "number" or src == "unsigned"
    else
        return src == ref
    end
end

local function _mk_keyfunc(index)
    local fun
    if #index.parts == 1 then
        fun = ("return function (t) return t and t[%s] or nil end"):format(index.parts[1].fieldno)
    else
        local rows = {}
        for k = 1, #index.parts do
            table.insert(
                rows,
                ("\tt[%s]==nil and NULL or t[%s],\n"):format(index.parts[k].fieldno, index.parts[k].fieldno)
            )
        end
        fun =
            "local NULL = require'msgpack'.NULL return function(t) return " ..
            "t and {\n" .. table.concat(rows, "") .. "} or nil end\n"
    end
    -- print(fun)
    return dostring(fun)
end

local function _callable(f)
    if type(f) == "function" then
        return true
    else
        local mt = debug.getmetatable(f)
        return mt and mt.__call
    end
end

local F = {}

function F:terminate()
    self._terminate = true
    self:stop_worker()
end

function F:start_watchdog(space_name)
    self._watcher =
        fiber.create(
        function(expiration)
            fiber.name("indexpiration-watchdog")
            local is_running = false
            while not expiration._terminate do
                if (box.info.ro == false or box.info.election.state == "leader") and is_running == false then
                    log.info("Starting indexpiration worker on %q for %q", box.info.listen, space_name)
                    expiration:start_worker()
                    is_running = true
                elseif (box.info.ro == true and box.info.election.state ~= "leader") and is_running == true then
                    log.info("Stopping indexpiration worker on %q for %q", box.info.listen, space_name)
                    expiration:stop_worker()
                    is_running = false
                end
                fiber.sleep(0.01)
            end
        end,
        self
    )
end

function F:stop_worker()
    if not self.running then
        return
    end
    self.running = false
    self._wait:put(true, 0)
end

function F:start_worker()
    if self._terminate or self.running then
        return
    end
    self.running = true
    self._worker =
        fiber.create(
        function(space, expiration, expire_index)
            local fname = space.name .. ".xpr"
            if package.reload then
                fname = fname .. "." .. package.reload.count
            end
            fiber.name(string.sub(fname, 1, 32))
            repeat
                fiber.sleep(0.001)
            until space.expiration
            log.info("Worker started")
            local curwait
            local collect = {}
            while box.space[space.name] and space.expiration == expiration and expiration.running do
                local r, e =
                    pcall(
                    function()
                        -- print("runat loop 2 ",box.time64())
                        local remaining
                        for _, t in expire_index:pairs({0}, {iterator = box.index.GT}) do
                            -- print("checking ",t)
                            local delta = expiration.check(t)

                            if delta <= 0 then
                                table.insert(collect, t)
                            else
                                remaining = delta
                                break
                            end

                            if #collect >= expiration.batch_size then
                                remaining = 0
                                break
                            end
                        end

                        if next(collect) then
                            -- print("batch collected", #collect)
                            if expiration.txn then
                                box.begin()
                            end
                            for _, t in pairs(collect) do
                                if not expiration.txn then
                                    if not expiration.running then
                                        remaining = 0
                                        break
                                    end
                                    t = box.space[space.name]:get(expiration._pk(t))
                                    if expiration.check(t) > 0 then
                                        t = nil
                                    end
                                end
                                if t then
                                    if expiration.on_delete then
                                        expiration.on_delete(t)
                                    end
                                    space:delete(expiration._pk(t))
                                end
                            end
                            if expiration.txn then
                                box.commit()
                            end
                        -- print("batch deleted")
                        end

                        if remaining then
                            if remaining >= 0 and remaining < 1 then
                                return remaining
                            end
                        end
                        return 1
                    end
                )

                table_clear(collect)

                if r then
                    curwait = e
                else
                    curwait = 1
                    log.error("Worker/ERR: %s", e)
                end
                -- log.info("Wait %0.2fs",curwait)
                if curwait == 0 then
                    fiber.sleep(0)
                end
                expiration._wait:get(curwait)
            end
            if expiration.running then
                log.info("Worker finished")
            else
                log.info("Worker stopped")
            end
        end,
        box.space[self.space],
        self,
        self.expire_index
    )
end

function M.upgrade(space, opts, depth)
    depth = depth or 0
    log.info("Indexpiration upgrade(%s,%s)", space.name, json.encode(opts))
    if not opts.field then
        error("opts.field required", 2)
    end

    local self = setmetatable({}, {__index = F})
    if space.expiration then
        self._wait = space.expiration._wait
        self._stat = space.expiration._stat
    else
        self._wait = fiber.channel(0)
    end
    self.debug = not (not opts.debug)
    if opts.txn ~= nil then
        self.txn = not (not opts.txn)
    else
        self.txn = true
    end
    self.on_delete = opts.on_delete
    self.precise = not (not self.precise)
    self.batch_size = opts.batch_size or 300

    local format_av = box.space._space.index.name:get(space.name)[7]
    local format = {}
    local have_format = false
    for no, f in pairs(format_av) do
        format[f.name] = {
            name = f.name,
            type = f.type,
            no = no
        }
        format[no] = format[f.name]
        have_format = true
    end
    for _, idx in pairs(space.index) do
        for _, part in pairs(idx.parts) do
            format[part.fieldno] = format[part.fieldno] or {no = part.fieldno}
            format[part.fieldno].type = part.type
        end
    end

    local expire_field_no
    if have_format then
        expire_field_no = format[opts.field].no
    else
        expire_field_no = opts.field
    end
    if type(expire_field_no) ~= "number" then
        error("Need correct `.field` option", 2 + depth)
    end

    -- 2. index check
    local expire_index
    local has_non_tree
    for _, index in pairs(space.index) do
        if index.parts[1].fieldno == expire_field_no then
            if index.type == "TREE" then
                expire_index = index
                break
            else
                has_non_tree = index
            end
        end
    end

    if not expire_index then
        if has_non_tree then
            error(string.format("index %s must be TREE (or create another)", has_non_tree.name), 2 + depth)
        else
            error(string.format("field %s requires tree index with it as first field", opts.field), 2 + depth)
        end
    end

    self._pk = _mk_keyfunc(space.index[0])

    -- if not self._stat then
    -- 	self._stat = {
    -- 		counts = {};
    -- 	}
    -- end

    if opts.kind == "time" or opts.kind == "time64" then
        if not typeeq(expire_index.parts[1].type, "num") then
            error(("Can't use field %s as %s"):format(opts.field, opts.kind), 2 + depth)
        end
        if opts.kind == "time" then
            self.check = function(t)
                return t[expire_field_no] - clock.realtime()
            end
        elseif opts.kind == "time64" then
            self.check = function(t)
                return tonumber(ffi.cast("int64_t", t[expire_field_no]) - ffi.cast("int64_t", clock.realtime64())) / 1e9
            end
        end
    elseif _callable(opts.kind) then
        self.check = opts.kind
    else
        error(("Unsupported kind: %s"):format(opts.kind), 2 + depth)
    end

    self.space = space.id
    self.expire_index = expire_index

    self._terminate = false
    self.running = false
    self:start_watchdog(space.name)

    if self.precise then
        self._on_repl =
            space:on_replace(
            function(old, new)
                self._wait:put(true, 0)
            end,
            self._on_repl
        )
    end
    rawset(space, "expiration", self)

    while self._wait:has_readers() do
        self._wait:put(true, 0)
    end

    log.info("Upgraded %s into indexpiration (status=%s)", space.name, box.info.status)
end

local function indexpiration(space, opts)
    M.upgrade(space, opts, 1)
end

--===================================================================
-- Centrifuge Tarantool module.
-- Provides PUB/SUB, ephemeral history streams and channel presence.
--===================================================================

local centrifuge = {}

local function on_disconnect()
    local id = box.session.storage.subscriber_id
    if id == nil then
        return
    end

    local channelsById = centrifuge.id_to_channels[id] or {}

    for key, _ in pairs(channelsById) do
        centrifuge.channel_to_ids[key][id] = nil
        if next(centrifuge.channel_to_ids[key]) == nil then
            centrifuge.channel_to_ids[key] = nil
        end

        channelsById[key] = nil
    end

    centrifuge.id_to_channels[id] = nil
    centrifuge.id_to_fiber[id]:close()
    centrifuge.id_to_fiber[id] = nil
    centrifuge.id_to_messages[id] = nil
end

centrifuge.init_spaces = function(opts)
    if not opts then
        opts = {}
    end

 
    local pubs_sync = opts.pubs_sync or false
    local meta_sync = opts.meta_sync or false

    local pubs_opts = opts.pubs_opts or {if_not_exists = true}
    if pubs_sync == true then
        pubs_opts.is_sync = true
    end
    box.schema.create_space("pubs", pubs_opts)
    box.space.pubs:format(
        {
            {name = "id", type = "unsigned"},
            {name = "channel", type = "string"},
            {name = "offset", type = "unsigned"},
            {name = "exp", type = "number"},
            {name = "data", type = "string"},
        }
    )
    box.space.pubs:create_index(
        "primary",
        {
            parts = {{field = "id", type = "unsigned"}},
            if_not_exists = true
        }
    )
    box.space.pubs:create_index(
        "channel",
        {
            parts = {{field = "channel", type = "string"}, {field = "offset", type = "unsigned"}},
            if_not_exists = true
        }
    )
    box.space.pubs:create_index(
        "exp",
        {
            parts = {{field = "exp", type = "number"}, {field = "id", type = "unsigned"}},
            if_not_exists = true
        }
    )

    local meta_opts = opts.meta_opts or {if_not_exists = true}
    if meta_sync == true then
        meta_opts.is_sync = true
    end
    box.schema.create_space("meta", meta_opts)
    box.space.meta:format(
        {
            {name = "channel", type = "string"},
            {name = "offset", type = "unsigned"},
            {name = "epoch", type = "string"},
            {name = "exp", type = "number"}
        }
    )
    box.space.meta:create_index(
        "primary",
        {
            parts = {{field = "channel", type = "string"}},
            if_not_exists = true
        }
    )
    box.space.meta:create_index(
        "exp",
        {
            parts = {{field = "exp", type = "number"}, {field = "channel", type = "string"}},
            if_not_exists = true
        }
    )

    box.schema.create_space("presence", {if_not_exists = true, temporary = true})
    box.space.presence:format(
        {
            {name = "channel", type = "string"},
            {name = "client_id", type = "string"},
            {name = "user_id", type = "string"},
            {name = "data", type = "string"},
            {name = "exp", type = "number"}
        }
    )
    box.space.presence:create_index(
        "primary",
        {
            parts = {{field = "channel", type = "string"}, {field = "client_id", type = "string"}},
            if_not_exists = true
        }
    )
    box.space.presence:create_index(
        "exp",
        {
            parts = {{field = "exp", type = "number"}},
            if_not_exists = true
        }
    )
end

centrifuge.start = function(opts)
    if not opts then
        opts = {}
    end

    indexpiration(
        box.space.pubs,
        {
            field = "exp",
            kind = "time",
            precise = true
        }
    )

    indexpiration(
        box.space.meta,
        {
            field = "exp",
            kind = "time",
            precise = true
        }
    )

    indexpiration(
        box.space.presence,
        {
            field = "exp",
            kind = "time",
            precise = true
        }
    )

    if not rawget(_G, "__centrifuge_cleanup_set") then
        box.session.on_disconnect(on_disconnect)
        rawset(_G, "__centrifuge_cleanup_set", true)
    end
end

centrifuge.id_to_channels = {}
centrifuge.channel_to_ids = {}
centrifuge.id_to_messages = {}
centrifuge.id_to_fiber = {}

function centrifuge.get_messages(id, use_polling, timeout)
    timeout = timeout or 0

    if not box.session.storage.subscriber_id then
        -- register poller connection. Connection will use this id
        -- to register or remove subscriptions.
        box.session.storage.subscriber_id = id
        centrifuge.id_to_fiber[id] = fiber.channel()
        return
    end
    assert(box.session.storage.subscriber_id == id)

    local now = fiber.time()
    while true do
        -- TODO recheck unsubscribe
        local messages = centrifuge.id_to_messages[id]
        centrifuge.id_to_messages[id] = nil
        if messages then
            if use_polling then
                -- write all messages to connection and return.
                return messages
            else
                local ok = box.session.push(messages)
                if ok ~= true then
                    error("Write error")
                end
            end
        else
            local left = (now + timeout) - fiber.time()
            if left <= 0 then
                -- timed out, client poller will call get_messages again.
                return
            end
            centrifuge.id_to_fiber[id]:get(left)
        end
    end
end

function centrifuge.subscribe(id, channels)
    for _, v in ipairs(channels) do
        local idChannels = centrifuge.id_to_channels[id] or {}
        idChannels[v] = true
        centrifuge.id_to_channels[id] = idChannels

        local channelIds = centrifuge.channel_to_ids[v] or {}
        channelIds[id] = true
        centrifuge.channel_to_ids[v] = channelIds
    end
end

function centrifuge.unsubscribe(id, channels)
    for _, v in ipairs(channels) do
        if centrifuge.id_to_channels[id] then
            centrifuge.id_to_channels[id][v] = nil
        end
        if centrifuge.channel_to_ids[v] then
            centrifuge.channel_to_ids[v][id] = nil
        end
        if centrifuge.id_to_channels[id] and next(centrifuge.id_to_channels[id]) == nil then
            centrifuge.id_to_channels[id] = nil
        end
        if centrifuge.channel_to_ids[v] and next(centrifuge.channel_to_ids[v]) == nil then
            centrifuge.channel_to_ids[v] = nil
        end
    end
end

local function publish_to_subscribers(channel, message_tuple)
    local channelIds = centrifuge.channel_to_ids[channel] or {}

    for k, _ in pairs(channelIds) do
        centrifuge.id_to_messages[k] = centrifuge.id_to_messages[k] or {}
        table.insert(centrifuge.id_to_messages[k], message_tuple)
    end
end

local function wake_up_subscribers(channel)
    local ids = centrifuge.channel_to_ids[channel] or {}
    for k, _ in pairs(ids) do
        local chan = centrifuge.id_to_fiber[k]
        if chan:has_readers() then
            chan:put(true, 0)
        end
    end
end

function centrifuge._publish(msg_type, channel, data, ttl, size, meta_ttl)
    local epoch = ""
    local offset = 0

    if ttl > 0 and size > 0 then
        local now = clock.realtime()
        local meta_exp = 0
        if meta_ttl > 0 then
            meta_exp = now + meta_ttl
        end
        local stream_meta = box.space.meta:get(channel)
        if stream_meta then
            offset = stream_meta[2] + 1
            epoch = stream_meta[3]
        else
            epoch = tostring(now)
            offset = 1
        end
        -- Need to use field numbers to work with Tarantool 1.10, otherwise we could write:
        -- box.space.meta:upsert({channel, offset, epoch, meta_exp}, {{'=', 'channel', channel}, {'+', 'offset', 1}, {'=', 'exp', meta_exp}})
        box.space.meta:upsert({channel, offset, epoch, meta_exp}, {{"=", 1, channel}, {"+", 2, 1}, {"=", 4, meta_exp}})
        box.space.pubs:auto_increment {channel, offset, clock.realtime() + tonumber(ttl), data}
        local max_offset_to_keep = offset - size
        if max_offset_to_keep > 0 then
            for _, v in box.space.pubs.index.channel:pairs({channel, max_offset_to_keep}, {iterator = box.index.LE}) do
                box.space.pubs:delete {v.id}
            end
        end
    end
    -- raise
    publish_to_subscribers(channel, {msg_type, channel, offset, epoch, data})
    -- raise
    wake_up_subscribers(channel)

    return {epoch = epoch, offset = offset}
end

function centrifuge.publish(msg_type, channel, data, ttl, size, meta_ttl)
    if not ttl then
        ttl = 0
    end
    if not size then
        size = 0
    end
    if not meta_ttl then
        meta_ttl = 0
    end

    box.begin()
    local rc, res = pcall(centrifuge._publish, msg_type, channel, data, ttl, size, meta_ttl)
    if not rc then
        log.warn(
            "Publish error: %q %q %q %q %q %q %q",
            tostring(res),
            msg_type,
            channel,
            data,
            ttl,
            size,
            meta_ttl
        )
        box.rollback()
        error("Publish error: " .. tostring(res))
    end
    box.commit()
    return res.offset, res.epoch
end

function centrifuge._history(channel, since_offset, limit, reverse, include_pubs, meta_ttl)
    if not meta_ttl then
        meta_ttl = 0
    end

    local meta_exp = 0
    local now = clock.realtime()
    if meta_ttl > 0 then
        meta_exp = now + meta_ttl
    end
    local epoch = tostring(now)
    -- Need to use field numbers to work with Tarantool 1.10, otherwise we could write:
    -- box.space.meta:upsert({channel, 0, epoch, meta_exp}, {{'=', 'channel', channel}, {'=', 'exp', meta_exp}})
    box.space.meta:upsert({channel, 0, epoch, meta_exp}, {{"=", 1, channel}, {"=", 4, meta_exp}})
    local stream_meta = box.space.meta:get(channel)

    if not include_pubs then
        return {stream_meta["offset"], stream_meta["epoch"], nil}
    end
    if reverse == false and stream_meta["offset"] == since_offset - 1 then
        return {stream_meta["offset"], stream_meta["epoch"], nil}
    end
    local num_entries = 0

    local get_offset = since_offset
    local iterator = box.index.GE
    if reverse == true then
        iterator = box.index.LE
        if since_offset == 0 then
            get_offset = stream_meta["offset"]
        end
    end
    local pubs =
        box.space.pubs.index.channel:pairs({channel, get_offset}, {iterator = iterator}):take_while(
        function(x)
            num_entries = num_entries + 1
            return x.channel == channel and (limit < 1 or num_entries < limit + 1)
        end
    ):totable()
    return {stream_meta["offset"], stream_meta["epoch"], pubs}
end

function centrifuge.history(channel, since_offset, limit, reverse, include_pubs, meta_ttl)
    if channel == nil then
        error("No channel specified")
    end
    box.begin()
    local rc, res = pcall(centrifuge._history, channel, since_offset, limit, reverse, include_pubs, meta_ttl)
    if not rc then
        log.warn("History error %q %q %q %q %q %q", channel, since_offset, limit, reverse, include_pubs, meta_ttl)
        box.rollback()
        error("History error: " .. tostring(res))
    end

    box.commit()
    return res[1], res[2], res[3]
end

function centrifuge.remove_history(channel)
    local batch_size = 10000
    box.begin()
    for _, v in box.space.pubs.index.channel:pairs {channel} do
        box.space.pubs:delete {v.id}

        batch_size = batch_size - 1
        if batch_size <= 0 then
            batch_size = 10000

            box.commit()
            fiber.yield()
            box.begin()
        end
    end
    box.commit()
end

function centrifuge.add_presence(channel, ttl, client_id, user_id, data)
    if not ttl then
        ttl = 0
    end
    local exp = clock.realtime() + ttl
    box.space.presence:put({channel, client_id, user_id, data, exp})
end

function centrifuge.remove_presence(channel, client_id)
    box.space.presence:delete {channel, client_id}
end

function centrifuge.presence(channel)
    if channel == nil then
        error("No specified channel")
    end

    return box.space.presence:select {channel}
end

function centrifuge.presence_stats(channel)
    if channel == nil then
        error("No specified channel")
    end

    local users = {}
    local num_clients = 0
    local num_users = 0
    for _, v in box.space.presence:pairs({channel}, {iterator = box.index.EQ}) do
        num_clients = num_clients + 1
        if not users[v.user_id] then
            num_users = num_users + 1
            users[v.user_id] = true
        end
    end
    return num_clients, num_users
end

function centrifuge.stop()
    box.session.on_disconnect(on_disconnect, nil)
    rawset(_G, "__centrifuge_cleanup_set", nil)
end

function centrifuge.publish_data(channel, json_data, ttl, size, meta_ttl)
    centrifuge.publish("p", channel, json_data, ttl, size, meta_ttl)

    return {res = "success", error = nil}
end

function centrifuge.publish_data_batch(batch, ttl, size, meta_ttl)
    for channel, json_data in pairs(batch) do
        centrifuge.publish("p", channel, json_data, ttl, size, meta_ttl)
    end

    return {res = "success", error = nil}
end

local function stop()
    centrifuge.stop()
    rawset(_G, "centrifuge", nil)
    rawset(_G, "publish_data", nil)
    rawset(_G, "publish_data_batch", nil)

    return true
end

local function init(opts) -- luacheck: no unused args
    if opts.is_master then
        opts.meta_opts = {if_not_exists=true, temporary = true}
        opts.pubs_opts = {if_not_exists=true, temporary = true}    
        
        centrifuge.init_spaces(opts)

        centrifuge.start(opts)

        box.schema.func.create('centrifuge', {if_not_exists = true})
        box.schema.func.create('publish_data', {if_not_exists = true})

    end
    rawset(_G, "centrifuge", centrifuge)
    rawset(_G, "publish_data", centrifuge.publish_data)
    rawset(_G, "publish_data_batch", centrifuge.publish_data_batch)

    return true
end



return {
    role_name = 'pubsub',
    init = init,
    stop = stop,
    utils = {
        centrifuge = centrifuge
    },

    publish_data = centrifuge.publish_data,
    publish_data_batch = centrifuge.publish_data_batch,
}