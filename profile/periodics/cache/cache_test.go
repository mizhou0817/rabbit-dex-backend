package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/cache"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"

	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/cache/mock"
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
)

func Test_Tier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	{
		marketsIds := []profile.MarketId{"BTC-USD", "ETH-USD"}
		profilesIds := []profile.ProfileId{0, 1, 2}
		extendedProfiles := []*model.ExtendedProfile{
			{
				Profile: model.Profile{
					ProfileId: 0,
					Type:      model.PROFILE_TYPE_TRADER,
					Status:    model.PROFILE_STATUS_ACTIVE,
					Wallet:    "0xinsurance",
					CreatedAt: 123,
				},
				Balance: *tdecimal.NewDecimal(decimal.NewFromFloat(123)),
			},
			{
				Profile: model.Profile{
					ProfileId: 1,
					Type:      model.PROFILE_TYPE_TRADER,
					Status:    model.PROFILE_STATUS_ACTIVE,
					Wallet:    "0xtrader1",
					CreatedAt: 1234,
				},
				Balance: *tdecimal.NewDecimal(decimal.NewFromFloat(1234)),
			},
			{
				Profile: model.Profile{
					ProfileId: 2,
					Type:      model.PROFILE_TYPE_TRADER,
					Status:    model.PROFILE_STATUS_ACTIVE,
					Wallet:    "0xtrader2",
					CreatedAt: 12345,
				},
				Balance: *tdecimal.NewDecimal(decimal.NewFromFloat(12345)),
			},
		}
		profilesMeta := profile.ProfilesMeta{}
		cumVolumes := map[profile.ProfileId]decimal.Decimal{
			0: decimal.NewFromFloat(5000001),
			1: decimal.NewFromFloat(4999999),
			2: decimal.NewFromFloat(10000000),
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
		}
		_ = profilesIds

		volCache := mock.NewMockVolumeCache(ctrl)
		marketMetaService := mock.NewMockMarketMetaService(ctrl)
		profileCacheService := mock.NewMockProfileCacheService(ctrl)
		notifyService := mock.NewMockNotifyService(ctrl)
		liquidateStrategy := mock.NewMockLiquidateStrategy(ctrl)
		tierCalc := mock.NewMockTierCalc(ctrl)

		volCache.EXPECT().GetVolume(gomock.Any()).Times(2 * len(cumVolumes)).DoAndReturn(func(p profile.ProfileId) (decimal.Decimal, bool) {
			v, ok := cumVolumes[p]
			return v, ok
		})
		marketMetaService.EXPECT().GetMarketsIds().AnyTimes().Return(marketsIds)
		marketMetaService.EXPECT().GetUpdatedProfilesMeta(ctx, gomock.Any()).AnyTimes().Return(profilesMeta, profile.MarketsLastTs{}, nil)
		profileCacheService.EXPECT().GetExtendedProfiles(ctx).AnyTimes().Return(extendedProfiles, nil)
		profileCacheService.EXPECT().UpdateProfilesCachesMetas(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(_ context.Context, data []*model.ProfileCacheMetas) error {
			return nil
		})
		notifyService.EXPECT().PublishExtendedProfiles(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(_ context.Context, data []*model.ExtendedProfileTierStatusData) error {
			return nil
		})
		liquidateStrategy.EXPECT().Process(gomock.Any()).AnyTimes().DoAndReturn(func(pc *model.ProfileCache) error {
			if pc.ProfileID == 1 {
				s := model.PROFILE_STATUS_LIQUIDATING
				pc.Status = &s
			}
			return nil
		})
		tierCalc.EXPECT().GetProfileTierStatus(gomock.Any()).AnyTimes().DoAndReturn(func(profileId profile.ProfileId) (model.TierStatusData, bool) {
			v, ok := expectedStatuses[profileId]
			return v, ok
		})

		options := cache.PeriodicsOptions{BatchSize: 2, PeriodicsInterval: time.Second, ParallelWorkers: 10}
		periodics := cache.NewPeriodics(volCache, marketMetaService, profileCacheService, notifyService, liquidateStrategy, tierCalc, options)

		require.NoError(t, periodics.RunOnce(ctx))
		require.NoError(t, periodics.RunOnce(ctx))
	}
}
