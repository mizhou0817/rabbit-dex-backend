package tests

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func TestBalanceOpsNotif(t *testing.T) {
	broker, err := model.GetBroker()
	assert.NoError(t, err)

	amount := tdecimal.NewDecimal(decimal.NewFromFloat(100.0))

	//create deposit
	b_ops, err := model.DataResponse[*model.BalanceOps]{}.Request(context.Background(), "profile", broker, "balance.create_deposit", []interface{}{
		1,
		"0xwallet",
		amount,
		"txhash",
		"",
		0,
	})
	assert.NoError(t, err)
	logrus.Info(b_ops)

	w_ids := []string{b_ops.OpsId}

	deposit_id := StringWithCharset(20, charset)

	// balance.process_deposit(profile_id, wallet, deposit_id, amount)
	b_ops, err = model.DataResponse[*model.BalanceOps]{}.Request(context.Background(), "profile", broker, "balance.process_deposit", []interface{}{
		1,
		"0xwallet",
		deposit_id,
		amount,
	})
	assert.NoError(t, err)
	logrus.Info(b_ops)

	// balance.process_deposit(profile_id, wallet, deposit_id, amount)
	b_ops, err = model.DataResponse[*model.BalanceOps]{}.Request(context.Background(), "profile", broker, "balance.acknowledge_withdrawals", []interface{}{
		w_ids,
	})
	logrus.Info(b_ops)

	b_ops, err = model.DataResponse[*model.BalanceOps]{}.Request(context.Background(), "profile", broker, "balance.completed_withdrawals", []interface{}{
		w_ids,
	})
	logrus.Info(b_ops)
}
