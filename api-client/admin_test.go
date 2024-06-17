package api_client

import (
	"bytes"
	"crypto/ecdsa"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
)

type AdminSuite struct {
	APITestSuite
}

func (s *AdminSuite) TestAddTier() {
	s.OnboardMarketMaker()

	// profile := s.onboardMarketMakerResp.Profile

	resp, err := s.Client().AddTier(&api.AddTierRequest{
		MarketId: "BTC-USD",
		TierId:   0,
		Title:    "test_tier",
	})
	s.NoError(err)
	logrus.Info(resp)
}

func (s *AdminSuite) TestWhichTier() {
	s.OnboardMarketMaker()

	// profile := s.onboardMarketMakerResp.Profile

	resp, err := s.Client().WhichTier(&api.WhichTierRequest{
		ProfileId: 0,
	})
	s.NoError(err)
	logrus.Info(resp)
}

func (s *AdminSuite) TestJWTSuperAdmin() {
	testPK := "e0ef370fffea97fd988d7c31919fad30d2eca1c57163fc3a81a15dae0415ca6a"
	privateKey, err := crypto.HexToECDSA(testPK)
	s.Require().NoError(err)
	s.Require().Equal("0x"+testPK, hexutil.Encode(crypto.FromECDSA(privateKey)))
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	s.Require().True(ok)
	wallet := crypto.PubkeyToAddress(*publicKey).String()
	s.T().Logf(
		"WARNING! Wallet %s has to be a member of the super_admins list in the config of"+
			" API running at %s", wallet, s.Client().apiUrl,
	)

	s.PrivateKey = privateKey
	s.Wallet = wallet
	s.onboardFrontendResp = nil
	s.onboardMarketMakerResp = nil
	s.client = nil

	_, err = s.Client().Onboarding()
	s.Require().NoError(err)

	s.Require().Equal(strings.ToLower(wallet), strings.ToLower(s.Client().Credentials.Wallet))
	s.Require().Equal(testPK, s.Client().Credentials.PrivateKey)

	req, err := http.NewRequest(
		http.MethodPost, s.Client().apiUrl+"/game_assets/blast",
		bytes.NewBuffer([]byte(`{"data": []}`)),
	)
	s.Require().NoError(err)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: s.Client().Credentials.Jwt})
	resp, err := s.Client().httpClient.Do(req)
	s.Require().NoError(err)
	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal(
		`{"success":true,"error":"","result":[{"max_batch_id":0,"replaced_record_count":0}]}`,
		string(respBody),
	)
}

func TestAdminSuite(t *testing.T) {
	suite.Run(t, &AdminSuite{})
}
