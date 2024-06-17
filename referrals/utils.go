package referrals

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"strings"
)

var ErrNetFee = errors.New("net fee seems to be wrong, critical")
var ErrPair = errors.New("fills should be in pairs")
var ErrTradeId = errors.New("trade id does not match for fills")

func GenShortCode(n uint) (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:n], nil
}

func newDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func GetLevel(currentVolume decimal.Decimal) ReferralLevelStatus {
	var current ReferralLevel
	var next *ReferralLevel
	var prev *ReferralLevel
	var neededVolume *decimal.Decimal

	max := len(REFERRAL_LEVELS)
	for i, tier := range REFERRAL_LEVELS {
		current = tier

		if i > 0 {
			prev = &REFERRAL_LEVELS[i-1]
		}

		if (i + 1) < max {
			next = &REFERRAL_LEVELS[i+1]
		} else {
			next = nil
			break
		}

		if currentVolume.GreaterThanOrEqual(current.Volume) &&
			currentVolume.LessThan(next.Volume) {
			break
		}
	}

	if next != nil {
		volumeDiff := next.Volume.Sub(currentVolume)
		neededVolume = &volumeDiff
	}

	return ReferralLevelStatus{
		Volume:       currentVolume,
		Current:      current,
		Next:         next,
		Prev:         prev,
		NeededVolume: neededVolume,
	}
}

func calculateBonus(previousVolume, currentVolume decimal.Decimal) (decimal.Decimal, []ReferralLevel) {
	var threshold decimal.Decimal

	startVolume := previousVolume
	bonus := decimal.Zero
	referralLevels := make([]ReferralLevel, 0)

	for currentVolume.GreaterThan(startVolume) {
		lvl := GetLevel(startVolume)

		if lvl.Next != nil {
			threshold = lvl.Next.Volume
		} else {
			threshold = currentVolume
		}

		if threshold.GreaterThan(currentVolume) {
			threshold = currentVolume
		}

		lvlAmount := threshold.Sub(startVolume)

		startVolume = startVolume.Add(lvlAmount)
		jumpedLvl := GetLevel(startVolume)
		if jumpedLvl.Current.Level > lvl.Current.Level {
			bonus = bonus.Add(jumpedLvl.Current.MilestoneBonus)
			referralLevels = append(referralLevels, jumpedLvl.Current)
		}
	}
	return bonus, referralLevels
}

func addToFees(fees map[uint64]decimal.Decimal, side referralFillRow, netFee *decimal.Decimal) {
	referrerId := side.ReferrerId
	if referrerId == nil {
		return
	}

	fee := side.Fee
	if fee.GreaterThan(decimal.Zero) {
		// this was a rebate, we don't pay out for rebates.
		return
	}

	if netFee != nil {
		// netFee is already positive.
		fee = *netFee
	} else {
		// fees are negative, rebates are positive.
		fee = fee.Mul(newDecimal(-1))
	}

	if !fee.GreaterThan(decimal.Zero) {
		return
	}

	var commissionPercent decimal.Decimal
	var commissionFee decimal.Decimal
	if *side.Model == ModelPercentage {
		lvl := GetLevel(side.ProfileVolume)
		commissionPercent = lvl.Current.CommissionPercent
	} else if *side.Model == ModelKOL {
		commissionPercent = *side.ModelFeePercent
	} else {
		commissionPercent = decimal.Zero
	}

	commissionFee = fee.Mul(commissionPercent)
	_, ok := fees[*referrerId]
	if !ok {
		fees[*referrerId] = decimal.Zero
	}

	// capping to not overpay
	commissionFee = decimal.Min(commissionFee, fee)

	fees[*referrerId] = fees[*referrerId].Add(commissionFee)
}

func calculateFees(fees map[uint64]decimal.Decimal, fills []referralFillRow) error {
	l := len(fills)
	if l == 0 {
		return nil
	}

	if l%2 != 0 {
		return ErrPair
	}

	for i := 0; i < l; i += 2 {
		sideA := fills[i]
		sideB := fills[i+1]

		if sideA.TradeId != sideB.TradeId {
			return ErrTradeId
		}

		// negative fee means we take from the customer (mostly taker)
		// positive fee means we credit the customer (maker)
		netFee := sideA.Fee.Add(sideB.Fee).Mul(newDecimal(-1))
		if netFee.LessThan(decimal.Zero) {
			// this should not happen, it means we paid more than we got from our customers
			// TODO: warn/error here or crash ? (monitoring)
			return ErrNetFee
		}

		if netFee.GreaterThan(decimal.Zero) {
			// both paid a fee
			if sideA.Fee.LessThan(decimal.Zero) && sideB.Fee.LessThan(decimal.Zero) {
				addToFees(fees, sideA, nil)
				addToFees(fees, sideB, nil)
				continue
			}

			// one of them is on rebate/0 fee
			if sideA.Fee.LessThan(decimal.Zero) {
				addToFees(fees, sideA, &netFee)
			} else {
				addToFees(fees, sideB, &netFee)
			}
		}
	}

	return nil
}

func CreateReferralLink(db *pgxpool.Pool, referrerShortCode string, invitedId uint64) error {
	sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	insertBuilder := sqlBuilder.
		Insert("app_referral_link").
		Columns("profile_id", "invited_id").
		Values(sq.Expr("(SELECT profile_id FROM app_referral_code WHERE LOWER(short_code) = ?)", strings.ToLower(referrerShortCode)), invitedId)

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}
