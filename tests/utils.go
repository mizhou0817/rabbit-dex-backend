package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/FZambia/tarantool"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"

	"github.com/strips-finance/rabbit-dex-backend/liqengine"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

const (
	decimalThreshold = 0.00001
)

var GLOBAL_LIQ_TRADE []model.Action = nil

var INSURANCE_ID uint = 0

var MARKET_ID_MAP = map[string]string{
	"btc": "BTC-USD",
	"eth": "ETH-USD",
	"sol": "SOL-USD",
}

type TraceStep struct {
	Action  string
	StepNum uint
	Order   *model.OrderData
	Book    *model.OrderbookData
}
type BookTrace struct {
	Steps []*TraceStep
}

func updateTotals(t *testing.T, apiTest *model.ApiModel, market_ids []string, traders []traderAccountItem) error {
	trader_ids := make([]uint, 0)
	for _, trader := range traders {
		profile, err := apiTest.GetProfileById(context.Background(), trader.TraderId)
		if err != nil || profile == nil {
			//logrus.Warnf("UpdateTotals skip profileId=%d", trader.TraderId)
			continue
		}

		trader_ids = append(trader_ids, trader.TraderId)

		_, err = apiTest.InvalidateCache(context.Background(), trader.TraderId)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	//UPDATE profile meta for each market
	for _, market_id := range market_ids {
		err := apiTest.TestUpdateProfiles(context.Background(), market_id, trader_ids)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	//UPDATE aggregated profile_meta on profile instance
	err := apiTest.TestPollProfileMeta(context.Background())
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = apiTest.TestPollExchangeData(context.Background())
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func checkResults(t *testing.T, ApiModel *model.ApiModel, market_ids []string, sequence *testJson, skip bool) {
	err := CheckFills(t, ApiModel, market_ids, sequence.Expected.Fills, skip)
	assert.NoError(t, err)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Info("SUCCESS: CheckFills")
	}

	err = CheckBooks(t, ApiModel, sequence.Expected.Orderbook, skip)
	assert.NoError(t, err)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Info("SUCCESS: CheckBooks")
	}

	err = CheckTraders(t, ApiModel, market_ids, sequence.Expected.TraderAccount, skip)
	assert.NoError(t, err)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Info("SUCCESS: CheckTraders")
	}

	err = CheckExchange(t, ApiModel, market_ids, sequence.Expected.TraderAccount, sequence.Expected.Exchange, skip)
	assert.NoError(t, err)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Info("SUCCESS: CheckExchange")
	}

	err = CheckINV3(t, ApiModel, market_ids, sequence.Expected.TraderAccount, sequence.Expected.INV3, skip)
	assert.NoError(t, err)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Info("SUCCESS: CheckINV3")
	}

	err = CheckOrderQueue(t, ApiModel, market_ids, sequence.Expected.OrderQueue)
	assert.NoError(t, err)
}

func doInsuranceUnwind(t *testing.T, broker *model.Broker, market_ids []string, sequence *testJson) uint {
	as, err := liqengine.NewTntAssistant(broker, "0", 0)
	if err != nil {
		logrus.Fatal(err)
	}

	insurance_service := liqengine.NewInsuranceService(1, as)
	assert.NotEmpty(t, insurance_service)

	total_actions := insurance_service.ProcessPositions(context.Background())

	return uint(total_actions)
}

func newLiqService(t *testing.T, broker *model.Broker) *liqengine.LiquidationService {
	as, err := liqengine.NewTntAssistant(broker, "0", 0)
	if err != nil {
		logrus.Fatal(err)
	}

	liq_service := liqengine.NewLiquidationService(1, as)
	assert.NotEmpty(t, liq_service)

	return liq_service
}

func doLiquidations(t *testing.T, broker *model.Broker, market_ids []string, sequence *testJson) uint {
	liq_service := newLiqService(t, broker)

	total, err, actions := liq_service.ProcessLiquidations(context.Background())
	assert.NoError(t, err)

	if len(GLOBAL_LIQ_TRADE) == 0 {
		GLOBAL_LIQ_TRADE = make([]model.Action, 0)
	}

	if len(actions) > 0 {
		GLOBAL_LIQ_TRADE = append(GLOBAL_LIQ_TRADE, actions...)
	}

	return total

}

func doScenario(t *testing.T, broker *model.Broker, scenario string, skip bool) (*model.ApiModel, []string, *testJson) {
	GLOBAL_LIQ_TRADE = nil
	ctx := context.Background()
	/*
		Create model
	*/
	ApiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, ApiModel)

	require.NoError(t, MakeIntegrityValid(ctx, broker.Pool["profile"]))

	sequence := LoadJson(TESTS[scenario]["json"])
	assert.NotNil(t, sequence)
	assert.NotEqual(t, 0, len(sequence.Sequence))

	market_ids, err := SetupMarkets(ApiModel, sequence.Markets)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(market_ids))

	traders := sequence.Expected.TraderAccount
	/*
		Create TRACE
	*/
	trace := new(BookTrace)
	trace.Steps = make([]*TraceStep, 0)

	for i, action := range sequence.Sequence {
		logrus.Warnf("ACTION FOUND=%s orderID=%d", action.Action, action.OrderId)
		err := updateTotals(t, ApiModel, market_ids, traders)
		if err != nil {
			logrus.Fatal(err)
		}
		err = DoAction(t, trace, ApiModel, i, action, sequence.Markets, sequence, market_ids, traders)
		if err != nil {
			logrus.Errorf("************ ERROR: action=%s marketId=%s traderId=%d orderId=%d", action.Action, MARKET_ID_MAP[action.MarketId], action.TraderId, action.OrderId)
			logrus.Error(err)
		}
		logrus.Infof("STEP=%d", i)

		{
			// currently SlipStopper service should execute conditional orders
			// but for tests determined behaviour of execution is required
			// so do execution here
			market_ids := make(map[string]struct{})
			for _, m := range sequence.Markets {
				id, ok := MARKET_ID_MAP[m.MarketId]
				if ok {
					market_ids[id] = struct{}{}
				}
			}
			for id := range market_ids {
				orders, err := ApiModel.TestGetAllOrders(ctx, []string{id})
				require.NoError(t, err)
				for _, o := range orders {
					if o.Status == model.PLACED {
						if _, err := ApiModel.OrderExecute(ctx, o.ProfileID, o.MarketID, o.OrderId); err != nil {
							logrus.Error(err)
						}
					}
				}
			}
		}
	}

	err = updateTotals(t, ApiModel, market_ids, traders)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(GLOBAL_LIQ_TRADE) > 0 {
		file, _ := json.MarshalIndent(GLOBAL_LIQ_TRADE, "", " ")
		_ = ioutil.WriteFile("liqudiations.json", file, 0644)
	}

	return ApiModel, market_ids, sequence
}

func SetupMarkets(apiTest *model.ApiModel, markets []marketItem) ([]string, error) {
	marketIds := make([]string, 0, len(markets))
	alreadyMet := make(map[string]struct{})

	for _, market := range markets {
		marketId, ok := MARKET_ID_MAP[market.MarketId]
		if !ok {
			logrus.Warnf("SKIP SETUP, NO MAPPING for market_id=%s", market.MarketId)
			continue
		}
		if _, ok := alreadyMet[marketId]; !ok {
			marketIds = append(marketIds, marketId)
		}
		alreadyMet[marketId] = struct{}{}

		if market.FairPrice <= 0 {
			logrus.Warnf("SKIP setup market_id=%s broken fair_price=%f", marketId, market.FairPrice)
			continue
		}

		new_fair_price, err := apiTest.TestUpdateFairPrice(context.Background(), marketId, market.FairPrice)
		if err != nil {
			logrus.Fatalf("SetupMarkets market_id=%s  err=%s", marketId, err)
		}

		if new_fair_price.InexactFloat64() != market.FairPrice {
			logrus.Fatalf("SetupMarkets market_id=%s  TestUpdateFairPrice new=%s required=%f", marketId, new_fair_price.String(), market.FairPrice)
		}
	}

	return marketIds, nil
}

func DoAction(t *testing.T,
	trace *BookTrace,
	apiTest *model.ApiModel,
	num int,
	action sequenceItem,
	markets []marketItem,
	sequence *testJson,
	market_ids []string,
	traders []traderAccountItem) error {

	var err error
	var (
		isOrder   bool
		ordersIds []string
	)

	switch action.Action {
	case "deposit":
		err = Deposit(apiTest, action.TraderId, action.Amount)
	case "withdraw":
		err = Withdraw(apiTest, action.TraderId, action.Amount)
	case "add", "match":
		_, err = MatchOrder(apiTest, MARKET_ID_MAP[action.MarketId], action, false)
		if err != nil {
			logrus.Errorf("MatchOrder error = %s", err.Error())
			return err
		}
		isOrder = true
	case "reduce":
		_, err = MatchOrder(apiTest, MARKET_ID_MAP[action.MarketId], action, true)
		if err != nil {
			logrus.Errorf("MatchOrder reduce=true error = %s", err.Error())
			return err
		}
		isOrder = true
	case "remove":
		_, err = CancelOrder(apiTest, MARKET_ID_MAP[action.MarketId], action)
		if err != nil {
			return err
		}
		isOrder = true
	case "amend":
		_, err = AmendOrder(apiTest, MARKET_ID_MAP[action.MarketId], action)
		if err != nil {
			return err
		}
		isOrder = true
	case "setFairPrice":
		logrus.Warnf("....... SETFAIRPRICE for market_id=%s found = %f", action.MarketId, action.Amount)
		err = SetFairPrice(apiTest, MARKET_ID_MAP[action.MarketId], action.Amount)
		if err != nil {
			logrus.Errorf("... SETFAIR PRice=%f for market_id=%s err=%s", action.Amount, action.MarketId, err.Error())
		}
	case "amend leverage":
		logrus.Warnf("action=%s", action.Action)
		err = SetLeverage(apiTest, MARKET_ID_MAP[action.MarketId], action.TraderId, uint(action.Leverage))
		return err
	case "amendPosLeverage":
		logrus.Warnf("action=%s", action.Action)
		err = SetLeverage(apiTest, MARKET_ID_MAP[action.MarketId], action.TraderId, uint(action.Amount))
		return err
	case "liquidate":
		broker, err := model.GetBroker()
		if err != nil {
			logrus.Fatal(err)
		}
		liqService := newLiqService(t, broker)
		ctx := context.Background()

		if action.UserId > 0 {
			profile, err := apiTest.GetProfileData(ctx, action.UserId)
			if err != nil {
				return err
			}
			err = liqService.CancelAllOrders(ctx, &profile.ProfileCache)
			if err != nil {
				return err
			}
			err = liqService.QueueLiquidateActions(ctx, &profile.ProfileCache)
			if err != nil {
				return err
			}
			for _, marketId := range MARKET_ID_MAP {
				orders, err := apiTest.GetAllOrders(ctx, profile.ProfileID, marketId, 999999)
				if err != nil {
					return err
				}
				for _, order := range orders {
					ordersIds = append(ordersIds, order.OrderId)
				}
			}
			break
		}

		for {
			total1 := doLiquidations(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			logrus.Warnf("total1 = %d", total1)

			total2 := doInsuranceUnwind(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			logrus.Warnf("total2 = %d", total2)

			if total1+total2 == 0 {
				break
			}
		}
	case "insurance_unwind":
		broker, err := model.GetBroker()
		if err != nil {
			logrus.Fatal(err)
		}
		for {
			total1 := doLiquidations(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			total2 := doInsuranceUnwind(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			if total1+total2 == 0 {
				break
			}
		}

	case "clawback":
		broker, err := model.GetBroker()
		if err != nil {
			logrus.Fatal(err)
		}
		for {
			total1 := doLiquidations(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			total2 := doInsuranceUnwind(t, broker, nil, nil)
			updateTotals(t, apiTest, market_ids, traders)

			if total1+total2 == 0 {
				break
			}
		}
	default:
		encoded, _ := json.Marshal(action)
		logrus.Fatalf("Unsuported action=%s num=%d | %s", action.Action, num, string(encoded))
	}

	if err != nil {
		return err
	}

	if isOrder || len(ordersIds) > 0 {
		if len(ordersIds) == 0 {
			customId := fmt.Sprintf("%s@%d", MARKET_ID_MAP[action.MarketId], action.OrderId)
			ordersIds = append(ordersIds, customId)
		}

		for _, orderId := range ordersIds {
			time.Sleep(10 * time.Millisecond)

			order, err := apiTest.GetOrderById(context.Background(), MARKET_ID_MAP[action.MarketId], orderId)
			if err != nil {
				logrus.Errorf("*********** apiTest.GetOrderById=%s", err.Error())
				return err
			}
			logrus.Infof("Order executed market_id=%s id=%s status=%s price=%s size=%s", order.MarketID, order.OrderId, order.Status, order.Price.String(), order.Size.String())
		}
	}

	{
		ctx := context.Background()
		if err := DoTierPeriodics(t, ctx, apiTest); err != nil {
			return err
		}
	}

	return nil
}

func Withdraw(apiTest *model.ApiModel, traderId uint, amount float64) error {
	wallet := fmt.Sprintf("0x-wallet-%d", traderId)
	profile, err := apiTest.GetProfileByWalletForExchangeId(context.Background(), wallet, model.EXCHANGE_DEFAULT)
	if err != nil {
		logrus.Fatalf("Withdraw: GetProfileByWallet error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
	}

	_, err = apiTest.WithdrawCredit(context.Background(), traderId, amount)
	if err != nil {
		logrus.Errorf("**** ERROR: Withdraw: WithdrawCredit error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
	}

	logrus.Infof("SUCCESS Withdraw, profileId=%d amount=%f", profile.ProfileId, amount)
	return nil
}

func Deposit(apiTest *model.ApiModel, traderId uint, amount float64) error {
	if traderId == 0 {
		_, err := apiTest.CreateInsuranceProfile(context.Background(), liqengine.INSURANCE_WALLET)
		if err != nil {
			logrus.Fatalf("Deposit: CreateTraderProfile error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
		}

		_, err = apiTest.DepositCredit(context.Background(), traderId, amount)
		if err != nil {
			logrus.Fatalf("Deposit: DepositCredit error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
		}

		return nil
	}

	wallet := fmt.Sprintf("0x-wallet-%d", traderId)

	profile, err := apiTest.GetProfileByWalletForExchangeId(context.Background(), wallet, model.EXCHANGE_DEFAULT)
	if err != nil {
		if err.Error() == model.PROFILE_NOT_FOUND_ERROR {
			profile, err = apiTest.CreateProfile(context.Background(), model.PROFILE_TYPE_TRADER, wallet, model.EXCHANGE_DEFAULT)
			if err != nil {
				logrus.Fatalf("Deposit: CreateTraderProfile error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
			}
		} else {
			logrus.Fatalf("Deposit: GetProfileByWallet error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
		}
	}

	_, err = apiTest.DepositCredit(context.Background(), traderId, amount)
	if err != nil {
		logrus.Fatalf("Deposit: DepositCredit error traderId=%d amount=%f error=%s", traderId, amount, err.Error())
	}

	_ = apiTest.WhiteListProfile(context.Background(), traderId)

	logrus.Infof("Deposit: SUCCESS profileId=%d amount=%f", profile.ProfileId, amount)
	return nil
}

func SetLeverage(apiTest *model.ApiModel, market_id string, trader_id, leverage uint) error {
	if leverage < 1 {
		return nil
	}
	ctx := context.TODO()

	profile_cache, err := apiTest.GetProfileCache(ctx, trader_id)
	if err != nil {
		return err
	}
	market_leverage, ok := profile_cache.Leverage[market_id]
	if !ok {
		return fmt.Errorf("leverage not found for market: %s, trader: %d", market_id, trader_id)
	}
	if leverage == uint(market_leverage.BigInt().Uint64()) {
		return nil
	}

	new_leverage, err := apiTest.UpdateLeverage(ctx, market_id, trader_id, leverage)
	if err != nil {
		return err
	}

	s := fmt.Sprintf("%d", leverage)
	if new_leverage.String() != s {
		return fmt.Errorf("new_leverage=%s required=%d", new_leverage.String(), leverage)
	}

	return nil
}

func MatchOrder(apiTest *model.ApiModel, market_id string, action sequenceItem, reduce bool) (*model.OrderCreateRes, error) {

	if !reduce {
		err := SetLeverage(apiTest, market_id, action.TraderId, uint(action.Leverage))
		if err != nil {
			logrus.Errorf("***** ERROR: MatchOrder SetLeverage err=%s", err.Error())
			return nil, err
		}
	}

	orderSide := "long"
	if action.Side == 1 {
		orderSide = "short"
	}

	custom_id := fmt.Sprintf("%s@%d", market_id, action.OrderId)

	res, err := apiTest.OrderCreateCustomId(
		context.Background(),
		action.TraderId,
		market_id,
		action.OrderType,
		orderSide,
		&action.Price,
		&action.Size,
		&action.TriggerPrice,
		&action.SizePercent,
		&action.TimeInForce,
		custom_id)

	if err != nil {
		logrus.Errorf("MatchOrder OrderCreate error: %s", err.Error())
	}

	logrus.Info("Order Create res:")
	logrus.Info(res)
	return &res, nil
}

func CancelOrder(apiTest *model.ApiModel, market_id string, action sequenceItem) (*model.OrderCancelRes, error) {
	err := SetLeverage(apiTest, market_id, action.TraderId, uint(action.Leverage))
	if err != nil {
		logrus.Fatalf("CancelOrder SetLeverage err=%s", err.Error())
	}

	custom_id := fmt.Sprintf("%s@%d", market_id, action.OrderId)
	res, err := apiTest.OrderCancel(context.Background(), action.TraderId, market_id, custom_id, "")
	if err != nil {
		logrus.Fatalf("CancelOrder error: %s", err.Error())
	}

	return &res, nil
}

func AmendOrder(apiTest *model.ApiModel, market_id string, action sequenceItem) (*model.OrderAmendRes, error) {
	err := SetLeverage(apiTest, market_id, action.TraderId, uint(action.Leverage))
	if err != nil {
		logrus.Errorf("****** ERROR: AmendOrder SetLeverage err=%s", err.Error())
		return nil, err
	}

	custom_id := fmt.Sprintf("%s@%d", market_id, action.OrderId)
	res, err := apiTest.OrderAmend(context.Background(), action.TraderId, market_id, custom_id, &action.Price, &action.Size,
		&action.TriggerPrice, &action.SizePercent)
	if err != nil {
		logrus.Errorf("********** AmendOrder error: %s", err.Error())
	}

	return &res, nil
}

func SetFairPrice(apiTest *model.ApiModel, market_id string, fair_price float64) error {
	new_fair_price, err := apiTest.TestUpdateFairPrice(context.Background(), market_id, fair_price)
	if err != nil {
		logrus.Fatalf("SetFairPrice: %s", err.Error())
	}
	if new_fair_price.InexactFloat64() != fair_price {
		logrus.Fatalf("SetFairPrice non-equal new_fair_price=%s require=%f", new_fair_price.String(), fair_price)
	}

	return err
}

func CheckFills(t *testing.T, apiTest *model.ApiModel, market_ids []string, fills []fillItem, skip bool) error {
	var err error
	var allFills []*model.FillData
	var allTrades []*model.TradeData

	fillsCountPerMarket := map[string]int{}
	for _, fill := range fills {
		marketID := MARKET_ID_MAP[fill.MarketId]
		val, ok := fillsCountPerMarket[marketID]
		if !ok {
			fillsCountPerMarket[marketID] = 1
		} else {
			fillsCountPerMarket[marketID] = val + 1
		}
	}

	for _, marketID := range market_ids {
		marketFills, marketTrades, err := apiTest.GetAllFills(context.Background(), []string{marketID})
		if err != nil {
			return err
		}

		if len(marketTrades) != fillsCountPerMarket[marketID] {
			text := fmt.Sprintf("***** CHECKFILLS ERROR: marketId=%s len(allTrades)=%d len(fills_expected)=%d", marketID, len(marketTrades), fillsCountPerMarket[marketID])
			if !skip {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		if len(marketTrades)*2 != len(marketFills) {
			text := fmt.Sprintf("****** INTERNAL CHECKFILLS ERROR: marketId=%s should be 2x len(allTrades)=%d len(allFills)=%d", marketID, len(marketTrades), len(marketFills))
			if skip {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		allFills = append(allFills, marketFills...)
		allTrades = append(marketTrades, marketTrades...)
	}

	totalTrades := 0
	totalFills := 0

	perMarket := make(map[string]int)
	perMarketWf := make(map[string]int)

	for _, trade := range allTrades {
		totalTrades += 1

		_, ok := perMarket[trade.MarketId]
		if !ok {
			perMarket[trade.MarketId] = 1
		} else {
			perMarket[trade.MarketId] += 1
		}
	}

	for _, fill := range allFills {
		if fill.TradeId != "wf3" && fill.TradeId != "clawback" {
			continue
		}
		totalFills += 1
		_, ok := perMarketWf[fill.MarketId]
		if !ok {
			perMarketWf[fill.MarketId] = 1
		} else {
			perMarketWf[fill.MarketId] += 1
		}
	}

	for k, pm := range perMarket {
		logrus.Warnf("CHECKFILLS INFO NORMAL for market=%s:  modified_trades=%d", k, pm)
	}

	for k, pm := range perMarketWf {
		logrus.Warnf("CHECKFILLS INFO Wf3CLAWBACK for market=%s:  modified_trades=%d", k, pm/2)
	}

	logrus.Warnf("CHECKFILLS INFO:  modified_trades=%d  fills_expected=%d", totalTrades+totalFills/2, len(fills))

	logrus.Warnf("CHECKFILLS INFO:  len(allTrades)=%d  len(fills_expected)=%d", len(allTrades), len(fills))
	logrus.Warnf("CHECKFILLS INFO:  len(allTrades)=%d  len(allFills)=%d", len(allTrades), len(allFills))

	file, _ := json.MarshalIndent(allFills, "", " ")
	_ = ioutil.WriteFile("fills_trace.json", file, 0644)

	file, _ = json.MarshalIndent(allTrades, "", " ")
	_ = ioutil.WriteFile("trades_trace.json", file, 0644)

	if len(allFills) == 0 {
		return errors.New("EMPTY fills")
	}

	type BidAskKey struct {
		BidOrderId string
		AskOrderId string
	}

	//It will be duplicated - but we don't care it's a test
	mappedFills := make(map[BidAskKey][]*model.FillData, len(allFills))

	for _, fill1 := range allFills {
		var key BidAskKey
		var bid *model.FillData
		var ask *model.FillData

		if fill1.Side == model.LONG {
			key.BidOrderId = fill1.OrderId
			bid = fill1
		} else {
			key.AskOrderId = fill1.OrderId
			ask = fill1
		}

		//find separate fill
		var found bool = false
		tradeId := fill1.TradeId
		for _, fill2 := range allFills {
			if fill2.TradeId == tradeId {
				if fill2.Side == model.LONG {
					key.BidOrderId = fill2.OrderId
					bid = fill2
				} else {
					key.AskOrderId = fill2.OrderId
					ask = fill2
				}

				found = true
				break
			}
		}

		if !found {
			text := fmt.Sprintf("**** CheckFills ERROR: separate fill not found for fillId=%s traderId=%d orderId=%s marketId=%s",
				fill1.Id,
				fill1.ProfileId,
				fill1.OrderId,
				fill1.MarketId)
			return errors.New(text)
		}

		if _, ok := mappedFills[key]; ok {
			//logrus.Warnf("CheckFills skip bidOrderID=%s  askOrderId=%s", key.BidOrderId, key.AskOrderId)
			continue
		}

		mappedFills[key] = make([]*model.FillData, 0, 2)
		mappedFills[key] = append(mappedFills[key], bid)
		mappedFills[key] = append(mappedFills[key], ask)
	}

	var hasError error = nil

	for _, fill := range fills {
		//logrus.Infof("checking fill = %d", i)
		bid_order_id := fmt.Sprintf("%s@%d", MARKET_ID_MAP[fill.MarketId], fill.BidOrderId)
		ask_order_id := fmt.Sprintf("%s@%d", MARKET_ID_MAP[fill.MarketId], fill.AskOrderId)

		actualFill, ok := mappedFills[BidAskKey{BidOrderId: bid_order_id, AskOrderId: ask_order_id}]
		if ok {
			const threshold = float64(0.0001)
			assert.Less(t, math.Abs(fill.Price-actualFill[0].Price.InexactFloat64()), threshold, fmt.Sprintf("Price: %s %s | expected=%f, actual=%f", bid_order_id, ask_order_id, fill.Price, actualFill[0].Price.InexactFloat64()))
			assert.Less(t, math.Abs(fill.Size-actualFill[0].Size.InexactFloat64()), threshold, fmt.Sprintf("Size: %s %s | expected=%f, actual=%f", bid_order_id, ask_order_id, fill.Size, actualFill[0].Size.InexactFloat64()))
			/*
				logrus.Infof("BidOrderId=%d TraderId=%d Price=%f Size=%f Fee=%f", item[0].OrderID, item[0].ProfileID, item[0].Price, item[0].Size, item[0].Fee)
				logrus.Infof("AskOrderId=%d TraderId=%d Price=%f Size=%f Fee=%f", item[1].OrderID, item[1].ProfileID, item[1].Price, item[1].Size, item[1].Fee)
			*/
		} else {
			text := fmt.Sprintf("**** CheckFills ERROR: fill not found for BidOrderId=%s AskOrderId=%s", bid_order_id, ask_order_id)

			//logrus.Warnf("searching for fills for BidOrderId=%s and for AskOrderId=%s", bid_order_id, ask_order_id)
			for _, f := range allFills {
				if f.OrderId == bid_order_id {
					// Find the same trad:
					for _, f1 := range allFills {
						if f1.TradeId == f.TradeId && f1.OrderId != f.OrderId {
							//logrus.Infof("TradeId=%s  order1=%s order2=%s", f1.TradeId, f.OrderId, f1.OrderId)
						}
					}
				}
			}

			for _, f := range allFills {
				if f.OrderId == ask_order_id {
					// Find the same trad:
					for _, f1 := range allFills {
						if f1.TradeId == f.TradeId && f1.OrderId != f.OrderId {
							//logrus.Infof("TradeId=%s  order1=%s order2=%s", f1.TradeId, f.OrderId, f1.OrderId)
						}
					}
				}
			}

			if hasError == nil {
				hasError = err
			}
			if !skip {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

	}

	return hasError
}

func CheckBooks(t *testing.T, apiTest *model.ApiModel, bookStates []bookStateItem, skip bool) error {

	for _, state := range bookStates {
		market_id, ok := MARKET_ID_MAP[state.MarketId]
		logrus.Infof("CHECKING books for market_id=%s", market_id)
		if !ok {
			logrus.Errorf("**** NO MAPPING FOR market=%s", state.MarketId)
			continue
		}

		book, err := apiTest.GetOrderbookData(context.Background(), market_id)
		if err != nil {
			logrus.Error(err)
			return err
		}

		var totalBids float64 = 0.0
		var totalAsks float64 = 0.0
		for _, bid := range book.Bids {
			totalBids += bid[1].InexactFloat64() // Size
		}
		logrus.Warnf("TOTALBIDS: marketId=%s totalBids(size)=%f", market_id, totalBids)

		for _, ask := range book.Asks {
			totalAsks += ask[1].InexactFloat64() // Size
		}
		logrus.Warnf("TOTALASKS: marketId=%s totalAsks(size)=%f", market_id, totalAsks)

		var bidMin, bidMax, askMin, askMax float64

		if len(book.Bids) > 0 {
			bidMin = book.Bids[len(book.Bids)-1][0].InexactFloat64() // Price
			bidMax = book.Bids[0][0].InexactFloat64()                // Price
		}

		if len(book.Asks) > 0 {
			askMin = book.Asks[0][0].InexactFloat64()                // Price
			askMax = book.Asks[len(book.Asks)-1][0].InexactFloat64() // Price
		}

		diff := math.Abs(totalBids - state.TotalBids)
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR MARKET=%s  totalBids=%f  expected=%f", state.MarketId, totalBids, state.TotalBids)
			logrus.Error(text)
			return errors.New(text)
		}

		diff = math.Abs(totalAsks - state.TotalOffers)
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR MARKET=%s  totalAsks=%f  expected=%f", state.MarketId, totalAsks, state.TotalOffers)
			logrus.Error(text)
			return errors.New(text)
		}

		_bidMax := bidMax
		if bidMin > bidMax {
			bidMax = bidMin
			bidMin = _bidMax
		}

		if bidMin != state.BidMin || bidMax != state.BidMax {
			text := fmt.Sprintf("**** ERROR MARKET=%s  bidMin=%f  bidMin expected=%f  bidMax=%f  bidMax expected=%f", state.MarketId, bidMin, state.BidMin, bidMax, state.BidMax)
			logrus.Error(text)
			return errors.New(text)
		}

		if askMin != state.OfferMin || askMax != state.OfferMax {
			text := fmt.Sprintf("****  ERROR MARKET=%s  askMin=%f  askMin expected=%f  askMax=%f  askMax expected=%f", state.MarketId, askMin, state.OfferMin, askMax, state.OfferMax)
			logrus.Error(text)
			return errors.New(text)
		}
	}

	return nil
}

func CheckTraders(t *testing.T, apiTest *model.ApiModel, market_ids []string, traders []traderAccountItem, skip bool) error {
	err := updateTotals(t, apiTest, market_ids, traders)
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, trader := range traders {
		profileData, err := apiTest.GetProfileData(context.Background(), trader.TraderId)
		if err != nil {
			//text := fmt.Sprintf("**** ERROR for traderId=%d err=%s", trader.TraderId, err.Error())
			//logrus.Warn(text)
			continue
		}

		diff := math.Abs(trader.WalletBalance - profileData.Balance.InexactFloat64())
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR traderId=%d profile.Balance=%s expected=%f diff=%f", trader.TraderId, profileData.Balance.String(), trader.WalletBalance, diff)
			if skip == false {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		diff = math.Abs(trader.CumVolume - profileData.CumTradingVolume.InexactFloat64())
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR traderId=%d profile.CumTradingVolume=%s expected=%f", trader.TraderId, profileData.CumTradingVolume.String(), trader.CumVolume)
			if skip == false {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		logrus.Warnf("traderId=%d balance=%f AE=%f Withdrawble=%f TotalPositionMargin=%f TotalOrderMargin=%f TotalNotional=%f AccountMargin=%f CumUnrealizedPnl=%f Health=%f AL=%f",
			profileData.ProfileID,
			profileData.Balance.InexactFloat64(),
			profileData.AccountEquity.InexactFloat64(),
			profileData.WithdrawableBalance.InexactFloat64(),
			profileData.TotalPositionMargin.InexactFloat64(),
			profileData.TotalOrderMargin.InexactFloat64(),
			profileData.TotalNotional.InexactFloat64(),
			profileData.AccountMargin.InexactFloat64(),
			profileData.CumUnrealizedPnl.InexactFloat64(),
			profileData.Health.InexactFloat64(),
			profileData.AccountLeverage.InexactFloat64(),
		)

		diff = math.Abs(trader.AccountEquity - profileData.AccountEquity.InexactFloat64())
		if diff >= 0.00001 {
			logrus.Warnf("traderId=%d profileData.AccountEquity=%s expected=%f", trader.TraderId, profileData.AccountEquity.String(), trader.AccountEquity)
			text := fmt.Sprintf("**** ERROR traderId=%d profilert.AccountEquity=%s expected=%f diff=%f", trader.TraderId, profileData.AccountEquity.String(), trader.AccountEquity, diff)
			if skip == false {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		diff = math.Abs(trader.UnrealizedPnl - profileData.CumUnrealizedPnl.InexactFloat64())
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR traderId=%d profilert.CumUnrealizedPnl=%s expected=%f diff=%f", trader.TraderId, profileData.CumUnrealizedPnl.String(), trader.UnrealizedPnl, diff)
			if skip == false {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}

		if trader.Margin >= float64(8e17) {
			logrus.Warnf("WARN traderId=%d profilert.AccountMargin=%s expected=%f", trader.TraderId, profileData.AccountMargin.String(), trader.Margin)
		} else {
			diff = math.Abs(trader.Margin - profileData.AccountMargin.InexactFloat64())
			if diff >= 0.00001 {
				text := fmt.Sprintf("**** ERROR traderId=%d profilert.AccountMargin=%s expected=%f diff=%f", trader.TraderId, profileData.AccountMargin.String(), trader.Margin, diff)
				if skip == false {
					return errors.New(text)
				} else {
					logrus.Error(text)
				}
			}
		}

		logrus.Warnf("traderId=%d TotalPositionMargin=%s TotalOrderMargin=%s", trader.TraderId, profileData.TotalPositionMargin.String(), profileData.TotalOrderMargin.String())

		diff = math.Abs(trader.Withdrawable - profileData.WithdrawableBalance.InexactFloat64())
		if diff >= 0.00001 {
			text := fmt.Sprintf("**** ERROR traderId=%d profilert.WithdrawbleBalance=%s expected=%f diff=%f", trader.TraderId, profileData.WithdrawableBalance.String(), trader.Withdrawable, diff)
			if skip == false {
				return errors.New(text)
			} else {
				logrus.Error(text)
			}
		}
	}

	return nil
}

func CheckExchange(t *testing.T, apiTest *model.ApiModel, market_ids []string, traders []traderAccountItem, exchange []exchangeItem, skip bool) error {
	err := updateTotals(t, apiTest, market_ids, traders)
	if err != nil {
		logrus.Error(err)
		return err
	}

	exchange_data, err := apiTest.GetExchangeData(context.Background())
	if err != nil {
		logrus.Error(err)
		return err
	}
	assert.NotEmpty(t, exchange_data)

	logrus.Infof("Exchange balance = %s", exchange_data.TotalBalance.String())
	logrus.Infof("Fee total = %s", exchange_data.TradingFee.String())

	insurance, err := apiTest.GetProfileData(context.Background(), INSURANCE_ID)
	if err != nil {
		return err
	}
	logrus.Infof("Insurance balance = %s", insurance.Balance.String())

	var cummulativeTradersBalance float64 = 0
	for _, trader := range traders {
		if trader.TraderId == 0 {
			continue
		}

		traderData, err := apiTest.GetProfileData(context.Background(), trader.TraderId)
		if err != nil && err.Error() == model.PROFILE_NOT_FOUND_ERROR {
			continue
		}

		if *traderData.ProfileType == model.PROFILE_TYPE_INSURANCE {
			continue
		}

		if err != nil {
			text := fmt.Sprintf("ERROR get TraderId=%d ERROR: %s", trader.TraderId, err.Error())
			return errors.New(text)
		}
		cummulativeTradersBalance += traderData.Balance.InexactFloat64()
	}

	var cummulativeTradersVolume float64 = 0
	var cummulativeTradersVolume_Q float64 = 0

	_, allTrades, err := apiTest.GetAllFills(context.Background(), market_ids)
	if err != nil {
		return err
	}

	for _, trade := range allTrades {
		cummulativeTradersVolume += trade.Price.InexactFloat64() * trade.Size.InexactFloat64()
		cummulativeTradersVolume_Q += trade.Size.InexactFloat64()
	}

	// logrus.Infof("cummulativeTradersBalance = %f  expected = %f", cummulativeTradersBalance, exchange[0].ExchangeBalanceExInsurance)
	logrus.Infof("cummulativeTradersVolume = %f   expected = %f", cummulativeTradersVolume, exchange[0].CumulativeVolume)
	logrus.Infof("cummulativeTradersVolume_Q = %f   expected = %f", cummulativeTradersVolume_Q, exchange[0].CumulativeVolume_Q)
	logrus.Infof("total_fee = %s   expected = %f", exchange_data.TradingFee.String(), exchange[0].TradingFee)

	diff := math.Abs(math.Abs(exchange_data.TradingFee.InexactFloat64()) - exchange[0].TradingFee)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: fee = %f  fee expected = %f", math.Abs(exchange_data.TradingFee.InexactFloat64()), exchange[0].TradingFee)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	diff = math.Abs(cummulativeTradersVolume - exchange[0].CumulativeVolume)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: CumulativeVolume has = %f  expected = %f", cummulativeTradersVolume, exchange[0].CumulativeVolume)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	diff = math.Abs(cummulativeTradersVolume_Q - exchange[0].CumulativeVolume_Q)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: CumulativeVolume_Q has = %f  expected = %f", cummulativeTradersVolume_Q, exchange[0].CumulativeVolume_Q)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	totalPhisicalMoney := exchange_data.TotalBalance.InexactFloat64() + cummulativeTradersBalance

	diff = math.Abs(totalPhisicalMoney - exchange[0].ExchangeBalanceExInsurance)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: cummulativeTradersBalance = %f  expected = %f diff=%f", totalPhisicalMoney, exchange[0].ExchangeBalanceExInsurance, diff)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}
	logrus.Infof("totalPhisicalMoney = %f   expected = %f", totalPhisicalMoney, exchange[0].ExchangeBalanceExInsurance)

	return nil
}

func CheckINV3(t *testing.T, apiTest *model.ApiModel, market_ids []string, traders []traderAccountItem, inv3 []inv3Item, skip bool) error {
	err := updateTotals(t, apiTest, market_ids, traders)
	if err != nil {
		logrus.Error(err)
		return err
	}

	exchange_data, err := apiTest.GetExchangeData(context.Background())
	if err != nil {
		logrus.Error(err)
		return err
	}
	assert.NotEmpty(t, exchange_data)

	logrus.Infof("Exchange balance = %s", exchange_data.TotalBalance.String())
	logrus.Infof("Fee total = %s", exchange_data.TradingFee.String())

	insurance, err := apiTest.GetProfileData(context.Background(), INSURANCE_ID)
	if err != nil {
		return err
	}
	logrus.Infof("Insurance balance = %s", insurance.Balance.String())

	aeTotal := 0.0
	// insuranceAe := 0.0

	for _, trader := range traders {
		profileData, err := apiTest.GetProfileData(context.Background(), trader.TraderId)
		if err != nil {
			//text := fmt.Sprintf("***** ERROR for traderId=%d err=%s", trader.TraderId, err.Error())
			//logrus.Warn(text)
			continue
		}

		if trader.TraderId == INSURANCE_ID {
			// insuranceAe = profileData.AccountEquity.InexactFloat64()
			continue
		} else {
			if profileData.AccountMargin.InexactFloat64() > 0.02 {
				aeTotal += profileData.AccountEquity.InexactFloat64()
			}
		}
	}

	sumAe := aeTotal
	diff := math.Abs(sumAe - inv3[0].SumAE)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: SumAE has = %f  expected = %f diff=%f", sumAe, inv3[0].SumAE, diff)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	exchangeBal_insuranceDeposit := exchange_data.TotalBalance.InexactFloat64() + insurance.Balance.InexactFloat64()
	for _, trader := range traders {
		if trader.TraderId == INSURANCE_ID {
			continue
		}

		profileData, err := apiTest.GetProfileData(context.Background(), trader.TraderId)
		if err != nil {
			//text := fmt.Sprintf("**** ERROR for traderId=%d err=%s", trader.TraderId, err.Error())
			//logrus.Warn(text)
			continue
		}

		exchangeBal_insuranceDeposit += profileData.Balance.InexactFloat64()
	}

	diff = math.Abs(exchangeBal_insuranceDeposit - inv3[0].ExchangeBal_insuranceDeposit)
	if diff >= 0.00001 {
		text := fmt.Sprintf("**** ERROR: ExchangeBal_insuranceDeposit = %f  expected = %f diff=%f", exchangeBal_insuranceDeposit, inv3[0].ExchangeBal_insuranceDeposit, diff)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	logrus.Infof("exchangeBal_insuranceDeposit = %f  expected = %f", exchangeBal_insuranceDeposit, inv3[0].ExchangeBal_insuranceDeposit)
	logrus.Infof("sumAe = %f   expected = %f", sumAe, inv3[0].SumAE)

	valid, err := apiTest.IsInv3Valid(context.Background(), 0)
	logrus.Infof("Inv3Status = %t   expected = %t err = %t", valid, inv3[0].Status, err)
	if valid != inv3[0].Status {
		text := fmt.Sprintf("**** ERROR: INV3 status = %t  expected = %t", valid, inv3[0].Status)
		if skip == false {
			return errors.New(text)
		} else {
			logrus.Error(text)
		}
	}

	return nil
}

func CheckOrderQueue(t *testing.T, apiTest *model.ApiModel, market_ids []string, queue []orderQueueItem) error {
	allOrders, err := apiTest.TestGetAllOrders(context.Background(), market_ids)
	if err != nil {
		return err
	}

	ordersMap := make(map[string]*model.OrderData)
	for _, order := range allOrders {
		if !isConditionalType(order.OrderType) {
			continue
		}
		ordersMap[order.OrderId] = order
	}

	for _, item := range queue {
		marketID, ok := MARKET_ID_MAP[item.MarketID]
		if !ok {
			return fmt.Errorf("market '%s' not configured for test", item.MarketID)
		}
		orderID := fmt.Sprintf("%s@%d", marketID, item.StopID)
		_ = orderID

		order, ok := ordersMap[orderID]
		if !ok {
			return fmt.Errorf("order '%s' not found in engine", orderID)
		}
		delete(ordersMap, orderID)

		if marketID != order.MarketID {
			return fmt.Errorf("order '%s' bad market-id: expected=%s, actual=%s", orderID, item.MarketID, order.MarketID)
		}
		if item.OrderType != order.OrderType {
			return fmt.Errorf("order '%s' bad order-type: expected=%s, actual=%s", orderID, item.OrderType, order.OrderType)
		}
		if item.TraderID != order.ProfileID {
			return fmt.Errorf("order '%s' bad trader-id: expected=%d, actual=%d", orderID, item.TraderID, order.ProfileID)
		}
		switch item.Status {
		case "active": // by Cheng: active = placed
			if order.Status != "placed" {
				return fmt.Errorf("order '%s' bad status: expected=%s, actual=%s", orderID, item.Status, order.Status)
			}
		case "triggered": // by Cheng: executed (will be reflected in orderbook, and trader OM, PM, etc.)
			if order.Status != "open" && order.Status != "closed" {
				return fmt.Errorf("order '%s' bad status: expected=%s, actual=%s", orderID, item.Status, order.Status)
			}
		case "rejected": // by Cheng: rejected by tarantool (after triggering)
			if order.Status != "rejected" {
				return fmt.Errorf("order '%s' bad status: expected=%s, actual=%s", orderID, item.Status, order.Status)
			}
		case "removed": // by Cheng: canceled by user
			fallthrough
		case "canceled": // by Cheng: canceled by system
			if order.Status != "canceled" {
				return fmt.Errorf("order '%s' bad status: expected=%s, actual=%s", orderID, item.Status, order.Status)
			}
		default:
			return fmt.Errorf("order '%s' unknown status: %s", orderID, item.Status)
		}

		switch item.OrderType {
		case "stop_loss", "take_profit":
			if math.Abs(item.TriggerPrice-order.TriggerPrice.InexactFloat64()) > decimalThreshold {
				return fmt.Errorf("order '%s' bad trigger-price: expected=%f, actual=%f", orderID, item.TriggerPrice, order.TriggerPrice.InexactFloat64())
			}
			if math.Abs(item.PercentPosition-order.SizePercent.InexactFloat64()) > decimalThreshold {
				return fmt.Errorf("order '%s' bad size-percent: expected=%f, actual=%f", orderID, item.PercentPosition, order.SizePercent.InexactFloat64())
			}
		default:
		}
		if item.TriggerPrice != order.TriggerPrice.InexactFloat64() {
			return fmt.Errorf("order '%s' bad trader-price: expected=%f, actual=%f", orderID, item.TriggerPrice, order.TriggerPrice.InexactFloat64())
		}
		if item.PercentPosition != order.SizePercent.InexactFloat64() {
			return fmt.Errorf("order '%s' bad size-percent: expected=%f, actual=%f", orderID, item.PercentPosition, order.SizePercent.InexactFloat64())
		}
	}

	for orderID := range ordersMap {
		return fmt.Errorf("order '%s' not found in test", orderID)
	}

	return nil
}

func MakeIntegrityValid(ctx context.Context, profileConn *tarantool.Connection) error {
	return evalScript(ctx, profileConn, `require('app.profile.integrity').make_valid()`)
}

var (
	conditionalOrders = []string{"stop_loss", "take_profit", "stop_loss_limit", "take_profit_limit", "stop_limit", "stop_market"}
)

func isConditionalType(t string) bool {
	for _, c := range conditionalOrders {
		if t == c {
			return true
		}
	}
	return false
}

func evalScript(ctx context.Context, conn *tarantool.Connection, cmd string) error {
	params := []interface{}{}
	_, err := conn.ExecContext(ctx, tarantool.Eval(cmd, params))
	return err
}

func minValue[T constraints.Integer](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func ToPtr[T any](v T) *T {
	return &v
}

func dumpJson(v any) {
	z, _ := json.MarshalIndent(v, "", "   ")
	fmt.Println(string(z))
}

func getShardId(instance string) string {
	return strings.SplitN(instance, ".", 2)[0]
}
