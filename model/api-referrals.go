package model

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	REFERRAL_CREATE_PAYOUT  = "balance.create_referral_payout"
	REFERRAL_PROCESS_PAYOUT = "profile.process_referral_payout"
)

func (api *ApiModel) CreateReferralPayout(ctx context.Context, id string, profileId uint64, marketId string, amount decimal.Decimal) (*BalanceOps, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[*BalanceOps]{}.Request(ctx, instance.Title, api.broker, REFERRAL_CREATE_PAYOUT, []interface{}{
		id,
		profileId,
		tdecimal.NewDecimal(amount),
	})

	return data, err
}

func (api *ApiModel) ProcessReferralPayout(ctx context.Context, marketId string) (bool, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		return false, err
	}

	data, err := DataResponse[bool]{}.Request(ctx, instance.Title, api.broker, REFERRAL_PROCESS_PAYOUT, []interface{}{
		marketId,
	})
	return data, err
}
