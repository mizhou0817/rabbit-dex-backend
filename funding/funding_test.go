package funding

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

func TestBasic(t *testing.T) {
	market_id := "BTC-USD"

	funding_service, err := NewFundingService(0)
	assert.NoError(t, err)
	assert.NotEmpty(t, funding_service)

	broker, err := model.GetBroker()
	assert.NoError(t, err)
	assert.NotEmpty(t, broker)
	apiModel := model.NewApiModel(broker)
	assert.NotEmpty(t, apiModel)

	err = apiModel.UpdateIndexPrice(context.TODO(), market_id, 21200.0)
	assert.NoError(t, err)

	marketData, err := apiModel.GetMarketData(context.Background(), market_id)
	assert.NoError(t, err)
	assert.NotEmpty(t, marketData)

	fundingMeta, err := apiModel.GetFundingMeta(context.Background(), market_id)
	assert.NoError(t, err)
	assert.NotEmpty(t, fundingMeta)

	marketPositions, err := apiModel.GetAllActivePositions(context.Background(), market_id, 0, MAX_POSITIONS)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(marketPositions))

	fundingPayments := make([]model.FundingPayment, 0, EXPECTED_POSITIONS)
	fundingUpdate := float64(1.0)
	d_funding_amount := tdecimal.NewDecimal(decimal.NewFromFloat(fundingUpdate))
	fundingPayments = append(fundingPayments, model.FundingPayment{
		MarketId:      market_id,
		ProfileId:     0,
		FundingAmount: d_funding_amount})

	logrus.Info(marketData.LastFundingUpdateTime)
	logrus.Info(marketData)

	err = apiModel.PayFunding(context.Background(), market_id, fundingPayments, marketData.LastFundingUpdateTime, 11.2, 12.2)
	assert.NoError(t, err)
}

//TODO : redone tests
/*
func TestFunding(t *testing.T) {
	instance := helpers.NewInstance()

	err := instance.Reset()
	if err != nil {
		t.Fatalf("Can't start tarantool err=%s", err.Error())
	}

	broker, err := model.GetBroker(model.BrokerConfig{Host: instance.Host, User: instance.User, Password: instance.Password})
	if err != nil {
		t.Fatalf("Can't create broker err=%s", err.Error())
	}

	exchangeModel, err := model.NewExchangeModel(broker)
	assert.NoError(t, err)
	assert.NotNil(t, exchangeModel)
	err = exchangeModel.Create(context.Background())
	assert.NoError(t, err)

	rtsModel, err := model.NewRtsModel(broker)
	assert.NoError(t, err)
	assert.NotNil(t, rtsModel)

	profileModel, err := model.NewProfileModel(broker)
	assert.NoError(t, err)
	assert.NotNil(t, profileModel)

	positionModel, err := model.NewPositionModel(broker)
	assert.NoError(t, err)
	assert.NotNil(t, positionModel)

	marketModel, err := model.NewMarketModel(broker)
	assert.NoError(t, err)
	assert.NotNil(t, marketModel)

	fundingService, err := NewFundingService(1, broker)
	assert.NoError(t, err)
	assert.NotNil(t, fundingService)

	var forced_margin float64 = 0.03
	var liquidation_margin float64 = 0.05

	market_ids := []string{"btc", "eth", "sol"}
	for _, market_id := range market_ids {
		err = marketModel.Create(context.Background(), &model.Market{
			MarketID:          market_id,
			ForcedMargin:      forced_margin,
			LiquidationMargin: liquidation_margin,
		})
		assert.NoError(t, err)
	}

	var positions map[string][]*model.Position = make(map[string][]*model.Position)
	for _, market_id := range market_ids {
		positions[market_id] = make([]*model.Position, 0)
	}

	// create profiles and positions
	profiles_count := 10
	middle := 5
	assert.Equal(t, 0, profiles_count%2, "profiles_count shold be % 2 == 0")
	for i := 1; i <= profiles_count; i++ {
		randomWallet := fmt.Sprintf("wallet-%d", i)
		profile, err := profileModel.Create(context.Background(), &model.Profile{
			Status: model.PROFILE_ACTIVE_STATUS,
			Wallet: randomWallet,
		})
		assert.NoError(t, err)
		assert.NotNil(t, profile)

		//Create position for profile on all markets
		side := model.LONG
		if i > middle {
			side = model.SHORT
		}
		price := 20.0
		size := 10.0
		for _, market_id := range market_ids {
			position := model.NewPosition(profile.ProfileID, market_id, price, size, side)

			_, err = positionModel.Replace(context.Background(), position)
			assert.NoError(t, err)

			positions[market_id] = append(positions[market_id], position)
		}
	}

	var market_price float64 = 10
	var index_price float64 = 10
	var fair_price float64 = 10

	var fundings map[int]float64 = map[int]float64{0: 0.1, 1: -0.002, 2: 0.003, 3: 0.0}

	current_time := time.Now().Unix()

	funding_rounds := 4
	for i := 0; i < funding_rounds; i++ {
		current_funding := fundings[i]
		current_time += 10
		for _, market_id := range market_ids {
			err = rtsModel.TestMarketRtUpdate(context.Background(), &model.MarketRt{
				MarketID:    market_id,
				LastUpdate:  current_time,
				MarketPrice: market_price,
				IndexPrice:  index_price,
				FairPrice:   fair_price,
			})

			err = rtsModel.TestMarketMetaUpdate(context.Background(), &model.MarketMeta{
				MarketID:              market_id,
				LastAvgUpdate:         current_time,
				LastFundingUpdateTime: current_time,
				FundingRate:           current_funding,
			})
		}

		fundingService.ProcessFunding(context.Background())

		for _, market_id := range market_ids {
			//Calc expected values
			total_longs := 0.0
			total_shorts := 0.0

			found := 0
			for _, position := range positions[market_id] {
				fundingUpdate := position.Size * fair_price * limit(current_funding)
				if position.Side == model.LONG {
					fundingUpdate = -fundingUpdate
					total_longs += fundingUpdate
				} else {
					total_shorts += fundingUpdate
				}
				found++
			}
			assert.NotEmpty(t, found)
			logrus.Infof("total_longs=%f total_shorts=%f", total_longs, total_shorts)

			marketFunding, err := marketModel.GetFunding(context.Background(), market_id)
			assert.NoError(t, err)
			assert.NotNil(t, marketFunding)

			if current_funding != 0 {
				assert.NotEqual(t, 0, total_longs, "ROUND=%d", i)
				assert.NotEqual(t, 0, total_shorts, "ROUND=%d", i)

				assert.Equal(t, current_time, marketFunding.LastUpdate, "ROUND=%d", i)
				assert.Equal(t, total_longs, marketFunding.TotalLong, "ROUND=%d", i)
				assert.Equal(t, total_shorts, marketFunding.TotalShort, "ROUND=%d", i)
			} else {
				assert.NotEqual(t, current_time, marketFunding.LastUpdate, "ROUND=%d", i)
				assert.NotEqual(t, total_longs, marketFunding.TotalLong, "ROUND=%d", i)
				assert.NotEqual(t, total_shorts, marketFunding.TotalShort, "ROUND=%d", i)
			}
		}
	}
}
*/
