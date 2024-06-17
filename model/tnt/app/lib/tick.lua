local checks = require('checks')
local decimal = require('decimal')

local ZERO = decimal.new(0)

local T = {}

function T.floor_decimal(number)
  if number == 0 then
    return number
  end

  number = number - 0.5
  return decimal.rescale(number, 0)
end

function T.ceil_decimal(number)
  if number == 0 then
    return number
  end

  number = number + 0.5
  return decimal.rescale(number, 0)
end

function T.round_to_nearest_tick(size, tick, min_tick)
  min_tick = min_tick or ZERO
  if decimal.is_decimal(size) == false then
    size = decimal.new(size)
  end

  if tick <= 0 then
        return size
  end

  -- TODO: there was a problem if size equal to 10, 100 .. and tick 1
  if tick == 1 then
    local _size = decimal.abs(size)
    if _size < 1 then
      return min_tick
    end

    local _conv = decimal.rescale(_size, 0)

    local diff = _conv - _size

    -- 3 scenarios:
    -- diff = 0 (equal - just return _conv)
    -- diff < 0 (round down already return _conv)
    -- diff > 0 (round up happened, need to substruct 1)
    if diff > 0 then 
      _conv = _conv - 1

      -- just integrity: never happened
      if _conv <= 0 then
        return ZERO
      end
    end

    size = decimal.abs(_conv)
    if size < min_tick then
        return min_tick
    end
  end

  local numTicks = decimal.abs(T.floor_decimal(size / tick))

  size = tick * numTicks
  if size < min_tick then
        return min_tick
  end
  -- some operations can change decimal internal structure so msgpuck not always correctly handles it
  -- _new(tostring)_ hack normalizes decimal
  return decimal.new(tostring(size))
end

function T.min(one, two)
  local min = one
  if two < one then
    min = two
  end

  return min
end

function T.max(one, two)
  local max = one
  if two > one then
    max = two
  end

  return max
end

function T.calc_middle_price(best_ask, best_bid, tick)
  return T.round_to_nearest_tick((best_ask + best_bid) / 2, tick)
end


function T.calc_mid_price(best_ask, tick)
  return T.round_to_nearest_tick(best_ask - tick, tick)
end

function T.is_mid_price_valid(mid_price, best_ask, best_bid)
  if mid_price <= 0 then
    return false
  end

  if mid_price >= best_ask or mid_price <= best_bid then
    return false
  end

  return true
end


function T.is_valid_rounding(value, delta)
    checks('decimal', 'decimal')

    if decimal.abs(value) <= delta then
        return true
    end
    return false
end

return T
