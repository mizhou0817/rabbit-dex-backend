package tests

import (
	"context"
	"testing"

	"time"

	"github.com/FZambia/tarantool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type TestDeadmanSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	api         *model.ApiModel
	gatewayConn *tarantool.Connection
	profile     *model.Profile
}

func (s *TestDeadmanSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Minute)

	broker := ClearAll(s.T())
	require.NotNil(s.T(), broker)

	s.gatewayConn = broker.Pool["api-gateway"]
	require.NotNil(s.T(), s.gatewayConn)

	s.api = model.NewApiModel(broker)
	require.NotNil(s.T(), s.api)

	err := s.api.UpdateIndexPrice(s.ctx, _marketId, _indexPrice)
	require.NoError(s.T(), err)

	var (
		wallet string = "0x123456"
	)

	// [Profile]
	profile, err := s.api.CreateProfile(s.ctx, model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), profile)
	s.profile = profile

	// [BalanceOps]
	_, err = s.api.DepositCredit(s.ctx, profile.ProfileId, _indexPrice)
	require.NoError(s.T(), err)
}

func (s *TestDeadmanSuite) TestDeadmanFlow() {
	data, err := s.api.DeadmanGet(s.ctx, s.profile.ProfileId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), data)

	var timeout uint = 10000
	data, err = s.api.DeadmanCreate(s.ctx, s.profile.ProfileId, timeout)
	require.NoError(s.T(), err)
	require.Equal(s.T(), data.ProfileId, s.profile.ProfileId)
	require.Equal(s.T(), data.Timeout, timeout)

	data, err = s.api.DeadmanDelete(s.ctx, s.profile.ProfileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), data.ProfileId, s.profile.ProfileId)
	require.Equal(s.T(), data.Timeout, timeout)

	data, err = s.api.DeadmanGet(s.ctx, s.profile.ProfileId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), data)
}

func TestDeadman(t *testing.T) {
	suite.Run(t, new(TestDeadmanSuite))
}
