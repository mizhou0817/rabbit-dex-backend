package api

import (
	"math/rand"
	"strings"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

type MarketResponse struct {
	BaseCurrency             string          `json:"base_currency"`
	QuoteCurrency            string          `json:"quote_currency"`
	ProductType              string          `json:"product_type"`
	OpenInterest             decimal.Decimal `json:"open_interest"`
	LongRatio                decimal.Decimal `json:"long_ratio"`
	ShortRatio               decimal.Decimal `json:"short_ratio"`
	NextFundingRateTimestamp int64           `json:"next_funding_rate_timestamp"`

	AverageDailyVolume              decimal.Decimal `json:"average_daily_volume"`
	LastTradePrice24High            decimal.Decimal `json:"last_trade_price_24high"`
	LastTradePrice24Low             decimal.Decimal `json:"last_trade_price_24low"`
	LastTradePrice24hChangePremium  decimal.Decimal `json:"last_trade_price_24h_change_premium"`
	LastTradePrice24hChangeBasis    decimal.Decimal `json:"last_trade_price_24h_change_basis"`
	AverageDailyVolumeChangePremium decimal.Decimal `json:"average_daily_volume_change_premium"`
	AverageDailyVolumeChangeBasis   decimal.Decimal `json:"average_daily_volume_change_basis"`
}

type LineChartsResponse struct {
	ChartPrices []decimal.Decimal `json:"chart_prices,omitempty"`
}

type ExtendedMarketResponse struct {
	model.MarketData
	MarketResponse
	LineChartsResponse
}

type MarketRequest struct {
	MarketId   string `form:"market_id"`
	LineCharts *bool  `form:"line_charts" binding:"omitempty"`
}

type MarketTradesRequest struct {
	MarketId string `form:"market_id" binding:"required"`
}

type MarketOrderBookRequest struct {
	MarketId string `form:"market_id" binding:"required"`
}

type MarketCoinsRequest struct {
	CoinIds string `form:"coin_ids" binding:"required"`
}

// TODO: move it to tarantool
var DEPOSITS_TO_LEVERAGE = decimal.NewFromFloat(5000000.0).Mul(decimal.NewFromFloat(20.0))

const TOTAL_MARKETS = 55

// FIXME: it's workaround to speedup /markets endpoint with data that can be cached
// i believe need to be refactored with common architecture practics
type marketViewOiType struct {
	totalVolume     decimal.Decimal
	perMarketVolume decimal.Decimal
}

var (
	marketViewOiFuncMut [TOTAL_MARKETS]sync.Mutex
	marketViewOiMapMut  sync.RWMutex
	marketViewOi        map[string]marketViewOiType = map[string]marketViewOiType{}
)

type marketViewDailyType struct {
	averageDailyVolume              decimal.Decimal
	lastTradePrice24High            decimal.Decimal
	lastTradePrice24Low             decimal.Decimal
	lastTradePrice24hChangePremium  decimal.Decimal
	lastTradePrice24hChangeBasis    decimal.Decimal
	averageDailyVolumeChangePremium decimal.Decimal
	averageDailyVolumeChangeBasis   decimal.Decimal
}

var (
	marketViewDailyFuncMut [TOTAL_MARKETS]sync.Mutex
	marketViewDailyMapMut  sync.RWMutex
	marketViewDaily        map[string]marketViewDailyType = map[string]marketViewDailyType{}
)

var (
	marketViewPriceFuncMut [TOTAL_MARKETS]sync.Mutex
	marketViewPriceMapMut  sync.RWMutex
	marketViewPrice        map[string][]decimal.Decimal = map[string][]decimal.Decimal{}
)

func HandleMarket(c *gin.Context) {
	var request MarketRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	response := make([]ExtendedMarketResponse, 0)
	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)
	db := ctx.TimeScaleDB

	filterMarketIds := strings.Split(request.MarketId, ",")

	//TODO: replace all to one SQL request.
	//for this no need to call GetMarketData from tarantool
	for ii, market := range ctx.Config.Service.Markets {
		marketIndex, market := ii, market // doing correct vars scope
		if request.MarketId != "" && filterMarketIds != nil && len(filterMarketIds) > 0 {
			var contains bool

			for _, filterMarketId := range filterMarketIds {
				if filterMarketId == market {
					contains = true
				}
			}

			if !contains {
				continue
			}
		}

		totalVolume := decimal.NewFromFloat(0.0)
		perMarketVolume := decimal.NewFromFloat(0.0)
		func() {
			marketViewOiMapMut.RLock()
			if m, ok := marketViewOi[market]; ok {
				totalVolume = m.totalVolume
				perMarketVolume = m.perMarketVolume
			}
			marketViewOiMapMut.RUnlock()

			go func() {
				if !marketViewOiFuncMut[marketIndex].TryLock() {
					return
				}
				defer marketViewOiFuncMut[marketIndex].Unlock()
				/*
					Calc OI data
				*/
				sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
				sqlTotal := sqlBuilder.
					Select("COALESCE(SUM(average_daily_volume), 0) as total").
					From("market_data_view")

				sqlPerMarket, argsOi, errOi := sqlBuilder.
					Select("COALESCE(average_daily_volume, 0) as per_market").
					From("market_data_view").
					Where("market_id = ?", market).
					Limit(1).ToSql()

				if errOi != nil {
					logrus.Error(errors.Wrap(errOi, "build permarket sql"))
					return
				}

				sqlOi, _, errOi := sqlBuilder.
					Select("a.total, b.per_market").
					FromSelect(sqlTotal, "a").
					JoinClause("CROSS JOIN (" + sqlPerMarket + ") AS b").
					ToSql()

				if errOi != nil {
					logrus.Error(errors.Wrap(errOi, "build oi sql"))
					return
				}

				var totalVolume decimal.Decimal
				var perMarketVolume decimal.Decimal
				errOi = db.QueryRow(c.Request.Context(), sqlOi, argsOi...).Scan(&totalVolume, &perMarketVolume)
				if errOi != nil && errOi != pgx.ErrNoRows {
					logrus.Error(errors.Wrap(errOi, "exec/scan oi sql"))
					return
				}
				marketViewOiMapMut.Lock()
				marketViewOi[market] = marketViewOiType{
					totalVolume:     totalVolume,
					perMarketVolume: perMarketVolume,
				}
				marketViewOiMapMut.Unlock()
			}()
		}()

		LongRatio := decimal.NewFromFloat(0.0)
		ShortRatio := decimal.NewFromFloat(0.0)
		OpenInterest := decimal.NewFromFloat(0.0)
		if !totalVolume.IsZero() {
			share := perMarketVolume.Div(totalVolume)

			//This one to estimate OI based on volume 10-15%
			coef1 := decimal.NewFromFloat(rand.Float64()*0.05 + 0.10)

			//This one is to randomize Deposits from 95% - 105% from what we have
			coef2 := decimal.NewFromFloat(rand.Float64()*0.1 + 0.95)

			OpenInterest = decimal.Min(
				DEPOSITS_TO_LEVERAGE.Mul(coef2).Mul(share),
				perMarketVolume.Mul(coef1),
			).Round(2)

			// in the range 45-55%
			LongRatio = decimal.NewFromFloat(rand.Float64()*0.1 + 0.45).Round(2)
			ShortRatio = decimal.Max(decimal.NewFromFloat(0.0), decimal.NewFromFloat(1.0).Sub(LongRatio)).Round(2)
		}

		res1, err := apiModel.GetMarketData(c.Request.Context(), market)
		if err != nil {
			ErrorResponse(c, err)
			return
		}

		baseCurrency := "unknown"
		quoteCurrency := "USD"
		currencies := strings.Split(market, "-")

		if len(currencies) >= 2 {
			baseCurrency = currencies[0]
			quoteCurrency = currencies[1]
		} else {
			logrus.Warnf("UNFORMATED market_id = %s", market)
		}

		//We pay funding at the first minute of each hour, so we can just align it to 1 hour
		res2 := MarketResponse{
			BaseCurrency:             baseCurrency,
			QuoteCurrency:            quoteCurrency,
			ProductType:              model.DEFAULT_INSTRUMENT_PRODUCT_TYPE,
			NextFundingRateTimestamp: NextHourTimestamp(),
			OpenInterest:             OpenInterest,
			LongRatio:                LongRatio,
			ShortRatio:               ShortRatio,
		}

		func() {
			marketViewDailyMapMut.RLock()
			if m, ok := marketViewDaily[market]; ok {
				res2.AverageDailyVolume = m.averageDailyVolume
				res2.LastTradePrice24High = m.lastTradePrice24High
				res2.LastTradePrice24Low = m.lastTradePrice24Low
				res2.LastTradePrice24hChangePremium = m.lastTradePrice24hChangePremium
				res2.LastTradePrice24hChangeBasis = m.lastTradePrice24hChangeBasis
				res2.AverageDailyVolumeChangePremium = m.averageDailyVolumeChangePremium
				res2.AverageDailyVolumeChangeBasis = m.averageDailyVolumeChangeBasis
			}
			marketViewDailyMapMut.RUnlock()

			go func() {
				if !marketViewDailyFuncMut[marketIndex].TryLock() {
					return
				}
				defer marketViewDailyFuncMut[marketIndex].Unlock()

				sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
				sql, args, err := sqlBuilder.
					Select("COALESCE(average_daily_volume, 0)",
						"COALESCE(last_trade_price_24high, 0)",
						"COALESCE(last_trade_price_24low, 0)",
						"COALESCE(last_trade_price_24h_change_premium, 0)",
						"COALESCE(last_trade_price_24h_change_basis, 0)",
						"COALESCE(average_daily_volume_change_premium, 0)",
						"COALESCE(average_daily_volume_change_basis, 0)").
					From("market_data_view").
					Where("market_id = ?", market).
					Limit(1).
					ToSql()
				if err != nil {
					logrus.Error(errors.Wrap(err, "build daily sql"))
					return
				}

				var row marketViewDailyType
				err = db.QueryRow(c.Request.Context(), sql, args...).Scan(
					&row.averageDailyVolume,
					&row.lastTradePrice24High,
					&row.lastTradePrice24Low,
					&row.lastTradePrice24hChangePremium,
					&row.lastTradePrice24hChangeBasis,
					&row.averageDailyVolumeChangePremium,
					&row.averageDailyVolumeChangeBasis,
				)
				if err != nil && err != pgx.ErrNoRows {
					logrus.Error(errors.Wrap(err, "exec/scan daily sql"))
					return
				}

				marketViewDailyMapMut.Lock()
				marketViewDaily[market] = row
				marketViewDailyMapMut.Unlock()
			}()
		}()

		res := ExtendedMarketResponse{
			MarketData:     *res1,
			MarketResponse: res2,
		}

		//TODO: change SQL for that
		// If line_charts flag presents - return 50 hourly closed prices
		if request.LineCharts != nil && *request.LineCharts == true {
			res.LineChartsResponse = LineChartsResponse{
				ChartPrices: make([]decimal.Decimal, 0),
			}
			func() {
				marketViewPriceMapMut.RLock()
				if m, ok := marketViewPrice[market]; ok {
					res.LineChartsResponse.ChartPrices = m
				}
				marketViewPriceMapMut.RUnlock()

				go func() {
					if !marketViewPriceFuncMut[marketIndex].TryLock() {
						return
					}
					defer marketViewPriceFuncMut[marketIndex].Unlock()

					sqlBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
					sql, args, err := sqlBuilder.
						Select("COALESCE(price, 0)").
						From("market_last_trade_view").
						Where("market_id = ?", market).
						Where("time >= CURRENT_TIMESTAMP - INTERVAL '24 hours'").
						OrderBy("time ASC").
						Limit(24).
						ToSql()
					if err != nil {
						logrus.Error(errors.Wrap(err, "build prices sql"))
						return
					}

					rows, err := db.Query(c.Request.Context(), sql, args...)
					if err != nil && err != pgx.ErrNoRows {
						logrus.Error(errors.Wrap(err, "exec prices sql"))
						return
					}
					defer rows.Close()

					var prices []decimal.Decimal
					for rows.Next() {
						var price decimal.Decimal

						if err := rows.Scan(&price); err != nil {
							logrus.Error(errors.Wrap(err, "scan prices result"))
							return
						}

						prices = append(prices, price)
					}

					marketViewPriceMapMut.Lock()
					marketViewPrice[market] = prices
					marketViewPriceMapMut.Unlock()
				}()
			}()
		}

		response = append(response, res)
	}

	SuccessResponse(c, response...)
}

func HandleMarketTrades(c *gin.Context) {
	var request MarketTradesRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	trades, err := apiModel.GetTrades(c.Request.Context(), request.MarketId, ctx.Pagination.Limit)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, trades...)
}

func HandleMarketOrderBook(c *gin.Context) {
	var request MarketOrderBookRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	ctx := GetRabbitContext(c)
	apiModel := model.NewApiModel(ctx.Broker)

	orderBook, err := apiModel.GetOrderbookData(c.Request.Context(), request.MarketId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, orderBook)
}

func HandleMarketCoins(c *gin.Context) {
	var request MarketCoinsRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, err)
		return
	}

	cg := NewCoinGecko()
	r, err := cg.GetMarkets(request.CoinIds)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, r...)
}
