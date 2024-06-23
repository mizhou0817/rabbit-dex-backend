local ffi = require('ffi')

local crbx = ffi.load('rbx')
ffi.cdef[[
void ffi_logger_init();
]]

local function logger_init()
    crbx.ffi_logger_init()
end

return {
    init = logger_init,
}
