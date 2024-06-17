package api

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"golang.org/x/exp/slices"
)

var _debugAvailable = []string{"dev", "testnet", "testing"}

func TestEnv(t *testing.T) {
	cfg, err := ReadConfig()
	if err != nil {
		panic(err)
	}

	f := false
	if slices.Contains(_debugAvailable, cfg.Service.EnvMode) {
		logrus.Info(cfg.Service.EnvMode)
		f = true
	}

	assert.Equal(t, true, f)

	//TODO: make it automatic

	broker, err := model.GetBroker()
	assert.NoError(t, err)

	apiModel := model.NewApiModel(broker)

	profile, err := apiModel.CreateProfile(context.Background(), model.PROFILE_TYPE_TRADER, "0xtestwaslesffffletss", model.EXCHANGE_DEFAULT)
	assert.NoError(t, err)

	logrus.Info(profile)
}
