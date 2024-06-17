package referrals

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"time"
)

type ReferralService struct {
	cfg       *Config
	dbManager *tsdb
}

func New(cfg *Config, dbPool *pgxpool.Pool) *ReferralService {
	dbManager := newTSDB(dbPool)

	return &ReferralService{
		cfg:       cfg,
		dbManager: dbManager,
	}
}

func (r *ReferralService) doVolumes() error {
	f := func(shardId string) error {
		ctx := context.Background()
		tx, err := r.dbManager.db.Begin(context.Background())
		if err != nil {
			return fmt.Errorf("db begin() failed: %w", err)
		}

		defer tx.Rollback(ctx)

		logrus.Info("Processing volumes for shard_id = ", shardId)
		volumes, window, err := r.dbManager.calculateVolumes(tx, shardId)
		if err != nil {
			return fmt.Errorf("calculateVolumes() failed: %w", err)
		}

		for profileId, v := range volumes {
			err = r.dbManager.updateVolume(tx, profileId, v.Volume)
			if err != nil {
				return fmt.Errorf("updateVolume() failed: %w", err)
			}

			if v.Model == ModelPercentage {
				bonus, levels := calculateBonus(v.ExistingVolume, v.ExistingVolume.Add(v.Volume))
				if !bonus.IsZero() {
					for _, level := range levels {
						err = r.dbManager.createReferralPayout(tx, profileId, shardId, level.MilestoneBonus)
						if err != nil {
							return fmt.Errorf("createReferralPayout() failed: %w", err)
						}

						err = r.dbManager.createBonusPayoutIntegrity(tx, profileId, level.Level)
						if err != nil {
							return fmt.Errorf("createBonusPayoutIntegrity() failed: %w", err)
						}
					}
				}
			}
		}

		err = r.dbManager.saveVolumePosition(tx, shardId, window.ArchiveIdStart, window.ArchiveIdEnd)
		if err != nil {
			return fmt.Errorf("saveVolumePosition() failed: %w", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("tx.Commit() failed: %w", err)
		}

		return nil
	}

	for _, shardId := range r.cfg.Service.ShardIds {
		// tx will be already rolled back at this point if something goes wrong.
		// we however do not want to stop the execution for the next shard.
		err := f(shardId)
		if err != nil {
			logrus.Error(err)
			continue
		}
	}
	return nil
}

func (r *ReferralService) doRefreshLeaderBoards() error {
	periods := []string{"lifetime", "monthly", "weekly"}
	for _, exchange := range model.SupportedExchangeIds {
		for _, period := range periods {
			err := r.dbManager.refreshLeaderBoard(exchange, period)
			if err != nil {
				return fmt.Errorf("refreshLeaderBoard(%s, %s) failed: %w", exchange, period, err)
			}
		}
	}

	return nil
}

func (r *ReferralService) createPayouts(tx pgx.Tx, shardId string) error {
	fills, window, err := r.dbManager.getReferralFills(tx, shardId)
	if err != nil {
		return err
	}

	l := len(fills)
	if l == 0 {
		return nil
	}

	fees := make(map[uint64]decimal.Decimal)
	err = calculateFees(fees, fills)
	if err != nil {
		return err
	}

	if len(fees) > 0 {
		for profileId, fee := range fees {
			err = r.dbManager.createReferralPayout(tx, profileId, shardId, fee)
			if err != nil {
				return fmt.Errorf("createReferralPayout() failed: %w", err)
			}
		}
	}

	err = r.dbManager.saveFillsPosition(tx, shardId, window.ArchiveIdStart, window.ArchiveIdEnd)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReferralService) processPayouts(tx pgx.Tx) error {
	payouts, err := r.dbManager.getUnProcessedPayouts(tx)
	if err != nil {
		return err
	}

	broker, err := model.GetBroker()
	if err != nil {
		return fmt.Errorf("GetBroker() failed: %w", err)
	}
	apiModel := model.NewApiModel(broker)

	ctx := context.Background()
	ids := make([]string, 0)
	for _, payout := range payouts {
		_, err = apiModel.CreateReferralPayout(ctx, payout.Id, payout.ProfileId, payout.MarketId, payout.Amount)
		// an expected error can be to get out of sync
		// and try to process twice the same payout, since tarantool is on a different process.
		// there is nothing wrong with this scenario, however we need to still mark this as processed
		if err != nil && err.Error() != model.ERR_REFERRAL_PAYOUT_ID_DUPLICATE {
			return fmt.Errorf("CreateReferralPayout() error: %w", err)
		}

		ids = append(ids, payout.Id)
	}

	err = r.dbManager.setToProcessed(tx, ids)
	if err != nil {
		return err
	}

	for _, marketId := range r.cfg.Service.ShardIds {
		_, err = apiModel.ProcessReferralPayout(ctx, marketId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReferralService) doPayouts() error {
	f := func(shardId string) error {
		logrus.Info("Creating payouts for shard_id = ", shardId)
		ctx := context.Background()
		tx, err := r.dbManager.db.Begin(context.Background())
		if err != nil {
			return fmt.Errorf("db begin() failed: %w", err)
		}
		defer tx.Rollback(ctx)

		err = r.createPayouts(tx, shardId)
		if err != nil {
			return fmt.Errorf("createPayouts() failed: %w", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("tx.Commit() failed: %w", err)
		}

		return nil
	}

	for _, shardId := range r.cfg.Service.ShardIds {
		err := f(shardId)
		if err != nil {
			logrus.Error(err)
			continue
		}
	}

	return nil
}

func (r *ReferralService) doProcessPayouts() error {
	ctx := context.Background()
	tx, err := r.dbManager.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("db begin() failed: %w", err)
	}
	defer tx.Rollback(ctx)

	err = r.processPayouts(tx)
	if err != nil {
		return fmt.Errorf("processPayouts() failed: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit() failed: %w", err)
	}

	return nil
}

func getRunnerInterval(regularInterval time.Duration, lastRunTime time.Time) time.Duration {
	now := time.Now().UTC()
	diffSecs := int64(now.Sub(lastRunTime) / time.Second)
	regularIntervalSecs := int64(regularInterval.Seconds())

	if diffSecs > regularIntervalSecs {
		// instant execution
		return 0
	}

	return time.Duration(regularIntervalSecs - diffSecs)
}

func (r *ReferralService) Run() {
	err := r.dbManager.setupRunners()
	if err != nil {
		logrus.Error("failed setupRunners(): ", err)
		return
	}

	runners, err := r.dbManager.getRunners()
	if err != nil {
		logrus.Error("failed getRunners(): ", err)
		return
	}

	volumesDefaultDuration := time.Duration(r.cfg.Service.VolumesInterval) * time.Second
	refreshLeaderBoardDefaultDuration := time.Duration(r.cfg.Service.LeaderBoardInterval) * time.Second
	createPayoutsDefaultDuration := time.Duration(r.cfg.Service.CreatePayoutsInterval) * time.Second
	processPayoutsDefaultDuration := time.Duration(r.cfg.Service.ProcessPayoutsInterval) * time.Second

	volumesDuration := getRunnerInterval(volumesDefaultDuration, runners[RUNNER_PROC_VOLUME])
	refreshLeaderBoardDuration := getRunnerInterval(refreshLeaderBoardDefaultDuration, runners[RUNNER_PROC_LEADERBOARD])
	createPayoutsDuration := getRunnerInterval(createPayoutsDefaultDuration, runners[RUNNER_PROC_CREATE_PAYOUT])
	processPayoutsDuration := getRunnerInterval(processPayoutsDefaultDuration, runners[RUNNER_PROC_PROCESS_PAYOUT])

	doVolumesTimer := time.NewTimer(volumesDuration)
	doLeaderBoardRefreshTimer := time.NewTimer(refreshLeaderBoardDuration)
	doCreatePayoutsTimer := time.NewTimer(createPayoutsDuration)
	doProcessPayoutsTimer := time.NewTimer(processPayoutsDuration)

	for {
		select {
		case <-doVolumesTimer.C:
			err := r.doVolumes()
			if err != nil {
				logrus.Error("failed updateVolumes(): ", err)
			}
			doVolumesTimer = time.NewTimer(volumesDefaultDuration)

			err = r.dbManager.saveRunner(RUNNER_PROC_VOLUME)
			if err != nil {
				logrus.Error("failed saveRunner(): ", RUNNER_PROC_VOLUME, err)
			}

		case <-doLeaderBoardRefreshTimer.C:
			err := r.doRefreshLeaderBoards()
			if err != nil {
				logrus.Error("failed doRefreshLeaderBoards(): ", err)
			}
			doLeaderBoardRefreshTimer = time.NewTimer(refreshLeaderBoardDefaultDuration)

			err = r.dbManager.saveRunner(RUNNER_PROC_LEADERBOARD)
			if err != nil {
				logrus.Error("failed saveRunner(): ", RUNNER_PROC_LEADERBOARD, err)
			}

		case <-doCreatePayoutsTimer.C:
			err := r.doPayouts()
			if err != nil {
				logrus.Error("failed doPayouts(): ", err)
			}
			doCreatePayoutsTimer = time.NewTimer(createPayoutsDefaultDuration)

			err = r.dbManager.saveRunner(RUNNER_PROC_CREATE_PAYOUT)
			if err != nil {
				logrus.Error("failed saveRunner(payout): ", RUNNER_PROC_CREATE_PAYOUT, err)
			}

		case <-doProcessPayoutsTimer.C:
			err := r.doProcessPayouts()
			if err != nil {
				logrus.Error("failed doProcessPayouts(): ", err)
			}
			doProcessPayoutsTimer = time.NewTimer(processPayoutsDefaultDuration)

			err = r.dbManager.saveRunner(RUNNER_PROC_PROCESS_PAYOUT)
			if err != nil {
				logrus.Error("failed saveRunner(): ", RUNNER_PROC_PROCESS_PAYOUT, err)
			}
		}
	}
}
