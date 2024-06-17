package api_client

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OnboardingFrontendSuite struct {
	APITestSuite
}

type OnboardingMarketMakerSuite struct {
	APITestSuite
}

type OnboardingVaultSuite struct {
	APITestSuite
}

func (s *OnboardingVaultSuite) TestOnboardVault() {
	resp, err := s.Client().onboardingVault()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)

	result := resp.Result[0]

	require.Equal(s.T(), strings.ToLower(s.Wallet), strings.ToLower(*result.Profile.Wallet))
	require.Greater(s.T(), len(result.RandomSecret), 0)
	require.Greater(s.T(), len(result.RefreshToken), 0)
	require.Greater(s.T(), len(result.Jwt), 0)
}

func (s *OnboardingMarketMakerSuite) TestOnboardMarketMaker() {
	resp, err := s.Client().Onboarding()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)

	result := resp.Result[0]

	require.Greater(s.T(), len(result.APISecret.Secret), 0)
	require.Greater(s.T(), len(result.APISecret.Key), 0)
	require.Greater(s.T(), len(result.Jwt), 0)

	prev_key := result.APISecret.Key

	resp, err = s.Client().Onboarding()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)

	result = resp.Result[0]
	require.Equal(s.T(), prev_key, result.APISecret.Key)

	logrus.Info(result.APISecret.Key)
}

func (s *OnboardingFrontendSuite) TestOnboardFrontend() {
	resp, err := s.Client().onboardingFrontend()
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)

	result := resp.Result[0]

	require.Equal(s.T(), strings.ToLower(s.Wallet), strings.ToLower(*result.Profile.Wallet))
	require.Greater(s.T(), len(result.RandomSecret), 0)
	require.Greater(s.T(), len(result.RefreshToken), 0)
	require.Greater(s.T(), len(result.Jwt), 0)
}

func TestOnboardingMarketMakerSuite(t *testing.T) {
	suite.Run(t, &OnboardingMarketMakerSuite{})
}

func TestOnboardingFrontendSuite(t *testing.T) {
	suite.Run(t, &OnboardingFrontendSuite{})
}

func TestOnboardingVaultSuite(t *testing.T) {
	suite.Run(t, &OnboardingVaultSuite{})
}
