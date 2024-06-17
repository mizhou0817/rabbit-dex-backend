package tests

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestBroker(t *testing.T) {
	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	profile := GetCreateProfile(t, apiModel, "0xwallet", 10000000.0)
	assert.NotEmpty(t, profile)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)
	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 1500.0)
	assert.NoError(t, err)

	go func(msg string) {
		logrus.Infof("started = %s", msg)
		for {
			apiModel.IsCancellAllAccepted(context.TODO(), profile.ProfileId)
			logrus.Info(msg)
		}
	}("dummy1")

	go func(msg string) {
		logrus.Infof("started = %s", msg)
		for {
			apiModel.IsCancellAllAccepted(context.TODO(), profile.ProfileId)
			logrus.Info(msg)
		}
	}("dummy2")

	select {}
}
