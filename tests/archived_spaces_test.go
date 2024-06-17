package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestArchivedSpaces(t *testing.T) {
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

	//Open orders on multiple markets and check that they have the right shard_id
	//ctx context.Context, profile_id uint, market_id, order_type, side string, price, size float64) (OrderCreateRes, error
	_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.LONG, ToPtr(20000.0), ToPtr(1.0), nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "ETH-USD", model.LIMIT, model.LONG, ToPtr(1300.0), ToPtr(1.0), nil, nil, nil, nil, nil)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)

	orders, err := apiModel.GetAllOrders(context.TODO(), profile.ProfileId, "BTC-USD", 1)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(orders))
	assert.Equal(t, "BTC-USD", orders[0].ShardId)
	assert.NotEqual(t, 0, orders[0].ArchiveId)
	logrus.Info(orders[0])

	orders, err = apiModel.GetAllOrders(context.TODO(), profile.ProfileId, "ETH-USD", 1)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(orders))
	assert.Equal(t, "ETH-USD", orders[0].ShardId)
	assert.NotEqual(t, 0, orders[0].ArchiveId)
	logrus.Info(orders[0])

	//Do a trade - close order - check archive_id
	_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "BTC-USD", model.LIMIT, model.SHORT, ToPtr(20000.0), ToPtr(1.0), nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	_, err = apiModel.OrderCreate(context.Background(), profile.ProfileId, "ETH-USD", model.LIMIT, model.SHORT, ToPtr(1500.0), ToPtr(1.0), nil, nil, nil, nil, nil)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)

	orders, err = apiModel.GetAllOrders(context.TODO(), profile.ProfileId, "BTC-USD", 1)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(orders))
	assert.Equal(t, "BTC-USD", orders[0].ShardId)
	assert.NotEqual(t, 1, orders[0].ArchiveId)
	logrus.Info(orders[0])

	orders, err = apiModel.GetAllOrders(context.TODO(), profile.ProfileId, "ETH-USD", 1)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(orders))
	assert.Equal(t, "ETH-USD", orders[0].ShardId)
	assert.NotEqual(t, 1, orders[0].ArchiveId)
	logrus.Info(orders[0])

	//Check trades and fills has the right shard_id and archive_id
	allFills, allTrades, err := apiModel.GetAllFills(context.TODO(), []string{"BTC-USD", "ETH-USD"})
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(allFills))
	assert.NotEqual(t, 0, len(allTrades))

	for _, fill := range allFills {
		logrus.Info(fill)
		assert.Equal(t, fill.MarketId, fill.ShardId)
		assert.NotEqual(t, 0, fill.ArchiveId)
	}

	for _, trade := range allTrades {
		logrus.Info(trade)
		assert.Equal(t, trade.MarketId, trade.ShardId)
		assert.NotEqual(t, 0, trade.ArchiveId)
	}

}
