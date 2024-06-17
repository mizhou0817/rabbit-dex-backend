package api_client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
)

type CandleSuite struct {
	APITestSuite
}

func (s *CandleSuite) TestCandleList() {
	// onboarding not required
	currentTimestamp := time.Now().Unix()
	timestampFrom := currentTimestamp - 3600*24*5
	timestampTo := currentTimestamp

	resp, err := s.Client().CandleList(&api.CandleListRequest{
		MarketId:      "BTC-USD",
		TimestampFrom: timestampFrom,
		TimestampTo:   timestampTo,
		Period:        5,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)
	require.NotEmpty(s.T(), resp.Result)

	result := resp.Result[0]

	require.NotZero(s.T(), len(resp.Result))

	require.Greater(s.T(), result.Close, 0.0)
	require.Greater(s.T(), result.High, 0.0)
	require.Greater(s.T(), result.Low, 0.0)
	require.Greater(s.T(), result.Open, 0.0)
	require.GreaterOrEqual(s.T(), result.Volume, 0.0)

	require.GreaterOrEqual(s.T(), result.Time, timestampFrom)
	require.LessOrEqual(s.T(), result.Time, timestampTo)
}

func TestCandleSuite(t *testing.T) {
	suite.Run(t, &CandleSuite{})
}
