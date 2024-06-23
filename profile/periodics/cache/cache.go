package cache

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/alitto/pond"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
	profilepkg "github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

var (
	DefaultPeriodicsOptions = PeriodicsOptions{
		BatchSize:         1000,
		ParallelWorkers:   10,
		PeriodicsInterval: 5 * time.Second,
	}
)

type PeriodicsOptions struct {
	// Update only caches for updated metas
	MetaOnly bool
	// Number of profiles to update with one call
	BatchSize uint
	// Number of workers to calculate profiles cache from meta
	ParallelWorkers uint
	// How often fetch/calculate/update profiles cache
	PeriodicsInterval time.Duration
}

type Periodics struct {
	volumeCache       VolumeCache
	metaService       MarketMetaService
	profileService    ProfileCacheService
	notifyService     NotifyService
	liquidateStrategy LiquidateStrategy
	tierCalc          TierCalc
	options           PeriodicsOptions
	marketsLastTs     profilepkg.MarketsLastTs
	profilesMeta      profilepkg.ProfilesMeta
	pool              *pond.WorkerPool
}

func NewPeriodics(
	volumeCache VolumeCache,
	metaService MarketMetaService,
	profileService ProfileCacheService,
	notifyService NotifyService,
	liquidateStrategy LiquidateStrategy,
	tierCalc TierCalc,
	options PeriodicsOptions,
) *Periodics {
	return &Periodics{
		volumeCache:       volumeCache,
		metaService:       metaService,
		profileService:    profileService,
		notifyService:     notifyService,
		liquidateStrategy: liquidateStrategy,
		tierCalc:          tierCalc,
		options:           options,
		marketsLastTs:     make(profilepkg.MarketsLastTs),
		pool:              pond.New(int(options.ParallelWorkers), 0),
	}
}

func (p *Periodics) Run(ctx context.Context) error {
	err := p.RunOnce(ctx)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		return errors.Wrap(err, "Cache.Periodics.RunOnce")
	}

	ticker := time.NewTicker(p.options.PeriodicsInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			tm := time.Now()
			err := p.RunOnce(ctx)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
				return errors.Wrap(err, "Cache.Periodics.RunOnce")
			}
			elapsed := time.Since(tm)
			if elapsed > p.options.PeriodicsInterval {
				logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("Cache.Periodics runs too long: %s", elapsed.String())
			}
		}
	}
}

func (p *Periodics) RunOnce(ctx context.Context) error {
	logrus.Info("Cache.Periodics.RunOnce starting")
	defer logrus.Info("Cache.Periodics.RunOnce finished")

	if err := context.Cause(ctx); err != nil {
		return err
	}

	profilesMeta, marketsLastTs, err := p.metaService.GetUpdatedProfilesMeta(ctx, p.marketsLastTs)
	if err != nil {
		return errors.Wrap(err, "GetUpdatedProfilesMeta")
	}

	if p.profilesMeta == nil {
		p.profilesMeta = profilesMeta
	} else {
		// merge with profiles meta from previous iteration
		// because meta from all markets is required to calculate profile cache
		for profileId, metas := range profilesMeta {
			destMetas := p.profilesMeta[profileId]
			if destMetas == nil {
				logrus.Warn(errors.Wrapf(ErrProfileMetaNotFound, "Merge Profile id=%d", profileId))
				destMetas = make(map[profilepkg.MarketId]*model.ProfileMeta)
				p.profilesMeta[profileId] = destMetas
			}
			for marketId, mt := range metas {
				destMetas[marketId] = mt
			}
		}
	}

	var profilesIds []profilepkg.ProfileId
	if p.options.MetaOnly {
		for profileId := range profilesMeta {
			profilesIds = append(profilesIds, profileId)
		}
	}
	profiles, err := p.profileService.GetExtendedProfiles(ctx, profilesIds...)
	if err != nil {
		return errors.Wrap(err, "GetExtendedProfiles")
	}

	group, ctx := p.pool.GroupContext(ctx)

	for i, n, size := 0, int(p.options.BatchSize), len(profiles); i < size; i += n {
		if i+n > size {
			n = size - i
		}

		profilesBatch := profiles[i : i+n]

		group.Submit(func() error {
			var (
				profilesCachesMetas []*model.ProfileCacheMetas
				extendedProfiles    []*model.ExtendedProfileTierStatusData
			)
			for _, profile := range profilesBatch {
				metas, ok := p.profilesMeta[profile.ProfileId]
				if !ok {
					metas = profilepkg.ProfileMarketsMeta{}
				}

				vol, _ := p.volumeCache.GetVolume(profile.ProfileId)
				pc := p.get(profile, vol, metas)
				if err := p.liquidateStrategy.Process(pc); err != nil {
					return errors.Wrap(err, "Liquidate.Process")
				}
				profilesCachesMetas = append(profilesCachesMetas, &model.ProfileCacheMetas{
					Cache: pc,
					Metas: metas,
				})

				tierStatus, _ := p.tierCalc.GetProfileTierStatus(profile.ProfileId)
				extendedProfiles = append(extendedProfiles, &model.ExtendedProfileTierStatusData{
					ExtendedProfileData: model.ExtendedProfileData{
						ProfileCache: *pc,
					},
					TierStatusData: tierStatus,
				})
			}

			if len(profilesCachesMetas) > 0 {
				if err := p.profileService.UpdateProfilesCachesMetas(ctx, profilesCachesMetas); err != nil {
					return errors.Wrap(err, "UpdateProfilesCachesMetas")
				}
			}

			if len(extendedProfiles) > 0 {
				if err := p.notifyService.PublishExtendedProfiles(ctx, extendedProfiles); err != nil {
					return errors.Wrap(err, "PublishExtendedProfiles")
				}
			}

			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	p.marketsLastTs = marketsLastTs
	return nil
}

func (p *Periodics) get(profile *model.ExtendedProfile, volume decimal.Decimal, metas map[profilepkg.MarketId]*model.ProfileMeta) *model.ProfileCache {
	decimalOne := decimal.NewFromInt(1)
	marketsIds := p.metaService.GetMarketsIds()
	var (
		profileType = profile.Type
		status      = model.PROFILE_STATUS_ACTIVE
		wallet      = profile.Wallet
		balance     = profile.Balance.Decimal

		accountEquity       = decimal.Zero
		totalPositionMargin = decimal.Zero
		totalOrderMargin    = decimal.Zero
		totalNotional       = decimal.Zero
		accountMargin       = decimalOne
		withdrawableBalance = decimal.Zero
		cumUnrealizedPnl    = decimal.Zero
		health              = decimal.Zero
		accountLeverage     = decimalOne
		leverage            = make(map[string]*tdecimal.Decimal, len(marketsIds))
	)
	for _, marketId := range marketsIds {
		leverage[marketId] = tdecimal.NewDecimal(decimalOne)
	}

	for _, mt := range metas {
		balance = balance.Add(mt.Balance.Decimal)
		totalPositionMargin = totalPositionMargin.Add(mt.TotalPositionMargin.Decimal)
		totalOrderMargin = totalOrderMargin.Add(mt.TotalOrderMargin.Decimal)
		totalNotional = totalNotional.Add(mt.TotalNotional.Decimal)
		cumUnrealizedPnl = cumUnrealizedPnl.Add(mt.CumUnrealizedPnl.Decimal)
		leverage[mt.MarketID] = tdecimal.NewDecimal(mt.MarketLeverage.Decimal)
	}

	accountEquity = balance.Add(cumUnrealizedPnl)
	withdrawableBalance = decimal.Min(accountEquity, balance).
		Sub(totalPositionMargin).
		Sub(totalOrderMargin)

	if !totalNotional.Equals(decimal.Zero) {
		accountMargin = accountEquity.Div(totalNotional)
	}

	if !accountMargin.Equals(decimal.Zero) {
		accountLeverage = decimalOne.Div(accountMargin)
		if accountMargin.GreaterThan(decimal.Zero) {
			health = decimal.Min(decimalOne, accountMargin)
		}
	}

	zero := int64(0)
	return &model.ProfileCache{
		ProfileID:           profile.ProfileId,
		ProfileType:         &profileType,
		Status:              &status,
		Wallet:              &wallet,
		LastUpdate:          &zero,
		Balance:             tdecimal.NewDecimal(balance),
		AccountEquity:       tdecimal.NewDecimal(accountEquity),
		TotalPositionMargin: tdecimal.NewDecimal(totalPositionMargin),
		TotalOrderMargin:    tdecimal.NewDecimal(totalOrderMargin),
		TotalNotional:       tdecimal.NewDecimal(totalNotional),
		AccountMargin:       tdecimal.NewDecimal(accountMargin),
		WithdrawableBalance: tdecimal.NewDecimal(withdrawableBalance),
		CumUnrealizedPnl:    tdecimal.NewDecimal(cumUnrealizedPnl),
		Health:              tdecimal.NewDecimal(health),
		AccountLeverage:     tdecimal.NewDecimal(accountLeverage),
		CumTradingVolume:    tdecimal.NewDecimal(volume),
		Leverage:            leverage,
		// currently this field should be ignored in profile.update_cache_and_meta func
		// because liqengine updates it on profile instance itself
		LastLiqCheck: &zero,
	}
}

var (
	ErrProfileMetaNotFound = errors.New("profile meta not found")
)
