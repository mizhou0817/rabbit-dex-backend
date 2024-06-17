package sources

import (
	"context"
	"time"
)

// interface containing a single func for creating new REST sources
// (coingecko, geckoterminal, coinmarketcap)
type RestSourceFactory interface {
	NewRestSource(ctx context.Context, coinTickers map[string]Ticker, marketId2 string, refCoinTickers []Ticker, maxDelays map[string]time.Duration, readInterval time.Duration, timeout time.Duration, multipliers map[string]float64, apiKey string) *SourceData
}

// interface containing a single func for creating new web socket sources
// (binance, okx, coinbase, kraken)
type WSSourceFactory interface {
	NewWSSource(ctx context.Context, coinTickers map[string]Ticker, maxDelays map[string]time.Duration, multipliers map[string]float64) *SourceData
}

type RestFactoryFunc func() RestSourceFactory
type WSFactoryFunc func() WSSourceFactory

var restFactoryMap = make(map[string]RestFactoryFunc)
var wsFactoryMap = make(map[string]WSFactoryFunc)

// register something that can create REST source factories, for example
// under the id "coingecko" register a creator of coingecko source factories
func RegisterRestFactory(id string, factory RestFactoryFunc) {
    restFactoryMap[id] = factory
}

// register something that can create WS source factories, for example
// under the id "binance" register a creator of binance source factories
func RegisterWSFactory(id string, factory WSFactoryFunc) {
    wsFactoryMap[id] = factory
}

// load a previously registered REST source factory by registered id
func LoadRestSourceFactory(id string) RestSourceFactory {
    if factory, exists := restFactoryMap[id]; exists {
        return factory()
    }
    return nil
}

// load a previously registered WS source factory by registered id
func LoadWSSourceFactory(id string) WSSourceFactory {
    if factory, exists := wsFactoryMap[id]; exists {
        return factory()
    }
    return nil
}