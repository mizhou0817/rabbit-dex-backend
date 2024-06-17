package settlement

import (
	"context"
	"math/big"
	"testing"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const BFX_EXCHANGE_ADDRESS = "0x0E4A0e095dDb34158D705C3094c9Fefd7dA40bF6"

func TestYieldClaim(t *testing.T) {
	logrus.Printf("TestYieldClaim")
	ctx := context.Background()

	config, err := ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	apiModel := &MockApiModel{}

	settlementService, err := ConstructSettlementService(
		config.Handlers, apiModel,
	)

	if err != nil {
		logrus.Fatal(err)
	}

	handler := settlementService.handlers[BFX_EXCHANGE_ADDRESS]
	logrus.Printf("ethereum_handler: %+v", handler)
	err = handler.ethereumHandler.Dial()
	if err != nil {
		logrus.Fatal(err)
	}
	// var txhash common.Hash
	// txhash, err = handler.ethereumHandler.TestYieldClaim()
	// logrus.Printf("txhash: %v", txhash)
	// if err != nil {
	// 	logrus.Printf("TestYieldClaim err %s", err.Error())
	// }
	err = handler.ethereumHandler.ClaimYield(ctx)
	assert.NoError(t, err)
	if err != nil {
		logrus.Printf("MakeYieldClaim err %s", err.Error())
	}
}

func TestYieldDistribution(t *testing.T) {
	logrus.Printf("TestYieldDistribution")
	ctx := context.Background()

	config, err := ReadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	apiModel := &MockApiModel{}

	settlementService, err := ConstructSettlementService(
		config.Handlers, apiModel,
	)

	if err != nil {
		logrus.Fatal(err)
	}

	handler := settlementService.handlers[BFX_EXCHANGE_ADDRESS]
	logrus.Printf("ethereum_handler: %+v", handler)
	err = handler.ethereumHandler.Dial()
	if err != nil {
		logrus.Fatal(err)
	}
	fromBlock := big.NewInt(3214752)
	toBlock := big.NewInt(3214754)
	handler.ethereumHandler.ProcessYieldEvents(ctx, fromBlock, toBlock)
	assert.Equal(t, 1, len(apiModel.ProcessedYield), "expected 1 yield event to be processed")
	assert.Equal(t, "11", apiModel.ProcessedYield[0].Amount.String(), "expected yield amount to be 11")
}
