package slipstopper

import (
	"context"
	"sync"

	avl "github.com/emirpasic/gods/trees/avltree"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/strips-finance/rabbit-dex-backend/pkg/log"

	"github.com/strips-finance/rabbit-dex-backend/model"
)

type Matcher struct {
	tree           *Tree
	nodesByOrderId map[string]*avl.Node
	broker         *model.Broker
	mu             sync.Mutex
}

func NewMatcher() *Matcher {
	b, err := model.GetBroker()
	if err != nil {
		logrus.Fatalln(err)
	}

	return &Matcher{
		tree:           NewAVLTree(),
		nodesByOrderId: make(map[string]*avl.Node),
		broker:         b,
	}
}

func (m *Matcher) Insert(order model.OrderData) {
	node := m.tree.Insert(order.TriggerPrice.Decimal, order)
	m.nodesByOrderId[order.OrderId] = node
}

func (m *Matcher) Remove(order model.OrderData) {
	node, ok := m.nodesByOrderId[order.OrderId]
	if ok {
		key := node.Key
		tup := node.Value.(Tuple)

		m.tree.Delete(key)
		delete(m.nodesByOrderId, order.OrderId)
		for _, o := range tup.Values {
			if o.OrderId != order.OrderId {
				m.Insert(o)
			}
		}
	} else {
		logrus.Infof("Remove: no node found with orderId=%v", order.OrderId)
		return
	}
}

func (m *Matcher) OnPriceUpdate(price decimal.Decimal) {
	// when a price update happens, we need to see if internally any orders are within this range.
	// if any orders are found to be within the price range, then gather these orders and send them
	// to the matching engine in tarantool to be executed.
	apiModel := model.NewApiModel(m.broker)

	ordersLTE := m.GetByLTE(price)
	for _, item := range ordersLTE {
		for _, val := range item.Values {
			order := val
			if (order.OrderType == model.STOP_LOSS && order.Side == model.LONG) ||
				(order.OrderType == model.STOP_LOSS_LIMIT && order.Side == model.LONG) ||
				(order.OrderType == model.TAKE_PROFIT && order.Side == model.SHORT) ||
				(order.OrderType == model.TAKE_PROFIT_LIMIT && order.Side == model.SHORT) ||
				(order.OrderType == model.STOP_LIMIT && order.Side == model.LONG) ||
				(order.OrderType == model.STOP_MARKET && order.Side == model.LONG) {
				logrus.Infof("[lte] found order to execute: id=%s trigger_price=%s fair_price=%s side=%s", order.OrderId, order.TriggerPrice, price, order.Side)
				resp, err := apiModel.OrderExecute(context.Background(), order.ProfileID, order.MarketID, order.OrderId)
				if err != nil {
					logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error OrderExecute: err=%v", err)
				} else {
					logrus.Infof("Executed order with: id=%s resp=%v", order.OrderId, resp)
				}
			}
		}
	}

	ordersGTE := m.GetByGTE(price)
	for _, item := range ordersGTE {
		for _, val := range item.Values {
			order := val
			if (order.OrderType == model.STOP_LOSS && order.Side == model.SHORT) ||
				(order.OrderType == model.STOP_LOSS_LIMIT && order.Side == model.SHORT) ||
				(order.OrderType == model.TAKE_PROFIT && order.Side == model.LONG) ||
				(order.OrderType == model.TAKE_PROFIT_LIMIT && order.Side == model.LONG) ||
				(order.OrderType == model.STOP_LIMIT && order.Side == model.SHORT) ||
				(order.OrderType == model.STOP_MARKET && order.Side == model.SHORT) {
				logrus.Infof("[gte] found order to execute: id=%s trigger_price=%s fair_price=%s side=%s", order.OrderId, order.TriggerPrice, price, order.Side)
				resp, err := apiModel.OrderExecute(context.Background(), order.ProfileID, order.MarketID, order.OrderId)
				if err != nil {
					logrus.WithField(log.AlertTag, log.AlertHigh).Errorf("Error OrderExecute: err=%v", err)
				} else {
					logrus.Infof("Executed order with: id=%s resp=%v", order.OrderId, resp)
				}
			}
		}
	}
}

func (m *Matcher) Size() uint64 {
	return m.tree.GetSize()
}

func (m *Matcher) ClearAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodesByOrderId = make(map[string]*avl.Node)
	m.tree.ClearAll()
}

func (m *Matcher) GetByRange(min, max decimal.Decimal) []Tuple {
	result := m.tree.RangeSearch(min, max)

	return result
}

func (m *Matcher) GetByLTE(key decimal.Decimal) []Tuple {
	result := m.tree.LTE(key)

	return result
}

func (m *Matcher) GetByGTE(key decimal.Decimal) []Tuple {
	result := m.tree.GTE(key)

	return result
}
