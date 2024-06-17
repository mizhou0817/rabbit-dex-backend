package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	CMC_URL = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"
)

type CoinmarketcapSource struct {
	latestPrices *sync.Map
	startedAt    time.Time
	lastUpdated  map[string]time.Time
	apiKey       string
	coinTickers  map[string]Ticker
	markets      map[string]string
	url          string
	shortUrl     string
	idParam      string
	refCoinTickers []Ticker
	maxUseAge    map[string]time.Duration
	readInterval time.Duration
	timeout      time.Duration
	multipliers  map[string]float64
}

type Status struct {
	Timestamp    string      `json:"timestamp"`
	ErrorCode    int         `json:"error_code"`
	ErrorMessage interface{} `json:"error_message"`
	Elapsed      int         `json:"elapsed"`
	CreditCount  int         `json:"credit_count"`
	Notice       interface{} `json:"notice"`
}

type Tag struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Platform struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Slug         string `json:"slug"`
	TokenAddress string `json:"token_address"`
}

type Quote struct {
	Price                 float64 `json:"price"`
	Volume24h             float64 `json:"volume_24h"`
	VolumeChange24h       float64 `json:"volume_change_24h"`
	PercentChange1h       float64 `json:"percent_change_1h"`
	PercentChange24h      float64 `json:"percent_change_24h"`
	PercentChange7d       float64 `json:"percent_change_7d"`
	PercentChange30d      float64 `json:"percent_change_30d"`
	PercentChange60d      float64 `json:"percent_change_60d"`
	PercentChange90d      float64 `json:"percent_change_90d"`
	MarketCap             float64 `json:"market_cap"`
	MarketCapDominance    float64 `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64 `json:"fully_diluted_market_cap"`
	TVL                   float64 `json:"tvl"`
	LastUpdated           string  `json:"last_updated"`
}

type Currency struct {
	ID                            int              `json:"id"`
	Name                          string           `json:"name"`
	Symbol                        string           `json:"symbol"`
	Slug                          string           `json:"slug"`
	NumMarketPairs                int              `json:"num_market_pairs"`
	DateAdded                     string           `json:"date_added"`
	Tags                          []Tag            `json:"tags"`
	MaxSupply                     int64            `json:"max_supply"`
	CirculatingSupply             float64          `json:"circulating_supply"`
	TotalSupply                   float64          `json:"total_supply"`
	Platform                      Platform         `json:"platform"`
	IsActive                      int              `json:"is_active"`
	InfiniteSupply                bool             `json:"infinite_supply"`
	CMCRank                       int              `json:"cmc_rank"`
	IsFiat                        int              `json:"is_fiat"`
	SelfReportedCirculatingSupply interface{}      `json:"self_reported_circulating_supply"`
	SelfReportedMarketCap         interface{}      `json:"self_reported_market_cap"`
	TvlRatio                      float64          `json:"tvl_ratio"`
	LastUpdated                   string           `json:"last_updated"`
	Quote                         map[string]Quote `json:"quote"`
}

type CMCResponse struct {
	Status Status              `json:"status"`
	Data   map[string]Currency `json:"data"`
}

type coinmarketcapSourceFactory struct{}

func init() {
	RegisterRestFactory("coinmarketcap", func() RestSourceFactory {
		return &coinmarketcapSourceFactory{}
	})
}

func (csf *coinmarketcapSourceFactory) NewRestSource(ctx context.Context, coinTickers map[string]Ticker, marketId2 string, refCoinTickers []Ticker, maxUseAge map[string]time.Duration, readInterval time.Duration, timeout time.Duration, multipliers map[string]float64, apiKey string) *SourceData {
	source := CoinmarketcapSource{
		coinTickers:  coinTickers,
		markets:      reverseInstrumentMap(coinTickers),
		latestPrices: &sync.Map{},
		lastUpdated:  make(map[string]time.Time),
		url:          CMC_URL,
		shortUrl:     CMC_URL[:33],
		idParam:      buildCMCIdParam(coinTickers),
		refCoinTickers: refCoinTickers,
		maxUseAge:    maxUseAge,
		readInterval: readInterval,
		timeout:      timeout,
		multipliers:  multipliers,
		apiKey:       apiKey,
	}
	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         []string{source.url},
	}
	return &sourceData
}

func (cmcs *CoinmarketcapSource) start(ctx context.Context) {
	go cmcs.reader(ctx)
}

func (cmcs *CoinmarketcapSource) listMarkets() (markets map[string]bool) {
	markets = make(map[string]bool, len(cmcs.coinTickers))
	for marketId := range cmcs.coinTickers {
		markets[marketId] = true
	}
	return markets
}

func (cmcs *CoinmarketcapSource) reader(ctx context.Context) {
	ticker := time.NewTicker(cmcs.readInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cmcs.readPriceData(ctx)
		}
	}
}

func (cmcs *CoinmarketcapSource) readPriceData(ctx context.Context) {

	prices, err := cmcs.readPrices(ctx)
	if err != nil {
		logrus.Warnf("error reading CMC prices: %s", err.Error())
		return
	}

	for coinId, currency := range prices {
		market, marketFound := cmcs.markets[coinId]
		if !marketFound {
			logrus.Warnf("no market found for coinId %s at url %s", coinId, cmcs.url)
		}
		multiplier, multFound := cmcs.multipliers[market]
		if !multFound {
			logrus.Warnf("no multiplier found for market %s at %s", market, cmcs.url)
			continue
		}

		quote := currency.Quote[cmcs.refCoinTickers[0].InstId]
		price := quote.Price * multiplier
		timestamp := quote.LastUpdated
		lastUpdated, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			logrus.Warnf("error parsing cmc timestamp %s, err: %s", timestamp, err.Error())
			return
		}

		priceTime := PriceTime{price, lastUpdated}
		stored := store(market, priceTime, cmcs.latestPrices, cmcs.shortUrl)
		if stored {
			cmcs.lastUpdated[market] = lastUpdated
		}
		if cmcs.getDataElapsed(market) > cmcs.maxUseAge[market] {
			logrus.Warnf("no data received for %s from %s since %v", market, cmcs.url, cmcs.lastUpdated[market])
		}
		elapsed := cmcs.getDataElapsed(market)
		maxUseAge := cmcs.maxUseAge[market]
		if elapsed > maxUseAge*3 {
			logrus.Warnf("%s received for %s from %s since %v", NO_PRICE_DATA, market, cmcs.url, cmcs.lastUpdated[market])
		}
	}
}

func (cmcs *CoinmarketcapSource) readPrices(ctx context.Context) (map[string]Currency, error) {
	var resp *http.Response
	var numReadAttempts int = 0
	for {
		client := &http.Client{
			Timeout: cmcs.timeout,
		}
		req, err := http.NewRequestWithContext(ctx, "GET", cmcs.url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating GET request %s gave error %s", cmcs.url, err.Error())
		}

		q := url.Values{}
		q.Add("id", cmcs.idParam)
		q.Add("convert_id", cmcs.refCoinTickers[0].InstId)

		req.Header.Set("Accepts", "application/json")
		req.Header.Add("X-CMC_PRO_API_KEY", cmcs.apiKey)
		req.URL.RawQuery = q.Encode()

		resp, err = client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			break
		} else {
			numReadAttempts++
			if numReadAttempts >= MAX_REST_READ_ATTEMPTS {
				return nil, fmt.Errorf("no price update response from %s, last error: %s", cmcs.url, err.Error())
			}
		}
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading GET response body, url %s, err %s", cmcs.url, err.Error())
	}

	var responseData CMCResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, fmt.Errorf("error deserializing GET response JSON %s, err: %s", string(body), err.Error())
	}
	return responseData.Data, nil
}

func (cmcs *CoinmarketcapSource) getDataElapsed(market string) time.Duration {
	if cmcs.lastUpdated[market].IsZero() {
		return time.Since(cmcs.startedAt)
	}
	return time.Since(cmcs.lastUpdated[market])
}

func buildCMCIdParam(coinTickers map[string]Ticker) string {
	var builder strings.Builder
	first := true
	for _, ticker := range coinTickers {
		if !first {
			builder.WriteString(",")
		}
		first = false
		builder.WriteString(ticker.InstId)
	}
	return builder.String()
}
