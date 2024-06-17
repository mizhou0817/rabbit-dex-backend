package slipstopper

import (
	"errors"
	avl "github.com/emirpasic/gods/trees/avltree"
	"github.com/shopspring/decimal"
	"github.com/strips-finance/rabbit-dex-backend/model"
)

var ErrKeyNotFound = errors.New("key not found")

type Tuple struct {
	Key    decimal.Decimal
	Values []model.OrderData
}

func DecimalComparator(a, b any) int {
	aCasted := a.(decimal.Decimal)
	bCasted := b.(decimal.Decimal)

	return aCasted.Cmp(bCasted)
}

type Tree struct {
	tree *avl.Tree
}

func NewAVLTree() *Tree {
	avlTree := avl.NewWith(DecimalComparator)
	return &Tree{tree: avlTree}
}

func (t *Tree) Insert(k decimal.Decimal, v any) *avl.Node {
	var node *avl.Node
	node = t.tree.GetNode(k)
	if node != nil {
		tup := node.Value.(Tuple)
		tup.Values = append(tup.Values, v.(model.OrderData))
		node.Value = tup
	} else {
		values := make([]model.OrderData, 0)
		values = append(values, v.(model.OrderData))
		tuple := Tuple{k, values}
		t.tree.Put(k, tuple)
		node = t.tree.GetNode(k)
	}
	return node
}

func (t *Tree) Delete(key any) {
	t.tree.Remove(key)
}

func (t *Tree) ClearAll() {
	t.tree.Clear()
}

func (t *Tree) GetSize() uint64 {
	return uint64(t.tree.Size())
}

func (t *Tree) Get(k any) (*Tuple, error) {
	item, found := t.tree.Get(k)
	if !found {
		return nil, ErrKeyNotFound
	}

	tup := item.(Tuple)
	return &tup, nil
}

func (t *Tree) RangeSearch(key1, key2 any) []Tuple {
	var tuples []Tuple
	it := t.tree.Iterator()
	cmp := t.tree.Comparator

	for it.Next() {
		currentKey := it.Key()
		if cmp(key1, currentKey) <= 0 && cmp(key2, currentKey) >= 0 {
			tuples = append(tuples, it.Value().(Tuple))
		}
	}

	return tuples
}

func (t *Tree) LTE(key1 any) []Tuple {
	var tuples []Tuple
	it := t.tree.Iterator()
	cmp := t.tree.Comparator

	for it.Next() {
		currentKey := it.Key()
		if cmp(currentKey, key1) <= 0 {
			tuples = append(tuples, it.Value().(Tuple))
		} else {
			// bailout early
			break
		}
	}

	return tuples
}

func (t *Tree) GTE(key1 any) []Tuple {
	var tuples []Tuple
	it := t.tree.Iterator()
	cmp := t.tree.Comparator

	for it.Next() {
		currentKey := it.Key()

		if cmp(currentKey, key1) >= 0 {
			tuples = append(tuples, it.Value().(Tuple))
		}
	}

	return tuples
}
