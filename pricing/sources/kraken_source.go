package sources

// see https://docs.kraken.com/websockets/#message-ohlc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

const (
	KN_URL        = "wss://ws.kraken.com"
	KN_BTC_USDT   = "XBT/USDT"
	KN_ETH_USDT   = "ETH/USDT"
	KN_SOL_USDT   = "SOL/USDT"
	KN_MATIC_USDT = "MATIC/USDT"
)

type KNSubscription struct {
	Name string `json:"name"`
}

type KNRequest struct {
	Event        string         `json:"event"`
	Pair         []string       `json:"pair"`
	Subscription KNSubscription `json:"subscription"`
}

type KNCandleData struct {
	ChannelID   int64
	Ohlc        [9]interface{}
	ChannelName string
	ProductID   string
}

type krakenSourceFactory struct{}

func init() {
	RegisterWSFactory("kraken", func() WSSourceFactory {
		return &krakenSourceFactory{}
	})
}

func (ksf *krakenSourceFactory) NewWSSource(ctx context.Context, coinTickers map[string]Ticker, maxUseAge map[string]time.Duration, multipliers map[string]float64) *SourceData {

	source := NewGobwasSource()
	source.url = KN_URL
	source.shortUrl = KN_URL
	source.coinTickers = coinTickers
	source.markets = reverseInstrumentMap(coinTickers)
	source.maxUseAge = maxUseAge
	source.multipliers = multipliers
	source.extractPrice = func(priceResponse []byte, multiplier float64) (PriceTime, error) {
		var dataArray []interface{}
		err := json.Unmarshal(priceResponse, &dataArray)
		if err != nil {
			return PriceTime{}, fmt.Errorf("error %s unmarshalling kraken candle data %s", err.Error(), string(priceResponse))
		}
		if len(dataArray) != 4 || len(dataArray[1].([]interface{})) != 9 {
			return PriceTime{}, fmt.Errorf("unexpected format for Kraken candle data \"%s\" from %s", string(priceResponse), source.url)
		}
		ohlc := dataArray[1].([]interface{})
		price, err := strconv.ParseFloat(ohlc[5].(string), 64)
		if err != nil {
			return PriceTime{}, fmt.Errorf("error %s parsing price in Kraken candle data %s", err.Error(), string(priceResponse))
		}
		closeTime, err := parseUnixTime(ohlc[0].(string))
		if err != nil {
			return PriceTime{}, fmt.Errorf("error %s parsing time in Kraken candle data %s", err.Error(), string(priceResponse))
		}
		return PriceTime{price * multiplier, closeTime}, nil
	}

	source.subscriptionBytes = func(coinTickers map[string]Ticker) []byte {
		pairs := make([]string, 0, len(coinTickers))
		for _, ticker := range coinTickers {
			pairs = append(pairs, ticker.InstId)
		}
		knSubscription := KNRequest{
			Event: "subscribe",
			Pair:  pairs,
			Subscription: KNSubscription{
				Name: "ohlc",
			},
		}
		bytes, err := json.Marshal(knSubscription)
		if err != nil {
			logrus.Warnf("Error marshalling Kraken market subscription request %v, err: %s", knSubscription, err)
		}
		return bytes
	}

	source.isPriceData = func(byteResponse []byte, op ws.OpCode) (bool, string) {
		if op != ws.OpText {
			return false, ""
		}
		var dataArray []interface{}
		err := json.Unmarshal(byteResponse, &dataArray)
		if err != nil {
			return false, ""
		}
		if len(dataArray) != 4 {
			return false, ""
		}
		marketValue, ok := dataArray[3].(string)
		if !ok {
			return false, ""
		}
		olhc, ok := dataArray[1].([]interface{})
		if !ok || len(olhc) != 9 {
			return false, ""
		}
		return true, marketValue
	}

	source.start(ctx)
	sourceData := SourceData{
		Markets:      source.listMarkets(),
		LatestPrices: source.latestPrices,
		Urls:         []string{source.url},
	}
	return &sourceData
}
