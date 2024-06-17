package pricing

/*
Receives config information and creates the price sources required for each market. Starts a market price service instance for each market, passing it its sources.
*/

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"

	"github.com/strips-finance/rabbit-dex-backend/pricing/sources"
)

const (
	DEFAULT_UPDATE_INTERVAL = 5 * time.Second
	DEFAULT_MAX_USE_AGE     = time.Minute
	DEFAULT_READ_INTERVAL   = 25 * time.Second
	DEFAULT_READ_TIMEOUT    = 20 * time.Second
	DEFAULT_REFERENCE_COIN  = "USDT"
)

type PricingService struct{}

type CoinData struct {
	CoinId  string           `yaml:"coin_id"`
	Tickers []sources.Ticker `yaml:"tickers"`
}

type InstId struct {
	ExchangeId string `yaml:"exchange_id"`
	InstId     string `yaml:"inst_id"`
}

type ExchangeData struct {
	ExchangeId   string `yaml:"exchange_id"`
	ReadInterval string `yaml:"read_interval"`
	ReadTimeout  string `yaml:"read_timeout"`
}

type MarketData struct {
	MarketId   string           `yaml:"market_id"`
	Sources    []sources.Ticker `yaml:"sources"`
	Multiplier float64          `yaml:"multiplier"`
	MaxUseAge  string           `yaml:"max_use_age"`
}

type PriceReceiver interface {
	UpdateIndexPrice(ctx context.Context, market_id string, index_price float64) error
}

// compile-time check that model.ApiModel implements PriceReceiver
var _ PriceReceiver = (*model.ApiModel)(nil)

func NewPricingService() *PricingService {
	// logrus.SetLevel(logrus.WarnLevel)
	return &PricingService{}
}

func (ps *PricingService) Run(ctx context.Context, config *Config) {
	broker, err := model.GetBroker()
	if err != nil {
		logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("PricingService, can't create broker err=%v", err)
		return
	}
	apiModel := model.NewApiModel(broker)
	ps.LoadConfig(ctx, apiModel, config)
}

func (ps *PricingService) LoadConfig(ctx context.Context, apiModel PriceReceiver, config *Config) {

	updateInterval, _ := timeFromStr(config.Service.UpdateInterval, "update_interval", DEFAULT_UPDATE_INTERVAL)

	defaultMaxUseAge, _ := timeFromStr(config.Service.DefaultMaxUseAge, "default_max_use_Age", DEFAULT_MAX_USE_AGE)

	// put exchange data into a map
	exchanges := make(map[string]ExchangeData)
	for _, data := range config.Service.ExchangeData {
		exchanges[data.ExchangeId] = data
	}
	// put coin data into a map
	coinTickers := make(map[string]map[string][]sources.Ticker)
	for _, data := range config.Service.CoinData {
		if coinTickers[data.CoinId] == nil {
			coinTickers[data.CoinId] = make(map[string][]sources.Ticker)
		}
		for _, ticker := range data.Tickers {
			tickerList := coinTickers[data.CoinId][ticker.ExchangeId]
			tickerList = append(tickerList, ticker)
			coinTickers[data.CoinId][ticker.ExchangeId] = tickerList
		}
	}

	// get the reference coin, which is used by the REST sources
	// to convert USD prices to a common currency (currently always USDT)
	refCoin := config.Service.ReferenceCoin
	if refCoin == "" {
		refCoin = DEFAULT_REFERENCE_COIN
	}

	marketData := config.Service.MarketData
	numMarkets := len(marketData)
	markets := make([]string, numMarkets)
	for i, data := range marketData {
		markets[i] = data.MarketId
	}
	logrus.Infof(
		"running pricing service with update interval %ds, default max price delay %ds, and %d markets %v",
		int64(updateInterval.Seconds()),
		int64(defaultMaxUseAge.Seconds()),
		numMarkets,
		markets)

	// create a price source for each exchange
	exchangeSources, maxUseAges := createExchangeSources(ctx, marketData, exchanges,
		coinTickers, refCoin, defaultMaxUseAge)

	// create the source map of marketId -> ChanWithSources
	// which is used by the aggregator
	// ChanWithSources is a struct containing, for that market:
	//    the set of sources from which the aggregator will receive price updates
	//    the channel on which the aggregator will send price updates
	//    the maximum age of a price which the aggregator will accept
	sourceMap := make(map[string]sources.ChanWithSources, len(exchangeSources))
	for _, sourceData := range exchangeSources {
		// add this source's data to the source map entry
		// for each market the source covers
		for marketId := range sourceData.Markets {
			// first find or create the map entry for this market
			var sourceMapEntry sources.ChanWithSources
			var exists bool
			if sourceMapEntry, exists = sourceMap[marketId]; !exists {
				sourceMapEntry = sources.ChanWithSources{
					Sources:  make([]*sources.SourceData, 0, 4),
					Channel:  make(chan []float64, 1),
					MaxDelay: maxUseAges[marketId],
				}
			}
			// add this source's data to the map entry for this market
			sourceMapEntry.Sources = append(sourceMapEntry.Sources, sourceData)
			sourceMap[marketId] = sourceMapEntry
		}
	}
	// create a market price service for each market, the market price services
	// receive the aggregated price updates from the sources, process them to
	// produce a market price, and send the market price to tarantool
	for marketId, cws := range sourceMap {
		mps := NewMarketPriceService(marketId, len(cws.Sources),
			cws.Channel, apiModel)
		mps.start(ctx)
		logrus.Infof("created price service for market_id %s", marketId)
	}
	// create the aggregator that will feed the data
	// from the exchange sources to the market price services
	sources.NewAggregator(ctx, sourceMap, updateInterval)
}

// creates a price source for each exchange,
// returns a slice of SourceData and a map of marketId -> maxUseAge
// SourceData is a struct containing:
//
//	the set of markets covered by the exchange source,
//	the map which will be updated with the latest prices,
//	the exchange URL
//
// maxUseAge is the oldest a price can be and still be used
func createExchangeSources(ctx context.Context, marketData []MarketData,
	exchanges map[string]ExchangeData, coinTickers map[string]map[string][]sources.Ticker,
	refCoin string,
	defaultMaxUseAge time.Duration) (sourceData []*sources.SourceData,
	maxUseAges map[string]time.Duration) {

	numMarkets := len(marketData)
	marketExchangeTickers := make(map[string]map[string]sources.Ticker, numMarkets)
	maxUseAges = make(map[string]time.Duration, numMarkets)
	multipliers := make(map[string]float64, numMarkets)
	for _, data := range marketData {
		marketExchangeTickers[data.MarketId] = make(map[string]sources.Ticker, len(data.Sources))
		for _, ticker := range data.Sources {
			marketExchangeTickers[data.MarketId][ticker.ExchangeId] = ticker
		}
		multiplier := data.Multiplier
		if multiplier == 0 {
			multiplier = 1.0
		}
		multipliers[data.MarketId] = multiplier
		maxUseAge, err := timeFromStr(data.MaxUseAge, fmt.Sprintf("%s max_use_age", data.MarketId), defaultMaxUseAge)
		if err != nil {
			logrus.Warnf("can't parse max_use_age %s for market_id=%s, err=%s, using default %ds", data.MaxUseAge, data.MarketId, err.Error(), int64(defaultMaxUseAge.Seconds()))
			maxUseAge = defaultMaxUseAge
		}
		maxUseAges[data.MarketId] = maxUseAge
	}
	exchangeMarketTickers := reorderTickerMap(marketExchangeTickers)
	sourceData = make([]*sources.SourceData, 0, len(exchangeMarketTickers))
	for exchangeId, marketTickers := range exchangeMarketTickers {
		exchangeSrcData := buildExchangeSource(ctx, exchangeId, marketTickers, maxUseAges, multipliers, exchanges, coinTickers, refCoin)
		if exchangeSrcData != nil {
			sourceData = append(sourceData, exchangeSrcData)
		}
	}
	return sourceData, maxUseAges
}

// `marketId` -> `exchangeId` -> `instId` is used as the config file
// format because grouping by market makes it easy to see the sources
// used for each market.
// `exchangeId` -> `marketId`> `instId` is the format needed to create the
// price sources. The conversion is done here.
func reorderTickerMap(marketExchangeTickers map[string]map[string]sources.Ticker) (exchangeMarketInstIds map[string]map[string]sources.Ticker) {
	exchangeMarketInstIds = make(map[string]map[string]sources.Ticker)
	for marketId, exchangeCoinIds := range marketExchangeTickers {
		for exchangeId, ticker := range exchangeCoinIds {
			if exchangeMarketInstIds[exchangeId] == nil {
				exchangeMarketInstIds[exchangeId] = make(map[string]sources.Ticker)
			}
			exchangeMarketInstIds[exchangeId][marketId] = ticker
		}
	}
	return exchangeMarketInstIds
}

// find a factory for this exchange and use it to create a price source
func buildExchangeSource(ctx context.Context, exchangeId string,
	marketTickers map[string]sources.Ticker, maxUseAges map[string]time.Duration, multipliers map[string]float64, exchanges map[string]ExchangeData,
	coinTickers map[string]map[string][]sources.Ticker, refCoin string) *sources.SourceData {

	// look for a web socket source factory first
	wsFact := sources.LoadWSSourceFactory(exchangeId)
	if wsFact != nil {
		return wsFact.NewWSSource(ctx, marketTickers, maxUseAges, multipliers)
	} else {
		// didn't find a web socket factory, look for a rest factory instead
		restFact := sources.LoadRestSourceFactory(exchangeId)
		if restFact != nil {
			// creating a rest factory requires a bit of additional information
			readInterval, readTimeout, refCoinTickers, apiKey := gatherAdditionalRestSrcInfo(exchangeId, exchanges, coinTickers, refCoin)
			return restFact.NewRestSource(ctx, marketTickers, refCoin,
				refCoinTickers, maxUseAges, readInterval, readTimeout, multipliers,
				apiKey)
		}
	}
	// couldn't find a source factory for this exchange
	logrus.Warnf("can't find factory for exchange_id=%s", exchangeId)
	return nil
}

func gatherAdditionalRestSrcInfo(exchangeId string,
	exchanges map[string]ExchangeData,
	coinTickers map[string]map[string][]sources.Ticker, refCoin string) (readInterval time.Duration,
	readTimeout time.Duration, refCoinTickers []sources.Ticker, apiKey string) {

	readInterval, _ = timeFromStr(exchanges[exchangeId].ReadInterval, "read_interval", DEFAULT_READ_INTERVAL)
	readTimeout, _ = timeFromStr(exchanges[exchangeId].ReadTimeout, "read_timeout", DEFAULT_READ_TIMEOUT)
	coinId2Map, ok := coinTickers[refCoin]
	if !ok {
		logrus.Warnf("can't find coin data for coin_id=%s", refCoin)
	}
	for _, tickers := range coinId2Map {
		for _, ticker := range tickers {
			if ticker.ExchangeId == exchangeId {
				refCoinTickers = append(refCoinTickers, ticker)
			}
		}
	}
	apiKey = getApiKey(exchangeId)
	return
}

func getApiKey(exchangeId string) string {
	if exchangeId == "coingecko" {
		return os.Getenv("COINGECKO_API_KEY")
	}
	if exchangeId == "geckoterminal" {
		return os.Getenv("COINGECKO_API_KEY")
	}
	if exchangeId == "coinmarketcap" {
		return os.Getenv("COINMARKETCAP_API_KEY")
	}
	return ""
}

func timeFromStr(intervalStr string, desc string,
	defaultValue time.Duration) (time.Duration, error) {

	if intervalStr == "" {
		return defaultValue, nil
	}
	intervalSeconds, err := strconv.ParseInt(intervalStr, 10, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("%s can't parse value %s err=%s", desc, intervalStr, err.Error())
	}
	return time.Second * time.Duration(intervalSeconds), nil
}
