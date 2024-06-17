package referrals

import "github.com/shopspring/decimal"

type ReferralLevel struct {
	Level             uint64          `json:"level"`
	Volume            decimal.Decimal `json:"volume"`
	CommissionPercent decimal.Decimal `json:"commission_percent"`
	MilestoneBonus    decimal.Decimal `json:"milestone_bonus"`
}

type ReferralLevelStatus struct {
	Volume       decimal.Decimal  `json:"volume"`
	Current      ReferralLevel    `json:"current"`
	Next         *ReferralLevel   `json:"next,omitempty"`
	Prev         *ReferralLevel   `json:"prev,omitempty"`
	NeededVolume *decimal.Decimal `json:"needed_volume,omitempty"`
}

var REFERRAL_LEVELS = []ReferralLevel{
	{1, newDecimal(0), newDecimal(0.20), newDecimal(0)},
	{2, newDecimal(200_000), newDecimal(0.20), newDecimal(5)},
	{3, newDecimal(400_000), newDecimal(0.22), newDecimal(10)},
	{4, newDecimal(1_000_000), newDecimal(0.24), newDecimal(30)},
	{5, newDecimal(3_500_000), newDecimal(0.26), newDecimal(100)},
	{6, newDecimal(8_750_000), newDecimal(0.28), newDecimal(250)},
	{7, newDecimal(23_000_000), newDecimal(0.30), newDecimal(700)},
	{8, newDecimal(65_000_000), newDecimal(0.32), newDecimal(1_000)},
	{9, newDecimal(200_000_000), newDecimal(0.34), newDecimal(5_000)},
	{10, newDecimal(500_000_000), newDecimal(0.36), newDecimal(10_000)},
	{11, newDecimal(1_000_000_000), newDecimal(0.38), newDecimal(25_000)},
	{12, newDecimal(10_000_000_000), newDecimal(0.40), newDecimal(100_000)},
}
