// +build run_pricing

// to run: go test -tags=go_tarantool_msgpack_v5,run_pricing -run TestPricingService

package pricing

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPricingService(t *testing.T) {
	pricingService := NewPricingService()
	dummyPriceReceiver := &DummyPriceReceiver{}
	config, err := ReadConfig()
	assert.NoError(t, err)
	godotenv.Load()
	pricingService.LoadConfig(context.Background(), dummyPriceReceiver, config)
	select {}
}

func TestPricingServiceWithTnt(t *testing.T) {
	pricingService := NewPricingService()
	config, err := ReadConfig()
	assert.NoError(t, err)
	pricingService.Run(context.Background(), config)
	select {}
}
