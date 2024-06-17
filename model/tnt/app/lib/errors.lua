local errors = require('errors')
local log = require('log')

local time = require('app.lib.time')

local error_class = getmetatable(errors.new_class('tricky way to get base error class'))
local error_class_new = error_class.new
local error_class_tostring = error_class.tostring

-- override error_class methods

error_class.new = function(self, err, ...)
    if type(err) == 'string' then
        err = table.concat({string.format(err, ...), string.format(':timestamp=%d', time.now())}, ' ')
    end

    local error_object = error_class_new(self, err, ...)
    log.error(error_object.backtrace and error_object:backtrace() or error_object:tostring())

    return error_object
end

error_class.tostring = function(err)
    return err.err
end

error_class.backtrace = function(err)
    return error_class_tostring(err)
end

local export = {}

function export.new_class(class_name, options)
    local err_class = errors.new_class(class_name, options)

    err_class.__instance_mt.__tostring = error_class.tostring
    err_class.__instance_mt.__index.tostring = error_class.tostring
    err_class.__instance_mt.__index.backtrace = error_class.backtrace

    return err_class
end

return export
