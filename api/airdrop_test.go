package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

func TestClaimOpsResponse(t *testing.T) {
	broker, err := model.GetBroker()
	assert.NoError(t, err)

	apiModel := model.NewApiModel(broker)

	ops, err := apiModel.TestCreateClaimOps(context.Background(), 1, "phase1", 202.111111)
	assert.NoError(t, err)
	assert.NotEmpty(t, ops)

	res, err := _signClaimOps(ops, "0x3ba925fdeae6b46d0bb4d424d829982cb2f7309e")
	assert.NoError(t, err)

	resp, err := json.Marshal(res)
	assert.NoError(t, err)
	assert.Equal(t, "202111111000000000000", res.BnAmount)
	fmt.Println(string(resp))

}
