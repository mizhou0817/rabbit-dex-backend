package model

import (
	"context"

	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	LIQUIDATION_BATCH           = "getters.liquidation_batch"
	PROFILE_UPDATE_STATUS       = "profile.update_status"
	PROFILE_UPDATE_LAST_CHECKED = "profile.update_last_checked"
	PROFILE_UPDATE_INSURANCE_ID = "profile.update_insurance_id"
	PROFILE_LIQUIDATED_VAULTS   = "profile.liquidated_vaults"
	QUEUE_LIQ_ACTIONS           = "internal.queue_liq_actions"
	IS_INV3_VALID               = "getters.is_inv3_valid"
	CACHED_IS_INV3_VALID        = "getters.cached_is_inv3_valid"
	GET_WINNING_POSITIONS       = "position.get_winning_positions"
	HIGH_PRIORITY_CANCEL_ALL    = "internal.high_priority_cancell_all"
	IS_CANCEL_ALL_ACCEPTED      = "internal.is_cancel_all_accepted"
)

type ActionType int

const (
	APlaceSellOrders ActionType = iota
	AInsTakeover
	AInsClawback
)

type Action struct {
	Kind     ActionType       `msgpack:"kind"`
	TraderId uint             `msgpack:"trader_id"`
	MarketId string           `msgpack:"market_id"`
	Size     tdecimal.Decimal `msgpack:"size"`
	Price    tdecimal.Decimal `msgpack:"price"`
}

func (a ActionType) Description() string {
	switch a {
	case APlaceSellOrders:
		return "placesellorders" // Always ASK (sell) order on behalf of the TraderId
	case AInsTakeover:
		return "instakeover" // Private match Insurance(buy order) Trader's(sell order) position of provided size with ZP price.
	case AInsClawback:
		return "insclawback" // Private match Insurance(sell order) with Trader's(buy order) with provided params
	default:
		return "Undefined"
	}
}

func (api *ApiModel) UpdateInsuranceId(ctx context.Context, insurance_id uint, market_ids []string) error {

	for _, market_id := range market_ids {
		instance, err := GetInstance().ByMarketID(market_id)
		if err != nil {
			return err
		}

		_, err = DataResponse[interface{}]{}.Request(ctx, instance.Title, api.broker, PROFILE_UPDATE_INSURANCE_ID, []interface{}{
			insurance_id,
			market_id,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (api *ApiModel) CreateInsuranceProfile(ctx context.Context, wallet string) (*Profile, error) {
	profile, err := DataResponse[*Profile]{}.Request(ctx, PROFILE_INSTANCE, api.broker, CREATE_PROFILE, []interface{}{
		PROFILE_TYPE_INSURANCE,
		PROFILE_STATUS_ACTIVE,
		wallet,
		EXCHANGE_DEFAULT,
	})

	return profile, err
}

func (api *ApiModel) HighPriorityCancelAll(ctx context.Context, profile_id uint, is_liquidation bool) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, API_INSTANCE, api.broker, HIGH_PRIORITY_CANCEL_ALL, []interface{}{
		profile_id,
		is_liquidation,
	})

	return err
}

func (api *ApiModel) IsCancellAllAccepted(ctx context.Context, profile_id uint) (bool, error) {
	accepted, err := DataResponse[bool]{}.Request(ctx, API_INSTANCE, api.broker, IS_CANCEL_ALL_ACCEPTED, []interface{}{
		profile_id,
	})

	return accepted, err
}

// WILL return positions for market_id with specified side, where unrealized_pnl of position > 0
func (api *ApiModel) GetWinningPositions(ctx context.Context, market_id, side string) ([]*PositionData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	res, err := DataResponse[[]*PositionData]{}.Request(ctx, instance.Title, api.broker, GET_WINNING_POSITIONS, []interface{}{
		side,
	})

	return res, err

}

func (api *ApiModel) CachedIsInv3Valid(ctx context.Context, inv3Buffer float64) (bool, error) {
	res, err := DataResponse[bool]{}.Request(ctx, PROFILE_INSTANCE, api.broker, CACHED_IS_INV3_VALID, []interface{}{
		inv3Buffer,
	})

	if err != nil {
		return false, err
	}

	return res, err
}

func (api *ApiModel) IsInv3Valid(ctx context.Context, inv3Buffer float64) (bool, error) {
	res, err := DataResponse[bool]{}.Request(ctx, PROFILE_INSTANCE, api.broker, IS_INV3_VALID, []interface{}{
		inv3Buffer,
	})

	return res, err
}

func (api *ApiModel) QueueLiqActions(ctx context.Context, actions []Action) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, API_INSTANCE, api.broker, QUEUE_LIQ_ACTIONS, []interface{}{
		actions,
	})

	return err
}

func (api *ApiModel) LiquidatedVaults(ctx context.Context, vaultProfileIds []uint) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, API_INSTANCE, api.broker, PROFILE_LIQUIDATED_VAULTS, []interface{}{
		vaultProfileIds,
	})

	return err
}

func (api *ApiModel) LiquidationBatch(ctx context.Context, last_id_checked *uint, limit int) ([]*ProfileCache, error) {
	res, err := DataResponse[[]*ProfileCache]{}.Request(ctx, PROFILE_INSTANCE, api.broker, LIQUIDATION_BATCH, []interface{}{
		last_id_checked,
		limit,
	})

	return res, err
}

func (api *ApiModel) ProfileUpdateStatus(ctx context.Context, profile_id uint, new_status string) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_UPDATE_STATUS, []interface{}{
		profile_id,
		new_status,
	})

	return err
}

func (api *ApiModel) ProfileUpdateLastLiqChecked(ctx context.Context, profile_id uint) (int64, error) {
	tm, err := DataResponse[int64]{}.Request(ctx, PROFILE_INSTANCE, api.broker, PROFILE_UPDATE_LAST_CHECKED, []interface{}{
		profile_id,
	})

	return tm, err
}
