local fiber = require('fiber')

local time = {}

function time.now()
    return tonumber(fiber.time64())
end

function time.now_sec()
    return math.ceil(fiber.time())
end

function time.now_milli()
    local micro = time.now()
    return tonumber(math.floor(micro / 1000))
end


return time