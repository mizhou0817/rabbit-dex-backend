package tdecimal

import (
	"math/big"

	"github.com/shopspring/decimal"
)

var (
	multipliers = make(map[int32]*decimal.Decimal, 2)
)

func createMultiplier(decimals int32) *decimal.Decimal {
	result := decimal.NewFromInt32(10).Pow(
		decimal.NewFromInt32(decimals))
	return &result
}

func TokenDecimalsToTDecimal(amount *big.Int, decimals int32) *Decimal {
	return NewDecimal(decimal.NewFromBigInt(amount, -decimals))
}

func TDecimalToTokenDecimals(amount *Decimal, decimals int32) *big.Int {
	multiplier := multipliers[decimals]
	if multiplier == nil {
		multiplier = createMultiplier(decimals)
		multipliers[decimals] = multiplier
	}
	return amount.Mul(*multiplier).Round(0).BigInt()
}

func intNumDigits(val int64) int32 {
	var count int32
	for val != 0 {
		val = val / 10
		count++
	}
	return count
}
