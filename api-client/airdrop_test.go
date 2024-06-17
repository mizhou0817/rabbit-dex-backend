package api_client

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/strips-finance/rabbit-dex-backend/api"
)

type AirdropSuite struct {
	APITestSuite
}

// One big flow for testing all
func (s *AirdropSuite) TestWholeFlow() {
	s.OnboardMarketMaker()

	profile := s.onboardMarketMakerResp.Profile

	/*
		Plan:
		1. Create 2 airdrops
		2. Setup 2 airdrops for 1 profile
		3. Check that all works with no error. Not doing real trades.
	*/

	airdrops := []string{"air1", "air2"}

	minTimestamp := int64(1682693219000000)
	totalRewards := float64(100)
	claimable := float64(1)

	for i, air := range airdrops {
		resp, err := s.Client().CreateAirdrop(&api.AirdropCreateRequest{
			Title:          air,
			StartTimestamp: minTimestamp + 100*int64(i+1),
			EndTimestamp:   minTimestamp + 200*int64(i+1),
		})
		s.NoError(err)
		s.True(resp.Success)
		s.Empty(resp.Error)

		result := resp.Result[0]
		s.Equal(result.Title, air)

		//setup
		resp2, err := s.Client().AirdropInit(&api.ProfileAirdropInitRequest{
			AirdropTitle: air,
			ProfileId:    profile.ProfileID,
			TotalRewards: totalRewards,
			Claimable:    claimable,
		})
		s.NoError(err)
		s.True(resp2.Success)
		s.Empty(resp2.Error)

		result2 := resp2.Result[0]
		s.Equal(result2.AirdropTitle, air)
		s.Equal(result2.ProfileId, profile.ProfileID)
		s.Equal(result2.TotalRewards, totalRewards)
		s.Equal(result2.Claimable, claimable)
	}

	//Get all airdrops
	resp, err := s.Client().GetAirdrops()
	s.NoError(err)
	s.True(resp.Success)
	s.Empty(resp.Error)

	result := resp.Result[0]
	logrus.Info(result)

	//Claim all
	resp1, err := s.Client().AirdropClaim()
	s.NoError(err)
	s.True(resp1.Success)
	s.Empty(resp1.Error)

	result2 := resp.Result[0]
	logrus.Info(result2)

}

func TestAirdropSuite(t *testing.T) {
	suite.Run(t, &AirdropSuite{})
}
