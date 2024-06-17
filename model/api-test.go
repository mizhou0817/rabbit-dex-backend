package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	TEST_UPDATE_FAIR_PRICE      = "fortest.update_fair_price"
	TEST_GET_FILLS              = "fortest.get_fills"
	TEST_GET_TRADES             = "fortest.get_trades"
	TEST_GET_ORDER_BY_ID        = "fortest.get_order_by_id"
	TEST_GET_ALL_ORDERS         = "fortest.get_all_orders"
	TEST_ENGINE_UPDATE_PROFILES = "periodics.update_profiles"
	TEST_PROFILE_POLL_META      = "periodics.update_profiles_meta"
	TEST_POLL_EXCHANGE_DATA     = "periodics.update_exchange_data"

	TEST_WHITE_LIST = "internal.white_list"

	TEST_STATS_SHOW_ALL = "stats.show_all"
	DEPOSIT_CREDIT      = "balance.deposit_credit"
	WITHDRAW_CREDIT     = "profile.withdraw_credit"
	TEST_REPLACE_OPS    = "balance.test_replace_ops"
)

func (api *ApiModel) TestPollExchangeData(ctx context.Context) error {
	return EmptyResponse{}.Request(ctx, PROFILE_INSTANCE, api.broker, TEST_POLL_EXCHANGE_DATA, []interface{}{})
}

func (api *ApiModel) TestPollProfileMeta(ctx context.Context) error {
	return EmptyResponse{}.Request(ctx, PROFILE_INSTANCE, api.broker, TEST_PROFILE_POLL_META, []interface{}{})
}

func (api *ApiModel) TestUpdateProfiles(ctx context.Context, market_id string, profile_ids []uint) error {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s", err.Error())
		return errors.New(text)
	}

	err = EmptyResponse{}.Request(ctx, instance.Title, api.broker, TEST_ENGINE_UPDATE_PROFILES, []interface{}{
		profile_ids,
	})

	return err
}

func (api *ApiModel) GetOrderById(ctx context.Context, market_id string, order_id string) (*OrderData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s", err.Error())
		return nil, errors.New(text)
	}

	order, err := DataResponse[*OrderData]{}.Request(ctx, instance.Title, api.broker, TEST_GET_ORDER_BY_ID, []interface{}{
		order_id,
	})

	return order, err
}

func (api *ApiModel) TestUpdateFairPrice(ctx context.Context, market_id string, fair_price float64) (*tdecimal.Decimal, error) {
	d_fair_price := tdecimal.NewDecimal(decimal.NewFromFloat(fair_price))

	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		text := fmt.Sprintf("GetInstance err=%s", err.Error())
		return nil, errors.New(text)
	}

	new_fair_price, err := DataResponse[*tdecimal.Decimal]{}.Request(ctx, instance.Title, api.broker, TEST_UPDATE_FAIR_PRICE, []interface{}{
		market_id,
		d_fair_price,
	})

	return new_fair_price, err
}

func (api *ApiModel) OrderCreateCustomId(ctx context.Context, profile_id uint, market_id, order_type, side string, price, size, trigger_price, size_percent *float64, time_in_force *string, custom_id string) (OrderCreateRes, error) {
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
		"",
		d_trigger_price,
		d_size_percent,
		time_in_force,
		custom_id,
		new(MatchingMeta),
	})

	return res, err
}

func (api *ApiModel) GetAllFills(ctx context.Context, market_ids []string) ([]*FillData, []*TradeData, error) {
	fill_data := make([]*FillData, 0)
	trade_data := make([]*TradeData, 0)

	logrus.Info(market_ids)
	for _, market_id := range market_ids {
		instance, err := GetInstance().ByMarketID(market_id)
		if err != nil {
			text := fmt.Sprintf("GetInstance market_id=%s err=%s", market_id, err.Error())
			logrus.Error(text)
			return nil, nil, errors.New(text)
		}

		fills, err := DataResponse[[]*FillData]{}.Request(ctx, instance.Title, api.broker, TEST_GET_FILLS, []interface{}{})
		if err != nil {
			return nil, nil, err
		}
		fill_data = append(fill_data, fills...)

		trades, err := DataResponse[[]*TradeData]{}.Request(ctx, instance.Title, api.broker, TEST_GET_TRADES, []interface{}{})
		if err != nil {
			return nil, nil, err
		}
		trade_data = append(trade_data, trades...)
	}

	return fill_data, trade_data, nil
}

func (api *ApiModel) TestGetAllOrders(ctx context.Context, market_ids []string) ([]*OrderData, error) {
	order_data := make([]*OrderData, 0)

	///api.broker.Pool
	for _, market_id := range market_ids {
		instance, err := GetInstance().ByMarketID(market_id)
		if err != nil {
			text := fmt.Sprintf("GetInstance market_id=%s err=%s", market_id, err.Error())
			logrus.Error(text)
			return nil, errors.New(text)
		}

		orders, err := DataResponse[[]*OrderData]{}.Request(ctx, instance.Title, api.broker, TEST_GET_ALL_ORDERS, []interface{}{})
		if err != nil {
			return nil, err
		}
		order_data = append(order_data, orders...)
	}

	return order_data, nil
}

func (api *ApiModel) GetUntypedOrderbookData(ctx context.Context, market_id string) (*UntypedOrderbookData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[*UntypedOrderbookData]{}.Request(ctx, instance.Title, api.broker, GET_ORDERBOOK_DATA, []interface{}{
		market_id,
	})

	return data, err
}

func (api *ApiModel) GetMarketStats(ctx context.Context, market_id string) ([]StatsData, error) {
	instance, err := GetInstance().ByMarketID(market_id)
	if err != nil {
		return nil, err
	}

	data, err := DataResponse[[]StatsData]{}.Request(ctx, instance.Title, api.broker, TEST_STATS_SHOW_ALL, []interface{}{
		10,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (api *ApiModel) GetApiStats(ctx context.Context) ([]StatsData, error) {
	data, err := DataResponse[[]StatsData]{}.Request(ctx, API_INSTANCE, api.broker, TEST_STATS_SHOW_ALL, []interface{}{
		10,
	})
	if err != nil {
		return nil, err
	}

	return data, nil

}

func (api *ApiModel) WhiteListProfile(ctx context.Context, profile_id uint) interface{} {
	res, err := DataResponse[interface{}]{}.Request(ctx, API_INSTANCE, api.broker, TEST_WHITE_LIST, []interface{}{
		profile_id,
	})
	if err != nil {
		logrus.Fatalf("error=%s", err.Error())
	}

	return res
}

func (api *ApiModel) DepositCredit(ctx context.Context, profileId uint, amount float64) (*BalanceOps, error) {
	d_amount := tdecimal.NewDecimal(decimal.NewFromFloat(amount))
	ops, err := DataResponse[*BalanceOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, DEPOSIT_CREDIT, []interface{}{
		profileId,
		d_amount,
	})

	return ops, err
}

func (api *ApiModel) WithdrawCredit(ctx context.Context, profileId uint, amount float64) (*BalanceOps, error) {
	d_amount := tdecimal.NewDecimal(decimal.NewFromFloat(amount))

	ops, err := DataResponse[*BalanceOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, WITHDRAW_CREDIT, []interface{}{
		profileId,
		d_amount,
	})

	return ops, err
}

func (api *ApiModel) TestReplaceOps(ctx context.Context, ops *BalanceOps) error {

	_, err := DataResponse[*BalanceOps]{}.Request(ctx, PROFILE_INSTANCE, api.broker, TEST_REPLACE_OPS, []interface{}{
		ops,
	})

	return err
}
