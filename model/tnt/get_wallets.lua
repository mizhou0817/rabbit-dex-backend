local net_box = require('net.box')
local conn = net_box.connect('localhost:3004')
local res = conn:eval([[
    local wallets = {}
    local tuples = box.space.profile:select()
    for _, tuple in ipairs(tuples) do
        table.insert(wallets, tuple[4])
    end
    return wallets
]])

for i, wallet in ipairs(res) do
    print(wallet)
end

conn:close()