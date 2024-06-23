package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/cache"
	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/onboard_tier"
	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/tier"
	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"

	"github.com/strips-finance/rabbit-dex-backend/cmd/profile/periodics/config"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "profile-periodics.yaml"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Error(errors.Wrap(err, "UserHomeDir"))
		return
	}

	cfg, err := config.ReadConfig(homeDir, DefaultConfigPath, DefaultConfigFile)
	if err != nil {
		logrus.Error(errors.Wrap(err, "ReadConfig"))
		return
	}

	broker, err := model.GetBroker()
	if err != nil {
		logrus.Error(errors.Wrap(err, "GetBroker"))
		return
	}
	api := model.NewApiModel(broker)

	notifyClient, err := profile.NewNotifyClient(cfg.Service.Notify.GrpcAddress, profile.NotifyClientOptions{
		Retries:    cfg.Service.Notify.Retries,
		RetryDelay: cfg.Service.Notify.RetryDelay,
	})
	if err != nil {
		logrus.Error(errors.Wrap(err, "NewNotifyClient"))
		return
	}

	marketsClient := profile.NewMarketsClient(api, cfg.Service.Markets.Ids, profile.MarketsClientOptions{
		Workers:    cfg.Service.Markets.Workers,
		Retries:    cfg.Service.Markets.Retries,
		RetryDelay: cfg.Service.Markets.RetryDelay,
	})

	profileClient := profile.NewProfileClient(api, profile.ProfileClientOptions{
		Retries: cfg.Service.Profile.Retries,
	})

	pool, err := pgxpool.New(ctx, cfg.Service.TimescaledbConnectionURI)
	if err != nil {
		logrus.Error(errors.Wrap(err, "TimescaleDB.New"))
		return
	}
	store := tsdb.NewStore(pool)

	volumeCache := profile.NewVolumeStoreCache(store)
	if err := volumeCache.Refresh(ctx); err != nil {
		logrus.Error(errors.Wrap(err, "VolumeCache.Refresh"))
		return
	}

	tierCalc := profile.NewTierCalc(volumeCache, store, api, profile.TierCalcOptions{})

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		ticker := time.NewTicker(cfg.Service.VolumeCache.RefreshInterval)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := volumeCache.Refresh(ctx); err != nil {
					return errors.Wrap(err, "VolumeCache.Refresh")
				}
				logrus.Info("Volume cache has been refreshed")
			}
		}
	})

	g.Go(func() error {
		options := cache.PeriodicsOptions{
			MetaOnly:          cfg.Service.Cache.MetaOnly,
			PeriodicsInterval: cfg.Service.Cache.PeriodicsInterval,
			ParallelWorkers:   cfg.Service.Cache.ParallelWorkers,
			BatchSize:         cfg.Service.Cache.BatchSize,
		}

		p := cache.NewPeriodics(
			volumeCache,
			marketsClient,
			profileClient,
			notifyClient,
			profile.MarginV1LiquidateStrategy{
				ForcedMargin: cfg.Service.Cache.Liquidate.MarginV1.ForcedMargin,
			},
			tierCalc,
			options,
		)
		if err := p.Run(ctx); err != nil {
			return errors.Wrap(err, "Cache.Periodics.Run")
		}

		return nil
	})

	g.Go(func() error {
		options := onboard_tier.PeriodicsOptions{
			PeriodicsInterval: cfg.Service.OnboardTier.PeriodicsInterval,
			BatchSize:         int(cfg.Service.OnboardTier.BatchSize),
		}

		p := onboard_tier.NewPeriodics(
			store,
			marketsClient,
			tierCalc,
			options,
		)
		if err := p.Run(ctx); err != nil {
			return errors.Wrap(err, "OnboardTier.Periodics.Run")
		}

		return nil
	})

	g.Go(func() error {
		options := tier.PeriodicsOptions{
			PeriodicsInterval: cfg.Service.Tier.PeriodicsInterval,
			BatchSize:         int(cfg.Service.Tier.BatchSize),
		}

		p := tier.NewPeriodics(
			store,
			marketsClient,
			tierCalc,
			options,
		)
		if err := p.Run(ctx); err != nil {
			return errors.Wrap(err, "Tier.Periodics.Run")
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		logrus.Error(err)
	}
}
