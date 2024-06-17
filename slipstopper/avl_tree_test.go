package slipstopper

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTree_GetSize(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(3), "dupthree")
	tree.Insert(decimal.NewFromFloat(4), "four")
	tree.Insert(decimal.NewFromFloat(4), "dupfour")

	assert.Equal(t, uint64(4), tree.GetSize())
}

func TestTree_Get(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(3), "threedup")

	v, _ := tree.Get(decimal.NewFromFloat(1))
	assert.Equal(t, 1, len(v.Values))
	assert.Equal(t, []any{"one"}, v.Values)

	v, _ = tree.Get(decimal.NewFromFloat(2))
	assert.Equal(t, 1, len(v.Values))
	assert.Equal(t, []any{"two"}, v.Values)

	v, _ = tree.Get(decimal.NewFromFloat(3))
	assert.Equal(t, 2, len(v.Values))
	assert.Equal(t, []any{"three", "threedup"}, v.Values)

	v, err := tree.Get(decimal.NewFromFloat(4))
	assert.Nil(t, v)
	assert.Error(t, err)
}

func TestTree_RangeSearch(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(4), "four")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(2), "twodup")
	tree.Insert(decimal.NewFromFloat(5), "five")

	tuples := tree.RangeSearch(decimal.NewFromFloat(1), decimal.NewFromFloat(4))

	// 1 2 3 4
	assert.Len(t, tuples, 4)

	assert.Equal(t, decimal.NewFromFloat(1), tuples[0].Key)
	assert.Equal(t, []any{"one"}, tuples[0].Values)
	assert.Equal(t, decimal.NewFromFloat(2), tuples[1].Key)
	assert.Equal(t, []any{"two", "twodup"}, tuples[1].Values)
	assert.Equal(t, decimal.NewFromFloat(3), tuples[2].Key)
	assert.Equal(t, []any{"three"}, tuples[2].Values)
	assert.Equal(t, decimal.NewFromFloat(4), tuples[3].Key)
	assert.Equal(t, []any{"four"}, tuples[3].Values)
}

func TestTree_LTE(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(4), "four")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(2), "twodup")
	tree.Insert(decimal.NewFromFloat(5), "five")

	tuples := tree.LTE(decimal.NewFromFloat(3))

	// 1 2 3
	assert.Len(t, tuples, 3)

	assert.Equal(t, decimal.NewFromFloat(1), tuples[0].Key)
	assert.Equal(t, []any{"one"}, tuples[0].Values)
	assert.Equal(t, decimal.NewFromFloat(2), tuples[1].Key)
	assert.Equal(t, []any{"two", "twodup"}, tuples[1].Values)
	assert.Equal(t, decimal.NewFromFloat(3), tuples[2].Key)
	assert.Equal(t, []any{"three"}, tuples[2].Values)
}

func TestTree_GTE(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(4), "four")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(2), "twodup")
	tree.Insert(decimal.NewFromFloat(5), "five")

	tuples := tree.GTE(decimal.NewFromFloat(3))

	// 3 4 5
	assert.Len(t, tuples, 3)

	assert.Equal(t, decimal.NewFromFloat(3), tuples[0].Key)
	assert.Equal(t, []any{"three"}, tuples[0].Values)
	assert.Equal(t, decimal.NewFromFloat(4), tuples[1].Key)
	assert.Equal(t, []any{"four"}, tuples[1].Values)
	assert.Equal(t, decimal.NewFromFloat(5), tuples[2].Key)
	assert.Equal(t, []any{"five"}, tuples[2].Values)
}

func TestTree_Delete(t *testing.T) {
	tree := NewAVLTree()

	tree.Insert(decimal.NewFromFloat(1), "one")
	tree.Insert(decimal.NewFromFloat(2), "two")
	tree.Insert(decimal.NewFromFloat(2), "duptwo")
	tree.Insert(decimal.NewFromFloat(3), "three")
	tree.Insert(decimal.NewFromFloat(3), "dupthree")

	assert.Equal(t, uint64(3), tree.GetSize())

	tree.Delete(decimal.NewFromFloat(2))
	assert.Equal(t, uint64(2), tree.GetSize())

	// try again to delete with key 2
	tree.Delete(decimal.NewFromFloat(2))
	assert.Equal(t, uint64(2), tree.GetSize())
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(m.HeapAlloc))
	fmt.Printf("\tHeapInuse = %v MiB", bToMb(m.HeapInuse))
	fmt.Println()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func XTestMemory(t *testing.T) {
	tree := NewAVLTree()
	for {
		for i := 0; i < 1000000; i++ {
			tree.Insert(decimal.NewFromInt(int64(i)), "foobar")
			tree.Delete(decimal.NewFromInt(int64(i)))
		}
		time.Sleep(3 * time.Second)
		PrintMemUsage()
	}
}
