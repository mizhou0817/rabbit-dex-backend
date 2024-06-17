package model

import "context"

const (
	DEBUG_PUSH_ORDERBOOK = "test_debug.push_orderbook"
	DEBUG_PUSH_PROFILE   = "test_debug.push_profile"
	DEBUG_CREATE_TRADE   = "candles.add_all_periods"
)

type DebugModel struct {
	broker *Broker
}

func NewDebugModel(broker *Broker) *DebugModel {
	return &DebugModel{
		broker: broker,
	}
}

func (d *DebugModel) PushOrderbook(ctx context.Context, market_id string, sequence uint, bid, bid_size, ask, ask_size float64) (interface{}, error) {
	res, err := DataResponse[interface{}]{}.Request(ctx, MARKET_INSTANCE, d.broker, DEBUG_PUSH_ORDERBOOK, []interface{}{
		market_id,
		sequence,
		bid,
		bid_size,
		ask,
		ask_size,
	})

	return res, err
}

func (d *DebugModel) PushProfile(ctx context.Context, profile_id uint) (interface{}, error) {
	res, err := DataResponse[interface{}]{}.Request(ctx, PROFILE_INSTANCE, d.broker, DEBUG_PUSH_PROFILE, []interface{}{
		profile_id,
	})

	return res, err
}

func (d *DebugModel) CreateCandle(ctx context.Context, market_id string, tm int64, price, size float64) (interface{}, error) {
	tm_micro := int64(tm * 1000000)

	res, err := DataResponse[interface{}]{}.Request(ctx, market_id, d.broker, DEBUG_CREATE_TRADE, []interface{}{
		price,
		size,
		tm_micro,
	})

	return res, err
}
