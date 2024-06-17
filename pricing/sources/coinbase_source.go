package sources

// see https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#ticker-batch-channel

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

const (
	CB_URL        = "wss://ws-feed.exchange.coinbase.com"
	CB_BTC_USDT   = "BTC-USDT"
	CB_BTC_USD    = "BTC-USD"
	CB_ETH_USDT   = "ETH-USDT"
	CB_ETH_USD    = "ETH-USD"
	CB_SOL_USDT   = "SOL-USDT"
	CB_SOL_USD    = "SOL-USD"
	CB_DOGE_USDT  = "DOGE-USDT"
	CB_DOGE_USD   = "DOGE-USD" // reference only
	CB_LDO_USD    = "LDO-USD"  // reference only
	CB_ARB_USD    = "ARB-USD"  // reference only
	CB_MATIC_USDT = "MATIC-USDT"
	CB_MATIC_USD  = "MATIC-USD"
)

type CBSubscription struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type CBTickerData struct {
	Type        string      `json:"type"`
	Sequence    int64       `json:"sequence"`
	ProductID   string      `json:"product_id"`
	Price       json.Number `json:"price"`
	Open24h     string      `json:"open_24h"`
	Volume24h   string      `json:"volume_24h"`
	Low24h      string      `json:"low_24h"`
	High24h     string      `json:"high_24h"`
	Volume30d   string      `json:"volume_30d"`
	BestBid     string      `json:"best_bid"`
	BestBidSize string      `json:"best_bid_size"`
	BestAsk     string      `json:"best_ask"`
	BestAskSize string      `json:"best_ask_size"`
	Side        string      `json:"side"`
	Time        time.Time   `json:"time"`
	TradeID     int64       `json:"trade_id"`
	LastSize    string      `json:"last_size"`
}

type coinbaseSourceFactory struct{}

func init() {
	RegisterWSFactory("cb", func() WSSourceFactory {
		return &coinbaseSourceFactory{}
	})
	RegisterWSFactory("cb2", func() WSSourceFactory {
		return &coinbaseSourceFactory{}
	})
}

func (csf *coinbaseSourceFactory) NewWSSource(ctx context.Context, coinTickers map[string]Ticker, maxUseAge map[string]time.Duration, multipliers map[string]float64) *SourceData {
	source := NewGobwasSource()
	source.url = CB_URL
	source.shortUrl = CB_URL
	source.coinTickers = coinTickers
	source.markets = reverseInstrumentMap(coinTickers)
	source.maxUseAge = maxUseAge
	source.multipliers = multipliers
	source.extractPrice = func(priceResponse []byte, multiplier float64) (PriceTime, error) {
		res := CBTickerData{}
		err := json.Unmarshal(priceResponse, &res)
		if err != nil {
			return PriceTime{}, err
		}
		price, err := res.Price.Float64()
		if err != nil {
			return PriceTime{}, err
		}
		return PriceTime{price * multiplier, res.Time}, nil
	}

	source.subscriptionBytes = func(coinTickers map[string]Ticker) []byte {
		products := make([]string, 0, len(coinTickers))
		for _, ticker := range coinTickers {
			products = append(products, ticker.InstId)
		}
		cbSubscription := CBSubscription{
			Type:       "subscribe",
			ProductIDs: products,
			Channels:   []string{"ticker_batch"},
		}
		bytes, err := json.Marshal(cbSubscription)
		if err != nil {
			logrus.Warnf("error marshalling Coinbase market subscription request %v, err: %s", cbSubscription, err)
		}
		return bytes
	}

	source.isPriceData = func(byteResponse []byte, op ws.OpCode) (bool, string) {
		if op != ws.OpText {
			return false, ""
		}
		res := CBTickerData{}
		err := json.Unmarshal(byteResponse, &res)
		return err == nil && res.Type == "ticker", res.ProductID
	}

	source.pingBytes = func() []byte {
		return []byte("ping")
	}

	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         []string{source.url},
	}
	return &sourceData
}
