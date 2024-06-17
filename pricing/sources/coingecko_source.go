package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	CG_USD          = "usd"
	LAST_UPDATED_AT = "last_updated_at"
	CG_URL          = "https://pro-api.coingecko.com/api/v3/simple/price"
)

type coingeckoSourceFactory struct{}

func init() {
	RegisterRestFactory("coingecko", func() RestSourceFactory {
		return &coingeckoSourceFactory{}
	})
}

func (csf *coingeckoSourceFactory) NewRestSource(ctx context.Context, coinTickers map[string]Ticker, marketId2 string, refCoinTickers []Ticker, maxUseAge map[string]time.Duration, readInterval time.Duration, timeout time.Duration, multipliers map[string]float64, apiKey string) *SourceData {
	urls := buildCGUrls(coinTickers, refCoinTickers[0])
	source := BasegeckoSource{
		latestPrices:     &sync.Map{},
		urls:             urls,
		markets:          buildMarkets(coinTickers, marketId2, refCoinTickers),
		marketId2:        marketId2,
		lastDataReceived: make(map[string]time.Time),
		coinTickers:      coinTickers,
		refCoinTickers:     refCoinTickers,
		maxUseAge:        maxUseAge,
		readInterval:     readInterval,
		timeout:          timeout,
		multipliers:      multipliers,
		apiKey:           apiKey,
	}

	source.extractPrices = func(responseBody []byte) (map[string]PriceTime, error) {
		var responseData map[string]map[string]float64
		err := json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return nil, fmt.Errorf("error deserializing GET response JSON %s, err: %s", string(responseBody), err.Error())
		}
		prices := make(map[string]PriceTime, len(responseData))
		for coinId, coinValue := range responseData {
			price := coinValue[CG_USD]
			timestamp := int64(coinValue[LAST_UPDATED_AT])
			lastUpdated := time.Unix(timestamp, 0)
			prices[coinId] = PriceTime{price, lastUpdated}
		}
		return prices, nil
	}

	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         source.urls,
	}
	return &sourceData

}

func buildCGUrls(coinTickers map[string]Ticker, refCoinTicker Ticker) []string {
	var urlBuilder strings.Builder
	urlBuilder.WriteString(CG_URL)
	urlBuilder.WriteString("?")
	var coinIdsBuilder strings.Builder
	coinIdsBuilder.WriteString(refCoinTicker.InstId)
	for _, ticker := range coinTickers {
		coinIdsBuilder.WriteString(",")
		coinIdsBuilder.WriteString(ticker.InstId)
	}
	q := url.Values{}
	q.Add("ids", coinIdsBuilder.String())
	q.Add("vs_currencies", USD)
	q.Add("include_last_updated_at", "true")
	q.Add("precision", "full")
	urlBuilder.WriteString(q.Encode())
	return []string{urlBuilder.String()}
}
