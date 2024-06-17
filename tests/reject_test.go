package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestRejectPostMatch(t *testing.T) {
	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	wallet := fmt.Sprintf("0xw%d", time.Now().UnixMicro())

	profile := GetCreateProfile(t, apiModel, wallet, 1000.0)
	assert.NotEmpty(t, profile)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)
	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 1500.0)
	assert.NoError(t, err)

	//Open orders on multiple markets and check that they have the right shard_id
	//ctx context.Context, profile_id uint, market_id, order_type, side string, price, size float64) (OrderCreateRes, error
	_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.LONG, ToPtr(20000.0), ToPtr(200.0), nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	err = apiModel.CancelAll(context.Background(), profile.ProfileId, false)
	assert.NoError(t, err)
}
