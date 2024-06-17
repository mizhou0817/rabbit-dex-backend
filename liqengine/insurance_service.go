package liqengine

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type InsuranceService struct {
	waterfallInterval time.Duration
	assistant         Assistant
	insuranceId       uint
	engine            LiquidationEngine
	stopf             context.CancelFunc
}

func NewInsuranceService(insuranceId uint, assistant Assistant) *InsuranceService {
	_insuranceId, err := assistant.GetOrCreateInsurance(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}

	is := InsuranceService{
		waterfallInterval: INSURANCE_WATERFALL_INTERVAL,
		assistant:         assistant,
		insuranceId:       _insuranceId,
		engine:            LiquidationEngine{},
	}
	return &is
}

func (is *InsuranceService) Run() context.CancelFunc {
	ctx, cancelf := context.WithCancel(context.Background())
	ticker := time.NewTicker(is.waterfallInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				is.ProcessPositions(ctx)
			}
		}
	}()
	is.stopf = cancelf
	return cancelf
}

func (is *InsuranceService) Stop() {
	if is.stopf != nil {
		is.stopf()
	}
}

func (is *InsuranceService) ProcessPositions(ctx context.Context) int {
	var total_actions int = 0

	var err error
	if is.assistant.ClawbackRequired(ctx) {
		total_actions, err = is.clawback(ctx)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("ClawbackRequired: Insurance service error processing positions:\n %s", err.Error())
		}
	} else {
		total_actions, err = is.sellOnMarket(ctx)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("sellOnMarket: Insurance service error processing positions:\n %s", err.Error())
		}
	}

	return total_actions
}

func (is *InsuranceService) sellOnMarket(ctx context.Context) (int, error) {

	insurance, err := is.assistant.GetInsuranceData(ctx, is.insuranceId)
	if err != nil {
		return 0, err
	}

	interval_passed := IsIntervalPassedForMicroseconds(*insurance.Cache.LastLiqCheck, INSURANCE_WATERFALL_INTERVAL)
	logrus.Infof("SellInMarket insuranceData received insurance.cache.LastLiqCheck = %d positions=%d interval_passed = %v",
		*insurance.Cache.LastLiqCheck,
		len(insurance.Positions),
		interval_passed)

	total_actions := 0
	if (len(insurance.Positions) > 0) && interval_passed {
		err = is.assistant.WaitForCancellAllAccepted(ctx, insurance.Cache.ProfileID)
		if err != nil {
			logrus.Error(err)
			return 0, err
		}

		// We remove orders when place actions
		actions := is.engine.insuranceSelloffActions(insurance)
		if len(actions) > 0 {
			is.assistant.Queue(ctx, actions)
		}
		is.assistant.UpdateLastChecked(ctx, insurance.Cache.ProfileID)

		total_actions = len(actions)
	}
	return total_actions, nil
}

func (is *InsuranceService) clawback(ctx context.Context) (int, error) {
	insurance, err := is.assistant.GetInsuranceData(ctx, is.insuranceId)
	if err != nil {
		return 0, err
	}

	total_actions := 0

	margin := insurance.Cache.AccountMargin.InexactFloat64()
	actions := make([]model.Action, 0, len(insurance.Positions))

	err = is.assistant.WaitForCancellAllAccepted(ctx, insurance.Cache.ProfileID)
	if err != nil {
		logrus.Error(err)
		return 0, err
	}

	for _, insurancePos := range insurance.Positions {
		requiredSide := FlipSide(insurancePos.Side)
		zeroPrice := calcZp(insurancePos, margin)
		winningTraders, err := is.assistant.GetWinningTraderPostns(ctx, insurancePos.MarketID, requiredSide, zeroPrice, 0)
		if err != nil {
			return 0, err
		}

		if len(winningTraders) > 0 {
			actions = append(
				actions,
				is.engine.clawbackActions(insurance, insurancePos, winningTraders)...,
			)
		}
	}

	total_actions = len(actions)
	logrus.Infof("... total_actions=%d", total_actions)

	if len(actions) > 0 {
		is.assistant.Queue(ctx, actions)
	}
	return total_actions, nil
}
