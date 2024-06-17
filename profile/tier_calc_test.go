package profile_test

import (
	"context"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/profile/mock"
	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

var (
	tiers []model.Tier = []model.Tier{
		{
			Tier:      0,
			Title:     "VIP 0 (Shrimp)",
			MakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0)),
			TakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0.0007)),
			MinVolume: *tdecimal.NewDecimal(decimal.NewFromFloat(0)),
			MinAssets: *tdecimal.NewDecimal(decimal.NewFromFloat(0)),
		},
		{
			Tier:      1,
			Title:     "VIP 1 (Herring)",
			MakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0)),
			TakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0.0005)),
			MinVolume: *tdecimal.NewDecimal(decimal.NewFromFloat(1000000)),
			MinAssets: *tdecimal.NewDecimal(decimal.NewFromFloat(1000000)),
		},
		{
			Tier:      2,
			Title:     "VIP 2 (Trout)",
			MakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0)),
			TakerFee:  *tdecimal.NewDecimal(decimal.NewFromFloat(0.0045)),
			MinVolume: *tdecimal.NewDecimal(decimal.NewFromFloat(5000000)),
			MinAssets: *tdecimal.NewDecimal(decimal.NewFromFloat(5000000)),
		},
	}

	affiliateProfilesTiers []model.AffiliateProfileTier = []model.AffiliateProfileTier{
		{ProfileId: 100, SpecialTierId: 3, ReplaceTierId: 1},
	}
)

func Test_TierCalc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	profilesIds := []profile.ProfileId{0, 1, 2, 3, 4, 5}
	referralLinks := []tsdb.ReferralLink{
		{ProfileId: 100, InvitedId: 3},
		{ProfileId: 100, InvitedId: 5},
	}
	profilesSpecialTiers := []model.ProfileSpecialTier{
		{ProfileId: 1, SpecialTierId: 1},
	}
	cumVolumes := map[profile.ProfileId]decimal.Decimal{
		0: decimal.Zero,
		1: decimal.NewFromFloat(5000001),
		2: decimal.NewFromFloat(4999999),
		3: decimal.NewFromFloat(10000000),
		4: decimal.NewFromFloat(999999999),
		5: decimal.NewFromFloat(999999),
	}

	volCache := mock.NewMockCalcVolumeCache(ctrl)
	tierStore := mock.NewMockTierStore(ctrl)
	profileTierService := mock.NewMockProfileTierService(ctrl)

	volCache.EXPECT().GetVolume(gomock.Any()).Times(len(cumVolumes)).DoAndReturn(func(p profile.ProfileId) (decimal.Decimal, bool) {
		v, ok := cumVolumes[p]
		return v, ok
	})
	tierStore.EXPECT().GetReferralsByInvitedProfiles(ctx, toAnys(profilesIds)...).Return(referralLinks, nil)

	profileTierService.EXPECT().GetTradingTiers(ctx).DoAndReturn(func(context.Context) ([]model.Tier, error) {
		tiersCopy := make([]model.Tier, len(tiers)) // will be sorted by volume in Recalculate
		copy(tiersCopy, tiers)
		return tiersCopy, nil
	})
	profileTierService.EXPECT().GetProfilesSpecialTiers(ctx).Return(profilesSpecialTiers, nil)
	profileTierService.EXPECT().GetAffiliateProfilesTiers(ctx, toAnys(profilesIds)...).Return(affiliateProfilesTiers, nil)

	options := profile.TierCalcOptions{BatchSize: 1}
	tc := profile.NewTierCalc(volCache, tierStore, profileTierService, options)

	require.NoError(t, tc.Recalculate(ctx, profilesIds))

	expectedTiers := []model.ProfileTier{
		{ProfileID: 0, TierID: 0},
		{ProfileID: 1, SpecialTierID: 1},
		{ProfileID: 2, TierID: 1},
		{ProfileID: 3, TierID: 2},
		{ProfileID: 4, TierID: 2},
		{ProfileID: 5, SpecialTierID: 3},
	}
	for _, profileId := range profilesIds {
		tier, okTier := tc.GetProfileTier(profileId)
		require.True(t, okTier)
		require.Equal(t, expectedTiers[profileId], tier)
	}

	expectedStatuses := map[profile.ProfileId]model.TierStatusData{
		profile.ProfileId(0): {
			Current: model.DummyTier{
				Tier:      tiers[0].Tier,
				Title:     tiers[0].Title,
				MinVolume: tiers[0].MinVolume.Decimal,
			},
			Next: &model.DummyTier{
				Tier:      tiers[1].Tier,
				Title:     tiers[1].Title,
				MinVolume: tiers[1].MinVolume.Decimal,
			},
			NeededVolume: func() *decimal.Decimal {
				vol := tiers[1].MinVolume.Decimal.Sub(cumVolumes[profile.ProfileId(0)])
				return &vol
			}(),
		},
		profile.ProfileId(1): {
			Current: model.DummyTier{
				Tier:      1,
				Title:     "Special tier",
				MinVolume: decimal.Zero,
			},
			Next:         nil,
			NeededVolume: nil,
		},
		profile.ProfileId(2): {
			Current: model.DummyTier{
				Tier:      tiers[1].Tier,
				Title:     tiers[1].Title,
				MinVolume: tiers[1].MinVolume.Decimal,
			},
			Next: &model.DummyTier{
				Tier:      tiers[2].Tier,
				Title:     tiers[2].Title,
				MinVolume: tiers[2].MinVolume.Decimal,
			},
			NeededVolume: func() *decimal.Decimal {
				vol := tiers[2].MinVolume.Decimal.Sub(cumVolumes[profile.ProfileId(2)])
				return &vol
			}(),
		},
		profile.ProfileId(3): {
			Current: model.DummyTier{
				Tier:      tiers[2].Tier,
				Title:     tiers[2].Title,
				MinVolume: tiers[2].MinVolume.Decimal,
			},
			// TierId(2) the last one
			Next:         nil,
			NeededVolume: nil,
		},
		profile.ProfileId(4): {
			Current: model.DummyTier{
				Tier:      tiers[2].Tier,
				Title:     tiers[2].Title,
				MinVolume: tiers[2].MinVolume.Decimal,
			},
			// TierId(2) the last one
			Next:         nil,
			NeededVolume: nil,
		},
		profile.ProfileId(5): {
			Current: model.DummyTier{
				Tier:      3,
				Title:     "Special tier",
				MinVolume: decimal.Zero,
			},
			Next: &model.DummyTier{
				Tier:      tiers[2].Tier,
				Title:     tiers[2].Title,
				MinVolume: tiers[2].MinVolume.Decimal,
			},
			NeededVolume: func() *decimal.Decimal {
				vol := tiers[2].MinVolume.Decimal.Sub(cumVolumes[profile.ProfileId(5)])
				return &vol
			}(),
		},
	}
	for _, profileId := range profilesIds {
		tierStatus, okTierStatus := tc.GetProfileTierStatus(profileId)
		require.True(t, okTierStatus)
		require.Nil(t, deep.Equal(expectedStatuses[profileId], tierStatus), "profileId=%d", profileId)
	}
}

// testing helpers

func toAnys[T any](s []T) (anys []any) {
	for _, v := range s {
		anys = append(anys, v)
	}
	return anys
}
