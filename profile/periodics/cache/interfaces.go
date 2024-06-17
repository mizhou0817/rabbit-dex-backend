//go:generate go run github.com/golang/mock/mockgen -source=$PWD/interfaces.go -destination=$PWD/mock/mocks.go -package=mock
package cache

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/profile"
)

type MarketMetaService interface {
	GetMarketsIds() []profile.MarketId
	GetUpdatedProfilesMeta(context.Context, profile.MarketsLastTs) (profile.ProfilesMeta, profile.MarketsLastTs, error)
}

type ProfileCacheService interface {
	GetExtendedProfiles(ctx context.Context, profilesIds ...profile.ProfileId) ([]*model.ExtendedProfile, error)
	UpdateProfilesCachesMetas(ctx context.Context, data []*model.ProfileCacheMetas) error
}

type NotifyService interface {
	PublishExtendedProfiles(ctx context.Context, data []*model.ExtendedProfileTierStatusData) error
}

type LiquidateStrategy interface {
	Process(pc *model.ProfileCache) error
}

type VolumeCache interface {
	GetVolume(profileId profile.ProfileId) (decimal.Decimal, bool)
}

type TierCalc interface {
	GetProfileTierStatus(profileId profile.ProfileId) (model.TierStatusData, bool)
}
