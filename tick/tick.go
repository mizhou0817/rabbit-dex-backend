package tick

import (
	"math"

	"github.com/shopspring/decimal"
)

const (
	USDT_TICK = 0.000001
)

var DECIMAL_USDT_TICK = decimal.NewFromFloat(0.000001)

func RoundDownToUsdtTick(size float64) float64 {
	return RoundDownToTick(size, USDT_TICK)
}

func RoundDownToTick(size float64, tick float64) float64 {
	if tick <= 0 {
		return size
	}
	milliTicks := int(math.Round(size * 1000 / tick))
	numTicks := float64(milliTicks / 1000)
	return numTicks * tick
}

func RoundDecimalDownToUsdtTick(size decimal.Decimal) decimal.Decimal {
	return RoundDecimalDownToTick(size, DECIMAL_USDT_TICK)
}

func RoundDecimalDownToTick(size decimal.Decimal, tick decimal.Decimal) decimal.Decimal {
	if tick.LessThanOrEqual(decimal.Zero) {
		return size
	}
	numTicks := size.Div(tick).Floor()
	return numTicks.Mul(tick)
}