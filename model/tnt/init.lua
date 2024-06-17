#!/usr/bin/env tarantool

require('strict').on()

-- configure path so that you can run application
-- from outside the root directory
if package.setsearchroot ~= nil then
    package.setsearchroot()
else
    -- Workaround for rocks loading in tarantool 1.10
    -- It can be removed in tarantool > 2.2
    -- By default, when you do require('mymodule'), tarantool looks into
    -- the current working directory and whatever is specified in
    -- package.path and package.cpath. If you run your app while in the
    -- root directory of that app, everything goes fine, but if you try to
    -- start your app with "tarantool myapp/init.lua", it will fail to load
    -- its modules, and modules from myapp/.rocks.
    local fio = require('fio')
    local app_dir = fio.abspath(fio.dirname(arg[0]))
    package.path = app_dir .. '/?.lua;' .. package.path
    package.path = app_dir .. '/?/init.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?/init.lua;' .. package.path
    package.cpath = app_dir .. '/?.so;' .. package.cpath
    package.cpath = app_dir .. '/?.dylib;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.so;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.dylib;' .. package.cpath
end

-- configure cartridge

local cartridge = require('cartridge')

local ok, err = cartridge.cfg({
    roles = {
        'cartridge.roles.vshard-storage',
        'cartridge.roles.vshard-router',
        'cartridge.roles.metrics',
        'migrator',
        'app.roles.markets.market-btc',
        'app.roles.markets.market-eth',
        'app.roles.markets.market-sol',
        'app.roles.markets.market-arb',
        'app.roles.markets.market-doge',
        'app.roles.markets.market-ldo',
        'app.roles.markets.market-sui',
        'app.roles.markets.market-pepe',
        'app.roles.markets.market-bch',
        'app.roles.markets.market-xrp',
        'app.roles.markets.market-wld',
        'app.roles.markets.market-ton',
        'app.roles.markets.market-stx',
        'app.roles.markets.market-matic',
        'app.roles.markets.market-trb',
        'app.roles.markets.market-apt',
        'app.roles.markets.market-inj',
        'app.roles.markets.market-aave',
        'app.roles.markets.market-link',
        'app.roles.markets.market-bnb',
        'app.roles.markets.market-rndr',
        'app.roles.markets.market-mkr',
        'app.roles.markets.market-rlb',
        'app.roles.markets.market-ordi',
        'app.roles.markets.market-stg',
        'app.roles.markets.market-sats',
        'app.roles.markets.market-tia',
        'app.roles.markets.market-blur',
        'app.roles.markets.market-jto',
        'app.roles.markets.market-meme',
        'app.roles.markets.market-sei',
        'app.roles.markets.market-yes',
        'app.roles.markets.market-wif',
        'app.roles.markets.market-strk',
        'app.roles.markets.market-shib',
        'app.roles.markets.market-bome',
        'app.roles.markets.market-slerf',
        'app.roles.markets.market-w',
        'app.roles.markets.market-ena',
        'app.roles.markets.market-pac',
        'app.roles.markets.market-maga',
        'app.roles.markets.market-trump',
        'app.roles.markets.market-mog',
        'app.roles.markets.market-not',
        'app.roles.markets.market-mother',
        'app.roles.markets.market-bonk',
        'app.roles.markets.market-taiko',
        'app.roles.markets.market-floki',
        'app.roles.pubsub',
        'app.roles.gateway',
        'app.roles.dev',
        'app.roles.profile',
        'app.roles.auth',

        'app.roles.tests.test_equeue',
        'app.roles.tests.test_debug'
    },
    vshard_groups = {
        'history',
        'profile'
    },
    bucket_count = 10000
},{
    too_long_threshold = 10.0, --xxx
    memtx_memory = 10 * 1024 * 1024 * 1024,
    log_format = 'json',
    -- log = "file:app.log"
})

assert(ok, tostring(err))

-- register admin function to use it with 'cartridge admin' command

-- local admin = require('app.admin')
-- admin.init()

-- local metrics_api = require('metrics.api')
-- metrics_api.register_callback(require('app.metrics.fibers').update)

local metrics = require('cartridge.roles.metrics')
metrics.enable_default_metrics()
metrics.set_export({
    {
        path = '/health',
        format = 'health'
    },
    {
        path = '/metrics/prometheus',
        format = 'prometheus'
    },
})
