package slipstopper

import (
	"golang.org/x/exp/maps"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

func TestMatcherFull(t *testing.T) {
	matcher := NewMatcher()

	price1, _ := tdecimal.NewDecimalFromString("1000.00")
	price2, _ := tdecimal.NewDecimalFromString("1100.00") // same price o2 and o3
	price3, _ := tdecimal.NewDecimalFromString("1100.00") // same price o2 and o3
	price4, _ := tdecimal.NewDecimalFromString("1300.00")

	o1 := model.OrderData{
		OrderId:      "1",
		OrderType:    "stop_loss",
		TriggerPrice: price1,
	}

	o2 := model.OrderData{
		OrderId:      "2",
		OrderType:    "stop_loss",
		TriggerPrice: price2,
	}

	o3 := model.OrderData{
		OrderId:      "3",
		OrderType:    "stop_loss",
		TriggerPrice: price3,
	}

	o4 := model.OrderData{
		OrderId:      "4",
		OrderType:    "stop_loss",
		TriggerPrice: price4,
	}

	matcher.Insert(o1)
	matcher.Insert(o2)
	matcher.Insert(o3)
	matcher.Insert(o4)

	// we have 2 orders with same pricing so should be 3 instead of 4
	assert.Equal(t, uint64(3), matcher.Size())

	// nodes however should be 4, since they are by order ids
	assert.Equal(t, 4, len(matcher.nodesByOrderId))
	keys := maps.Keys(matcher.nodesByOrderId)
	sort.Strings(keys)
	assert.Equal(t, keys, []string{"1", "2", "3", "4"})

	// check that the same price orders are linked inside the same node.
	orders2 := matcher.nodesByOrderId["2"].Value.(Tuple).Values
	assert.Equal(t, 2, len(orders2))
	assert.Equal(t, "2", orders2[0].OrderId)
	assert.Equal(t, "3", orders2[1].OrderId)

	orders3 := matcher.nodesByOrderId["3"].Value.(Tuple).Values
	assert.Equal(t, 2, len(orders3))
	assert.Equal(t, "2", orders3[0].OrderId)
	assert.Equal(t, "3", orders3[1].OrderId)

	// check node is same instance and not a copy
	assert.Equal(t, *matcher.nodesByOrderId["2"], *matcher.nodesByOrderId["3"])

	// let's now remove some items
	matcher.Remove(o2)
	assert.Equal(t, uint64(3), matcher.Size())
	assert.Equal(t, 3, len(matcher.nodesByOrderId))
	keys = maps.Keys(matcher.nodesByOrderId)
	sort.Strings(keys)
	assert.Equal(t, keys, []string{"1", "3", "4"})

	// remove one more element
	matcher.Remove(o3)
	assert.Equal(t, uint64(2), matcher.Size())
	assert.Equal(t, 2, len(matcher.nodesByOrderId))
	keys = maps.Keys(matcher.nodesByOrderId)
	sort.Strings(keys)
	assert.Equal(t, keys, []string{"1", "4"})

	// clear all
	matcher.ClearAll()
	assert.Equal(t, uint64(0), matcher.Size())
	assert.Equal(t, 0, len(matcher.nodesByOrderId))
}
