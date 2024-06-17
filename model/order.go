package model

import (
	"context"
	"errors"

	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

type Task struct {
	TaskId    int         `msgpack:"task_id"  json:"id"`
	Status    string      `msgpack:"status"  json:"status"`
	Data      interface{} `msgpack:"data"  json:"data"`
	OrderId   string      `msgpack:"order_id"  json:"order_id"`
	ProfileId string      `msgpack:"profile_id"  json:"profile_id"`
	OrderType string      `msgpack:"order_type"  json:"order_type"`
	IPriority int         `msgpack:"i_priority"  json:"i_priority"`
	OPriority int         `msgpack:"o_priority"  json:"o_priority"`
	Timestamp int64       `msgpack:"timestamp"  json:"timestamp"`
}

type OrderCreateRes struct {
	OrderId       string            `msgpack:"order_id"  json:"id"`
	MarketId      string            `msgpack:"market_id"  json:"market_id"`
	ProfileId     uint              `msgpack:"profile_id"  json:"profile_id"`
	Status        string            `msgpack:"status"  json:"status"`
	Size          *tdecimal.Decimal `msgpack:"order_size"  json:"size"`
	Price         *tdecimal.Decimal `msgpack:"order_price"  json:"price"`
	Side          string            `msgpack:"order_side"  json:"side"`
	Type          string            `msgpack:"order_type"  json:"type"`
	IsLiquidation bool              `msgpack:"is_liquidation"  json:"is_liquidation"`
	ClientOrderId *string           `msgpack:"client_order_id"  json:"client_order_id"`
	TriggerPrice  *tdecimal.Decimal `msgpack:"trigger_price"  json:"trigger_price"`
	SizePercent   *tdecimal.Decimal `msgpack:"size_percent"  json:"size_percent"`
	TimeInForce   *string           `msgpack:"time_in_force"  json:"time_in_force"`
}

type OrderExecuteRes struct {
	OrderId   string `msgpack:"order_id"  json:"id"`
	MarketId  string `msgpack:"market_id"  json:"market_id"`
	ProfileId uint   `msgpack:"profile_id"  json:"profile_id"`
}

type OrderCancelRes struct {
	OrderId       string `msgpack:"order_id"  json:"id"`
	MarketId      string `msgpack:"market_id"  json:"market_id"`
	ProfileId     uint   `msgpack:"profile_id"  json:"profile_id"`
	Status        string `msgpack:"status"  json:"status"`
	ClientOrderId string `msgpack:"client_order_id"  json:"client_order_id"`
}

type OrderAmendRes struct {
	OrderId      string            `msgpack:"order_id"  json:"id"`
	MarketId     string            `msgpack:"market_id"  json:"market_id"`
	ProfileId    uint              `msgpack:"profile_id"  json:"profile_id"`
	Status       string            `msgpack:"status"  json:"status"`
	Size         *tdecimal.Decimal `msgpack:"order_size"  json:"size,omitempty"`
	Price        *tdecimal.Decimal `msgpack:"order_price"  json:"price,omitempty"`
	TriggerPrice *tdecimal.Decimal `msgpack:"trigger_price"  json:"trigger_price,omitempty"`
	SizePercent  *tdecimal.Decimal `msgpack:"size_percent"  json:"size_percent,omitempty"`
}

type CancellAllRes struct {
	MarketId  string `msgpack:"market_id"  json:"market_id"`
	ProfileId uint   `msgpack:"profile_id"  json:"profile_id"`
	Status    string `msgpack:"status"  json:"status"`
}

type OrderResponse[T any] struct {
	Task  *Task  `msgpack:"task"`
	Order T      `msgpack:"order"`
	Error string `msgpack:"error"`
}

func (r OrderResponse[C]) request(ctx context.Context, instance string, broker *Broker, fn string, params []interface{}) (*Task, C, error) {
	var res []OrderResponse[C]
	err := broker.Execute(instance, ctx, fn, params, &res)
	if err != nil {
		return nil, *new(C), err
	}
	if len(res) == 0 {
		return nil, *new(C), errors.New("UNKNOWN ERROR")
	}

	if res[0].Error != "" {
		return nil, *new(C), errors.New(res[0].Error)
	}

	return res[0].Task, res[0].Order, nil
}
