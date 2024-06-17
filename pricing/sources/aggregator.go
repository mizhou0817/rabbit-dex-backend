package sources

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	MARKET_SERVICE_STOPPED = "MARKET_SERVICE_STOPPED"
)

type SourceData struct {
	Markets      map[string]bool
	LatestPrices *sync.Map
	Urls         []string
}

type ChanWithSources struct {
	Channel  chan []float64
	Sources  []*SourceData
	MaxDelay time.Duration
}

type Aggregator struct {
	SourceMap map[string]ChanWithSources
	Interval  time.Duration
}

func NewAggregator(ctx context.Context, sourceMap map[string]ChanWithSources, interval time.Duration) {
	aggregator := Aggregator{
		SourceMap: sourceMap,
		Interval:  interval,
	}
	aggregator.start(ctx)
}

func (a *Aggregator) start(ctx context.Context) {
	go a.streamPrices(ctx)
}

func (a *Aggregator) streamPrices(ctx context.Context) {
	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.sendPriceBatches(ctx)
		}
	}
}

func (a *Aggregator) sendPriceBatches(ctx context.Context) {
	for market, cws := range a.SourceMap {
		prices := make([]float64, 0, len(cws.Sources))
		for _, srcData := range cws.Sources {
			pt := load(market, srcData.LatestPrices)
			if (pt != PriceTime{}) && time.Since(pt.time) <= cws.MaxDelay {
				prices = append(prices, pt.price)
			}
		}
		if len(prices) > 0 {
			select {
			case cws.Channel <- prices:
			default:
				logrus.Errorf("%s market_id = %s", MARKET_SERVICE_STOPPED, market)
				panic(MARKET_SERVICE_STOPPED)
			}
		}
	}
}
