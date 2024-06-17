package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestSequence(t *testing.T) {
	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)

	wallet := fmt.Sprintf("0x%d", time.Now().UnixMicro())
	profile := GetCreateProfile(t, apiModel, wallet, 10000.0)
	assert.NotEmpty(t, profile)

	i := 1
	for {
		_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.LONG, ToPtr(20000.0), ToPtr(200.0), nil, nil, nil, nil, nil)
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)

		if i%5 == 0 {
			_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.LONG, ToPtr(20000.0), ToPtr(0.1), nil, nil, nil, nil, nil)
			assert.NoError(t, err)
		}
		i++
	}
}
