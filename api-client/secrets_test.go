package api_client

import (
	// "math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/auth"
	// "github.com/strips-finance/rabbit-dex-backend/signer"
)

type SecretsSuite struct {
	APITestSuite
}

func (s *SecretsSuite) TestCreateFlow() {
	s.OnboardMarketMaker()

	//profile := s.onboardMarketMakerResp.Profile

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

	expiration := time.Now().Unix() + 100000
	resp, err := s.Client().SecretCreate(api.SecretCreateRequest{
		Tag:        "tag",
		Expiration: expiration,
		AllowedIpList: []string{
			"192.0.2.1",
			"192.0.2.3",
		},
	}, pkSignature, pkTimestamp)
	s.NoError(err)
	logrus.Info(resp)

	//Try to withdraw
	resp1, err := s.Client().ListSecrets(pkSignature, pkTimestamp)
	s.NoError(err)
	logrus.Info(resp1)
}

func (s *SecretsSuite) TestGet() {
	s.OnboardMarketMaker()

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

	expiration := time.Now().Unix() + 100000
	_, err = s.Client().SecretCreate(api.SecretCreateRequest{
		Tag:        "tag1",
		Expiration: expiration,
		AllowedIpList: []string{
			"192.0.2.1",
			"192.0.2.3",
		},
	}, pkSignature, pkTimestamp)
	s.NoError(err)

	_, err = s.Client().SecretCreate(api.SecretCreateRequest{
		Tag:        "tag2",
		Expiration: expiration,
		AllowedIpList: []string{
			"212.0.2.1",
			"212.0.2.3",
		},
	}, pkSignature, pkTimestamp)

	//Try to withdraw
	resp1, err := s.Client().ListSecrets(pkSignature, pkTimestamp)
	s.NoError(err)
	logrus.Info(resp1.Result)

}

func TestSecretsSuite(t *testing.T) {
	suite.Run(t, &SecretsSuite{})
}
