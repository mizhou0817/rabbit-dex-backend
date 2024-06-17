package onboard_tier

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/pkg/log"
)

var (
	DefaultPeriodicsOptions = PeriodicsOptions{
		PeriodicsInterval: 2 * time.Second,
	}
)

type PeriodicsOptions struct {
	PeriodicsInterval time.Duration
	BatchSize         int
}

type Periodics struct {
	options            PeriodicsOptions
	store              Store
	marketsTierService MarketsTierService
	tierCalc           TierCalc
}

func NewPeriodics(
	store Store,
	marketsTierService MarketsTierService,
	tierCalc TierCalc,
	options PeriodicsOptions,
) *Periodics {
	return &Periodics{
		options:            options,
		store:              store,
		marketsTierService: marketsTierService,
		tierCalc:           tierCalc,
	}
}

func (p *Periodics) Run(ctx context.Context) error {
	var lastTs int64

	err := p.RunOnce(ctx, &lastTs)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		return errors.Wrap(err, "OnboardTier.Periodics.RunOnce")
	}

	ticker := time.NewTicker(p.options.PeriodicsInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			tm := time.Now()
			err := p.RunOnce(ctx, &lastTs)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
				return errors.Wrap(err, "OnboardTier.Periodics.RunOnce")
			}
			elapsed := time.Since(tm)
			if elapsed > p.options.PeriodicsInterval {
				logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("OnboardTier.Periodics runs too long: %s", elapsed.String())
			}
		}
	}
}

func (p *Periodics) RunOnce(ctx context.Context, lastTs *int64) error {
	logrus.Info("OnboardTier.Periodics.RunOnce starting")
	defer logrus.Info("OnboardTier.Periodics.RunOnce finished")

	if err := context.Cause(ctx); err != nil {
		return err
	}

	now := time.Now().UnixMicro()

	if lastTs == nil {
		return errors.Wrapf(ErrArgumentRequired, "lastTs should be specified")
	}
	ts := *lastTs

	profilesIds, err := p.store.GetProfilesIdsAfterCreatedAt(ctx, ts)
	if err != nil {
		return errors.Wrapf(err, "GetProfilesIdsAfterCreatedAt")
	}
	if len(profilesIds) == 0 {
		return nil
	}
	logrus.Infof("OnboardTier.Periodics Profiles found: %d, afterTs: %d", len(profilesIds), ts)

	if err := p.tierCalc.Recalculate(ctx, profilesIds); err != nil {
		return errors.Wrapf(err, "Recalculate")
	}

	for i, n, size := 0, p.options.BatchSize, len(profilesIds); i < size; i += n {
		if i+n > size || n == 0 {
			n = size - i
		}

		var profilesTiers []model.ProfileTier
		for _, profileId := range profilesIds[i : i+n] {
			tier, ok := p.tierCalc.GetProfileTier(profileId)
			if !ok {
				return errors.Wrapf(ErrProfileTierNotFound, "GetProfileTier: profile-id=%d", profileId)
			}
			profilesTiers = append(profilesTiers, tier)
		}

		if err := p.marketsTierService.UpdateProfilesToTiers(ctx, profilesTiers); err != nil {
			return errors.Wrap(err, "UpdateProfilesToTiers")
		}
	}

	*lastTs = now
	return nil
}

var (
	ErrArgumentRequired    = errors.New("argument is required")
	ErrProfileTierNotFound = errors.New("profile tier not found")
)
