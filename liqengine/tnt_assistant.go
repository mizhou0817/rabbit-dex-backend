package liqengine

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

const (
	WAIT_CANCEL_ALL_INTERVAL = 100 * time.Millisecond
	POLL_INTERVAL            = 20 * time.Millisecond
	MAX_EXPECTED_POSITIONS   = 1000
	INSURANCE_WALLET         = "0xinsurance"
)

type TntAssistant struct {
	broker     *model.Broker
	exchangeId string
	inv3Buffer float64
	apiModel   *model.ApiModel
}

func NewTntAssistant(broker *model.Broker, exchangeId string, inv3Buffer float64) (*TntAssistant, error) {
	apiModel := model.NewApiModel(broker)
	ta := &TntAssistant{
		broker:     broker,
		exchangeId: exchangeId,
		inv3Buffer: inv3Buffer,
		apiModel:   apiModel,
	}
	return ta, nil
}

func (ta *TntAssistant) Queue(ctx context.Context, actions []model.Action) error {
	return ta.apiModel.QueueLiqActions(ctx, actions)
}

func (ta *TntAssistant) LiquidatedVaults(ctx context.Context, vaults []uint) error {
	return ta.apiModel.LiquidatedVaults(ctx, vaults)
}

func (ta *TntAssistant) CompletedLiquidation(ctx context.Context, traderId uint) error {
	return ta.apiModel.ProfileUpdateStatus(ctx, uint(traderId), model.PROFILE_STATUS_ACTIVE)
}

func (ta *TntAssistant) UpdateLastChecked(ctx context.Context, traderId uint) error {
	_, err := ta.apiModel.ProfileUpdateLastLiqChecked(ctx, traderId)
	return err
}

func (ta *TntAssistant) GetNextLiqBatch(ctx context.Context, last_id *uint, limit int) ([]*model.ProfileCache, error) {
	return ta.apiModel.LiquidationBatch(ctx, last_id, limit)
}

func (ta *TntAssistant) GetInsuranceData(ctx context.Context, insurance_id uint) (*AccountData, error) {
	cache, err := ta.apiModel.GetProfileCache(ctx, insurance_id)
	if err != nil {
		return nil, err
	}

	return ta.GetAccountData(ctx, cache)
}

func (ta *TntAssistant) GetAccountData(ctx context.Context, profile *model.ProfileCache) (*AccountData, error) {
	data := AccountData{
		Cache: profile,
	}

	positions, err := ta.apiModel.GetOpenPositions(ctx, profile.ProfileID)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return &data, nil
	}

	data.Positions = positions
	data.Markets = make(map[string]*model.MarketData)

	for _, pos := range positions {
		marketData, err := ta.apiModel.GetMarketData(ctx, pos.MarketID)
		if err != nil {
			return &data, err
		}

		data.Markets[pos.MarketID] = marketData
	}

	return &data, nil
}

func (ta *TntAssistant) ClawbackRequired(ctx context.Context) bool {
	valid, err := ta.apiModel.CachedIsInv3Valid(ctx, ta.inv3Buffer)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		return false
	}

	if valid {
		return false
	}

	data, err := ta.GetInsuranceData(ctx, uint(0))
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		return false
	}

	if data.Cache.AccountMargin.InexactFloat64() > 0.001 {
		return false
	}

	return true
}

func (ta *TntAssistant) GetWinningTraderPostns(ctx context.Context,
	marketId string,
	side string,
	atPrice float64,
	insuranceId uint) (map[uint]*model.PositionData, error) {

	winPositions, err := ta.apiModel.GetWinningPositions(ctx, marketId, side)
	if err != nil {
		return nil, err
	}
	winningPositions := make(map[uint]*model.PositionData)
	for _, modelPos := range winPositions {
		traderId := uint(modelPos.ProfileID)
		if traderId == insuranceId {
			continue
		}

		pnl := calcUnrealizedPnl(modelPos.Size.InexactFloat64(), modelPos.EntryPrice.InexactFloat64(), atPrice, modelPos.Side)
		if pnl <= 0 {
			continue
		}
		winningPositions[traderId] = modelPos
	}
	return winningPositions, nil
}

func (ta *TntAssistant) GetNextLiquidationServiceId() ServiceId {
	return 0 // @todo
}

// DEPRECATED: use WaitForCancellAllAccepted if you need the guarantee of cancel all orders before moving forward
func (ta *TntAssistant) CancelExistingOrders(ctx context.Context, traderId uint) error {
	return ta.apiModel.HighPriorityCancelAll(ctx, traderId, true)
}

func (ta *TntAssistant) WaitForCancellAllAccepted(ctx context.Context, traderId uint) error {
	err := ta.apiModel.HighPriorityCancelAll(ctx, traderId, true)
	if err != nil {
		return err
	}

	attempts := 0
	for {
		is_accepted, err := ta.apiModel.IsCancellAllAccepted(ctx, traderId)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		} else {
			if is_accepted {
				return nil
			}
		}
		attempts += 1
		if attempts > 20 {
			return errors.New("WaitForCancellAllAccepted exceed 20 attempts")
		}
		time.Sleep(WAIT_CANCEL_ALL_INTERVAL)
	}
}

// TODO: done it for prod use case
func (ta *TntAssistant) GetOrCreateInsurance(ctx context.Context) (uint, error) {

	profile, err := ta.apiModel.GetProfileById(context.Background(), 0)
	if err != nil {
		if err.Error() == model.PROFILE_NOT_FOUND_ERROR {
			profile, err = ta.apiModel.CreateInsuranceProfile(context.Background(), INSURANCE_WALLET)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	}

	if profile.ProfileId != 0 {
		logrus.WithField(log.AlertTag, log.AlertCrit).Fatalf("Insurance profileId must be always 0. but have insurance_id=%d", profile.ProfileId)
	}

	//SET INSURANCE FOR EACH MARKET - replace with config params
	/*
		err = ta.apiModel.UpdateInsuranceId(context.Background(), profile.ProfileId, []string{"BTC-USD", "ETH-USD", "SOL-USD"})
		if err != nil {
			return profile.ProfileId, err
		}*/

	return profile.ProfileId, nil
}
