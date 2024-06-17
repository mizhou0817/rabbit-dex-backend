package tests

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"testing"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	_marketId   = "BTC-USD"
	_indexPrice = 20000.0
)

type TestSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	api         *model.ApiModel
	profileConn *tarantool.Connection
	btcConn     *tarantool.Connection

	profile *model.Profile
}

func (s *TestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Minute)

	broker := ClearAll(s.T(), SkipInstances("api-gateway"))
	require.NotNil(s.T(), broker)

	s.profileConn = broker.Pool["profile"]
	require.NotNil(s.T(), s.profileConn)
	s.btcConn = broker.Pool["BTC-USD"]
	require.NotNil(s.T(), s.btcConn)

	s.api = model.NewApiModel(broker)
	require.NotNil(s.T(), s.api)

	err := s.api.UpdateIndexPrice(s.ctx, _marketId, _indexPrice)
	require.NoError(s.T(), err)

	//AdHoc: Ensure that tarantool knows about addresses
	_, err = s.api.AddContractMap(context.Background(),
		"",
		0,
		model.EXCHANGE_DEFAULT)
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

func (s *TestSuite) TearDownTest() {
	s.cancel()
}

type TestAPIPublicSuite struct {
	TestSuite
}

func (s *TestAPIPublicSuite) TestOrderAPI() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.LIMIT
		side      string  = model.LONG
		price     float64 = 200.0
		size      float64 = 2.0
	)

	// [OrderCreateRes]
	createdOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), createdOrder)
	time.Sleep(time.Second)

	// [OrderData]
	orders, err := s.api.GetAllOrders(s.ctx, profileId, marketId, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order := orders[0]
	require.Equal(s.T(),
		[]any{createdOrder.OrderId, createdOrder.MarketId, createdOrder.ProfileId, createdOrder.Price.String(), createdOrder.Size.String(), createdOrder.Side, createdOrder.Type},
		[]any{order.OrderId, order.MarketID, order.ProfileID, order.Price.String(), order.Size.String(), order.Side, order.OrderType},
	)

	orders, err = s.api.GetOpenOrders(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order = orders[0]
	require.Equal(s.T(),
		[]any{createdOrder.OrderId, createdOrder.MarketId, createdOrder.ProfileId, createdOrder.Price.String(), createdOrder.Size.String(), createdOrder.Side, createdOrder.Type},
		[]any{order.OrderId, order.MarketID, order.ProfileID, order.Price.String(), order.Size.String(), order.Side, order.OrderType},
	)

	var (
		price2 float64 = 100.0
		size2  float64 = 1.0
	)
	// [OrderAmendRes]
	amendOrder, err := s.api.OrderAmend(s.ctx, profileId, marketId, order.OrderId, &price2, &size2, nil, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{order.OrderId, order.MarketID, order.ProfileID, fmt.Sprint(price2), fmt.Sprint(size2), "amending"},
		[]any{amendOrder.OrderId, amendOrder.MarketId, amendOrder.ProfileId, amendOrder.Price.String(), amendOrder.Size.String(), amendOrder.Status},
	)

	// [OrderCancelRes]
	cancelOrder, err := s.api.OrderCancel(s.ctx, profileId, marketId, order.OrderId, "")
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{order.OrderId, order.MarketID, order.ProfileID, "canceling"},
		[]any{cancelOrder.OrderId, cancelOrder.MarketId, cancelOrder.ProfileId, cancelOrder.Status},
	)

	logrus.Info("*****.....TESTING cancel order:")
	logrus.Info(cancelOrder)

	// [-]
	err = s.api.CancelAll(s.ctx, profileId, true)
	require.NoError(s.T(), err)
}

func (s *TestAPIPublicSuite) TestCancelByCoid() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.LIMIT
		side      string  = model.LONG
		price     float64 = 200.0
		size      float64 = 2.0
	)

	//Test deduplication and cancel by client_order_id
	coids := make([]string, 0)
	for i := 0; i < 6; i++ {
		client_order_id := fmt.Sprintf("%d-%d", i, time.Now().UnixMicro())
		coids = append(coids, client_order_id)
	}

	s.api.WhiteListProfile(s.ctx, profileId)

	//Can't create 2 orders with same client_order_id
	newOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &coids[0], nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), newOrder)
	time.Sleep(time.Second)

	newOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &coids[0], nil, nil, nil, nil)
	logrus.Info(err)
	require.Error(s.T(), err)

	for _, coid := range coids[1:] {
		newOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &coid, nil, nil, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), newOrder)
		time.Sleep(time.Second)
	}

	orders, err := s.api.GetAllOrders(s.ctx, profileId, marketId, 100)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, len(coids))

	for i, order := range orders {
		require.Equal(s.T(), "open", order.Status)
		require.Equal(s.T(), coids[i], *order.ClientOrderId)
	}

	for _, coid := range coids {
		cancelOrder, err := s.api.OrderCancel(s.ctx, profileId, marketId, "", coid)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), cancelOrder)
		time.Sleep(time.Second)
	}

	orders, err = s.api.GetAllOrders(s.ctx, profileId, marketId, 100)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, len(coids))

	for _, order := range orders {
		require.Equal(s.T(), "canceled", order.Status)
	}

}

func (s *TestAPIPublicSuite) TestPositionAPI() {
	var (
		wallet string = "0x234567"
	)

	// [Profile]
	profile2, err := s.api.CreateProfile(s.ctx, model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), profile2)

	// [BalanceOps]
	_, err = s.api.DepositCredit(s.ctx, profile2.ProfileId, _indexPrice)
	require.NoError(s.T(), err)

	var (
		profileId1 uint    = s.profile.ProfileId
		profileId2 uint    = profile2.ProfileId
		marketId   string  = _marketId
		orderType  string  = model.LIMIT
		price      float64 = 20000.0
		size       float64 = 0.001
	)

	// [OrderCreateRes] do a trade
	order1, err := s.api.OrderCreate(s.ctx, profileId1, marketId, orderType, model.LONG, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order1)
	time.Sleep(time.Second)
	order2, err := s.api.OrderCreate(s.ctx, profileId2, marketId, orderType, model.SHORT, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order2)
	time.Sleep(time.Second)

	// [PositionData]
	positions, err := s.api.GetOpenPositions(s.ctx, profileId1)
	require.NoError(s.T(), err)
	require.Len(s.T(), positions, 1)
	pos := positions[0]
	require.Equal(s.T(),
		[]any{order1.MarketId, order1.ProfileId, order1.Side, order1.Size.String(), order1.Price.String()},
		[]any{pos.MarketID, pos.ProfileID, pos.Side, fmt.Sprint(size), pos.EntryPrice.String()},
	)

	positions, err = s.api.GetAllActivePositions(s.ctx, marketId, 0, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), positions, 2)
	pos = positions[0]
	require.Equal(s.T(),
		[]any{order1.MarketId, order1.ProfileId, order1.Side, order1.Size.String(), order1.Price.String()},
		[]any{pos.MarketID, pos.ProfileID, pos.Side, fmt.Sprint(size), pos.EntryPrice.String()},
	)
	pos = positions[1]
	require.Equal(s.T(),
		[]any{order2.MarketId, order2.ProfileId, order2.Side, order2.Size.String(), order2.Price.String()},
		[]any{pos.MarketID, pos.ProfileID, pos.Side, fmt.Sprint(size), pos.EntryPrice.String()},
	)

	{
		var (
			slPrice float64 = 100.0
			slSize  float64 = 0.01
			slSide  string  = model.LONG
		)
		createdSlOrder, err := s.api.OrderCreate(s.ctx, profileId1, marketId, model.STOP_LOSS, slSide, nil, nil, nil, &slPrice, &slSize, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), createdSlOrder)
		time.Sleep(time.Second)

		placedOrders, err := s.api.GetPlacedOrders(s.ctx, marketId, nil)
		require.NoError(s.T(), err)
		require.Len(s.T(), placedOrders, 1)
	}
}

func (s *TestAPIPublicSuite) TestMarketAPI() {
	var (
		marketId string = _marketId
	)

	// [MarketData]
	market, err := s.api.GetMarketData(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), market)
	require.Equal(s.T(),
		[]any{marketId, "active", fmt.Sprint(_indexPrice)},
		[]any{market.MarketID, *market.Status, market.IndexPrice.String()},
	)
}

func (s *TestAPIPublicSuite) TestOrderbookAPI() {
	var (
		marketId string = _marketId
	)

	// [OrderbookData]
	orderbook, err := s.api.GetOrderbookData(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), marketId, orderbook.MarketID)
}

func (s *TestAPIPublicSuite) TestProfileAPI() {
	var (
		profileId uint   = s.profile.ProfileId
		wallet    string = s.profile.Wallet
	)

	// [ProfileData]
	profile, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{s.profile.ProfileId, s.profile.Type, s.profile.Status, s.profile.Wallet},
		[]any{profile.ProfileID, *profile.ProfileType, *profile.Status, *profile.Wallet},
	)

	// [ProfileCache]
	cache, err := s.api.GetProfileCache(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{profile.ProfileID, profile.ProfileType, profile.Status, profile.Wallet, profile.Balance.String(), profile.AccountEquity.String(), profile.TotalPositionMargin.String(), profile.TotalOrderMargin.String(), profile.TotalNotional.String(), profile.AccountMargin.String(), profile.WithdrawableBalance.String(), profile.CumUnrealizedPnl.String(), profile.Health.String(), profile.AccountLeverage.String(), profile.CumTradingVolume.String(), profile.LastLiqCheck},
		[]any{cache.ProfileID, cache.ProfileType, cache.Status, cache.Wallet, cache.Balance.String(), cache.AccountEquity.String(), cache.TotalPositionMargin.String(), cache.TotalOrderMargin.String(), cache.TotalNotional.String(), cache.AccountMargin.String(), cache.WithdrawableBalance.String(), cache.CumUnrealizedPnl.String(), cache.Health.String(), cache.AccountLeverage.String(), cache.CumTradingVolume.String(), cache.LastLiqCheck},
	)

	cache, err = s.api.InvalidateCache(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{profile.ProfileID, profile.ProfileType, profile.Status, profile.Wallet, profile.Balance.String(), profile.AccountEquity.String(), profile.TotalPositionMargin.String(), profile.TotalOrderMargin.String(), profile.TotalNotional.String(), profile.AccountMargin.String(), profile.WithdrawableBalance.String(), profile.CumUnrealizedPnl.String(), profile.Health.String(), profile.AccountLeverage.String(), profile.CumTradingVolume.String(), profile.LastLiqCheck},
		[]any{cache.ProfileID, cache.ProfileType, cache.Status, cache.Wallet, cache.Balance.String(), cache.AccountEquity.String(), cache.TotalPositionMargin.String(), cache.TotalOrderMargin.String(), cache.TotalNotional.String(), cache.AccountMargin.String(), cache.WithdrawableBalance.String(), cache.CumUnrealizedPnl.String(), cache.Health.String(), cache.AccountLeverage.String(), cache.CumTradingVolume.String(), cache.LastLiqCheck},
	)

	cache, err = s.api.InvalidateCacheAndNotify(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{profile.ProfileID, profile.ProfileType, profile.Status, profile.Wallet, profile.Balance.String(), profile.AccountEquity.String(), profile.TotalPositionMargin.String(), profile.TotalOrderMargin.String(), profile.TotalNotional.String(), profile.AccountMargin.String(), profile.WithdrawableBalance.String(), profile.CumUnrealizedPnl.String(), profile.Health.String(), profile.AccountLeverage.String(), profile.CumTradingVolume.String(), profile.LastLiqCheck},
		[]any{cache.ProfileID, cache.ProfileType, cache.Status, cache.Wallet, cache.Balance.String(), cache.AccountEquity.String(), cache.TotalPositionMargin.String(), cache.TotalOrderMargin.String(), cache.TotalNotional.String(), cache.AccountMargin.String(), cache.WithdrawableBalance.String(), cache.CumUnrealizedPnl.String(), cache.Health.String(), cache.AccountLeverage.String(), cache.CumTradingVolume.String(), cache.LastLiqCheck},
	)

	// [Profile]
	byWallet, err := s.api.GetProfileByWalletForExchangeId(s.ctx, wallet, model.EXCHANGE_DEFAULT)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{s.profile.ProfileId, s.profile.Type, s.profile.Status, s.profile.Wallet},
		[]any{byWallet.ProfileId, byWallet.Type, byWallet.Status, byWallet.Wallet},
	)

	byId, err := s.api.GetProfileById(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{s.profile.ProfileId, s.profile.Type, s.profile.Status, s.profile.Wallet},
		[]any{byId.ProfileId, byId.Type, byId.Status, byId.Wallet},
	)
}

func (s *TestAPIPublicSuite) TestExtendedProfileAPI_None() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId

		price float64 = 20000.0
		size  float64 = 0.001

		wallet2 string = "0x234567"
	)

	profile2 := s.createDepositedTrader(wallet2, _indexPrice)
	_, _ = s.doTrade(marketId, profileId, profile2.ProfileId, price, size)

	time.Sleep(time.Second)

	profile1, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), profile1.Orders, 0)
	require.Len(s.T(), profile1.Positions, 1)
}

func (s *TestAPIPublicSuite) TestExtendedProfileAPI_SL() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId

		price float64 = 20000.0
		size  float64 = 0.001

		wallet2 string = "0x234567"
	)

	profile2 := s.createDepositedTrader(wallet2, _indexPrice)
	_, _ = s.doTrade(marketId, profileId, profile2.ProfileId, price, size)

	time.Sleep(time.Second)

	profile1, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), profile1.Orders, 0)
	require.Len(s.T(), profile1.Positions, 1)

	{
		var (
			slPrice float64 = 100.0
			slSize  float64 = 0.01
			slSide  string  = ""
		)
		createdSlOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, model.STOP_LOSS, slSide, nil, nil, nil, &slPrice, &slSize, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), createdSlOrder)
		time.Sleep(time.Second)

		placedOrders, err := s.api.GetPlacedOrders(s.ctx, marketId, nil)
		require.NoError(s.T(), err)
		require.Len(s.T(), placedOrders, 1)

		exProfile1, err := s.api.GetExtendedProfileData(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Len(s.T(), exProfile1.Orders, 1)
		require.Len(s.T(), exProfile1.Positions, 1)
		require.NotNil(s.T(), exProfile1.Positions[0].StopLoss)
		require.Nil(s.T(), exProfile1.Positions[0].TakeProfit)
	}
}

func (s *TestAPIPublicSuite) TestExtendedProfileAPI_TP() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId

		price float64 = 20000.0
		size  float64 = 0.001

		wallet2 string = "0x234567"
	)

	profile2 := s.createDepositedTrader(wallet2, _indexPrice)
	_, _ = s.doTrade(marketId, profileId, profile2.ProfileId, price, size)

	time.Sleep(time.Second)

	profile1, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), profile1.Orders, 0)
	require.Len(s.T(), profile1.Positions, 1)

	{
		var (
			tpPrice float64 = 100000.0
			tpSize  float64 = 0.01
			tpSide  string  = model.LONG
		)
		createdTpOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, model.TAKE_PROFIT, tpSide, nil, nil, nil, &tpPrice, &tpSize, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), createdTpOrder)
		time.Sleep(time.Second)

		placedOrders, err := s.api.GetPlacedOrders(s.ctx, marketId, nil)
		require.NoError(s.T(), err)
		require.Len(s.T(), placedOrders, 1)

		exProfile1, err := s.api.GetExtendedProfileData(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Len(s.T(), exProfile1.Orders, 1)
		require.Len(s.T(), exProfile1.Positions, 1)
		require.Nil(s.T(), exProfile1.Positions[0].StopLoss)
		require.NotNil(s.T(), exProfile1.Positions[0].TakeProfit)
	}
}

func (s *TestAPIPublicSuite) TestExtendedProfileAPI_SLTP() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId

		price float64 = 20000.0
		size  float64 = 0.001

		wallet2 string = "0x234567"
	)

	profile2 := s.createDepositedTrader(wallet2, _indexPrice)
	_, _ = s.doTrade(marketId, profileId, profile2.ProfileId, price, size)

	time.Sleep(time.Second)

	profile1, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), profile1.Orders, 0)
	require.Len(s.T(), profile1.Positions, 1)

	{
		var (
			slPrice float64 = 100.0
			slSize  float64 = 0.01
			slSide  string  = model.LONG
		)
		createdSlOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, model.STOP_LOSS, slSide, nil, nil, nil, &slPrice, &slSize, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), createdSlOrder)

		var (
			tpPrice float64 = 100000.0
			tpSize  float64 = 0.01
			tpSide  string  = model.LONG
		)
		createdTpOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, model.TAKE_PROFIT, tpSide, nil, nil, nil, &tpPrice, &tpSize, nil, nil)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), createdTpOrder)
		time.Sleep(time.Second)

		placedOrders, err := s.api.GetPlacedOrders(s.ctx, marketId, nil)
		require.NoError(s.T(), err)
		require.Len(s.T(), placedOrders, 2)

		exProfile1, err := s.api.GetExtendedProfileData(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Len(s.T(), exProfile1.Orders, 2)
		require.Len(s.T(), exProfile1.Positions, 1)
		require.NotNil(s.T(), exProfile1.Positions[0].StopLoss)
		require.NotNil(s.T(), exProfile1.Positions[0].TakeProfit)
	}
}

func (s *TestAPIPublicSuite) TestExchangeAPI() {
	evalScript(s.ctx, s.profileConn, `
	local trading_fee = require("decimal").new(100)
	local total_balance = require("decimal").new(200)
	box.space.exchange_total:replace({1, trading_fee, total_balance})
	`)

	// [ExchangeData]
	exchange, err := s.api.GetExchangeData(s.ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{int64(1), "100", "200"},
		[]any{exchange.Id, exchange.TradingFee.String(), exchange.TotalBalance.String()},
	)
}

func (s *TestAPIPublicSuite) TestLeverageAPI() {
	var (
		marketId  string = _marketId
		profileId uint   = s.profile.ProfileId
		leverage  uint   = 12
	)

	value, err := s.api.UpdateLeverage(s.ctx, marketId, profileId, leverage)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fmt.Sprint(leverage), value.String())
}

func (s *TestAPIPublicSuite) TestCandlesAPI() {
	evalScript(s.ctx, s.btcConn, `
	local candles = require('app.engine.candles')
	local price = require("decimal").new(100)
	local size = require("decimal").new(2)

	local res = candles.add_all_periods(price, size, 0)
	if res.error ~= nil then
		error(res.error)
	end
	`)

	var (
		marketId string = _marketId
		period   uint   = 1440
	)

	// [CandleData]
	candles, err := s.api.GetCandles(s.ctx, marketId, period, 0, time.Now().Unix())
	require.NoError(s.T(), err)
	require.Len(s.T(), candles, 301)
	for _, c := range candles {
		require.Equal(s.T(),
			[]any{"0", "100", "100", "100", "100"},
			[]any{c.Volume.String(), c.Low.String(), c.High.String(), c.Open.String(), c.Close.String()},
		)
	}
}

func TestAPIPublic(t *testing.T) {
	suite.Run(t, new(TestAPIPublicSuite))
}

type TestPingLimitSuite struct {
	TestSuite
}

func (s *TestPingLimitSuite) TestPingLimit() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.PING_LIMIT
		side      string  = model.LONG
		price     float64 = 200.0
		size      float64 = 0.01
	)

	var err error
	var createdOrder model.OrderCreateRes

	err = s.api.CancelAll(s.ctx, profileId, false)
	require.NoError(s.T(), err)
	time.Sleep(time.Second)

	s.api.TestUpdateFairPrice(s.ctx, marketId, float64(18000))

	// [OrderCreateRes]
	createdOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.ErrorContains(s.T(), err, "mid_price not possible best_ask=0 best_bid=0")
	require.Empty(s.T(), createdOrder)
	logrus.Info(err)

	//Create bid only
	bestbid := float64(18810)
	makerOrder, err1 := s.api.OrderCreate(s.ctx, profileId, marketId, model.LIMIT, model.LONG, &bestbid, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err1)
	require.NotEmpty(s.T(), makerOrder)
	time.Sleep(time.Second)

	orders, err2 := s.api.GetOpenOrders(s.ctx, marketId, profileId)
	require.NoError(s.T(), err2)
	require.Len(s.T(), orders, 1)

	createdOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.ErrorContains(s.T(), err, "mid_price not possible best_ask=0 best_bid=18810.0")
	require.Empty(s.T(), createdOrder)
	logrus.Info(err)

	//Create ask
	bestask := float64(18811)
	makerOrder1, err1 := s.api.OrderCreate(s.ctx, profileId, marketId, model.LIMIT, model.SHORT, &bestask, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err1)
	require.NotEmpty(s.T(), makerOrder1)
	time.Sleep(time.Second)

	createdOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.ErrorContains(s.T(), err, "mid_price not possible has=18810.0 best_ask=18811.0 best_bid=18810.0")
	require.Empty(s.T(), createdOrder)
	logrus.Info(err)

	_, err3 := s.api.OrderCancel(s.ctx, profileId, marketId, makerOrder1.OrderId, "")
	require.NoError(s.T(), err3)
	time.Sleep(time.Second)

	bestask = float64(18820)
	makerOrder2, err1 := s.api.OrderCreate(s.ctx, profileId, marketId, model.LIMIT, model.SHORT, &bestask, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err1)
	require.NotEmpty(s.T(), makerOrder2)
	time.Sleep(time.Second)

	createdOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), createdOrder)
	time.Sleep(time.Second)
	logrus.Infof("ping_limit id = %s", createdOrder.OrderId)

	orders, err = s.api.GetAllOrders(s.ctx, profileId, marketId, 100)
	require.NoError(s.T(), err)

	count := 0
	for _, order := range orders {
		logrus.Infof("OrderId=%s type=%s status=%s price=%s size=%s side=%s",
			order.OrderId,
			order.OrderType,
			order.Status,
			order.Price.String(),
			order.Size.String(),
			order.Side)
		if order.OrderType == orderType {
			count += 1
			require.Equal(s.T(), model.CLOSED, order.Status)
		}
	}
	require.Equal(s.T(), 2, count)

	err = s.api.CancelAll(s.ctx, profileId, false)
	require.NoError(s.T(), err)
	time.Sleep(time.Second)

}

func TestPingLimit(t *testing.T) {
	suite.Run(t, new(TestPingLimitSuite))
}

type TestAPIPublicOrderSuite struct {
	TestSuite
}

func (s *TestAPIPublicOrderSuite) TestOrderAPI() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.LIMIT
		side      string  = model.LONG
		price     float64 = 200.0
		size      float64 = 2.0
	)

	// [OrderCreateRes]
	createdOrder, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), createdOrder)
	time.Sleep(time.Second)

	// [OrderData]
	orders, err := s.api.GetAllOrders(s.ctx, profileId, marketId, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order := orders[0]
	require.Equal(s.T(),
		[]any{createdOrder.OrderId, createdOrder.MarketId, createdOrder.ProfileId, createdOrder.Price.String(), createdOrder.Size.String(), createdOrder.Side, createdOrder.Type},
		[]any{order.OrderId, order.MarketID, order.ProfileID, order.Price.String(), order.Size.String(), order.Side, order.OrderType},
	)

	orders, err = s.api.GetOpenOrders(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order = orders[0]
	require.Equal(s.T(),
		[]any{createdOrder.OrderId, createdOrder.MarketId, createdOrder.ProfileId, createdOrder.Price.String(), createdOrder.Size.String(), createdOrder.Side, createdOrder.Type},
		[]any{order.OrderId, order.MarketID, order.ProfileID, order.Price.String(), order.Size.String(), order.Side, order.OrderType},
	)

	var (
		price2 float64 = 100.0
		size2  float64 = 1.0
	)
	// [OrderAmendRes]
	amendOrder, err := s.api.OrderAmend(s.ctx, profileId, marketId, order.OrderId, &price2, &size2, nil, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{order.OrderId, order.MarketID, order.ProfileID, fmt.Sprint(price2), fmt.Sprint(size2), "amending"},
		[]any{amendOrder.OrderId, amendOrder.MarketId, amendOrder.ProfileId, amendOrder.Price.String(), amendOrder.Size.String(), amendOrder.Status},
	)

	// [OrderCancelRes]
	cancelOrder, err := s.api.OrderCancel(s.ctx, profileId, marketId, order.OrderId, "")
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{order.OrderId, order.MarketID, order.ProfileID, "canceling"},
		[]any{cancelOrder.OrderId, cancelOrder.MarketId, cancelOrder.ProfileId, cancelOrder.Status},
	)

	logrus.Info("*****.....TESTING cancel order:")
	logrus.Info(cancelOrder)

	// [-]
	err = s.api.CancelAll(s.ctx, profileId, false)
	require.NoError(s.T(), err)
	time.Sleep(time.Second)

	//Amend by price only scenario

	createdOrder, err = s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), createdOrder)
	time.Sleep(time.Second)
	logrus.Info("created1:")
	logrus.Info(createdOrder.OrderId)
	logrus.Info(createdOrder.Price)
	logrus.Info(createdOrder.Size)

	// [OrderData]
	orders, err = s.api.GetOpenOrders(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order = orders[0]
	require.Equal(s.T(),
		[]any{createdOrder.OrderId, createdOrder.MarketId, createdOrder.ProfileId, createdOrder.Price.String(), createdOrder.Size.String(), createdOrder.Side, createdOrder.Type},
		[]any{order.OrderId, order.MarketID, order.ProfileID, order.Price.String(), order.Size.String(), order.Side, order.OrderType},
	)
	//profile_id uint, market_id, order_id string, new_price, new_size, new_trigger_price, new_size_percent *float64
	new_price := float64(201)
	amendedOrder, err := s.api.OrderAmend(s.ctx, profileId, marketId, createdOrder.OrderId, &new_price, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), amendedOrder)
	logrus.Info("amended1:")
	logrus.Info(amendedOrder.OrderId)
	logrus.Info(amendedOrder.Price)
	logrus.Info(amendedOrder.Size)
	require.Empty(s.T(), amendedOrder.Size)
	time.Sleep(time.Second)

	orders, err = s.api.GetOpenOrders(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), orders, 1)
	order = orders[0]
	require.Equal(s.T(),
		[]any{amendedOrder.OrderId, amendedOrder.MarketId, amendedOrder.ProfileId, amendedOrder.Price.String(), "2"},
		[]any{order.OrderId, order.MarketID, order.ProfileID, "201", order.Size.String()},
	)

	err = s.api.CancelAll(s.ctx, profileId, false)
	require.NoError(s.T(), err)
	time.Sleep(time.Second)

}

func TestAPIPublicOrder(t *testing.T) {
	suite.Run(t, new(TestAPIPublicOrderSuite))
}

type TestAPITestSuite struct {
	TestSuite
}

func (s *TestAPITestSuite) TestFillAPI() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.LIMIT
		price     float64 = 20000.0
		size      float64 = 0.01
	)

	client_order_id := "some_order_id"
	logrus.Infof("... client_order_id = %s", client_order_id)

	client_order_id2 := "some_order_id2"
	logrus.Infof("... client_order_id2 = %s", client_order_id2)

	// [OrderCreateRes] do a trade
	order1, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, model.LONG, &price, &size, &client_order_id, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order1)
	time.Sleep(time.Second)
	order2, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, model.SHORT, &price, &size, &client_order_id2, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order2)
	time.Sleep(time.Second)

	// [FillData, TradeData]
	fills, trades, err := s.api.GetAllFills(s.ctx, []string{marketId})
	require.NoError(s.T(), err)

	require.Len(s.T(), fills, 2)
	require.Equal(s.T(),
		[]any{order1.ProfileId, order1.MarketId, order1.OrderId, order1.Price.String(), order1.Size.String(), order1.Side, true, false},
		[]any{fills[0].ProfileId, fills[0].MarketId, fills[0].OrderId, fills[0].Price.String(), fills[0].Size.String(), fills[0].Side, fills[0].IsMaker, fills[0].Liquidation},
	)
	require.Equal(s.T(),
		[]any{order2.ProfileId, order2.MarketId, order2.OrderId, order2.Price.String(), order2.Size.String(), order2.Side, false, false},
		[]any{fills[1].ProfileId, fills[1].MarketId, fills[1].OrderId, fills[1].Price.String(), fills[1].Size.String(), fills[1].Side, fills[1].IsMaker, fills[1].Liquidation},
	)
	require.Equal(s.T(), []any{fills[0].TradeId}, []any{fills[1].TradeId})

	require.Equal(s.T(), client_order_id, *fills[0].ClientOrderId)
	require.Equal(s.T(), client_order_id2, *fills[1].ClientOrderId)

	require.Len(s.T(), trades, 1)
	trade := trades[0]
	require.Equal(s.T(),
		[]any{order2.MarketId, order2.Price.String(), order2.Size.String(), order2.Side, false},
		[]any{trade.MarketId, trade.Price.String(), trade.Size.String(), trade.TakerSide, trade.Liquidation},
	)
}

func TestAPITest(t *testing.T) {
	suite.Run(t, new(TestAPITestSuite))
}

type TestAPILiqSuite struct {
	TestSuite
}

func (s *TestAPILiqSuite) TestUpdateInsuranceId() {
	var (
		insuranceId uint   = s.profile.ProfileId
		marketId    string = _marketId
	)

	err := s.api.UpdateInsuranceId(s.ctx, insuranceId, []string{marketId})
	require.NoError(s.T(), err)
}

func (s *TestAPILiqSuite) TestCreateInsuranceProfile() {
	var (
		wallet string = "0xinsurance"
	)

	profile, err := s.api.CreateInsuranceProfile(s.ctx, wallet)
	require.NoError(s.T(), err)
	require.Equal(s.T(),
		[]any{uint(1), "insurance", "active"},
		[]any{profile.ProfileId, profile.Type, profile.Status},
	)
}

func (s *TestAPILiqSuite) TestIsInv3Valid() {
	err := MakeIntegrityValid(s.ctx, s.profileConn)
	require.NoError(s.T(), err)

	ok, err := s.api.IsInv3Valid(s.ctx, 0.0)
	require.NoError(s.T(), err)
	require.True(s.T(), ok)
}

func (s *TestAPILiqSuite) TestHighPriorityCancelAll() {
	var (
		profileId     uint = s.profile.ProfileId
		isLiquidation bool = false
	)

	err := s.api.HighPriorityCancelAll(s.ctx, profileId, isLiquidation)
	require.NoError(s.T(), err)
}

func (s *TestAPILiqSuite) TestIsCancellAllAccepted() {
	var (
		profileId uint = s.profile.ProfileId
	)

	ok, err := s.api.IsCancellAllAccepted(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.True(s.T(), ok)
}

func (s *TestAPILiqSuite) TestGetWinningPositions() {
	var (
		marketId string = _marketId
		side     string = "long"
	)
	profiles, err := s.api.GetWinningPositions(s.ctx, marketId, side)
	require.NoError(s.T(), err)
	require.Len(s.T(), profiles, 0) //TODO: add condition to get some profiles
}

func (s *TestAPILiqSuite) TestQueueLiqActions() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId

		price float64 = 20000.0
		size  float64 = 0.001

		wallet2 string = "0x234567"
	)

	profile2 := s.createDepositedTrader(wallet2, _indexPrice)
	_, _ = s.doTrade(marketId, profileId, profile2.ProfileId, price, size)

	time.Sleep(time.Second)

	upd1Profile, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), upd1Profile.Orders, 0)
	require.Len(s.T(), upd1Profile.Positions, 1)

	sellAll := model.Action{
		Kind:     model.APlaceSellOrders,
		TraderId: profileId,
		MarketId: marketId,
		Size:     *tdecimal.NewDecimal(decimal.NewFromFloat(size)),
		Price:    *tdecimal.NewDecimal(decimal.NewFromFloat(price / 2)),
	}
	err = s.api.QueueLiqActions(s.ctx, []model.Action{sellAll})
	require.NoError(s.T(), err)

	// trade for liquidation trade to happen
	_, _ = s.doTrade(marketId, profile2.ProfileId, profileId, math.Floor(price/1.01), size)

	time.Sleep(time.Second)

	upd2Profile, err := s.api.GetProfileData(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Len(s.T(), upd2Profile.Orders, 1)
	require.Len(s.T(), upd2Profile.Positions, 0)

	looseAmount := decimal.NewFromFloat(price / 2 * size)
	require.Equal(s.T(), 0, upd1Profile.Balance.Cmp(upd2Profile.Balance.Decimal.Add(looseAmount)))
}

func (s *TestAPILiqSuite) TestLiquidationBatch() {
	var (
		lastLiqChecked uint = 0
		limit          int  = 1
	)

	err := MakeIntegrityValid(s.ctx, s.profileConn)
	require.NoError(s.T(), err)

	caches, err := s.api.LiquidationBatch(s.ctx, &lastLiqChecked, limit)
	require.NoError(s.T(), err)
	require.Len(s.T(), caches, 0) //TODO: add condition to get some profiles
}

func (s *TestAPILiqSuite) TestProfileUpdateStatus() {
	var (
		profileId uint   = s.profile.ProfileId
		status    string = "liquidating"
	)

	err := s.api.ProfileUpdateStatus(s.ctx, profileId, status)
	require.NoError(s.T(), err)
}

func (s *TestAPILiqSuite) TestProfileUpdateLastLiqChecked() {
	var (
		profileId uint = s.profile.ProfileId
	)

	liqChecked, err := s.api.ProfileUpdateLastLiqChecked(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.Less(s.T(), int64(0), liqChecked)
}

func TestAPILiq(t *testing.T) {
	suite.Run(t, new(TestAPILiqSuite))
}

type TestAPIDepositWithdrawSuite struct {
	TestSuite
}

func (s *TestAPIDepositWithdrawSuite) TestAcquireWithdrawLock() {
	var (
		profileId = s.profile.ProfileId
	)

	lock, err := s.api.AcquireWithdrawLock(s.ctx, profileId)
	require.NoError(s.T(), err)
	_ = lock
}

func (s *TestAPIDepositWithdrawSuite) TestReleaseWithdrawLock() {
	var (
		profileId = s.profile.ProfileId
	)

	lock, err := s.api.ReleaseWithdrawLock(s.ctx, profileId)
	require.NoError(s.T(), err)
	_ = lock
}

func (s *TestAPIDepositWithdrawSuite) TestCheckWithdrawAllowed() {
	var (
		profileId = s.profile.ProfileId
	)

	ok := s.api.CheckWithdrawAllowed(s.ctx, profileId)
	require.True(s.T(), ok)
}

func (s *TestAPIDepositWithdrawSuite) TestCreateDeposit() {
	var (
		profileId = s.profile.ProfileId
		wallet    = s.profile.Wallet
		amount    = tdecimal.NewDecimal(decimal.New(100, 0))
		txhash    = "hash1"
	)

	ops, err := s.api.CreateDeposit(s.ctx, profileId, wallet, amount, txhash, "", 0)
	require.NoError(s.T(), err)
	_ = ops
}

func (s *TestAPIDepositWithdrawSuite) TestProcessDeposit() {
	var (
		profileId = s.profile.ProfileId
		deposit   = model.Deposit{Id: "deposit1", Wallet: s.profile.Wallet, Amount: tdecimal.NewDecimal(decimal.New(100, 0)), Tx: "hash1"}
	)

	err := s.api.ProcessDeposit(s.ctx, profileId, deposit, false)
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawSuite) TestProcessDepositUnknown() {
	var (
		deposit = model.Deposit{Id: "deposit1", Wallet: s.profile.Wallet, Amount: tdecimal.NewDecimal(decimal.New(100, 0))}
	)

	err := s.api.ProcessDepositUnknown(s.ctx, deposit)
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawSuite) TestCreateWithdrawal() {
	var (
		profileId = s.profile.ProfileId
		wallet    = s.profile.Wallet
		amount    = tdecimal.NewDecimal(decimal.New(100, 0))
	)

	ops, err := s.api.CreateWithdrawal(s.ctx, profileId, wallet, amount, model.EXCHANGE_DEFAULT)
	require.NoError(s.T(), err)
	_ = ops
}

func (s *TestAPIDepositWithdrawSuite) TestGetPendingWithdrawals() {
	w, err := s.api.GetPendingWithdrawals(s.ctx, model.EXCHANGE_DEFAULT, 0)
	require.NoError(s.T(), err)
	_ = w
}

func (s *TestAPIDepositWithdrawSuite) TestCompletedWithdrawals() {
	err := s.api.CompletedWithdrawals(s.ctx, []*model.WithdrawalTxInfo{})
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawSuite) TestGetWithdrawalsSuspended() {
	val, err := s.api.GetWithdrawalsSuspended(s.ctx)
	require.NoError(s.T(), err)
	require.False(s.T(), val)
}

func (s *TestAPIDepositWithdrawSuite) TestSuspendWithdrawals() {
	err := s.api.SuspendWithdrawals(s.ctx)
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawSuite) TestGetLastProcessedBlockNumber() {
	block, err := s.api.GetLastProcessedBlockNumber(s.ctx, "", 0, "")
	require.Error(s.T(), errors.New("Unable to parse last processed block number"), err)
	require.Nil(s.T(), block, nil)
}

func (s *TestAPIDepositWithdrawSuite) TestSetLastProcessedBlockNumber() {
	var lastProcessed *big.Int
	err := s.api.SetLastProcessedBlockNumber(s.ctx, lastProcessed, "", 0, "")
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawSuite) TestGetLastProcessedDepositBlockNumber() {
	block, err := s.api.GetLastProcessedBlockNumber(s.ctx, "", 0, "")
	require.Error(s.T(), errors.New("Unable to parse last processed block number"), err)
	require.Nil(s.T(), block, nil)
}

func (s *TestAPIDepositWithdrawSuite) TestSetLastProcessedDepositBlockNumber() {
	var lastProcessed *big.Int
	err := s.api.SetLastProcessedBlockNumber(s.ctx, lastProcessed, "", 0, "")
	require.NoError(s.T(), err)
}

func TestAPIDepositWithdraw(t *testing.T) {
	suite.Run(t, new(TestAPIDepositWithdrawSuite))
}

type TestAPIDepositWithdrawDebugSuite struct {
	TestSuite
}

func (s *TestAPIDepositWithdrawDebugSuite) TestGetSettlementState() {
	state, err := s.api.GetSettlementState(s.ctx)
	require.NoError(s.T(), err)
	_ = state
}

func (s *TestAPIDepositWithdrawDebugSuite) TestGetProcessingOps() {
	ops, err := s.api.GetProcessingOps(s.ctx)
	require.NoError(s.T(), err)
	_ = ops
}

func (s *TestAPIDepositWithdrawDebugSuite) TestMergeUnknown() {
	_, err := s.api.MergeUnknown(s.ctx)
	require.NoError(s.T(), err)
}

func (s *TestAPIDepositWithdrawDebugSuite) TestDeleteUnknown() {
	_, err := s.api.DeleteUnknown(s.ctx)
	require.NoError(s.T(), err)
}

func TestAPIDepositWithdrawDebug(t *testing.T) {
	suite.Run(t, new(TestAPIDepositWithdrawDebugSuite))
}

func (s *TestSuite) createDepositedTrader(wallet string, amount float64) *model.Profile {
	// [Profile]
	profile, err := s.api.CreateProfile(s.ctx, model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), profile)

	// [BalanceOps]
	_, err = s.api.DepositCredit(s.ctx, profile.ProfileId, amount)
	require.NoError(s.T(), err)

	return profile
}

func (s *TestSuite) doTrade(marketId string, traderLong, traderShort uint, price, size float64) (bid model.OrderCreateRes, ask model.OrderCreateRes) {
	var (
		orderType string = model.LIMIT
	)

	// [OrderCreateRes] do a trade
	order1, err := s.api.OrderCreate(s.ctx, traderLong, marketId, orderType, model.LONG, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order1)
	//time.Sleep(time.Second)

	order2, err := s.api.OrderCreate(s.ctx, traderShort, marketId, orderType, model.SHORT, &price, &size, nil, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order2)

	return order1, order2
}

type TestAdminApiSuite struct {
	TestSuite
}

func (s *TestAdminApiSuite) TestAdminApi() {
	var (
		marketId string = _marketId
	)

	// [MarketData]
	market, err := s.api.GetMarketData(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), market)
	require.Equal(s.T(),
		[]any{marketId, "active", fmt.Sprint(_indexPrice)},
		[]any{market.MarketID, *market.Status, market.IndexPrice.String()},
	)

	newUrl := "https://new_url.com"
	market, err = s.api.MarketUpdateIconUrl(s.ctx, marketId, newUrl)
	require.NoError(s.T(), err)
	require.Equal(s.T(), newUrl, market.IconUrl)

	// Invalid URL return errr
	newUrl = "new_url@com"
	market, err = s.api.MarketUpdateIconUrl(s.ctx, marketId, newUrl)
	require.Error(s.T(), err)

	newTitle := "SomeNewTitle"
	market, err = s.api.MarketUpdateTitle(s.ctx, marketId, newTitle)
	require.NoError(s.T(), err)
	require.Equal(s.T(), newTitle, market.MarketTitle)

}

func TestAdminApi(t *testing.T) {
	suite.Run(t, new(TestAdminApiSuite))
}

type TestAPIAirdropSuite struct {
	TestSuite
}

func (s *TestAPIAirdropSuite) TestWholeAirdropFlow() {
	var (
		profileId = s.profile.ProfileId
	)

	airdrops := []string{"air1", "air2"}

	minTimestamp := int64(1682693219000000)
	totalRewards := float64(100)
	claimable := float64(1)

	for _, air := range airdrops {
		airdrop, err := s.api.CreateAirdrop(s.ctx, air, minTimestamp+100, minTimestamp+200)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), airdrop)
		require.Equal(s.T(), air, airdrop.Title)

		profileAirdrop, err := s.api.SetProfileTotal(s.ctx, profileId, air, totalRewards, claimable)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), airdrop)
		require.Equal(s.T(), profileId, profileAirdrop.ProfileId)
		require.Equal(s.T(), air, profileAirdrop.AirdropTitle)
		require.Equal(s.T(), uint64(totalRewards), profileAirdrop.TotalRewards.BigInt().Uint64())
		require.Equal(s.T(), uint64(claimable), profileAirdrop.Claimable.BigInt().Uint64())

	}

	//get all airdrops
	all, err := s.api.GetProfileAirdrops(s.ctx, profileId)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), all)

	for i, pa := range all {
		require.Equal(s.T(), profileId, pa.ProfileId)
		require.Equal(s.T(), airdrops[i], pa.AirdropTitle)
		require.Equal(s.T(), uint64(totalRewards), pa.TotalRewards.BigInt().Uint64())
		require.Equal(s.T(), uint64(claimable), pa.Claimable.BigInt().Uint64())
	}

	//update claimable
	for _, air := range airdrops {
		r, err := s.api.UpdateProfileClaimable(s.ctx, profileId, air)
		if err == nil {
			require.NoError(s.T(), err)
			require.NotEmpty(s.T(), r)
			require.Equal(s.T(), air, r.AirdropTitle)
		} else {
			//require.Equal(s.T(), "NO_FILLS_FOR_PERIOD", err.Error())
		}
	}

	err = s.api.UpdateAllProfileAirdrops(s.ctx, profileId)
	require.NoError(s.T(), err)

	was := false
	//Claim all
	for _, air := range airdrops {
		r, err := s.api.ProfileClaimAll(s.ctx, profileId, air)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), r)
		require.Equal(s.T(), model.AIRDROP_CLAIMING_STATUS, r.Status)

		//Get pending
		r, err = s.api.PendingClaimOps(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Equal(s.T(), model.AIRDROP_CLAIMING_STATUS, r.Status)

		all, err = s.api.GetProfileAirdrops(s.ctx, profileId)

		for i, pa := range all {
			if pa.AirdropTitle == r.AirdropTitle {
				all[i].Claimable = r.Amount
				all[i].Claimed = *tdecimal.NewDecimal(all[i].Claimed.Sub(r.Amount.Decimal))
			}
		}

		r, err = s.api.FinishClaim(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Equal(s.T(), model.AIRDROP_CLAIMED_STATUS, r.Status)

		r, err = s.api.PendingClaimOps(s.ctx, profileId)
		require.NoError(s.T(), err)
		require.Empty(s.T(), r)

		was = true
	}

	require.Equal(s.T(), true, was)

}

func TestAPIAirdrop(t *testing.T) {
	suite.Run(t, new(TestAPIAirdropSuite))
}

func TestTiers(t *testing.T) {
	suite.Run(t, new(TestTiersSuite))
}

type TestTiersSuite struct {
	TestSuite
}

func (s *TestTiersSuite) TestTiersFlow() {
	var (
		profileId uint   = s.profile.ProfileId
		marketId  string = _marketId
	)

	// [OrderCreateRes] do a trade
	tiers, err := s.api.GetTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tiers)
	i_len := len(tiers)

	special_tiers, err := s.api.GetSpecialTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), special_tiers)

	profile_tiers, err := s.api.GetProfileTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), profile_tiers)

	new_id := uint(i_len + 2)

	new_tier, err := s.api.AddTier(s.ctx, marketId, new_id, "new_test_tier", 0.1, 0.1, 10, 20)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "new_test_tier", new_tier.Title)
	require.Equal(s.T(), new_id, uint(new_tier.Tier))

	dMakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(0.2))
	dTakerFee := tdecimal.NewDecimal(decimal.NewFromFloat(0.2))
	dMinVolume := tdecimal.NewDecimal(decimal.NewFromFloat(100.0))
	dAssets := tdecimal.NewDecimal(decimal.NewFromFloat(200.0))

	// Edit test
	new_tier, err = s.api.EditTier(s.ctx, marketId, new_id, "new_test_tier", 0.2, 0.2, 100, 200)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "new_test_tier", new_tier.Title)
	require.True(s.T(), dMakerFee.Equal(new_tier.MakerFee.Decimal))
	require.True(s.T(), dTakerFee.Equal(new_tier.TakerFee.Decimal))
	require.True(s.T(), dAssets.Equal(new_tier.MinAssets.Decimal))
	require.True(s.T(), dMinVolume.Equal(new_tier.MinVolume.Decimal))

	tiers, err = s.api.GetTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), i_len+1, len(tiers))

	special_tier_id := uint(1)
	p_tier, err := s.api.AddProfileToTier(s.ctx, marketId, profileId, uint(special_tier_id))
	require.Error(s.T(), err)
	require.Empty(s.T(), p_tier)

	new_special_tier, err := s.api.AddSpecialTier(s.ctx, marketId, uint(special_tier_id), "new_test_tier", 0.1, 0.1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "new_test_tier", new_special_tier.Title)

	dMakerFee = tdecimal.NewDecimal(decimal.NewFromFloat(0.2))
	dTakerFee = tdecimal.NewDecimal(decimal.NewFromFloat(0.2))

	// Edit test
	new_special_tier, err = s.api.EditSpecialTier(s.ctx, marketId, uint(special_tier_id), "new_test_tier", 0.2, 0.2)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), new_special_tier)
	require.True(s.T(), dMakerFee.Equal(new_special_tier.MakerFee.Decimal))
	require.True(s.T(), dTakerFee.Equal(new_special_tier.TakerFee.Decimal))

	p_tier, err = s.api.AddProfileToTier(s.ctx, marketId, profileId, uint(special_tier_id))
	require.NoError(s.T(), err)
	require.Equal(s.T(), profileId, p_tier.ProfileID)
	require.Equal(s.T(), special_tier_id, uint(p_tier.TierID))

	special_tiers, err = s.api.GetSpecialTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(special_tiers))

	profile_tiers, err = s.api.GetProfileTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, len(profile_tiers))

	//which Tier
	tier, err := s.api.WhichTier(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "new_test_tier", tier.Title)

	//Remove tiers
	_, err = s.api.RemoveTier(s.ctx, marketId, new_id)
	require.NoError(s.T(), err)

	_, err = s.api.RemoveSpecialTier(s.ctx, marketId, uint(special_tier_id))
	require.NoError(s.T(), err)

	_, err = s.api.RemoveProfileTier(s.ctx, marketId, profileId)
	require.NoError(s.T(), err)

	tiers, err = s.api.GetTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), tiers)

	special_tiers, err = s.api.GetSpecialTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), special_tiers)

	profile_tiers, err = s.api.GetProfileTiers(s.ctx, marketId)
	require.NoError(s.T(), err)
	require.Empty(s.T(), profile_tiers)

}

func TestMissingData(t *testing.T) {
	suite.Run(t, new(TestMissingDataSuite))
}

type TestMissingDataSuite struct {
	TestSuite
}

func (s *TestMissingDataSuite) TestFlow() {
	var (
		profileId uint    = s.profile.ProfileId
		marketId  string  = _marketId
		orderType string  = model.LIMIT
		side      string  = model.LONG
		price     float64 = 200.0
		size      float64 = 2.0
	)

	// Create order without client_order_id
	// [OrderCreateRes]
	client_order_id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixMicro())
	order, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &client_order_id, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order)
	time.Sleep(time.Second)

	cancelOrder, err := s.api.OrderCancel(s.ctx, profileId, marketId, order.OrderId, client_order_id)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), cancelOrder)
	require.Equal(s.T(), order.OrderId, cancelOrder.OrderId)
	require.Equal(s.T(), *order.ClientOrderId, cancelOrder.ClientOrderId)
	require.Equal(s.T(), order.OrderId, cancelOrder.OrderId)

	client_order_id2 := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixMicro())
	order2, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &client_order_id2, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order2)
	time.Sleep(time.Second)

	cancelOrder2, err := s.api.OrderCancel(s.ctx, profileId, marketId, "", client_order_id2)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), cancelOrder2)
	require.Equal(s.T(), *order2.ClientOrderId, cancelOrder2.ClientOrderId)
	require.Equal(s.T(), order2.OrderId, cancelOrder2.OrderId)

	logrus.Info(cancelOrder2)

	client_order_id3 := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixMicro())
	order3, err := s.api.OrderCreate(s.ctx, profileId, marketId, orderType, side, &price, &size, &client_order_id3, nil, nil, nil, nil)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order2)
	time.Sleep(time.Second)

	cancelOrder3, err := s.api.OrderCancel(s.ctx, profileId, marketId, order3.OrderId, "")
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), cancelOrder3)
	require.Equal(s.T(), *order3.ClientOrderId, cancelOrder3.ClientOrderId)
	require.Equal(s.T(), order3.OrderId, cancelOrder3.OrderId)

	logrus.Info(cancelOrder3)

	_, err = s.api.OrderCancel(s.ctx, profileId, marketId, "", "")
	require.EqualError(s.T(), err, "ORDER_ID_OR_CLIENT_ORDER_ID_REQUIRED")

	logrus.Info(err)
}

func TestVaultPermissions(t *testing.T) {
	suite.Run(t, new(TestVaultPermissionsSuite))
}

type TestVaultPermissionsSuite struct {
	TestSuite
}

func (s *TestVaultPermissionsSuite) TestFlow() {

	// [OrderCreateRes] do a trade
	vault := "0XSOMEVAULT"
	wallet := "0XSOMEWALLET"
	role := 1

	err := s.api.IsValidSigner(s.ctx, vault, wallet, uint(role))
	require.Error(s.T(), err)

	err = s.api.AddPermission(s.ctx, vault, wallet, uint(role))
	require.NoError(s.T(), err)

	err = s.api.IsValidSigner(s.ctx, vault, wallet, uint(role))
	require.NoError(s.T(), err)

	err = s.api.IsValidSigner(s.ctx, vault, wallet, uint(role+1))
	require.Error(s.T(), err)

}
