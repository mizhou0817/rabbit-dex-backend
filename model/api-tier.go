package model

import (
	"context"

	"github.com/pkg/errors"
)

// the most part of tier logic in api-admin.go
// need to do audit for it

type TierId = uint

type AffiliateProfileTier struct {
	ProfileId     uint   `msgpack:"profile_id"`
	SpecialTierId TierId `msgpack:"special_tier_id"`
	ReplaceTierId TierId `msgpack:"replace_tier_id"`
}

type ProfileSpecialTier struct {
	ProfileId     uint   `msgpack:"profile_id"`
	SpecialTierId TierId `msgpack:"special_tier_id"`
}

func (api *ApiModel) GetTradingTiers(ctx context.Context) ([]Tier, error) {
	tiers, err := DataResponse[[]Tier]{}.Request(ctx, PROFILE_INSTANCE, api.broker, "getters.get_tiers", []any{})

	return tiers, err
}

func (api *ApiModel) GetAffiliateProfilesTiers(ctx context.Context, profilesIds ...uint) ([]AffiliateProfileTier, error) {
	tiers, err := DataResponse[[]AffiliateProfileTier]{}.Request(ctx, PROFILE_INSTANCE, api.broker, "getters.get_affiliate_profiles_tiers", []any{
		profilesIds,
	})

	return tiers, err
}

func (api *ApiModel) GetProfilesSpecialTiers(ctx context.Context) ([]ProfileSpecialTier, error) {
	tiers, err := DataResponse[[]ProfileSpecialTier]{}.Request(ctx, PROFILE_INSTANCE, api.broker, "getters.get_profiles_special_tiers", []any{})

	return tiers, err
}

func (api *ApiModel) UpdateMarketProfilesToTiers(ctx context.Context, marketId string, profilesToTiers []ProfileTier) error {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		return errors.Wrapf(err, "GetInstance")
	}

	_, err = DataResponse[any]{}.Request(ctx, instance.Title, api.broker, "setters.update_profiles_to_tiers", []any{
		profilesToTiers,
	})

	return err
}
