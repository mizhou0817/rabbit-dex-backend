package pricing

/*
Calculates new prices for tarantool from values received on a channel.
The new price is calculated from the individual prices from each source.
Checks for consistency between the prices from the sources.
*/

import (
	"context"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"
)

const (
	FIVE_PERCENT            = 0.05
	TEN_PERCENT             = 0.1
	MAX_REJECTED_RUN_LENGTH = 10
)

type MarketPriceService struct {
	marketId          string
	mostRecentPrice   float64
	priceChan         chan []float64
	rejectedRunLength uint
	allData           []float64
	useableData       []float64
	scratch           []float64
	apiModel          PriceReceiver
}

func NewMarketPriceService(marketId string, numSources int, priceChan chan []float64, apiModel PriceReceiver) *MarketPriceService {
	ps := MarketPriceService{
		marketId:    marketId,
		priceChan:   priceChan,
		allData:     make([]float64, numSources+1),
		useableData: make([]float64, 0, numSources),
		scratch:     make([]float64, numSources+1),
		apiModel:    apiModel,
	}
	return &ps
}

func (mps *MarketPriceService) start(ctx context.Context) {
	go mps.priceUpdater(ctx)
}

func (mps *MarketPriceService) priceUpdater(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case inputPrices, ok := <-mps.priceChan:
			if !ok {
				return
			}
			mps.updateModelPrice(ctx, inputPrices)
		}
	}
}

func (mps *MarketPriceService) updateModelPrice(ctx context.Context, inputPrices []float64) error {
	price, numSources, err := mps.processInputs(ctx, inputPrices)
	if err != nil {
		logrus.Warnf("%v in processPrice for market_id %s", err, mps.marketId)
		return err
	}

	err = mps.apiModel.UpdateIndexPrice(ctx, mps.marketId, price)
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertCrit).Errorf("error %v calling apiModel.UpdatePriceIndex", err)
		return err
	}

	logrus.Infof("Index price updated for market_id=%s price=%f based on %d sources", mps.marketId, price, numSources)

	return nil
}

func (mps *MarketPriceService) processInputs(ctx context.Context, inputPrices []float64) (price float64, numSources int, err error) {
	latest, numSources, err := mps.getCombinedPrice(ctx, inputPrices)
	if err != nil {
		return 0.0, 0, err
	} else if mps.mostRecentPrice == 0.0 || mps.rejectedRunLength > MAX_REJECTED_RUN_LENGTH || closeEnough(latest, mps.mostRecentPrice, TEN_PERCENT) {
		mps.mostRecentPrice = latest
		mps.rejectedRunLength = 0
		return latest, numSources, nil
	} else {
		mps.rejectedRunLength++
		return latest, numSources,
			fmt.Errorf("price jump in market %s, last accepted %v, rejected %v, run length %v",
				mps.marketId, mps.mostRecentPrice, latest, mps.rejectedRunLength)
	}
}

func (mps *MarketPriceService) getCombinedPrice(ctx context.Context, inputPrices []float64) (float64, int, error) {
	var availableData []float64
	if mps.mostRecentPrice == 0.0 {
		availableData = mps.allData[1:1]
	} else {
		mps.allData[0] = mps.mostRecentPrice
		availableData = mps.allData[:1]
	}
	availableData = append(availableData, inputPrices...)
	var median1 float64
	numSources := len(inputPrices)
	if numSources > 1 {
		var consistent bool
		median1, consistent = mps.checkDataConsistency(availableData)
		if !consistent && mps.mostRecentPrice != 0.0 {
			median1, consistent = mps.checkDataConsistency(availableData[1:])
		}
		if !consistent {
			return 0.0, len(availableData), fmt.Errorf(
				"market %s has inconsistent price data, fewer than 2 values within 5%% of median: %v",
				mps.marketId, median1)
		}
	}
	var rawData []float64
	if mps.mostRecentPrice == 0.0 {
		rawData = availableData
	} else {
		rawData = availableData[1:]
	}
	mps.useableData = mps.useableData[:0]
	for _, price := range rawData {
		if numSources == 1 || closeEnough(price, median1, FIVE_PERCENT) {
			mps.useableData = append(mps.useableData, price)
		}
	}
	if len(mps.useableData) == 0 {
		return 0.0, 0, fmt.Errorf("found no useable data")
	}
	return mps.median(mps.useableData), len(mps.useableData), nil
}

func (mps *MarketPriceService) checkDataConsistency(values []float64) (median1 float64, consistent bool) {
	median1 = mps.median(values)
	if len(values) < 2 {
		return median1, true
	}
	var count uint
	for _, value := range values {
		if closeEnough(value, median1, FIVE_PERCENT) {
			count++
			if count >= 2 {
				return median1, true
			}
		}
	}
	return median1, false
}

func (mps *MarketPriceService) median(values []float64) float64 {
	mps.scratch = mps.scratch[:len(values)]
	copy(mps.scratch, values)
	sort.Float64s(mps.scratch)
	var median float64
	l := len(values)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (mps.scratch[l/2-1] + mps.scratch[l/2]) / 2
	} else {
		median = mps.scratch[l/2]
	}
	return median
}

func closeEnough(value float64, target float64, tolerance float64) bool {
	if target == 0.0 {
		return value == 0.0
	}
	fracDiff := (value - target) / target
	if fracDiff < 0.0 {
		fracDiff = -fracDiff
	}
	return fracDiff <= tolerance
}
