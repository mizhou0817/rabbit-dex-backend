package tests

import (
	"context"
	"testing"

	// "github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/liqengine"
)
func TestVaultLiquidation(t *testing.T) {
	ctx := context.Background()
	assistant := &DummyAssistant{}
	ls := liqengine.NewLiquidationService(0, assistant)
	ls.ProcessLiquidations(ctx)
	assert.Equal(t, 1, len(assistant.VaultsSeen))
	assert.Equal(t, uint(123), assistant.VaultsSeen[0])
}