package api_client

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	"github.com/strips-finance/rabbit-dex-backend/model"
	// "github.com/strips-finance/rabbit-dex-backend/signer"
)

type BalanceSuite struct {
	APITestSuite
}

func (s *BalanceSuite) TestDepositWithdrawFlow() {
	s.OnboardMarketMaker()

	profile := s.onboardMarketMakerResp.Profile

	amount := 1000.0
	randomHash := fmt.Sprintf("0x%d", time.Now().UnixMicro())
	resp, err := s.Client().Deposit(&api.DepositRequest{
		TxHash: randomHash,
		Amount: amount,
	})
	s.NoError(err)
	logrus.Info(resp)

	s.True(resp.Success)
	s.Empty(resp.Error)

	result := resp.Result[0]
	s.Equal(result.Status, model.BALANCE_OPS_STATUS_PENDING)

	//Deposit TEST credits to be able to test withdraw
	broker, err := model.GetBroker()
	s.NoError(err)

	apiModel := model.NewApiModel(broker)

	logrus.Info("1**** making test deposit")
	deposit := 10000.0
	_, err = apiModel.DepositCredit(context.Background(), profile.ProfileID, deposit)
	s.NoError(err)

	pkTimestamp := s.client.getExpirationTimestamp()
	signReq := &auth.MetamaskSignRequest{
		Message:   ONBOARDING_MESSAGE,
		Timestamp: pkTimestamp,
	}

	privateKey, err := s.client.Credentials.GetPrivateKey()
	s.NoError(err)

	// encoder := signer.NewEIP712Encoder(
	// 	"RabbitXId",
	// 	"1",
	// 	"",
	// 	big.NewInt(int64(31337)),
	// )
	// pkSignature, err := auth.EthSign(auth.EIP_712, signReq, privateKey, encoder)
	pkSignature, err := auth.EthSign(auth.EIP_191, signReq, privateKey, nil)
	s.NoError(err)

	//Try to withdraw
	resp, err = s.Client().Withdraw(&api.WithdrawalRequest{
		Amount: 1.0,
	}, pkSignature, pkTimestamp)
	s.NoError(err)
	logrus.Info(resp)
	result = resp.Result[0]
	s.Equal(result.Status, model.BALANCE_OPS_STATUS_PENDING)

	//Try to withdraw second time, not allowed
	resp, err = s.Client().Withdraw(&api.WithdrawalRequest{
		Amount: 1.0,
	}, pkSignature, pkTimestamp)
	s.NoError(err)
	s.False(resp.Success)
	s.Equal("WITHDRAW_UNAVAILABLE_PENDING_EXIST", resp.Error)

	_, err = s.Client().CancelWithdrawal()
	s.NoError(err)

	//Should work now
	resp, err = s.Client().Withdraw(&api.WithdrawalRequest{
		Amount: amount,
	}, pkSignature, pkTimestamp)
	s.NoError(err)
	result = resp.Result[0]
	s.Equal(result.Status, model.BALANCE_OPS_STATUS_PENDING)

	logrus.Info("3**** process pending withdraw")
	// this should set the due block to 3630723
	err = apiModel.UpdatePendingWithdrawals(context.Background(), big.NewInt(3630723), big.NewInt(3630723), "")
	s.NoError(err)
	// this should set the status to claimable since the due block has passed
	err = apiModel.UpdatePendingWithdrawals(context.Background(), big.NewInt(3630724), big.NewInt(3630724), "")
	s.NoError(err)
}

func TestBalanceSuite(t *testing.T) {
	suite.Run(t, &BalanceSuite{})
}
