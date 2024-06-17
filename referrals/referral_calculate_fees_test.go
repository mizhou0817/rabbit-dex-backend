package referrals

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/strips-finance/rabbit-dex-backend/tests"
)

func ToPtr[T any](v T) *T {
	return tests.ToPtr(v)
}

func TestCalculateFeesNoRows(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	var fills []referralFillRow

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 0)
}

func TestCalculateFeesSingleFill(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{{}}

	err := calculateFees(fees, fills)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPair)
	assert.Len(t, fees, 0)
}

func TestCalculateFeesNetFeeErr(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(-10)},
		{ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(10.001)},
	}

	err := calculateFees(fees, fills)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNetFee)
	assert.Len(t, fees, 0)
}

func TestCalculateFeesTradeIdErr(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{TradeId: "111"},
		{TradeId: "222"},
	}

	err := calculateFees(fees, fills)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTradeId)
	assert.Len(t, fees, 0)
}

func TestCalculateFeesSimpleNoRebate(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(-10)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-5)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 1)

	fee := fees[1]
	expected := newDecimal(10*0.2 + 5*0.2)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesSimpleEqualRebate(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(-10)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 0)
}

func TestCalculateFeesSimpleWithRebate(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 1)

	fee := fees[1]
	expected := newDecimal((12 - 10) * 0.2)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesSimpleWithRebateReverse(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 1)

	fee := fees[1]
	expected := newDecimal((12 - 10) * 0.2)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesMultiple(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},

		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(4)), ProfileVolume: decimal.Zero, Fee: newDecimal(-20)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(5)), ProfileVolume: decimal.Zero, Fee: newDecimal(12)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 2)

	fee := fees[1]
	expected := newDecimal((12 - 10) * 0.2)
	assert.Equal(t, expected.String(), fee.String())

	fee = fees[2]
	expected = newDecimal((20 - 12) * 0.2)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesKOLModel(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.5)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.5)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},

		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.4)), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(4)), ProfileVolume: decimal.Zero, Fee: newDecimal(-20)},
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.4)), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(5)), ProfileVolume: decimal.Zero, Fee: newDecimal(12)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 2)

	fee := fees[1]
	expected := newDecimal((12 - 10) * 0.5)
	assert.Equal(t, expected.String(), fee.String())

	fee = fees[2]
	expected = newDecimal((20 - 12) * 0.4)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesMixedModel(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.5)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(10)},
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(0.5)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},

		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(4)), ProfileVolume: decimal.Zero, Fee: newDecimal(-20)},
		{Model: ToPtr(ModelPercentage), ReferrerId: ToPtr(uint64(2)), InvitedId: ToPtr(uint64(5)), ProfileVolume: decimal.Zero, Fee: newDecimal(12)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 2)

	fee := fees[1]
	expected := newDecimal((12 - 10) * 0.5)
	assert.Equal(t, expected.String(), fee.String())

	fee = fees[2]
	expected = newDecimal((20 - 12) * 0.2)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesKOLModelOverPayment(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(1)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(-10)},
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(1.2)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(-12)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 1)

	fee := fees[1]
	expected := newDecimal((12 + 10) * 1)
	assert.Equal(t, expected.String(), fee.String())
}

func TestCalculateFeesZeroPayout(t *testing.T) {
	fees := make(map[uint64]decimal.Decimal)
	fills := []referralFillRow{
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(1)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(2)), ProfileVolume: decimal.Zero, Fee: newDecimal(0)},
		{Model: ToPtr(ModelKOL), ModelFeePercent: ToPtr(newDecimal(1.2)), ReferrerId: ToPtr(uint64(1)), InvitedId: ToPtr(uint64(3)), ProfileVolume: decimal.Zero, Fee: newDecimal(0)},
	}

	err := calculateFees(fees, fills)
	assert.NoError(t, err)
	assert.Len(t, fees, 0)
}
