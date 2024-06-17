function table.slice(tbl, first, last)
    local sliced = {}

    for i = first or 1, last or #tbl, 1 do
        sliced[#sliced+1] = tbl[i]
    end

    return sliced
end

function table.extend(tbl, other_tbl)
    for _, value in pairs(other_tbl or {}) do
        table.insert(tbl, value)
    end
end


return {
    slice = table.slice,
    extend = table.extend,
}

