local setters = {}

function setters.refresh_profile_data(profile_id)

end

function setters.next_order_id(market_id)
    local n =  box.sequence.order_sequence:next()
    local oid = market_id .. "@" .. tostring(n)

    return oid
end


return setters