package model

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	BALANCE_OPS_LIST    = "getters.list_operations"
	UPDATE_MARKET_URL   = "market.update_icon_url"
	UPDATE_MARKET_TITLE = "market.update_market_title"

	ADD_TIER    = "profile.add_tier"
	REMOVE_TIER = "profile.remove_tier"
	GET_TIERS   = "profile.get_tiers"
	EDIT_TIER   = "profile.edit_tier"

	ADD_SPECIAL_TIER    = "profile.add_special_tier"
	REMOVE_SPECIAL_TIER = "profile.remove_special_tier"
	GET_SPECIAL_TIERS   = "profile.get_special_tiers"
	EDIT_SPECIAL_TIER   = "profile.edit_special_tier"

	ADD_PROFILE_TIER    = "profile.add_profile_to_special_tier"
	REMOVE_PROFILE_TIER = "profile.remove_profile_from_tier"
	GET_PROFILE_TIERS   = "profile.get_profile_tiers"

	WHICH_TIER = "profile.getter_which_tier"
)

type Tier struct {
	Tier      uint             `msgpack:"tier" json:"tier"`
	Title     string           `msgpack:"title" json:"title"`
	MakerFee  tdecimal.Decimal `msgpack:"maker_fee" json:"maker_fee"`
	TakerFee  tdecimal.Decimal `msgpack:"taker_fee" json:"taker_fee"`
	MinVolume tdecimal.Decimal `msgpack:"min_volume" json:"min_volume"`
	MinAssets tdecimal.Decimal `msgpack:"min_assets" json:"min_assets"`
}

type SpecialTier struct {
	Tier     uint             `msgpack:"tier" json:"tier"`
	Title    string           `msgpack:"title" json:"title"`
	MakerFee tdecimal.Decimal `msgpack:"maker_fee" json:"maker_fee"`
	TakerFee tdecimal.Decimal `msgpack:"taker_fee" json:"taker_fee"`
}

type ProfileTier struct {
	ProfileID     uint `msgpack:"profile_id" json:"profile_id"`
	SpecialTierID uint `msgpack:"special_tier_id" json:"special_tier_id"`
	TierID        uint `msgpack:"tier_id" json:"tier_id"`
}

type TierData struct {
	MarketId string      `json:"market_id"`
	TierData SpecialTier `json:"tier_data"`
}

/*
function PM.add_tier(tier_id, title, maker_fee, taker_fee, min_volume, min_assets)
function PM.remove_tier(tier_id)
function PM.add_special_tier(tier_id, title, maker_fee, taker_fee)
function PM.remove_special_tier(tier_id)
function PM.add_profile_to_special_tier(profile_id, tier_id)
function PM.remove_profile_from_tier(profile_id)
function PM.get_tiers()
function PM.get_special_tiers()
function PM.get_profile_tiers()
*/

/* Tiers management */
func (api *ApiModel) GetTiers(ctx context.Context, marketId string) ([]Tier, error) {
	title := PROFILE_INSTANCE
	if marketId != title {
		instance, err := GetInstance().ByMarketID(marketId)
		if err != nil {
			text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
			return nil, errors.New(text)
		}
		title = instance.Title
	}

	return DataResponse[[]Tier]{}.Request(ctx, title, api.broker, GET_TIERS, []interface{}{})
}

func (api *ApiModel) GetSpecialTiers(ctx context.Context, marketId string) ([]SpecialTier, error) {
	title := PROFILE_INSTANCE
	instance, err := GetInstance().ByMarketID(marketId)
	if marketId != title {
		if err != nil {
			text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
			return nil, errors.New(text)
		}
		title = instance.Title
	}

	return DataResponse[[]SpecialTier]{}.Request(ctx, title, api.broker, GET_SPECIAL_TIERS, []interface{}{})
}

func (api *ApiModel) GetProfileTiers(ctx context.Context, marketId string) ([]ProfileTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[[]ProfileTier]{}.Request(ctx, instance.Title, api.broker, GET_PROFILE_TIERS, []interface{}{})
}

func (api *ApiModel) AddTier(ctx context.Context, marketId string, tierId uint, title string, makerFee, takerFee, minVolume, minAssets float64) (*Tier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	dMakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(makerFee))
	dTakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(takerFee))
	dMinVolume := tdecimal.NewDecimal(decimal.NewFromFloat(minVolume))
	dMinAssets := tdecimal.NewDecimal(decimal.NewFromFloat(minAssets))

	return DataResponse[*Tier]{}.Request(ctx, instance.Title, api.broker, ADD_TIER, []interface{}{
		tierId,
		title,
		dMakerFee,
		dTakerFee,
		dMinVolume,
		dMinAssets,
	})
}

func (api *ApiModel) AddSpecialTier(ctx context.Context, marketId string, tierId uint, title string, makerFee, takerFee float64) (*SpecialTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	dMakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(makerFee))
	dTakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(takerFee))

	return DataResponse[*SpecialTier]{}.Request(ctx, instance.Title, api.broker, ADD_SPECIAL_TIER, []interface{}{
		tierId,
		title,
		dMakerFee,
		dTakerFee,
	})
}

func (api *ApiModel) AddProfileToTier(ctx context.Context, marketId string, profileId, tierId uint) (*ProfileTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*ProfileTier]{}.Request(ctx, instance.Title, api.broker, ADD_PROFILE_TIER, []interface{}{
		profileId,
		tierId,
	})
}

func (api *ApiModel) EditTier(ctx context.Context, marketId string, tierId uint, title string, makerFee, takerFee, minVolume, minAssets float64) (*Tier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	dMakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(makerFee))
	dTakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(takerFee))
	dMinVolume := tdecimal.NewDecimal(decimal.NewFromFloat(minVolume))
	dMinAssets := tdecimal.NewDecimal(decimal.NewFromFloat(minAssets))

	return DataResponse[*Tier]{}.Request(ctx, instance.Title, api.broker, EDIT_TIER, []interface{}{
		tierId,
		title,
		dMakerFee,
		dTakerFee,
		dMinVolume,
		dMinAssets,
	})
}

func (api *ApiModel) EditSpecialTier(ctx context.Context, marketId string, tierId uint, title string, makerFee, takerFee float64) (*SpecialTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	dMakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(makerFee))
	dTakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(takerFee))

	return DataResponse[*SpecialTier]{}.Request(ctx, instance.Title, api.broker, EDIT_SPECIAL_TIER, []interface{}{
		tierId,
		title,
		dMakerFee,
		dTakerFee,
	})
}

func (api *ApiModel) RemoveTier(ctx context.Context, marketId string, tierId uint) (*Tier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*Tier]{}.Request(ctx, instance.Title, api.broker, REMOVE_TIER, []interface{}{
		tierId,
	})
}

func (api *ApiModel) RemoveSpecialTier(ctx context.Context, marketId string, tierId uint) (*SpecialTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*SpecialTier]{}.Request(ctx, instance.Title, api.broker, REMOVE_SPECIAL_TIER, []interface{}{
		tierId,
	})
}

func (api *ApiModel) RemoveProfileTier(ctx context.Context, marketId string, profileId uint) (*ProfileTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*ProfileTier]{}.Request(ctx, instance.Title, api.broker, REMOVE_PROFILE_TIER, []interface{}{
		profileId,
	})
}

func (api *ApiModel) BalanceOpsList(ctx context.Context, profileId uint, offset, limit uint) ([]*BalanceOps, error) {
	ops, err := DataResponse[[]*BalanceOps]{}.Request(
		ctx,
		PROFILE_INSTANCE,
		api.broker,
		BALANCE_OPS_LIST,
		[]interface{}{
			profileId,
			offset,
			limit,
		},
	)

	return ops, err
}

func (api *ApiModel) MarketUpdateIconUrl(ctx context.Context, marketId string, newUrl string) (*MarketData, error) {
	_, err := url.ParseRequestURI(newUrl)
	if err != nil {
		return nil, err
	}

	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*MarketData]{}.Request(ctx, instance.Title, api.broker, UPDATE_MARKET_URL, []interface{}{
		marketId,
		newUrl,
	})
}

func (api *ApiModel) MarketUpdateTitle(ctx context.Context, marketId string, newTitle string) (*MarketData, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return nil, errors.New(text)
	}

	return DataResponse[*MarketData]{}.Request(ctx, instance.Title, api.broker, UPDATE_MARKET_TITLE, []interface{}{
		marketId,
		newTitle,
	})

}

func (api *ApiModel) WhichTier(ctx context.Context, marketId string, profileId uint) (SpecialTier, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), marketId)
		return SpecialTier{}, errors.New(text)
	}

	return DataResponse[SpecialTier]{}.Request(ctx, instance.Title, api.broker, WHICH_TIER, []interface{}{
		profileId,
	})

}
