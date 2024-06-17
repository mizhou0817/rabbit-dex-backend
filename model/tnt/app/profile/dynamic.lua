local checks = require('checks')
local decimal = require('decimal')
local log = require('log')

local config = require('app.config')
local time = require('app.lib.time')

require("app.config.constants")


local function init_spaces()
    local update_params = box.schema.space.create('update_params', {temporary=false, if_not_exists = true})
    update_params:format({
        {name = 'id', type = 'string'},
        {name = 'skip_update', type = 'boolean'},
        {name = 'interval', type = 'number'},
    })

    update_params:create_index('primary', {
        unique = true,
        parts = {{field = 'id'}},
        if_not_exists = true })

    local meta_fiber = box.space.update_params:get("meta_fiber")
    if meta_fiber == nil then
        box.space.update_params:replace({"meta_fiber", false, 60})
    end

end

local function get_config(id)
    checks('string')

    return box.space.update_params:get(id)
end

return {
    init_spaces = init_spaces,
    get_config = get_config
}
