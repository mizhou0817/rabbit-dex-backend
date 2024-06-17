package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestPeriodics(t *testing.T) {
	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)
	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 1500.0)
	assert.NoError(t, err)
	err = apiModel.UpdateIndexPrice(context.TODO(), "SOL-USD", 15.0)
	assert.NoError(t, err)

	wallet := fmt.Sprintf("0xw%d", time.Now().UnixMicro())
	profile := GetCreateProfile(t, apiModel, wallet, 1000.0)
	assert.NotEmpty(t, profile)

	cache, err := apiModel.InvalidateCacheAndNotify(context.TODO(), profile.ProfileId)
	assert.NoError(t, err)
	assert.NotEmpty(t, cache)
	logrus.Info(cache)

	/*
		//UPDATE one profile cache
		cache, err := apiModel.InvalidateCache(context.TODO(), profile.ProfileId)
		assert.NoError(t, err)
		assert.NotEmpty(t, cache.LastUpdate)
		assert.Equal(t, profile.ProfileId, cache.ProfileID)
		assert.Equal(t, 1000.0, cache.AccountEquity.InexactFloat64())

		//Create 22000 profiles and update cache for them
		for i := 0; i < 3000; i++ {
			wallet = fmt.Sprintf("0xw%d%d", i, time.Now().UnixMicro())
			profile = GetCreateProfile(t, apiModel, wallet, 1000000.0)
			assert.NotEmpty(t, profile)

			//Open orders
			_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.LONG, 20000.0, 0.5, nil)
			assert.NoError(t, err)
		}

		for {
			// balance.process_deposit(profile_id, wallet, deposit_id, amount)
			res, err := model.DataResponse[string]{}.Request(context.Background(), "profile", broker, "periodics.test_update_profiles_meta", []interface{}{})
			assert.NoError(t, err)
			assert.Empty(t, res)

			time.Sleep(10 * time.Second)
		}

		select {}
	*/
}
