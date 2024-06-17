package sources

// see https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-streams.md#klinecandlestick-streams

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gobwas/ws"
)

const (
	BINANCE_URL   = "wss://stream.binance.com:9443/stream?streams="
	BSC_BTC_USDT  = "btcusdt@kline_1m"
	BSC_ETH_USDT  = "ethusdt@kline_1m"
	BSC_SOL_USDT  = "solusdt@kline_1m"
	BSC_SUI_USDT  = "suiusdt@kline_1m"
	BSC_ARB_USDT  = "arbusdt@kline_1m"
	BSC_DOGE_USDT = "dogeusdt@kline_1m"
	BSC_LDO_USDT  = "ldousdt@kline_1m"
	BSC_PEPE_USDT = "pepeusdt@kline_1m"
)

type binanceSourceFactory struct{}

func init() {
	RegisterWSFactory("binance", func() WSSourceFactory {
		return &binanceSourceFactory{}
	})
}

func (bsf *binanceSourceFactory) NewWSSource(ctx context.Context, coinTickers map[string]Ticker, maxUseAge map[string]time.Duration, multipliers map[string]float64) *SourceData {
	source := NewGobwasSource()
	source.coinTickers = coinTickers
	source.markets = reverseInstrumentMap(coinTickers)
	source.url = constructURL(coinTickers)
	source.shortUrl = BINANCE_URL[:29]
	source.maxUseAge = maxUseAge
	source.multipliers = multipliers
	source.isPriceData = func(byteResponse []byte, op ws.OpCode) (bool, string) {
		if op != ws.OpText {
			return false, ""
		}
		res := StreamWrapper{}
		err := json.Unmarshal(byteResponse, &res)
		if err != nil || res.Data.Lk == (KlineData{}) {
			return false, ""
		}
		return true, res.Stream
	}
	source.extractPrice = func(priceResponse []byte, multiplier float64) (PriceTime, error) {
		res := StreamWrapper{}
		err := json.Unmarshal(priceResponse, &res)
		if err != nil {
			return PriceTime{}, err
		}
		closeTime := time.Unix(0, int64(res.Data.Lk.T)*int64(time.Millisecond))
		now := time.Now()
		if closeTime.After(now) {
			closeTime = now
		}
		price, err := strconv.ParseFloat(res.Data.Lk.Lc, 64)
		if err != nil {
			return PriceTime{}, err
		}
		return PriceTime{price * multiplier, closeTime}, nil
	}
	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         []string{source.url},
	}
	return &sourceData
}

type StreamWrapper struct {
	Stream string       `json:"stream"`
	Data   KlineWrapper `json:"data"`
}

type KlineWrapper struct {
	Le string    `json:"e"` // Event type
	E  int       `json:"E"` // Event time
	Ls string    `json:"s"` // Symbol
	Lk KlineData `json:"k"`
}

type KlineData struct {
	Lt int    `json:"t"` // Kline start time
	T  int    `json:"T"` // Kline close time
	Ls string `json:"s"` // Symbol
	Li string `json:"i"` // Interval ("1m", "5m", "1h" etc )
	Lf int    `json:"f"` // First trade ID
	L  int    `json:"L"` // Last trade ID
	Lo string `json:"o"` // Open price (floating point)
	Lc string `json:"c"` // Close price (floating point)
	Lh string `json:"h"` // High price (floating point)
	Ll string `json:"l"` // Low price (floating point)
	Lv string `json:"v"` // Base asset volume (integer)
	Ln int    `json:"n"` // Number of trades
	Lx bool   `json:"x"` // Is this kline closed?
	Lq string `json:"q"` // Quote asset volume (floating point)
	V  string `json:"V"` // Taker buy base asset volume (floating point)
	Q  string `json:"Q"` // Taker buy quote asset volume (floating point)
	B  string `json:"B"` // Ignore
}

func constructURL(coinTickers map[string]Ticker) string {
	var builder strings.Builder
	builder.WriteString(BINANCE_URL)
	first := true
	for _, ticker := range coinTickers {
		if !first {
			builder.WriteString("/")
		}
		first = false
		builder.WriteString(ticker.InstId)
	}
	return builder.String()
}

// URL is built as, for example:
// "wss://stream.binance.com:9443/stream?streams=" "btcusdt@kline_1m" "/" "ethusdt@kline_1m" "/" "solusdt@kline_1m" "/" "suiusdt@kline_1m" "/" "arbusdt@kline_1m" "/" "dogeusdt@kline_1m" "/" "ldousdt@kline_1m" "/" "pepeusdt@kline_1m"
