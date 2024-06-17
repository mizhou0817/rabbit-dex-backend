package liqengine

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

const (
	CANCEL_INTERVAL = 1 * time.Second
	CHECK_INTERVAL  = 2 * time.Second
	LIQ_BATCH_LIMIT = 10
)

type ServiceId int

type LiquidationService struct {
	serviceId     ServiceId
	insuranceId   uint
	checkInterval time.Duration
	assistant     Assistant
	engine        LiquidationEngine
	stopf         context.CancelFunc
}

func NewLiquidationService(insuranceId uint, assistant Assistant) *LiquidationService {
	_insuranceId, err := assistant.GetOrCreateInsurance(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}

	serviceId := assistant.GetNextLiquidationServiceId()
	ls := LiquidationService{
		serviceId:     serviceId,
		insuranceId:   _insuranceId,
		checkInterval: CHECK_INTERVAL,
		assistant:     assistant,
		engine:        LiquidationEngine{},
	}
	return &ls
}

func (ls *LiquidationService) Run() context.CancelFunc {
	ctx, cancelf := context.WithCancel(context.Background())
	ticker := time.NewTicker(ls.checkInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ls.ProcessLiquidations(ctx)
			}
		}
	}()
	ls.stopf = cancelf
	return cancelf
}

func (ls *LiquidationService) Stop() {
	if ls.stopf != nil {
		ls.stopf()
	}
}

func (ls *LiquidationService) ProcessLiquidations(ctx context.Context) (uint, error, []model.Action) {
	var last_id *uint = nil
	var account *AccountData

	all_actions := make([]model.Action, 0)

	total := uint(0)
	total_liq := 0
	for {
		batch, err := ls.assistant.GetNextLiqBatch(ctx, last_id, LIQ_BATCH_LIMIT)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
			return 0, err, nil
		}

		logrus.Infof(".... ProcessLiquidations round batch=%d", len(batch))

		if len(batch) == 0 {
			break
		}

		for _, profile := range batch {
			last_id = &profile.ProfileID

			if *profile.ProfileType == model.PROFILE_TYPE_INSURANCE {
				logrus.Warn("CAN'T LIQUIDATE INSURANCE")
				continue
			}

			total += 1
			select {
			case <-ctx.Done():
				return 0, nil, nil
			default:
			}
			accountMargin := profile.AccountMargin.InexactFloat64()
			if accountMargin < -0.1 || (accountMargin > 0.03 && !ls.engine.isLiquidationEnding(profile)) {
				logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("MARGIN_ERROR profile %d account margin %f", profile.ProfileID, accountMargin)
				continue
			}
			logrus.Warnf("....profileId=%d  margin=%f ae=%f total_notioanl=%f", profile.ProfileID, profile.AccountMargin.InexactFloat64(), profile.AccountEquity.InexactFloat64(), profile.TotalNotional.InexactFloat64())

			if ls.engine.shouldLiquidationHaveMoreTime(profile) {
				continue
			}

			if ls.engine.isLiquidationEnding(profile) {
				err = ls.assistant.WaitForCancellAllAccepted(ctx, profile.ProfileID)
				if err != nil {
					logrus.Error(err)
					continue
				}

				ls.assistant.CompletedLiquidation(ctx, profile.ProfileID)
				continue
			}

			if ls.engine.belowLiquidationMargin(profile.AccountMargin.InexactFloat64()) {
				//TODO: can be replaced with CancelAllOrders (risky to do it now)
				err = ls.assistant.WaitForCancellAllAccepted(ctx, profile.ProfileID)
				if err != nil {
					logrus.Error(err)
					continue
				}

				//TODO: can be replaced with QueueLiquidateActions (risky to do it now)
				total_liq += 1
				account, err = ls.assistant.GetAccountData(ctx, profile)
				if err != nil {
					logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("Liquidation service, error getting account %d:\n%s", profile.ProfileID, err.Error())
					continue
				}

				actions, liquidatedVaults := ls.engine.requiredActions(account)

				if len(actions) > 0 {
					all_actions = append(all_actions, actions...)
					ls.assistant.Queue(ctx, actions)
				}
				if len(liquidatedVaults) > 0 {
					err := ls.assistant.LiquidatedVaults(ctx, liquidatedVaults)
					if err != nil {
						logrus.Errorf("LIQUIDATION SERVICE: error liquidating vaults: %v", err)
					}
				}

				ls.assistant.UpdateLastChecked(ctx, profile.ProfileID)
			}
		}
	}
	if total != 0 || total_liq != 0 {
		logrus.Warnf(".... LIQ total=%d", total)
		logrus.Warnf(".... for liquidation total=%d", total_liq)
	}
	return total, nil, all_actions
}

func (ls *LiquidationService) CancelAllOrders(ctx context.Context, profile *model.ProfileCache) error {
	err := ls.assistant.WaitForCancellAllAccepted(ctx, profile.ProfileID)
	if err != nil {
		logrus.Errorf("Liquidation service: account: %d, cancelall error: %v", profile.ProfileID, err)
		return err
	}

	return nil
}

func (ls *LiquidationService) QueueLiquidateActions(ctx context.Context, profile *model.ProfileCache) error {
	account, err := ls.assistant.GetAccountData(ctx, profile)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("Liquidation service: error getting account %d:\n%s", profile.ProfileID, err.Error())
		return err
	}

	actions, liquidatedVaults := ls.engine.requiredActions(account)
	if len(actions) > 0 {
		ls.assistant.Queue(ctx, actions)
	}
	if len(liquidatedVaults) > 0 {
		err := ls.assistant.LiquidatedVaults(ctx, liquidatedVaults)
		if err != nil {
		logrus.Errorf("LIQUIDATION SERVICE: error liquidating vaults: %v", err)
	}
	}

	return nil
}
