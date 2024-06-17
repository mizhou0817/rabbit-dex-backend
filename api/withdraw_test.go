package api

import (
	// "context"
	// "fmt"
	// "testing"
	// "time"

	// "github.com/shopspring/decimal"
	// "github.com/sirupsen/logrus"
	// "github.com/stretchr/testify/assert"
	// "github.com/strips-finance/rabbit-dex-backend/model"
	// "github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

// TODO - update this test for settlement v2

// func TestWithdrawTarantoolLowLevel(t *testing.T) {
// 	broker, err := model.GetBroker()
// 	assert.NoError(t, err)

// 	apiModel := model.NewApiModel(broker)

// 	//Setup market prices - just to not see errors in the logs
// 	err = apiModel.UpdateIndexPrice(context.TODO(), "BTC-USD", 20000.0)
// 	assert.NoError(t, err)
// 	err = apiModel.UpdateIndexPrice(context.TODO(), "ETH-USD", 1500.0)
// 	assert.NoError(t, err)
// 	err = apiModel.UpdateIndexPrice(context.TODO(), "SOL-USD", 15.0)
// 	assert.NoError(t, err)

// 	randomWallet := fmt.Sprintf("0x%d", time.Now().UnixMicro())
// 	profile, err := apiModel.CreateTraderProfile(context.Background(), randomWallet)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, profile)

// 	// Deposit some credits
// 	deposit := 10000.0
// 	_, err = apiModel.DepositCredit(context.Background(), profile.ProfileId, deposit)
// 	assert.NoError(t, err)

// 	// Check withdraw lock
// 	lock, err := apiModel.AcquireWithdrawLock(context.TODO(), profile.ProfileId)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, lock)
// 	assert.Equal(t, true, lock.Locked)
// 	assert.Equal(t, profile.ProfileId, lock.ProfileId)

// 	tm := lock.Timestamp

// 	//can't acquire already locked lock
// 	lock, err = apiModel.AcquireWithdrawLock(context.TODO(), profile.ProfileId)
// 	assert.Error(t, err)
// 	assert.Equal(t, "ALREADY_LOCKED", err.Error())
// 	assert.Empty(t, lock)

// 	lock, err = apiModel.ReleaseWithdrawLock(context.TODO(), profile.ProfileId)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, lock)
// 	assert.Equal(t, false, lock.Locked)
// 	assert.Equal(t, profile.ProfileId, lock.ProfileId)
// 	assert.NotEqual(t, tm, lock.Timestamp) // timestamp changed

// 	tm1 := lock.Timestamp

// 	// it's ok to release whatever amount of times
// 	lock, err = apiModel.ReleaseWithdrawLock(context.TODO(), profile.ProfileId)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, lock)
// 	assert.NotEqual(t, tm1, lock.Timestamp) // time changed

// 	//CHECK withdraw allowed
// 	is_allowed := apiModel.CheckWithdrawAllowed(context.TODO(), profile.ProfileId)
// 	assert.Equal(t, true, is_allowed)

// 	//Create pending TX
// 	d_amount := tdecimal.NewDecimal(decimal.NewFromFloat(10.0))
// 	b_ops, err := model.DataResponse[*model.BalanceOps]{}.Request(context.Background(), "profile", broker, "balance.create_withdrawal", []interface{}{
// 		profile.ProfileId,
// 		profile.Wallet,
// 		d_amount,
// 		"txhash",
// 	})
// 	assert.NoError(t, err)
// 	logrus.Info(b_ops)

// 	//CHECK withdraw allowed
// 	is_allowed = apiModel.CheckWithdrawAllowed(context.TODO(), profile.ProfileId)
// 	assert.Equal(t, false, is_allowed)

// 	//Process TX
// 	err = apiModel.AcknowledgeWithdrawals(context.TODO(), []string{b_ops.OpsId})
// 	assert.NoError(t, err)

// 	//CHECK withdraw allowed again
// 	is_allowed = apiModel.CheckWithdrawAllowed(context.TODO(), profile.ProfileId)
// 	assert.Equal(t, true, is_allowed)
// }
