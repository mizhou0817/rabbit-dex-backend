package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GT_URL_PART_1 = "https://api.geckoterminal.com/api/v2/networks/"
	GT_URL_PART_2 = "/tokens/multi"
	// GT_URL = "https://api.geckoterminal.com/api/v2/networks/eth/tokens/multi"

)

type Response struct {
	Data []DataItem `json:"data"`
}

type DataItem struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"`
	Attributes    Attributes    `json:"attributes"`
	Relationships Relationships `json:"relationships"`
}

type Attributes struct {
	Address           string    `json:"address"`
	Name              string    `json:"name"`
	Symbol            string    `json:"symbol"`
	CoingeckoCoinID   string    `json:"coingecko_coin_id"`
	Decimals          int       `json:"decimals"`
	TotalSupply       string    `json:"total_supply"`
	PriceUSD          string    `json:"price_usd"`
	FdvUSD            string    `json:"fdv_usd"`
	TotalReserveInUSD string    `json:"total_reserve_in_usd"`
	VolumeUSD         VolumeUSD `json:"volume_usd"`
	MarketCapUSD      string    `json:"market_cap_usd"`
}

type VolumeUSD struct {
	H24 string `json:"h24"`
}

type Relationships struct {
	TopPools TopPools `json:"top_pools"`
}

type TopPools struct {
	Data []TopPoolData `json:"data"`
}

type TopPoolData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type geckoterminalSourceFactory struct{}

func init() {
	RegisterRestFactory("geckoterminal", func() RestSourceFactory {
		return &geckoterminalSourceFactory{}
	})
}

func (gtsf *geckoterminalSourceFactory) NewRestSource(ctx context.Context, coinTickers map[string]Ticker, marketId2 string, refCoinTickers []Ticker, maxDelays map[string]time.Duration, readInterval time.Duration, timeout time.Duration, multipliers map[string]float64, apiKey string) *SourceData {
	urls := buildGTUrls(coinTickers, refCoinTickers)
	source := BasegeckoSource{
		latestPrices:     &sync.Map{},
		urls:             urls,
		markets:          buildMarkets(coinTickers, marketId2, refCoinTickers),
		marketId2:        marketId2,
		lastDataReceived: make(map[string]time.Time),
		coinTickers:      coinTickers,
		refCoinTickers:   refCoinTickers,
		maxUseAge:        maxDelays,
		readInterval:     readInterval,
		timeout:          timeout,
		multipliers:      multipliers,
		apiKey:           apiKey,
	}

	source.extractPrices = func(responseBody []byte) (map[string]PriceTime, error) {
		var response Response
		err := json.Unmarshal(responseBody, &response)
		if err != nil {
			return nil, fmt.Errorf("error deserializing GET response JSON %s, err: %s", string(responseBody), err.Error())
		}
		prices := make(map[string]PriceTime, len(response.Data))
		now := time.Now()
		for _, dataItem := range response.Data {
			price, err := strconv.ParseFloat(dataItem.Attributes.PriceUSD, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing price %s in response %s, err: %s", dataItem.Attributes.PriceUSD, string(responseBody), err.Error())
			}
			prices[dataItem.Attributes.Address] = PriceTime{price, now}
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

func buildGTUrls(coinTickers map[string]Ticker, refCoinTickers []Ticker) []string {
	tickersByNetwork := make(map[string][]Ticker)
	for _, ticker := range coinTickers {
		tickersByNetwork[ticker.Network] = append(tickersByNetwork[ticker.Network], ticker)
	}
	refTickersByNetwork := make(map[string]Ticker)
	for _, ticker := range refCoinTickers {
		refTickersByNetwork[ticker.Network] = ticker
	}
	gtUrls := make([]string, 0, len(tickersByNetwork))
	for network, tickers := range tickersByNetwork {
		var urlBuilder strings.Builder
		urlBuilder.WriteString(GT_URL_PART_1)
		urlBuilder.WriteString(network)
		urlBuilder.WriteString(GT_URL_PART_2)
		urlBuilder.WriteString("/")
		first := true
		for _, ticker := range tickers {
			if !first {
				urlBuilder.WriteString(",")
			}
			first = false
			urlBuilder.WriteString(ticker.InstId)
		}
		if !first {
			urlBuilder.WriteString(",")
		}
		refTicker, ok := refTickersByNetwork[network]
		if !ok {
			panic(fmt.Sprintf("geckoterminal config has no reference ticker for network %s", network))
		}
		urlBuilder.WriteString(refTicker.InstId)
		gtUrls = append(gtUrls, urlBuilder.String())
	}
	return gtUrls
}
