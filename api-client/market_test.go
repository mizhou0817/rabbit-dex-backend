package api_client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MarketSuite struct {
	APITestSuite
}

func (s *MarketSuite) TestMarket() {

	resp, err := s.Client().MarketList(nil)
	require.NoError(s.T(), err)
	require.True(s.T(), resp.Success)
	require.Empty(s.T(), resp.Error)
	require.GreaterOrEqual(s.T(), len(resp.Result), 3)

	requiredMarkets := []string{
		"BTC-USD",
		"ETH-USD",
		"SOL-USD",
	}

	for _, market := range resp.Result {
		s.Contains(requiredMarkets, market.MarketID)

		require.Greater(s.T(), *market.MinTick, 0.0)
		require.Greater(s.T(), *market.MinOrder, 0.0)
		require.Greater(s.T(), market.FairPrice.InexactFloat64(), 0.0)
		require.Greater(s.T(), market.IndexPrice.InexactFloat64(), 0.0)
	}
}

func TestMarketSuite(t *testing.T) {
	suite.Run(t, &MarketSuite{})
}
