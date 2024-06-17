package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestEngineLatency(t *testing.T) {
	market_id := "BTC-USD"

	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	profile := GetCreateProfile(t, apiModel, "test-wallet", 10000000.0)
	assert.NotEmpty(t, profile)

	apiModel.WhiteListProfile(context.Background(), profile.ProfileId)

	ids := make([]string, 0)
	for i := 1; i < 10000; i++ {
		order, err := apiModel.OrderCreate(
			context.Background(),
			profile.ProfileId,
			market_id,
			"limit",
			"long",
			ToPtr(10.0),
			ToPtr(1.0),
			nil,
			nil,
			nil,
			nil,

			nil,
		)
		assert.NoError(t, err)
		ids = append(ids, order.OrderId)

		if i%100 == 0 {
			for _, o_id := range ids {
				_, err := apiModel.OrderCancel(
					context.Background(),
					order.ProfileId,
					market_id,
					o_id,
					"")
				assert.NoError(t, err)
			}
			ids = ids[:0]
		}
	}
}
