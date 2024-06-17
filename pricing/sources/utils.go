package sources

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	MAX_REST_READ_ATTEMPTS = 8
	NO_PRICE_DATA          = "NO_PRICE_DATA"
)

type Ticker struct {
	ExchangeId   string `yaml:"exchange_id"`
	InstId       string `yaml:"inst_id"`
	Network      string `yaml:"network"`
	MaxUseAge    string `yaml:"max_use_age"`
	ReconnectAge string `yaml:"reconnect_age"`
}

type PriceTime struct {
	price float64
	time  time.Time
}

func reverseInstrumentMap(original map[string]Ticker) map[string]string {
	reversed := make(map[string]string, len(original))
	for k, v := range original {
		reversed[v.InstId] = k
	}
	return reversed
}

func store(market string, priceTime PriceTime, latestPrices *sync.Map, url string) bool {
	var v any
	var prevAv *atomic.Value
	var prevPt PriceTime
	var havePrevPt bool
	// does the map already contain an entry for this market?
	if v, havePrevPt = latestPrices.Load(market); havePrevPt {
		// does the entry have type *atomic.Value?
		if prevAv, havePrevPt = v.(*atomic.Value); havePrevPt {
			// does the atomic value contain a PriceTime?
			prevPt, havePrevPt = prevAv.Load().(PriceTime)
		}
	}

	if havePrevPt && prevPt.price == priceTime.price {
		logrus.Infof("received unchanged price %f for %s from %s", priceTime.price, market, url)
		return false
	}

	logrus.Infof("received price %f for %s from %s", priceTime.price, market, url)
	if havePrevPt {
		prevAv.Store(priceTime)
	} else {
		var newVal atomic.Value
		newVal.Store(priceTime)
		latestPrices.Store(market, &newVal)
	}
	return true
}

func load(market string, latestPrices *sync.Map) PriceTime {
	v, ok := latestPrices.Load(market)
	if !ok {
		return PriceTime{}
	}
	av, ok := v.(*atomic.Value)
	if !ok {
		return PriceTime{}
	}
	priceTime, ok := av.Load().(PriceTime)
	if !ok {
		return PriceTime{}
	}
	return priceTime
}

// take a string like "1542057314.748512345" and return
// a time.Time interpreting the string as a decimal fraction
// of seconds since 1970-01-01 00:00:00 UTC
func parseUnixTime(input string) (time.Time, error) {
	ftime, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing float from %s, err: %s", input, err.Error())
	}
	intPart, fracPart := math.Modf(ftime)
	secs := int64(intPart)
	nanos := int64(fracPart * 1e9)
	return time.Unix(secs, nanos), nil
}
