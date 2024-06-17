package sources

// see https://www.okx.com/docs-v5/en/#public-data-websocket-index-candlesticks-channel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

const (
	OKX_URL        = "wss://ws.okx.com:8443/ws/v5/business"
	OKX_BTC_USDT   = "BTC-USDT"
	OKX_ETH_USDT   = "ETH-USDT"
	OKX_SOL_USDT   = "SOL-USDT"
	OKX_DOGE_USDT  = "DOGE-USDT"
	OKX_LDO_USDT   = "LDO-USDT"
	OKX_ARB_USDT   = "ARB-USDT"
	OKX_MATIC_USDT = "MATIC-USDT"
	OKX_SUI_USDT   = "SUI-USDT"
	OKX_PEPE_USDT  = "PEPE-USDT"
)

type OkxPriceData struct {
	Event string          `json:"event"`
	Arg   OkxArg          `json:"arg"`
	Data  [][]json.Number `json:"data"`
}

type OkxSubscription struct {
	Op   string   `json:"op"`
	Args []OkxArg `json:"args"`
}

type OkxArg struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type okxSourceFactory struct{}

func init() {
	RegisterWSFactory("okx", func() WSSourceFactory {
		return &okxSourceFactory{}
	})
}

func (oks *okxSourceFactory) NewWSSource(ctx context.Context, coinTickers map[string]Ticker, maxUseAge map[string]time.Duration, multipliers map[string]float64) *SourceData {
	source := NewGobwasSource()
	source.url = OKX_URL
	source.shortUrl = OKX_URL
	source.coinTickers = coinTickers
	source.markets = reverseInstrumentMap(coinTickers)
	source.maxUseAge = maxUseAge
	source.multipliers = multipliers
	source.extractPrice = func(priceResponse []byte, multiplier float64) (PriceTime, error) {
		res := OkxPriceData{}
		err := json.Unmarshal(priceResponse, &res)
		if err != nil {
			return PriceTime{}, err
		}
		data := res.Data
		if len(data) < 1 {
			return PriceTime{}, fmt.Errorf("no price update data in response \"%s\" from %s", priceResponse, source.url)
		}
		prices := data[0]
		if len(prices) != 6 {
			return PriceTime{}, fmt.Errorf("unexpected price update data %s in response from %s", prices, source.url)
		}
		price, err := prices[4].Float64()
		if err != nil {
			return PriceTime{}, err
		}

		var closeTime time.Time
		closeMillis, err := prices[0].Int64()
		if err == nil {
			closeTime = time.Unix(0, closeMillis*int64(time.Millisecond))
			now := time.Now()
			if closeTime.After(now) {
				closeTime = now
			}
		} else {
			logrus.Warnf("error %s parsing okx close time %s", err.Error(), prices[0].String())
			closeTime = time.Now()
		}
		return PriceTime{price * multiplier, closeTime}, nil
	}
	source.subscriptionBytes = func(coinTickers map[string]Ticker) []byte {
		args := make([]OkxArg, 0, len(coinTickers))
		for _, ticker := range coinTickers {
			args = append(args, OkxArg{
				Channel: "index-candle1m",
				InstId:  ticker.InstId,
			})
		}
		okxSubscription := OkxSubscription{
			Op:   "subscribe",
			Args: args,
		}
		bytes, err := json.Marshal(okxSubscription)
		if err != nil {
			logrus.Warnf("Error marshalling OKX subscription request %v, err: %s", coinTickers, err)
		}
		return bytes
	}

	source.pingBytes = func() []byte {
		return []byte("ping")
	}

	source.isPriceData = func(byteResponse []byte, op ws.OpCode) (bool, string) {
		if op != ws.OpText {
			return false, ""
		}
		res := OkxPriceData{}
		err := json.Unmarshal(byteResponse, &res)
		isPrice := err == nil &&
			res.Event == "" &&
			len(res.Data) > 0 &&
			res.Arg != (OkxArg{})
		if !isPrice {
			return false, ""
		}
		return true, res.Arg.InstId
	}

	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         []string{source.url},
	}
	return &sourceData
}
