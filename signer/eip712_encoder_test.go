package signer

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {

	name := "RabbitXId"
	version := "1"
	chainId := big.NewInt(int64(11155111))
	verifyingContract := ""

	encoder := NewEIP712Encoder(name, version, verifyingContract, chainId)
	assert.NotEmpty(t, encoder)

	message := "Welcome to RabbitX!\n\nClick to sign in and on-board your wallet for trading perpetuals.\n\nThis request will not trigger a blockchain transaction or cost any gas fees. This signature only proves you are the true owner of this wallet.\n\nBy signing this message you agree to the terms and conditions of the exchange."
	timestamp := big.NewInt(time.Now().Unix())

	msg, err := encoder.EncodeData(
		"signin",
		apitypes.Types{
			"signin": []apitypes.Type{
				{Name: "message", Type: "string"},
				{Name: "timestamp", Type: "uint256"},
			},
		},
		apitypes.TypedDataMessage{
			"message":   message,
			"timestamp": (*math.HexOrDecimal256)(timestamp),
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, len(msg), 32)
}
