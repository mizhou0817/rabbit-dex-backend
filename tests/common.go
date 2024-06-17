package tests

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/helpers"
	"github.com/strips-finance/rabbit-dex-backend/model"
	profilepkg "github.com/strips-finance/rabbit-dex-backend/profile"
	tierpkg "github.com/strips-finance/rabbit-dex-backend/profile/periodics/tier"
	"github.com/strips-finance/rabbit-dex-backend/profile/tsdb"
)

type ClearOption func(*clearOptions)

type clearOptions struct {
	skipSpaces    []string
	skipInstances []string
}

func SkipSpaces(spaces ...string) ClearOption {
	return func(o *clearOptions) {
		o.skipSpaces = spaces
	}
}

func SkipInstances(instances ...string) ClearOption {
	return func(o *clearOptions) {
		o.skipInstances = instances
	}
}

func ClearAll(t *testing.T, opts ...ClearOption) *model.Broker {
	options := clearOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	broker, err := model.GetBroker()
	assert.NoError(t, err)

	mode, err := model.DataResponse[string]{}.Request(context.Background(), "api-gateway", broker, "mode", []interface{}{})
	assert.NoError(t, err)

	if mode != "sync" {
		logrus.Warnf("SYNC mode required: check model/tnt/app/config/init.lua  sys.MODE must be SYNC.  Current mode=%s", mode)
	}

	instance := helpers.NewInstance(broker)
	assert.NotEmpty(t, instance)

	skipInstances := options.skipInstances
	for id := range broker.Pool {
		if id != getShardId(id) {
			// it's read-only replica
			skipInstances = append(skipInstances, id)
		}
	}

	skipSpaces := append(options.skipSpaces,
		"market",
		"tier",
		"orderbook_sequence",
		"pubs",
		"meta",
		"presence",
	)
	err = instance.Reset(skipSpaces, skipInstances)
	assert.NoError(t, err)

	logrus.Infof("... running tnt in mode=%s", mode)

	return broker
}

func GetCreateProfile(t *testing.T, apiModel *model.ApiModel, wallet string, credits float64) *model.Profile {
	profile, err := apiModel.GetProfileByWalletForExchangeId(context.Background(), wallet, model.EXCHANGE_DEFAULT)
	logrus.Info("*** get profile")
	logrus.Error(err)
	logrus.Info(profile)
	if err != nil || profile == nil {
		logrus.Info("creating profile")
		profile, err = apiModel.CreateProfile(context.Background(), model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
		assert.NoError(t, err)
		assert.NotNil(t, profile)

		_, err = apiModel.DepositCredit(context.Background(), profile.ProfileId, credits)
		assert.NoError(t, err)
	}

	return profile
}

type profileInstance struct {
	api *model.ApiModel
}

func (i *profileInstance) GetProfilesIdsAfterCreatedAt(ctx context.Context, afterTsMicro int64) ([]profilepkg.ProfileId, error) {
	var ids []profilepkg.ProfileId

	profiles, err := i.api.GetExtendedProfiles(ctx)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		ids = append(ids, profile.ProfileId)
	}

	return ids, nil
}

func (i *profileInstance) GetVolumesAggregatesLast30d(ctx context.Context) ([]tsdb.CumVolume, error) {
	var volumes []tsdb.CumVolume

	profiles, err := i.api.GetExtendedProfiles(ctx)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		data, err := i.api.GetProfileData(ctx, profile.ProfileId)
		if err != nil {
			return nil, err
		}
			volumes = append(volumes, tsdb.CumVolume{ProfileId: data.ProfileID, CumVolume: data.CumTradingVolume.Decimal})
	}

	return volumes, nil
}

func (i *profileInstance) GetReferralsByInvitedProfiles(context.Context, ...profilepkg.ProfileId) ([]tsdb.ReferralLink, error) {
	return []tsdb.ReferralLink{}, nil
}

func DoTierPeriodics(t *testing.T, ctx context.Context, apiModel *model.ApiModel) error {
	store := &profileInstance{api: apiModel}
	cache := profilepkg.NewVolumeStoreCache(store)
	cache.Refresh(ctx)

	var marketsIds []profilepkg.MarketId
	for _, id := range MARKET_ID_MAP {
		marketsIds = append(marketsIds, id)
	}

	marketsClient := profilepkg.NewMarketsClient(apiModel, marketsIds, profilepkg.DefaultMarketsClientOptions)
	tc := profilepkg.NewTierCalc(cache, store, apiModel, profilepkg.TierCalcOptions{BatchSize: 200})

	tp := tierpkg.NewPeriodics(store, marketsClient, tc, tierpkg.DefaultPeriodicsOptions)
	if err := tp.RunOnce(ctx); err != nil {
		return err
	}

	return nil
}
