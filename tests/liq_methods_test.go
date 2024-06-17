package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/liqengine"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestLiqMethods(t *testing.T) {
	broker := ClearAll(t, SkipInstances("api-gateway"))
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	profiles := make([]*model.Profile, 0)
	wallets := []string{"w1", "w2", "w3", "w4", "w5"}
	credits := 10000000.0
	status := model.PROFILE_STATUS_LIQUIDATING

	//Create profiles and put them to liquidating status
	for _, wallet := range wallets {
		profile := GetCreateProfile(t, apiModel, wallet, credits)
		assert.NotEmpty(t, profile)

		err := apiModel.ProfileUpdateStatus(context.Background(), profile.ProfileId, status)
		assert.NoError(t, err)

		tm, err := apiModel.ProfileUpdateLastLiqChecked(context.Background(), profile.ProfileId)
		assert.NoError(t, err)
		assert.NotEmpty(t, tm)

		profiles = append(profiles, profile)
	}
	assert.Equal(t, len(wallets), len(profiles))

	for i, profile := range profiles {
		assert.Equal(t, wallets[i], profile.Wallet)

		cache, err := apiModel.GetProfileCache(context.Background(), profile.ProfileId)
		assert.NoError(t, err)
		assert.Equal(t, profiles[i].Wallet, *cache.Wallet)
		assert.Equal(t, status, *cache.Status)
	}

	limit := 2
	var last_id *uint = nil
	i := 0
	total := len(wallets)/limit + 1
	for {
		batch, err := apiModel.LiquidationBatch(context.Background(), last_id, limit)
		assert.NoError(t, err)
		if i >= int(total) {
			assert.Equal(t, 0, len(batch))
			break
		}
		for j, profile := range batch {
			assert.Equal(t, wallets[i*limit+j], *profile.Wallet)

			last_id = &profile.ProfileID
		}
		i++
	}
}

func TestNewLiquidationMethods(t *testing.T) {
	broker := ClearAll(t, SkipInstances("api-gateway"))
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)
	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 1500.0)
	assert.NoError(t, err)

	profiles := make([]*model.Profile, 0)
	wallets := []string{"0xw1", "0xw2", "0xw3", "0xw4", "0xw5"}
	credits := 10000000.0

	//Create profiles and put them to liquidating status
	for _, wallet := range wallets {
		profile := GetCreateProfile(t, apiModel, wallet, credits)
		assert.NotEmpty(t, profile)

		profiles = append(profiles, profile)

		//Create 1000 orders for profileId 0
		if profile.ProfileId == 0 {
			apiModel.WhiteListProfile(context.TODO(), profile.ProfileId)
			for i := 0; i < 10; i++ {
				_, err := apiModel.OrderCreate(context.TODO(), profile.ProfileId, "ETH-USD", model.LIMIT, model.LONG, ToPtr(1500.0), ToPtr(0.01), nil, nil, nil, nil, nil)
				assert.NoError(t, err)
			}
		}
	}
	assert.Equal(t, len(wallets), len(profiles))

	//wait for all orders created
	time.Sleep(2 * time.Second)
	for i, profile := range profiles {
		assert.Equal(t, wallets[i], profile.Wallet)

		cache, err := apiModel.GetProfileCache(context.Background(), profile.ProfileId)
		assert.NoError(t, err)
		assert.Equal(t, profiles[i].Wallet, *cache.Wallet)
	}

	as, err := liqengine.NewTntAssistant(broker, "0", 0)
	assert.NoError(t, err)

	//logrus.Info("Cancel existing orders")
	//err = as.CancelExistingOrders(context.TODO(), profiles[0].ProfileId)
	//assert.NoError(t, err)

	logrus.Info("Wait for cancel all")
	err = as.WaitForCancellAllAccepted(context.TODO(), profiles[0].ProfileId)
	assert.NoError(t, err)
}
