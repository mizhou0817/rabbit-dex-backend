package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
	"golang.org/x/exp/slices"
)

const (
	CREATE_PROFILE                        = "profile.create"
	GET_PROFILE_BY_ID                     = "profile.get"
	GET_PROFILE_BY_WALLET_FOR_EXCHANGE_ID = "profile.get_by_wallet_for_exchange_id"
	PROFILE_NOT_FOUND                     = "PROFILE_NOT_FOUND"
	PAY_FUNDING                           = "engine.pay_funding"

	UPDATE_INDEX_PRICE        = "market.update_index_price"
	GET_PROFILE_DATA          = "getters.get_profile_data"
	GET_EXTENDED_PROFILE_DATA = "getters.get_extended_profile_data"
	GET_EXTENDED_PROFILES     = "getters.get_extended_profiles"
	GET_OPEN_POSITIONS        = "getters.get_open_positions"
	GET_REQUESTED_UNSTAKES    = "getters.get_requested_unstakes"
	GET_ALL_ACTIVE_POSITIONS  = "position.get_all_active_positions"
	CANCEL_ALL                = "public.cancel_all"

	GET_ALL_ORDERS  = "order.get_all_orders"
	GET_ALL_ORDERS2 = "order.get_all_orders2"

	GET_OPEN_ORDERS = "getters.get_open_orders"
	UPDATE_LEVERGAE = "profile.update_leverage"

	GET_PROFILES_META_AFTER_TS = "profile.get_profiles_meta_after_ts"

	GET_MARKET_DATA  = "market.get_market_data"
	GET_FUNDING_META = "market.get_funding_meta"

	GET_ORDERBOOK_DATA = "engine.get_orderbook_data"
	GET_TRADE_DATA     = "trade.get_trade_data"

	ORDER_CREATE  = "public.new_order"
	ORDER_AMEND   = "public.amend_order"
	ORDER_CANCEL  = "public.cancel_order"
	ORDER_EXECUTE = "internal.execute_order"

	GET_CANDLES       = "candles.get_candles"
	GET_EXCHANGE_DATA = "getters.get_exchange_data"

	GET_PROFILE_CACHE           = "cache.get_cache"
	INVALIDATE_CACHE            = "cache.invalidate_cache"
	INVALIDATE_CACHE_AND_NOTIFY = "cache.invalidate_cache_and_notify"

	UPDATE_PROFILES_CACHES_AND_METAS = "setters.update_profiles_caches_and_metas"

	ARCHIVER_GET_NEXT_BATCH = "archiver.get_next_batch"

	WRITE_FRONTEND_STORAGE = "write_frontend_storage"
	READ_FRONTEND_STORAGE  = "read_frontend_storage"
)

type NextBatchResponse struct {
	Columns   []ColumnDescriptor `msgpack:"columns"`
	Data      []any              `msgpack:"data"`
	Timestamp uint64             `msgpack:"timestamp"`
}

type ColumnDescriptor struct {
	Name string `msgpack:"name"`
	Type string `msgpack:"type"`
}

type FundingMeta struct {
	MarketID   string            `msgpack:"market_id"`
	LastUpdate int64             `msgpack:"last_update"`
	TotalLong  *tdecimal.Decimal `msgpack:"total_long"`
	TotalShort *tdecimal.Decimal `msgpack:"total_short"`
}

type FundingPayment struct {
	MarketId      string
	ProfileId     uint
	FundingAmount *tdecimal.Decimal
}

type ApiModel struct {
	broker *Broker
}

func NewApiModel(broker *Broker) *ApiModel {
	return &ApiModel{
		broker: broker,
	}
}

func (api *ApiModel) WriteFrontendStorage(ctx context.Context, profile_id uint, data []byte) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, AUTH_INSTANCE, api.broker, WRITE_FRONTEND_STORAGE, []interface{}{
		profile_id,
		data,
	})

	return err
}

func (api *ApiModel) ReadFrontendStorage(ctx context.Context, profile_id uint) ([]byte, error) {
	res, err := DataResponse[[]byte]{}.Request(ctx, AUTH_INSTANCE, api.broker, READ_FRONTEND_STORAGE, []interface{}{
		profile_id,
	})

	return res, err
}

func (api *ApiModel) CancelAll(ctx context.Context, profile_id uint, is_liquidation bool) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, API_INSTANCE, api.broker, CANCEL_ALL, []interface{}{
		profile_id,
		is_liquidation,
	})

	return err
}

func (api *ApiModel) OrderCancel(ctx context.Context, profile_id uint, market_id, order_id, client_order_id string) (OrderCancelRes, error) {
	_, res, err := OrderResponse[OrderCancelRes]{}.request(ctx, API_INSTANCE, api.broker, ORDER_CANCEL, []interface{}{
		profile_id,
		market_id,
		order_id,
		client_order_id,
	})

	return res, err
}

func (api *ApiModel) OrderAmend(ctx context.Context, profile_id uint, market_id, order_id string, new_price, new_size, new_trigger_price, new_size_percent *float64) (OrderAmendRes, error) {
	var d_price *tdecimal.Decimal
	if new_price != nil {
		d_price = tdecimal.NewDecimal(decimal.NewFromFloat(*new_price))
	}
	var d_size *tdecimal.Decimal
	if new_size != nil {
		d_size = tdecimal.NewDecimal(decimal.NewFromFloat(*new_size))
	}
	var d_trigger_price *tdecimal.Decimal
	if new_trigger_price != nil {
		d_trigger_price = tdecimal.NewDecimal(decimal.NewFromFloat(*new_trigger_price))
	}
	var d_size_percent *tdecimal.Decimal
	if new_size_percent != nil {
		d_size_percent = tdecimal.NewDecimal(decimal.NewFromFloat(*new_size_percent))
	}

	_, res, err := OrderResponse[OrderAmendRes]{}.request(ctx, API_INSTANCE, api.broker, ORDER_AMEND, []interface{}{
		profile_id,
		market_id,
		order_id,
		d_price,
		d_size,
		d_trigger_price,
		d_size_percent,
	})

	return res, err
}

func (api *ApiModel) OrderCreate(ctx context.Context, profile_id uint, market_id, order_type, side string, price, size *float64, client_order_id *string, trigger_price, size_percent *float64, time_in_force *string, meta *MatchingMeta) (OrderCreateRes, error) {
	var d_price *tdecimal.Decimal
	if price != nil {
		d_price = tdecimal.NewDecimal(decimal.NewFromFloat(*price))
	}
	var d_size *tdecimal.Decimal
	if size != nil {
		d_size = tdecimal.NewDecimal(decimal.NewFromFloat(*size))
	}
	var d_trigger_price *tdecimal.Decimal
	if trigger_price != nil {
		d_trigger_price = tdecimal.NewDecimal(decimal.NewFromFloat(*trigger_price))
	}
	var d_size_percent *tdecimal.Decimal
	if size_percent != nil {
		d_size_percent = tdecimal.NewDecimal(decimal.NewFromFloat(*size_percent))
	}

	_, res, err := OrderResponse[OrderCreateRes]{}.request(ctx, API_INSTANCE, api.broker, ORDER_CREATE, []interface{}{
		profile_id,
		market_id,
		order_type,
		side,
		d_price,
		d_size,
		client_order_id,
		d_trigger_price,
		d_size_percent,
		time_in_force,
		nil, // custom_order_id

		meta,
	})

	return res, err
}

func (api *ApiModel) OrderExecute(ctx context.Context, profile_id uint, market_id, order_id string) (OrderExecuteRes, error) {
	_, res, err := OrderResponse[OrderExecuteRes]{}.request(ctx, API_INSTANCE, api.broker, ORDER_EXECUTE, []interface{}{
		profile_id,
		market_id,
		order_id,
	})

	return res, err
}

func (api *ApiModel) CreateProfile(ctx context.Context, profile_type, wallet, exchange_id string) (*Profile, error) {
	if !slices.Contains(supportedProfileTypes, profile_type) {
		return nil, fmt.Errorf("unsupported profile type = %s", profile_type)
	}

	profile, err := DataResponse[*Profile]{}.Request(ctx, PROFILE_INSTANCE, api.broker, CREATE_PROFILE, []interface{}{
		profile_type,
		PROFILE_STATUS_ACTIVE,
		wallet,
		exchange_id,
	})

	return profile, err
}

func (api *ApiModel) GetProfileById(ctx context.Context, profile_id uint) (*Profile, error) {
	profile, err := DataResponse[*Profile]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_PROFILE_BY_ID, []interface{}{
		profile_id,
	})

	return profile, err
}

func (api *ApiModel) GetProfileByWalletForExchangeId(ctx context.Context, wallet, exchange_id string) (*Profile, error) {
	profile, err := DataResponse[*Profile]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_PROFILE_BY_WALLET_FOR_EXCHANGE_ID, []interface{}{
		wallet,
		exchange_id,
	})

	return profile, err
}

func (api *ApiModel) UpdateIndexPrice(ctx context.Context, market_id string, index_price float64) error {
	d_index_price := tdecimal.NewDecimal(decimal.NewFromFloat(index_price))

	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s for market_id=%s", err.Error(), market_id)
		return errors.New(text)
	}

	_, err = DataResponse[*tdecimal.Decimal]{}.Request(ctx, instance.Title, api.broker, UPDATE_INDEX_PRICE, []interface{}{
		market_id,
		d_index_price,
	})

	return err
}
func (api *ApiModel) GetMarketData(ctx context.Context, market_id string) (*MarketData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[*MarketData]{}.Request(ctx, instance.Title, api.broker, GET_MARKET_DATA, []interface{}{
		market_id,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (api *ApiModel) GetFundingMeta(ctx context.Context, market_id string) (*FundingMeta, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[*FundingMeta]{}.Request(ctx, instance.Title, api.broker, GET_FUNDING_META, []interface{}{
		market_id,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (api *ApiModel) PayFunding(ctx context.Context,
	market_id string,
	fundingPayments []FundingPayment,
	LastFundingUpdateTime int64,
	totalLong, totalShort float64) error {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return err
	}

	d_total_long := tdecimal.NewDecimal(decimal.NewFromFloat(totalLong))
	d_total_short := tdecimal.NewDecimal(decimal.NewFromFloat(totalShort))
	_, err = DataResponse[interface{}]{}.Request(ctx, instance.Title, api.broker, PAY_FUNDING, []interface{}{
		market_id,
		fundingPayments,
		LastFundingUpdateTime,
		d_total_long,
		d_total_short,
	})

	return err
}

func (api *ApiModel) GetProfileData(ctx context.Context, profile_id uint) (*ProfileData, error) {
	// cannot be ReadOnly() because cache can be updated inside tarantool
	data, err := DataResponse[*ProfileData]{}.Request(ctx, PROFILE_INSTANCE, api.broker, GET_PROFILE_DATA, []interface{}{
		profile_id,
	})

	return data, err
}

func (api *ApiModel) GetExtendedProfileData(ctx context.Context, profile_id uint) (*ExtendedProfileData, error) {
	// cannot be ReadOnly() because cache can be updated inside tarantool
	data, err := DataResponse[*ExtendedProfileData]{}.Request(ctx, PROFILE_INSTANCE, api.broker, GET_EXTENDED_PROFILE_DATA, []interface{}{
		profile_id,
	})

	return data, err
}

func (api *ApiModel) GetOrderbookData(ctx context.Context, market_id string) (*OrderbookData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[*OrderbookData]{}.Request(ctx, instance.Title, api.broker, GET_ORDERBOOK_DATA, []interface{}{
		market_id,
	})

	return data, err
}

func (api *ApiModel) GetTrades(ctx context.Context, marketId string, limit int64) ([]*TradeData, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[[]*TradeData]{}.Request(ctx, instance.Title, api.broker, GET_TRADE_DATA, []interface{}{
		limit,
	})

	return data, err
}

func (api *ApiModel) GetAllActivePositions(ctx context.Context, market_id string, offset, limit uint) ([]*PositionData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[[]*PositionData]{}.Request(ctx, instance.Title, api.broker, GET_ALL_ACTIVE_POSITIONS, []interface{}{
		market_id,
		offset,
		limit,
	})

	return data, err
}

func (api *ApiModel) GetOpenPositions(ctx context.Context, profileId uint) ([]*PositionData, error) {
	data, err := DataResponse[[]*PositionData]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_OPEN_POSITIONS, []interface{}{
		profileId,
	})

	return data, err
}

func (api *ApiModel) GetRequestedUnstakes(ctx context.Context, profileId uint) ([]*BalanceOps, error) {
	data, err := DataResponse[[]*BalanceOps]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_REQUESTED_UNSTAKES, []interface{}{
		profileId,
	})

	return data, err
}

func (api *ApiModel) GetAllOrders(ctx context.Context, profile_id uint, marketID string, limit uint) ([]*OrderData, error) {
	instance, err := GetInstance().ByMarketID(marketID)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[[]*OrderData]{}.Request(ctx, ReadOnly(instance.Title), api.broker, GET_ALL_ORDERS, []interface{}{
		profile_id,
		limit,
	})

	return data, err
}

/*
	 FIXME:
		1) used in tests only
		2) actual <profile/getters.get_open_orders> tarantool func has just <profileId> argument and iterates over all available markets, so <marketId> is ignored
*/
func (api *ApiModel) GetOpenOrders(ctx context.Context, marketId string, profileId uint) ([]*OrderData, error) {
	data, err := DataResponse[[]*OrderData]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_OPEN_ORDERS, []interface{}{
		profileId,
		marketId,
	})

	return data, err
}

func (api *ApiModel) GetPlacedOrders(ctx context.Context, marketID string, profileID *uint) ([]*OrderData, error) {
	instance, err := GetInstance().ByMarketID(marketID)
	if err != nil {
		return nil, err
	}

	return DataResponse[[]*OrderData]{}.Request(ctx, ReadOnly(instance.Title), api.broker, GET_ALL_ORDERS2, []interface{}{
		profileID,
		PLACED,
		nil, // all order types
	})
}

func (api *ApiModel) GetCandles(ctx context.Context, market_id string, period uint, time_from, time_to int64) ([]*CandleData, error) {
	data, err := DataResponse[[]*CandleData]{}.Request(ctx, ReadOnly(market_id), api.broker, GET_CANDLES, []interface{}{
		period,
		time_from,
		time_to,
	})

	return data, err
}

func (api *ApiModel) UpdateLeverage(ctx context.Context, market_id string, profile_id uint, leverage uint) (*tdecimal.Decimal, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	d_leverage := tdecimal.NewDecimal(decimal.NewFromInt(int64(leverage)))
	new_leverage, err := DataResponse[*tdecimal.Decimal]{}.Request(ctx, instance.Title, api.broker, UPDATE_LEVERGAE, []interface{}{
		profile_id,
		market_id,
		d_leverage,
	})

	if err == nil {
		_, err = DataResponse[*ProfileCache]{}.Request(ctx, PROFILE_INSTANCE, api.broker, INVALIDATE_CACHE_AND_NOTIFY, []interface{}{
			profile_id,
		})

		if err != nil {
			return nil, err
		}
	}

	return new_leverage, err
}

func (api *ApiModel) GetExchangeData(ctx context.Context) (*ExchangeData, error) {
	data, err := DataResponse[*ExchangeData]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_EXCHANGE_DATA, []interface{}{})

	return data, err
}

func (api *ApiModel) GetProfileCache(ctx context.Context, profile_id uint) (*ProfileCache, error) {
	data, err := DataResponse[*ProfileCache]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_PROFILE_CACHE, []interface{}{
		profile_id,
	})

	return data, err
}

func (api *ApiModel) InvalidateCache(ctx context.Context, profileId uint) (*ProfileCache, error) {
	data, err := DataResponse[*ProfileCache]{}.Request(ctx, PROFILE_INSTANCE, api.broker, INVALIDATE_CACHE, []interface{}{
		profileId,
	})

	return data, err
}

func (api *ApiModel) InvalidateCacheAndNotify(ctx context.Context, profileId uint) (*ProfileCache, error) {
	data, err := DataResponse[*ProfileCache]{}.Request(ctx, PROFILE_INSTANCE, api.broker, INVALIDATE_CACHE_AND_NOTIFY, []interface{}{
		profileId,
	})

	return data, err
}

func (api *ApiModel) GetProfilesMetaAfterTs(ctx context.Context, marketId string, unixMicroTs int64) ([]*ProfileMeta, error) {
	instance, err := GetInstance().ByMarketID(marketId)
	if err != nil {
		return nil, err
	}

	return DataResponse[[]*ProfileMeta]{}.Request(ctx, ReadOnly(instance.Title), api.broker, GET_PROFILES_META_AFTER_TS, []interface{}{
		float64(unixMicroTs),
	})
}

func (api *ApiModel) GetExtendedProfiles(ctx context.Context, profilesIds ...uint) ([]*ExtendedProfile, error) {
	return DataResponse[[]*ExtendedProfile]{}.Request(ctx, ReadOnly(PROFILE_INSTANCE), api.broker, GET_EXTENDED_PROFILES, []interface{}{
		profilesIds,
	})
}

func (api *ApiModel) UpdateProfilesCachesAndMetas(ctx context.Context, data []*ProfileCacheMetas) error {
	_, err := DataResponse[interface{}]{}.Request(ctx, PROFILE_INSTANCE, api.broker, UPDATE_PROFILES_CACHES_AND_METAS, []interface{}{
		data,
	})

	return err
}

func (api *ApiModel) ArchiverGetNextBatch(ctx context.Context, instance string, space string, fromArchiveId uint64, limit uint64) (*NextBatchResponse, error) {
	data, err := DataResponse[*NextBatchResponse]{}.Request(ctx, instance, api.broker, ARCHIVER_GET_NEXT_BATCH, []interface{}{
		space,
		fromArchiveId,
		limit,
	})

	if err != nil {
		return nil, err
	}

	return data, err
}
