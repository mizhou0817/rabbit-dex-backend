package funding

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	MAX_MARKETS        = 10000
	MAX_POSITIONS      = 100000
	EXPECTED_MARKETS   = 10
	EXPECTED_POSITIONS = 1000
)

type FundingService struct {
	apiModel        *model.ApiModel
	interval        time.Duration
	fundingPayments []model.FundingPayment
	stopf           context.CancelFunc
	cfg             *Config
}

type MarketFunding struct {
	FairPrice float64
	Rate      float64
}

func NewFundingService(interval time.Duration) (*FundingService, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Can't read funding config err=%s", err.Error())
	}

	broker, err := model.GetBroker()
	if err != nil {
		return nil, fmt.Errorf("Error obtaining Tarantool broker: %s", err.Error())
	}
	apiModel := model.NewApiModel(broker)

	fs := &FundingService{
		apiModel:        apiModel,
		interval:        interval,
		fundingPayments: make([]model.FundingPayment, 0, EXPECTED_POSITIONS),
		cfg:             config,
	}
	return fs, nil
}

func (fs *FundingService) Run() (context.CancelFunc, error) {
	ctx, cancelf := context.WithCancel(context.Background())
	ticker := time.NewTicker(fs.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fs.ProcessFunding(ctx)
			}
		}
	}()
	fs.stopf = cancelf
	return cancelf, nil
}

func (fs *FundingService) Stop() {
	if fs.stopf != nil {
		fs.stopf()
	}
}

// TODO: process should return error
func (fs *FundingService) ProcessFunding(ctx context.Context) {
	for _, market_id := range fs.cfg.Service.Markets {

		marketData, err := fs.apiModel.GetMarketData(ctx, market_id)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Funding service, error marketData for market %s: %v", market_id, err)
			continue
		}

		fundingMeta, err := fs.apiModel.GetFundingMeta(ctx, market_id)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Funding service, error retrieving funding_meta for market %s: %v", market_id, err)
			continue
		}
		if marketData.LastFundingUpdateTime <= fundingMeta.LastUpdate || marketData.LastFundingRate.InexactFloat64() == 0.0 {
			text := fmt.Sprintf("SKIP funding payment for market_id=%s now=%d marketMeta.LastFundingUpdateTime=%d marketFunding.LastUpdate=%d marketMeta.FundingRate=%f",
				market_id,
				time.Now().Unix(),
				marketData.LastFundingUpdateTime,
				fundingMeta.LastUpdate,
				marketData.LastFundingRate.InexactFloat64(),
			)
			logrus.Warnf(text)
			continue
		}

		marketPositions, err := fs.apiModel.GetAllActivePositions(ctx, market_id, 0, MAX_POSITIONS)
		if err != nil {
			logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Funding service, error retrieving positions for market %s: %v", market_id, err)
			return
		}
		fs.fundingPayments = fs.fundingPayments[:0]
		var totalLong, totalShort float64
		for _, position := range marketPositions {
			fundingUpdate := position.Size.InexactFloat64() * marketData.FairPrice.InexactFloat64() * limit(marketData.LastFundingRate.InexactFloat64())
			if position.Side == model.LONG {
				fundingUpdate = -fundingUpdate
				totalLong += fundingUpdate
			} else {
				totalShort += fundingUpdate
			}

			d_funding_amount := tdecimal.NewDecimal(decimal.NewFromFloat(fundingUpdate))

			fs.fundingPayments = append(fs.fundingPayments, model.FundingPayment{
				MarketId:      position.MarketID,
				ProfileId:     position.ProfileID,
				FundingAmount: d_funding_amount})
		}

		if len(fs.fundingPayments) > 0 {
			err = fs.apiModel.PayFunding(ctx, market_id, fs.fundingPayments, marketData.LastFundingUpdateTime, totalLong, totalShort)
			if err != nil {
				logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Funding service, error paying funding for market %s: %v", market_id, err)
				return
			}
		}
	}

}

func limit(rate float64) float64 {
	if rate < -0.01 {
		return -0.01
	}
	if rate > 0.01 {
		return 0.01
	}
	return rate
}
