package airdrop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testnetContract = "0x01cB633ACdbe00459D97F6213793fe69eBe275cd"
	testnetProvider = "https://eth-goerli.g.alchemy.com/v2/QySg3apqK5Qrom_jS8nCUMnY3zh0dXON"
)

func TestProvider(t *testing.T) {
	provider, err := NewAirdropProvider(testnetProvider)
	assert.NoError(t, err)
	assert.NotEmpty(t, provider)

	claimed, err := provider.ProcessedClaims(11)
	assert.NoError(t, err)
	assert.Equal(t, false, claimed)
}
