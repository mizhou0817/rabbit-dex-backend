package tier

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
		PeriodicsInterval: time.Hour,
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
	err := p.RunOnce(ctx)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Error(err)
		return errors.Wrap(err, "Tier.Periodics.RunOnce")
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
				return errors.Wrap(err, "Tier.Periodics.RunOnce")
			}
			elapsed := time.Since(tm)
			if elapsed > p.options.PeriodicsInterval {
				logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("Tier.Periodics runs too long: %s", elapsed.String())
			}
		}
	}
}

func (p *Periodics) RunOnce(ctx context.Context) error {
	logrus.Info("Tier.Periodics.RunOnce starting")
	defer logrus.Info("Tier.Periodics.RunOnce finished")

	if err := context.Cause(ctx); err != nil {
		return err
	}

	profilesIds, err := p.store.GetProfilesIdsAfterCreatedAt(ctx, 0)
	if err != nil {
		return errors.Wrapf(err, "GetProfilesIdsAfterCreatedAt")
	}
	if len(profilesIds) == 0 {
		return nil
	}
	logrus.Infof("Tier.Periodics Profiles found: %d", len(profilesIds))

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

	return nil
}

var (
	ErrProfileTierNotFound = errors.New("profile tier not found")
)
