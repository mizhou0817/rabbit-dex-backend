package signer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type WithdrawalSigner struct {
	SignerEIP712
}

func NewWithdrawalSigner(domainName, exchangeAddress string, chainId *big.Int, kmsKeyId string) (signer *WithdrawalSigner, err error) {
	signer712, err := NewEIP712Signer(
		domainName,
		"1",
		exchangeAddress,
		chainId,
		kmsKeyId,
	)
	if err != nil {
		return nil, err
	}
	signer = &WithdrawalSigner{
		SignerEIP712: *signer712,
	}
	return signer, err
}

func (ws *WithdrawalSigner) Sign(id *big.Int, wallet string, amount *big.Int) (r []byte, s []byte, v uint, err error) {
	r, s, v, err = ws.KmsSignTypedData("withdrawal", apitypes.Types{
		"withdrawal": []apitypes.Type{
			{Name: "id", Type: "uint256"},
			{Name: "trader", Type: "address"},
			{Name: "amount", Type: "uint256"},
		},
	},
		apitypes.TypedDataMessage{
			"id":     id,
			"trader": wallet,
			"amount": amount,
		},
	)
	return r, s, v, err
}
