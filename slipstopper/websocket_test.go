package slipstopper

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tests"
)

func sleep() {
	time.Sleep(time.Millisecond * 1000)
}

func TestReceiveInitialData(t *testing.T) {
	broker := tests.ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = tests.SetFairPrice(apiModel, "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 100.0)
	assert.NoError(t, err)

	profile1 := tests.GetCreateProfile(t, apiModel, "0xwallet1", 10000.0)
	assert.NotEmpty(t, profile1)

	profile2 := tests.GetCreateProfile(t, apiModel, "0xwallet2", 10000.0)
	assert.NotEmpty(t, profile2)

	//TODO: discuss this deposit to insurance, seems there is a bug (rounding error?)
	err = tests.Deposit(apiModel, 0, 50000)
	assert.NotEmpty(t, profile2)

	price := 1200.0
	size := 1.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile1.ProfileId,
		"BTC-USD",
		"limit",
		"short",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	size = 2.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"limit",
		"long",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	// wait for the matching engine to match.
	sleep()

	triggerPrice := 1500.0
	sizePercent := 1.0
	size = 1
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"take_profit",
		"long",
		&price,
		&size,
		nil,
		&triggerPrice,
		&sizePercent,
		nil,

		nil,
	)
	assert.NoError(t, err)
	sleep()

	cfg, err := ReadConfig()
	if err != nil {
		logrus.Fatalln(err)
	}

	wsClient := NewWSClient(cfg)
	readyChan := make(chan bool)
	wsClient.Run(readyChan)
	<-readyChan

	sleep()
}

func TestSimpleConditionalOrderCreate(t *testing.T) {
	cfg, err := ReadConfig()
	if err != nil {
		logrus.Fatalln(err)
	}

	wsClient := NewWSClient(cfg)
	readyChan := make(chan bool)
	go wsClient.Run(readyChan)
	<-readyChan

	broker := tests.ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	err = apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = tests.SetFairPrice(apiModel, "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 100.0)
	assert.NoError(t, err)

	profile1 := tests.GetCreateProfile(t, apiModel, "0xwallet1", 10000.0)
	assert.NotEmpty(t, profile1)

	profile2 := tests.GetCreateProfile(t, apiModel, "0xwallet2", 10000.0)
	assert.NotEmpty(t, profile2)

	//TODO: discuss this deposit to insurance, seems there is a bug (rounding error?)
	err = tests.Deposit(apiModel, 0, 50000)
	assert.NotEmpty(t, profile2)

	price := 1200.0
	size := 1.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile1.ProfileId,
		"BTC-USD",
		"limit",
		"short",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	size = 2.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"limit",
		"long",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	// wait for the matching engine to match.
	sleep()

	// we should not have any orders at this stage.
	assert.Equal(t, uint64(0), wsClient.matcherByMarket["BTC-USD"].Size())

	triggerPrice := 1500.0
	sizePercent := 1.0
	size = 1
	sltpOrder, err := apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"take_profit",
		"long",
		&price,
		&size,
		nil,
		&triggerPrice,
		&sizePercent,
		nil,

		nil,
	)
	assert.NoError(t, err)
	sleep()

	// we should now have 1 order
	assert.Equal(t, uint64(1), wsClient.matcherByMarket["BTC-USD"].Size())
	tup, err := wsClient.matcherByMarket["BTC-USD"].tree.Get(sltpOrder.TriggerPrice.Decimal)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(tup.Values))
	assert.Equal(t, tup.Key.String(), sltpOrder.TriggerPrice.String())

	err = apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 1500.0)
	assert.NoError(t, err)

	err = tests.SetFairPrice(apiModel, "BTC-USD", 1500.0)
	assert.NoError(t, err)
	sleep()

	wsClient.matcherByMarket["BTC-USD"].OnPriceUpdate(decimal.NewFromFloat(1500.0))
	sleep()
}

func TestSimpleConditionalOrderCancel(t *testing.T) {
	cfg, err := ReadConfig()
	if err != nil {
		logrus.Fatalln(err)
	}

	wsClient := NewWSClient(cfg)
	readyChan := make(chan bool)
	go wsClient.Run(readyChan)
	<-readyChan

	broker := tests.ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	err = apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = tests.SetFairPrice(apiModel, "BTC-USD", 1200.0)
	assert.NoError(t, err)

	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 100.0)
	assert.NoError(t, err)

	profile1 := tests.GetCreateProfile(t, apiModel, "0xwallet1", 10000.0)
	assert.NotEmpty(t, profile1)

	profile2 := tests.GetCreateProfile(t, apiModel, "0xwallet2", 10000.0)
	assert.NotEmpty(t, profile2)

	//TODO: discuss this deposit to insurance, seems there is a bug (rounding error?)
	err = tests.Deposit(apiModel, 0, 50000)
	assert.NotEmpty(t, profile2)

	price := 1200.0
	size := 1.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile1.ProfileId,
		"BTC-USD",
		"limit",
		"short",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	size = 2.0
	_, err = apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"limit",
		"long",
		&price,
		&size,
		nil,
		nil,
		nil,
		nil,

		nil,
	)
	assert.NoError(t, err)

	// wait for the matching engine to match.
	sleep()

	// we should not have any orders at this stage.
	assert.Equal(t, uint64(0), wsClient.matcherByMarket["BTC-USD"].Size())

	triggerPrice := 1500.0
	sizePercent := 1.0
	size = 1
	sltpOrder, err := apiModel.OrderCreate(
		context.TODO(),
		profile2.ProfileId,
		"BTC-USD",
		"take_profit",
		"long",
		&price,
		&size,
		nil,
		&triggerPrice,
		&sizePercent,
		nil,

		nil,
	)
	assert.NoError(t, err)
	sleep()

	// we should now have 1 order
	assert.Equal(t, uint64(1), wsClient.matcherByMarket["BTC-USD"].Size())
	tup, err := wsClient.matcherByMarket["BTC-USD"].tree.Get(sltpOrder.TriggerPrice.Decimal)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(tup.Values))
	assert.Equal(t, tup.Key.String(), sltpOrder.TriggerPrice.String())

	// cancel the SLTP order
	_, err = apiModel.OrderCancel(context.TODO(), sltpOrder.ProfileId, sltpOrder.MarketId, sltpOrder.OrderId, "")
	assert.NoError(t, err)

	sleep()
	// we should now have 0 order
	assert.Equal(t, uint64(0), wsClient.matcherByMarket["BTC-USD"].Size())
	_, err = wsClient.matcherByMarket["BTC-USD"].tree.Get(sltpOrder.TriggerPrice.Decimal)
	assert.Error(t, err)
}
