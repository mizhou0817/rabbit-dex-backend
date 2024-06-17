package sources

/*
Handles the CoinGecko and GeckoTerminal REST APIs to receive prices.

Used by coingecko_source.go and geckoterminal_source.go.

Prices from the APIs are in USD, since they are required by RabbitX in USDT the
USDT price is also retrieved from the API and used to convert the price to USDT.
*/

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	USD = "usd"
)

type BasegeckoSource struct {
	latestPrices     *sync.Map
	startedAt        time.Time
	lastDataReceived map[string]time.Time
	urls             []string
	apiKey           string
	markets          map[string]string
	marketId2        string
	coinTickers      map[string]Ticker
	refCoinTickers   []Ticker
	maxUseAge        map[string]time.Duration
	readInterval     time.Duration
	timeout          time.Duration
	multipliers      map[string]float64
	extractPrices    func(responseBody []byte) (map[string]PriceTime, error)
}

func (bgs *BasegeckoSource) start(ctx context.Context) {
	go bgs.reader(ctx)
}

func (bgs *BasegeckoSource) listMarkets() (markets map[string]bool) {
	markets = make(map[string]bool, len(bgs.coinTickers))
	for marketId := range bgs.coinTickers {
		markets[marketId] = true
	}
	return markets
}

func (bgs *BasegeckoSource) reader(ctx context.Context) {
	bgs.startedAt = time.Now()
	ticker := time.NewTicker(bgs.readInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bgs.readPriceData(ctx)
		}
	}
}

func (bgs *BasegeckoSource) readPriceData(ctx context.Context) {
	for _, url := range bgs.urls {
		prices, err := bgs.findPrices(ctx, url)
		if err != nil {
			logrus.Warnf("error reading price from %s: %s", url, err.Error())
			return
		}
		price2, ok := prices[bgs.marketId2]
		if !ok {
			logrus.Warnf("no price returned for %s from %s", bgs.refCoinTickers[0].InstId, url)
			return
		}
		if price2.price < 0.01 {
			logrus.Warnf("price for %s is %v", bgs.refCoinTickers[0].InstId, price2.price)
			return
		}
		for market, priceTime := range prices {
			var multiplier float64
			var multFound bool
			var maxUseAge time.Duration
			if market == bgs.marketId2 {
				multiplier = 1
				multFound = true
				maxUseAge = 10 * time.Minute
			} else {
				multiplier, multFound = bgs.multipliers[market]
				maxUseAge = bgs.maxUseAge[market]
			}
			if !multFound {
				logrus.Warnf("no multiplier found for %s", market)
				continue
			}
			priceTime.price = (priceTime.price * multiplier) / price2.price
			stored := store(market, priceTime, bgs.latestPrices, url[:29])
			if stored {
				bgs.lastDataReceived[market] = priceTime.time
			}
			elapsed := bgs.getDataElapsed(market)
			if elapsed > maxUseAge*3 {
				logrus.Warnf("%s received for %s from %s since %v", NO_PRICE_DATA, market, url, bgs.lastDataReceived[market])
			}
		}
	}
}

func (bgs *BasegeckoSource) findPrices(ctx context.Context, url string) (map[string]PriceTime, error) {
	var resp *http.Response
	var numReadAttempts int = 0
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("findPrice context canceled or expired: %v", ctx.Err())
		default:
		}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating GET request %s gave error %s", url, err.Error())
		}

		req.Header.Set("x-cg-pro-api-key", bgs.apiKey)
		req.Header.Set("accept", "application/json")
		req.Header.Set("User-Agent", "curl")

		client := &http.Client{
			Timeout: bgs.timeout,
		}
		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			break
		} else {
			numReadAttempts++
			if numReadAttempts >= MAX_REST_READ_ATTEMPTS {
				err = fmt.Errorf("no price update response from %s, last error: %s", url, err.Error())
				return nil, err
			}
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GET response body, url %s, err %s", url, err.Error())
	}
	pricesByCoinId, err := bgs.extractPrices(body)
	if err != nil {
		return nil, fmt.Errorf("error extracting prices from response body %s, err %s", string(body), err.Error())
	}
	result := make(map[string]PriceTime, len(pricesByCoinId))
	for coinId, priceTime := range pricesByCoinId {
		market := bgs.markets[coinId]
		if market == "" {
			return nil, fmt.Errorf("no market found for coinId %s", coinId)
		}
		result[market] = priceTime
	}
	return result, nil
}

func (bgs *BasegeckoSource) getDataElapsed(market string) time.Duration {
	if bgs.lastDataReceived[market].IsZero() {
		return time.Since(bgs.startedAt)
	}
	return time.Since(bgs.lastDataReceived[market])
}

func buildMarkets(coinTickers1 map[string]Ticker, marketId2 string, refCoinTickers []Ticker) map[string]string {
	markets := make(map[string]string, len(coinTickers1)+1)
	for market, ticker := range coinTickers1 {
		markets[ticker.InstId] = market
	}
	for _, refCoinTicker := range refCoinTickers {
		markets[refCoinTicker.InstId] = marketId2
	}
	return markets
}
