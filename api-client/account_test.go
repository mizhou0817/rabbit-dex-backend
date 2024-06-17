package api_client

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
)

type AccountSuite struct {
	APITestSuite
}

func (s *AccountSuite) TestAccount() {
	s.OnboardMarketMaker()

	resp, err := s.Client().Account()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)

	result := resp.Result[0]

	require.Equal(s.T(), s.Wallet, *result.Wallet)
}

func (s *AccountSuite) TestAccountValidateError() {
	s.client = nil

	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMyIsImV4cCI6MTY2MjgzNDMyMX0.kpRWp6mUEN45Iz0ukUJ-3-OqasR0ka0-BZx7SXwaRP0"
	resp, err := s.Client().AccountValidate(invalidToken)
	require.NoError(s.T(), err)
	require.False(s.T(), resp.Success)
	require.NotEmpty(s.T(), resp.Error)
}

func (s *AccountSuite) TestAccountValidate() {
	s.OnboardMarketMaker()

	resp, err := s.Client().AccountValidate(s.Client().Credentials.Jwt)
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)
}

func (s *AccountSuite) TestAccountLeverage() {
	s.OnboardMarketMaker()

	new_leverage := uint(5)
	resp, err := s.Client().AccountUpdateLeverage(&api.AccountSetLeverageRequest{
		MarketId: "BTC-USD",
		Leverage: new_leverage,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)

	str := fmt.Sprintf("%d", new_leverage)
	logrus.Info(resp.Result[0].Leverage["BTC-USD"].String())
	require.Equal(s.T(), str, resp.Result[0].Leverage["BTC-USD"].String())
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, &AccountSuite{})
}
