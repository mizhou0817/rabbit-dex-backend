package api_client

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
	"testing"
)

type JWTFrontendSuite struct {
	APITestSuite
}

type JWTMarketMakerSuite struct {
	APITestSuite
}

func (s *JWTFrontendSuite) TestJWTUpdate() {
	s.OnboardFrontend()

	resp, err := s.Client().JwtUpdate(&api.JwtRequest{
		IsClient:     true,
		RefreshToken: s.onboardFrontendResp.RefreshToken,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)
	require.NotEmpty(s.T(), resp.Result[0].Jwt)

	require.Equal(s.T(), resp.Result[0].Jwt, s.Client().Credentials.Jwt)
}

func (s *JWTFrontendSuite) TestJWTUpdateError() {
	s.OnboardFrontend()

	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMyIsImV4cCI6MTY2MjgzNDMyMX0.kpRWp6mUEN45Iz0ukUJ-3-OqasR0ka0-BZx7SXwaRP0"
	resp, err := s.Client().JwtUpdate(&api.JwtRequest{
		IsClient:     true,
		RefreshToken: invalidToken,
	})
	require.NoError(s.T(), err)
	require.False(s.T(), resp.Success)
	require.NotEmpty(s.T(), resp.Error)
	require.Len(s.T(), resp.Result, 0)
}

func (s *JWTMarketMakerSuite) TestJWTUpdate() {
	s.OnboardMarketMaker()

	resp, err := s.Client().JwtUpdate(&api.JwtRequest{
		IsClient:     false,
		RefreshToken: s.onboardMarketMakerResp.APISecret.Secret,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)
	require.NotEmpty(s.T(), resp.Result[0].Jwt)

	require.Equal(s.T(), resp.Result[0].Jwt, s.Client().Credentials.Jwt)
}

func TestJWTFrontendSuite(t *testing.T) {
	suite.Run(t, &JWTFrontendSuite{})
}

func TestJWTMarketMakerSuite(t *testing.T) {
	suite.Run(t, &JWTMarketMakerSuite{})
}
