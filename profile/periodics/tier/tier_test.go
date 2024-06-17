package tier_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/tier"

	"github.com/strips-finance/rabbit-dex-backend/profile/periodics/tier/mock"
)

func Test_Tier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	{
		profilesIds := []profile.ProfileId{0, 1, 2}
		profilesToTiers := map[profile.ProfileId]model.ProfileTier{
			profile.ProfileId(0): {ProfileID: 0, TierID: 2},
			profile.ProfileId(1): {ProfileID: 1, SpecialTierID: 9},
			profile.ProfileId(2): {ProfileID: 2, TierID: 0},
		}

		store := mock.NewMockStore(ctrl)
		marketsTierService := mock.NewMockMarketsTierService(ctrl)
		tierCalc := mock.NewMockTierCalc(ctrl)

		store.EXPECT().GetProfilesIdsAfterCreatedAt(ctx, gomock.Any()).Times(2).DoAndReturn(func(_ context.Context, afterTsMicro int64) ([]profile.ProfileId, error) {
			return profilesIds, nil
		})
		marketsTierService.EXPECT().UpdateProfilesToTiers(ctx, gomock.Any()).Times(4).DoAndReturn(func(_ context.Context, p2t []model.ProfileTier) error {
			for _, v := range p2t {
				require.Equal(t, profilesToTiers[v.ProfileID], v)
			}
			return nil
		})
		tierCalc.EXPECT().Recalculate(ctx, gomock.Any()).Times(2).Return(nil)
		tierCalc.EXPECT().GetProfileTier(gomock.Any()).Times(2 * len(profilesIds)).DoAndReturn(func(profileId profile.ProfileId) (model.ProfileTier, bool) {
			v, ok := profilesToTiers[profileId]
			return v, ok
		})

		options := tier.PeriodicsOptions{time.Second, 2}
		periodics := tier.NewPeriodics(store, marketsTierService, tierCalc, options)

		require.NoError(t, periodics.RunOnce(ctx))
		require.NoError(t, periodics.RunOnce(ctx))
	}
}
