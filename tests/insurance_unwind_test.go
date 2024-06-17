package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

const INSURANCE_WATERFALL_INTERVAL = time.Duration(6 * time.Second)

func TestInsuranceUnwind(t *testing.T) {

	broker := ClearAll(t)
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

	insuranceFound := false
	//Create profiles and put them to liquidating status
	for _, wallet := range wallets {
		profile := GetCreateProfile(t, apiModel, wallet, credits)
		assert.NotEmpty(t, profile)

		if profile.ProfileId == 0 {
			insuranceFound = true
		}

		profiles = append(profiles, profile)
	}
	assert.Equal(t, len(wallets), len(profiles))
	assert.Equal(t, true, insuranceFound)

	for _, profile := range profiles {
		if profile.ProfileId == 0 {
			_, err := apiModel.OrderCreate(context.TODO(), profile.ProfileId, "ETH-USD", model.LIMIT, model.LONG, ToPtr(1500.0), ToPtr(0.01), nil, nil, nil, nil, nil)
			assert.NoError(t, err)
		} else {
			_, err := apiModel.OrderCreate(context.TODO(), profile.ProfileId, "ETH-USD", model.LIMIT, model.SHORT, ToPtr(1500.0), ToPtr(0.01), nil, nil, nil, nil, nil)
			assert.NoError(t, err)
		}
	}
	//wait for all orders created
	time.Sleep(2 * time.Second)
	cache, err := apiModel.GetProfileCache(context.TODO(), uint(0))
	assert.NoError(t, err)
	assert.Equal(t, uint(0), cache.ProfileID)

	positions, err := apiModel.GetOpenPositions(context.TODO(), uint(0))
	assert.NoError(t, err)
	assert.NotEmpty(t, positions)

	logrus.Info(positions[0])
	marketData, err := apiModel.GetMarketData(context.TODO(), positions[0].MarketID)
	assert.NoError(t, err)
	assert.NotEmpty(t, marketData)
}
