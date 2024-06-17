package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestOrderCreateClientOrderId(t *testing.T) {
	market_id := "BTC-USD"

	broker := ClearAll(t)
	assert.NotEmpty(t, broker)

	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	profile := GetCreateProfile(t, apiModel, "0xwallet", 10000000.0)
	assert.NotEmpty(t, profile)

	//Setup market prices
	err := apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
	assert.NoError(t, err)

	client_order_id := "12345678901234567890123456789012345678901234567890"
	order1, err := apiModel.OrderCreate(
		context.Background(),
		profile.ProfileId,
		market_id,
		"limit",
		"long",
		ToPtr(20000.0),
		ToPtr(1.0),
		&client_order_id,
		nil,
		nil,
		nil,

		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, *order1.ClientOrderId, client_order_id)

	order2, err := apiModel.OrderCreate(
		context.Background(),
		profile.ProfileId,
		market_id,
		"limit",
		"long",
		ToPtr(20000.0),
		ToPtr(1.0),
		nil,
		nil,
		nil,
		nil,

		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, order2.ClientOrderId, (*string)(nil))

	client_order_id = "123456789012345678901234567890123456789012345678901"
	_, err = apiModel.OrderCreate(
		context.Background(),
		profile.ProfileId,
		market_id,
		"limit",
		"long",
		ToPtr(20000.0),
		ToPtr(1.0),
		&client_order_id,
		nil,
		nil,
		nil,

		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, string(err.Error()), "CLIENT_ORDER_ID_TOO_LARGE")
}
