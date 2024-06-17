package api_client

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"github.com/strips-finance/rabbit-dex-backend/auth"
)

type APITestSuite struct {
	suite.Suite

	Cfg        *api.Config
	Router     *gin.Engine
	PrivateKey *ecdsa.PrivateKey
	Wallet     string

	onboardMarketMakerResp *auth.OnboardMarketMakerResult
	onboardFrontendResp    *auth.OnboardFrontendResult
	client                 *Client
}

// var APIRunning = false

func (s *APITestSuite) SetupTest() {
	s.Cfg = api.ReadDefaultConfig()

	// Generate PrivateKey to perform message signing
	privateKey, err := crypto.GenerateKey()

	s.NoError(err)
	s.PrivateKey = privateKey

	// Derive PublicKey from PrivateKey
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	s.True(ok)

	// Derive Wallet address from PublicKey
	s.Wallet = crypto.PubkeyToAddress(*publicKey).String()

	s.Router = api.Router()

	// if !APIRunning {
	//	go func() {
	//		logrus.Info("Starting REST-API Service")
	//
	//		cfg := api.ReadDefaultConfig()
	//		addr := fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port)
	//
	//		if err := s.Router.Run(addr); err != nil {
	//			logrus.Error(err)
	//		}
	//	}()
	//
	//	APIRunning = true
	// }
}

func (s *APITestSuite) TearDownTest() {
	time.Sleep(time.Second)
}

func (s *APITestSuite) Client() *Client {
	if s.client != nil {
		return s.client
	}

	privateKeyBytes := crypto.FromECDSA(s.PrivateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)
	credentials := &ClientCredentials{PrivateKey: privateKeyHex}

	client, err := NewClient(LOCAL_URL, credentials)
	require.NoError(s.T(), err)

	s.client = client

	return client
}

func (s *APITestSuite) OnboardMarketMaker() {
	if s.onboardMarketMakerResp != nil {
		return
	}

	resp, err := s.Client().Onboarding()
	s.NoError(err)

	require.True(s.T(), resp.Success)
	require.Greater(s.T(), len(resp.Result), 0)
	result := resp.Result[0]

	require.Equal(s.T(), strings.ToLower(s.Wallet), *result.Profile.Wallet)
	require.Greater(s.T(), len(result.APISecret.Secret), 0)
	require.Greater(s.T(), len(result.APISecret.Key), 0)
	require.Greater(s.T(), len(result.Jwt), 0)

	s.onboardMarketMakerResp = &result
}

func (s *APITestSuite) OnboardFrontend() {
	if s.onboardFrontendResp != nil {
		return
	}

	resp, err := s.Client().onboardingFrontend()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)

	result := resp.Result[0]

	require.Equal(s.T(), strings.ToLower(s.Wallet), *result.Profile.Wallet)
	require.Greater(s.T(), len(result.RandomSecret), 0)
	require.Greater(s.T(), len(result.RefreshToken), 0)
	require.Greater(s.T(), len(result.Jwt), 0)

	s.onboardFrontendResp = &result
}
