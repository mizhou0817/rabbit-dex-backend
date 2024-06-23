local checks = require('checks')
local ffi = require('ffi')

local crbx = ffi.load('rbx')
ffi.cdef[[
/* should be the same as ffi::COptions in rbx/src/pubsub/mod.rs */
typedef struct {
    const char* address; // host:port
    const char* rt_name;
    uint32_t rt_workers;
} ffi_pubsub_options_t;

typedef void* pubsub_app_t;

pubsub_app_t ffi_pubsub_start(ffi_pubsub_options_t* opts);
void ffi_pubsub_stop(pubsub_app_t app);
int ffi_pubsub_publish(pubsub_app_t app, const char* channel, const char* data, size_t data_len);
]]

local function pubsub_start(opts)
    checks({address = 'string'})

    local ffi_opts = ffi.new('ffi_pubsub_options_t[1]', {{
        address = ffi.cast('uint8_t*', opts.address),
    }})
    local app = crbx.ffi_pubsub_start(ffi_opts)
    if app == box.NULL then
        return nil, box.error.last()
    end

    return app
end

local function pubsub_stop(app)
    checks('cdata')

    crbx.ffi_pubsub_stop(app)
end

local function pubsub_publish(app, channel, data)
    checks('cdata', 'string', 'string')

    local err = crbx.ffi_pubsub_publish(
        ffi.cast('pubsub_app_t', app),
        ffi.cast('uint8_t*', channel),
        ffi.cast('const char*', data),
        ffi.cast('size_t', string.len(data))
    )
    if err < 0 then
        return box.error.last()
    end
end

return {
    start = pubsub_start,
    stop = pubsub_stop,
    publish = pubsub_publish,
}
