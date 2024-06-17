package signer

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type EIP712Encoder struct {
	domainTypes []apitypes.Type
	domainData  apitypes.TypedDataDomain
}

const (
	DOMAIN_KEY = "EIP712Domain"
)
 var (
	_eip191Header = []byte("\x19\x01")
 )

// can encode data for EIP 712 but does not handle signing of the encoded data
func NewEIP712Encoder(domainName, version, verifyingContract string, chainId *big.Int) *EIP712Encoder {

	includeVerifyingContract := common.IsHexAddress(verifyingContract)

	if includeVerifyingContract {
		return &EIP712Encoder{
			domainTypes: []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			domainData: apitypes.TypedDataDomain{
				Name:              domainName,
				Version:           version,
				ChainId:           (*math.HexOrDecimal256)(chainId),
				VerifyingContract: verifyingContract,
			},
		}
	} else {
		return &EIP712Encoder{
			domainTypes: []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			domainData: apitypes.TypedDataDomain{
				Name:    domainName,
				Version: version,
				ChainId: (*math.HexOrDecimal256)(chainId),
			},
		}
	}
}

func (en *EIP712Encoder) EncodeData(primaryType string, types apitypes.Types, message apitypes.TypedDataMessage) ([]byte, error) {

	if primaryType == "" {
		return nil, fmt.Errorf("invalid primaryType: %s", primaryType)
	} else if len(types) == 0 {
		return nil, errors.New("empty types")
	} else if len(message) == 0 {
		return nil, errors.New("empty message")
	}

	data := apitypes.TypedData{
		PrimaryType: primaryType,
		Types: apitypes.Types{
			DOMAIN_KEY: en.domainTypes,
		},
		Domain:  en.domainData,
		Message: message,
	}

	for k, v := range types {
		data.Types[k] = v
	}

	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, err
	}

	domainSeparator, err := data.HashStruct(DOMAIN_KEY, data.Domain.Map())
	if err != nil {
		return nil, err
	}

	rawData := append(_eip191Header, domainSeparator...)
	rawData = append(rawData, typedDataHash...)

	encodedDataHash := crypto.Keccak256Hash(rawData)
	return encodedDataHash.Bytes(), nil
}
