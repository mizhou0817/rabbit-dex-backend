//go:generate go run github.com/golang/mock/mockgen -source=$PWD/interfaces.go -destination=$PWD/mock/mocks.go -package=mock
package tier

import (
	"context"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"
)

type Store interface {
	GetProfilesIdsAfterCreatedAt(ctx context.Context, afterTsMicro int64) ([]profile.ProfileId, error)
	GetVolumesAggregatesLast30d(context.Context) ([]tsdb.CumVolume, error)
	GetReferralsByInvitedProfiles(context.Context, ...profile.ProfileId) ([]tsdb.ReferralLink, error)
}

type MarketsTierService interface {
	UpdateProfilesToTiers(ctx context.Context, profilesToTiers []model.ProfileTier) error
}

type TierCalc interface {
	Recalculate(ctx context.Context, profilesIds []profile.ProfileId) error
	GetProfileTier(profileId profile.ProfileId) (model.ProfileTier, bool)
}
