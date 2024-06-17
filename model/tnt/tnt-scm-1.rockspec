package = 'tnt'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'vshard == 0.1.21',
    'lua >= 5.1',
    'checks >= 3.1.0-1',
    'cartridge == 2.7.4-1',
    'metrics <= 0.15.1',
    'cartridge-cli-extensions == 1.1.1-1',
    'migrations == 0.4.2-1',
    -- for testing
    'luatest'
}
build = {
    type = 'none';
}
