package signer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/strips-finance/rabbit-dex-backend/tdecimal"
)

const (
	KEY_ID   = "c04c633d-84a6-4fd6-8148-53c1a41129e6"
	KEY_ID_2 = "e8b5dc35-e63d-47e9-bf7e-5cdc26174606"
)

var (
	TOKEN_DECIMALS                      = int32(18)
	Decimals_Multiplier decimal.Decimal = decimal.NewFromInt32(10).Pow(
		decimal.NewFromInt32(TOKEN_DECIMALS))
	ZeroBigInt = big.NewInt(0)
)

func TokenDecimalsToTDecimal(amount *big.Int) *tdecimal.Decimal {
	return tdecimal.NewDecimal(decimal.NewFromBigInt(amount, -TOKEN_DECIMALS))
}

func TDecimalToTokenDecimals(amount *tdecimal.Decimal) *big.Int {
	return amount.Mul(Decimals_Multiplier).Round(0).BigInt()
}

func TestSigner(t *testing.T) {

	/*
		createKeyOutput, err := CreateSignVerifyKey("test_key1")
		assert.NoError(t, err)
		keyID := *createKeyOutput.KeyMetadata.KeyId
		keyARN := *createKeyOutput.KeyMetadata.Arn
		fmt.Println("Key Id:", keyID)
		fmt.Println("Key ARN:", keyARN)
	*/
	name := "RabbitxAirdrop"
	version := "1"
	chainId := big.NewInt(int64(31337))
	verifyingContract := "0x9d0d84B2b061126E56D1c0ce6f9F14085812fAe1"

	signer, err := NewEIP712Signer(name, version, verifyingContract, chainId, KEY_ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, signer)

	id := big.NewInt(2)
	trader := "0xb95cDF9d692729aBA12F713F4222777c98D6eed1"
	amount := TDecimalToTokenDecimals(&tdecimal.Decimal{decimal.NewFromInt(int64(10))})
	address, _ := GetEthAddress(KEY_ID)
	logrus.Infof("key_id ETH address = %s", *address)

	/*
		INPUT:
		KEY_ID = c04c633d-84a6-4fd6-8148-53c1a41129e6
		eth address = 0x5D8f5F9b0DFdaf2D1b691b3bf6a1ceE2F6c7FDc9

		name := "RabbitxAirdrop"
		version := "1"
		chainId := big.NewInt(int64(31337))
		verifyingContract := "0x9d0d84B2b061126E56D1c0ce6f9F14085812fAe1"
		id := big.NewInt(2)
		trader := "0xb95cDF9d692729aBA12F713F4222777c98D6eed1"
		amount := TDecimalToTokenDecimals(&tdecimal.Decimal{decimal.NewFromInt(int64(10))})


		FOR debug purpose check that (inside KmsSignTypedData) as expected:
		* typedDataHash
		* domainSeparator
		* digest

		logrus.Infof("Expected typedDataHash = %s", "0xac030ba9969d91e52d78bb7872ee853fe3da9bea44c8669e53bf625fe1415904")
		logrus.Infof("mhash = %s", typedDataHash)

		logrus.Infof("expected domainSeparator = %s", "0xfada7ed6602d0a52e936eaac816e82b9bf026b1023b9e6bb200520055aa4002c")
		logrus.Infof("domainSeparator = %s", domainSeparator)

		logrus.Infof("expected digest = %s", "0xb6766e2d911806f16354e2896389c64c3a04e7e81f0618187388d965b421fbcf")
		logrus.Infof("digest = %s", encodedDataHash)

	*/

	r, s, v, err := signer.KmsSignTypedData("claim", apitypes.Types{
		"claim": []apitypes.Type{
			{Name: "id", Type: "uint"},
			{Name: "trader", Type: "address"},
			{Name: "amount", Type: "uint"},
		},
	},
		apitypes.TypedDataMessage{
			"id":     id,
			"trader": trader,
			"amount": amount,
		},
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, r)
	assert.NotEmpty(t, s)
	assert.NotEmpty(t, v)
	assert.Subset(t, []uint{27, 28}, []uint{v})

	logrus.Infof("r = 0x%x", r)
	logrus.Infof("s = 0x%x", s)
	logrus.Infof("v = %d", v)

	r, s, v, err = signer.KmsSignTypedData("claim", apitypes.Types{
		"claim": []apitypes.Type{
			{Name: "id", Type: "uint"},
			{Name: "trader", Type: "address"},
			{Name: "amount", Type: "uint"},
		},
	},
		apitypes.TypedDataMessage{
			"id":     id,
			"trader": trader,
			"amount": amount,
		},
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, r)
	assert.NotEmpty(t, s)
	assert.NotEmpty(t, v)
	assert.Subset(t, []uint{27, 28}, []uint{v})

	logrus.Infof("r = 0x%x", r)
	logrus.Infof("s = 0x%x", s)
	logrus.Infof("v = %d", v)

}
