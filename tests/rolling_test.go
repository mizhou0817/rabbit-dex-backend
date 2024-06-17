package tests

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

func TestRolling(t *testing.T) {
	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	profile := GetCreateProfile(t, apiModel, "0xwallet", 10000000.0)
	assert.NotEmpty(t, profile)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)

	//create deposit
	d_amount := tdecimal.NewDecimal(decimal.NewFromFloat(11.0))

	for i := 0; i < 20; i++ {
		err = model.EmptyResponse{}.Request(context.Background(), "BTC-USD", broker, "market.update_roll_value", []interface{}{
			"dummy_rolling",
			"BTC-USD",
			d_amount,
			1,
			3,
			true})
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)

		err = model.EmptyResponse{}.Request(context.Background(), "BTC-USD", broker, "market.get_roll_avg", []interface{}{
			"dummy_rolling",
			"BTC-USD"})

		assert.NoError(t, err)

		err = model.EmptyResponse{}.Request(context.Background(), "BTC-USD", broker, "periodics._do_market_periodics", []interface{}{})
		assert.NoError(t, err)
	}

	err = model.EmptyResponse{}.Request(context.Background(), "BTC-USD", broker, "market.reset_roll_value", []interface{}{
		"dummy_rolling",
		"BTC-USD"})

}
