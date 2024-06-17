package referrals

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestCalculateFees(t *testing.T) {
	var testCases = []struct {
		prevVolume        decimal.Decimal
		currentVolume     decimal.Decimal
		expected          decimal.Decimal
		expectedNumLevels uint
	}{
		// all
		{newDecimal(0), newDecimal(0), newDecimal(0), 0},
		{newDecimal(0), newDecimal(100_000), newDecimal(0), 0},
		{newDecimal(0), newDecimal(200_000), newDecimal(5), 1},
		{newDecimal(0), newDecimal(400_000), newDecimal(15), 2},
		{newDecimal(0), newDecimal(1_000_000), newDecimal(45), 3},
		{newDecimal(0), newDecimal(3_500_000), newDecimal(145), 4},
		{newDecimal(0), newDecimal(8_750_000), newDecimal(395), 5},
		{newDecimal(0), newDecimal(23_000_000), newDecimal(1_095), 6},
		{newDecimal(0), newDecimal(65_000_000), newDecimal(2_095), 7},
		{newDecimal(0), newDecimal(200_000_000), newDecimal(7_095), 8},
		{newDecimal(0), newDecimal(500_000_000), newDecimal(17_095), 9},
		{newDecimal(0), newDecimal(1_000_000_000), newDecimal(42_095), 10},
		{newDecimal(0), newDecimal(10_000_000_000), newDecimal(142_095), 11},

		// exotic

		{newDecimal(0), newDecimal(1), newDecimal(0), 0},
		{newDecimal(0), newDecimal(100_001), newDecimal(0), 1},
		{newDecimal(0), newDecimal(200_001), newDecimal(5), 1},
		{newDecimal(0), newDecimal(1_000_000_000), newDecimal(42_095), 10},

		{newDecimal(999_999), newDecimal(1_000_000), newDecimal(30), 1},
		{newDecimal(999_999), newDecimal(1_000_001), newDecimal(30), 1},

		{newDecimal(10_000_000_000), newDecimal(100_000_000_000), newDecimal(0), 0},
	}

	for i, test := range testCases {
		amountExpected := test.expected
		amountCalculated, _ := calculateBonus(test.prevVolume, test.currentVolume)

		if !amountExpected.Equal(amountCalculated) {
			t.Errorf("loop = %d, for prevVolume = %s currentVolume = %s, expected %s. got %s.",
				i, test.prevVolume.String(), test.currentVolume.String(), amountExpected.String(), amountCalculated.String())
		}
	}
}
