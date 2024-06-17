package pricing

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type DummyPriceReceiver struct{}
type PriceRec struct {
	count           int
	batchStart      time.Time
	lastRecvd       time.Time
	lastPrice       float64
	longestInterval time.Duration
}

var trackers map[string]PriceRec = make(map[string]PriceRec)
var mu sync.Mutex
var lastOutput time.Time = time.Now()

func (dpr *DummyPriceReceiver) UpdateIndexPrice(ctx context.Context, marketId string, price float64) error {
	logrus.Debugf("DummyPriceReceiver: marketId: %s, price: %f", marketId, price)
	mu.Lock()
	priceRec := trackers[marketId]
	now := time.Now()
	if priceRec.lastRecvd.IsZero() {
		priceRec.batchStart = lastOutput
		priceRec.lastRecvd = lastOutput
		priceRec.lastPrice = price
		priceRec.longestInterval = 0
	}
	interval := now.Sub(priceRec.lastRecvd)
	if interval > priceRec.longestInterval {
		priceRec.longestInterval = interval
	}
	priceRec.count = priceRec.count + 1
	priceRec.lastRecvd = now
	trackers[marketId] = priceRec
	if now.Sub(lastOutput) > 15*time.Minute {
		for marketId, priceRec := range trackers {
			logrus.Warnf("got %d %s prices in %v, last %f, longest interval %v", priceRec.count, marketId[:7], now.Sub(priceRec.batchStart), priceRec.lastPrice, priceRec.longestInterval)
			priceRec.count = 0
			priceRec.batchStart = now
			priceRec.longestInterval = 0
			trackers[marketId] = priceRec
		}
		lastOutput = now
	}
	mu.Unlock()
	return nil
}

// compile-time check that DummyPriceReceiver implements PriceReceiver
var _ PriceReceiver = (*DummyPriceReceiver)(nil)
