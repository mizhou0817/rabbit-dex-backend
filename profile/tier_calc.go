//go:generate go run github.com/golang/mock/mockgen -source=$PWD/tier_calc.go -destination=$PWD/mock/tier_calc.go -package=mock
package profile

import (
	"context"
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"
)

type CalcVolumeCache interface {
	GetVolume(ProfileId) (decimal.Decimal, bool)
}

type ProfileTierService interface {
	GetTradingTiers(context.Context) ([]model.Tier, error)
	GetProfilesSpecialTiers(context.Context) ([]model.ProfileSpecialTier, error)
	GetAffiliateProfilesTiers(ctx context.Context, profilesIds ...uint) ([]model.AffiliateProfileTier, error)
}

type TierStore interface {
	GetReferralsByInvitedProfiles(context.Context, ...ProfileId) ([]tsdb.ReferralLink, error)
}

type TierCalcOptions struct {
	BatchSize int
}

type TierCalc struct {
	options            TierCalcOptions
	cache              CalcVolumeCache
	store              TierStore
	profileTierService ProfileTierService

	mu        sync.RWMutex
	tiersData map[ProfileId]tierCalcEntry
}

type tierCalcEntry struct {
	*model.TierStatusData
	*model.ProfileTier
}

func NewTierCalc(cache CalcVolumeCache, store TierStore, profileTierService ProfileTierService, options TierCalcOptions) *TierCalc {
	return &TierCalc{
		options:            options,
		cache:              cache,
		store:              store,
		profileTierService: profileTierService,

		tiersData: map[ProfileId]tierCalcEntry{},
	}
}

func (c *TierCalc) Recalculate(ctx context.Context, profilesIds []ProfileId) error {
	const (
		specialTierName = "Special tier"
	)
	tiers, err := c.profileTierService.GetTradingTiers(ctx)
	if err != nil {
		return errors.Wrapf(err, "GetTradingTiers")
	}
	slices.SortFunc(tiers, func(a, b model.Tier) int {
		if a.MinVolume.LessThan(b.MinVolume.Decimal) {
			return 1
		}
		if a.MinVolume.GreaterThan(b.MinVolume.Decimal) {
			return -1
		}
		return 0
	})

	type referralsType map[ProfileId]ProfileId
	var referralsMap referralsType
	{
		referrals, err := c.store.GetReferralsByInvitedProfiles(ctx, profilesIds...)
		if err != nil {
			return errors.Wrap(err, "GetReferralsByInvitedProfiles")
		}
		referralsMap = make(referralsType, len(referrals))
		for _, v := range referrals {
			referralsMap[v.InvitedId] = v.ProfileId
		}
	}

	type specTiersType map[ProfileId]model.ProfileSpecialTier
	var specTiersMap specTiersType
	{
		specTiers, err := c.profileTierService.GetProfilesSpecialTiers(ctx)
		if err != nil {
			return errors.Wrap(err, "GetProfilesSpecialTiers")
		}
		specTiersMap = make(specTiersType, len(specTiers))
		for _, v := range specTiers {
			specTiersMap[v.ProfileId] = v
		}
	}

	type affiliateTiersType map[ProfileId]model.AffiliateProfileTier
	var affiliateTiersMap affiliateTiersType
	{
		affiliateTiers, err := c.profileTierService.GetAffiliateProfilesTiers(ctx, profilesIds...)
		if err != nil {
			return errors.Wrap(err, "GetAffiliateProfilesTiers")
		}
		affiliateTiersMap = make(affiliateTiersType, len(affiliateTiers))
		for _, v := range affiliateTiers {
			affiliateTiersMap[v.ProfileId] = v
		}
	}

	for i, n, size := 0, c.options.BatchSize, len(profilesIds); i < size; i += n {
		if i+n > size || n == 0 {
			n = size - i
		}

		for _, profileId := range profilesIds[i : i+n] {
			curVolume, _ := c.cache.GetVolume(profileId)

			for {
				if tier, ok := specTiersMap[profileId]; ok {
					c.add(profileId, tierCalcEntry{
						ProfileTier: &model.ProfileTier{
							ProfileID:     profileId,
							SpecialTierID: tier.SpecialTierId,
						},
						TierStatusData: &model.TierStatusData{
							Current: model.DummyTier{
								Tier:      tier.SpecialTierId,
								Title:     specialTierName,
								MinVolume: decimal.Zero,
							},
						},
					})
					break // special case
				}

				if affiliateId, ok := referralsMap[profileId]; ok {
					affTier, ok := affiliateTiersMap[affiliateId]
					if ok {
						var current model.DummyTier
						var next *model.DummyTier
						var neededVolume *decimal.Decimal

						var specialTierId model.TierId
						var tierId model.TierId

						for _, tier := range tiers {
							if tier.Tier == affTier.ReplaceTierId {
								specialTierId = affTier.SpecialTierId
								current = model.DummyTier{
									Tier:      specialTierId,
									Title:     specialTierName,
									MinVolume: decimal.Zero,
								}
								break
							}
							if curVolume.GreaterThanOrEqual(tier.MinVolume.Decimal) {
								tierId = tier.Tier
								current = model.DummyTier{
									Tier:      tierId,
									Title:     tier.Title,
									MinVolume: tier.MinVolume.Decimal,
								}
								break
							}
							next = &model.DummyTier{
								Tier:      tier.Tier,
								Title:     tier.Title,
								MinVolume: tier.MinVolume.Decimal,
							}
						}

						if next != nil {
							volDiff := next.MinVolume.Sub(curVolume)
							neededVolume = &volDiff
						}

						c.add(profileId, tierCalcEntry{
							ProfileTier: &model.ProfileTier{
								ProfileID:     profileId,
								SpecialTierID: specialTierId,
								TierID:        tierId,
							},
							TierStatusData: &model.TierStatusData{
								Current:      current,
								Next:         next,
								NeededVolume: neededVolume,
							},
						})
						break // affiliate case
					}
				}

				{
					var current model.DummyTier
					var next *model.DummyTier
					var neededVolume *decimal.Decimal

					for _, tier := range tiers {
						if curVolume.GreaterThanOrEqual(tier.MinVolume.Decimal) {
							current = model.DummyTier{
								Tier:      tier.Tier,
								Title:     tier.Title,
								MinVolume: tier.MinVolume.Decimal,
							}
							break
						}
						next = &model.DummyTier{
							Tier:      tier.Tier,
							Title:     tier.Title,
							MinVolume: tier.MinVolume.Decimal,
						}
					}

					if next != nil {
						volDiff := next.MinVolume.Sub(curVolume)
						neededVolume = &volDiff
					}

					c.add(profileId, tierCalcEntry{
						ProfileTier: &model.ProfileTier{
							ProfileID: profileId,
							TierID:    current.Tier,
						},
						TierStatusData: &model.TierStatusData{
							Current:      current,
							Next:         next,
							NeededVolume: neededVolume,
						},
					})
				}
				break // default case
			}
		}
	}

	return nil
}

func (c *TierCalc) GetProfileTierStatus(profileId ProfileId) (model.TierStatusData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if tierEntry, ok := c.tiersData[profileId]; ok {
		return *tierEntry.TierStatusData, ok
	}

	return model.TierStatusData{}, false
}

func (c *TierCalc) GetProfileTier(profileId ProfileId) (model.ProfileTier, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if tierEntry, ok := c.tiersData[profileId]; ok {
		return *tierEntry.ProfileTier, ok
	}
	return model.ProfileTier{}, false
}

func (c *TierCalc) add(profileId ProfileId, tier tierCalcEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tiersData[profileId] = tier
}
