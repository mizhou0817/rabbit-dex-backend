package auth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

/*
Can be returned from MetamaskVerify function call if
invalid signature provided so recovered address differs
from expected one.
*/
type InvalidMetamaskSignatureError struct {
	ExpectedWallet  common.Address
	RecoveredWallet common.Address
}

func (e InvalidMetamaskSignatureError) Error() string {
	return fmt.Sprintf(
		"invalid ecdsa signature from metamask (expected address: %s, recovered address: %s)",
		e.ExpectedWallet.String(),
		e.RecoveredWallet.String(),
	)
}

type InvalidPayloadSignatureError struct {
	Expected string
	Actual   string
}

func (e InvalidPayloadSignatureError) Error() string {
	return fmt.Sprintf(
		"invalid payload signature (expected: %s, actual: %s)",
		e.Expected,
		e.Actual,
	)
}

/*
Returned from MetamaskVerify function when signature
has been expired. This check performed to ensure malicious
user cannot reuse signature old signature.
*/
type ExpiredSignatureError struct {
	SignatureTimestamp int64
	CurrentTimestamp   int64
}

func (e ExpiredSignatureError) Error() string {
	return fmt.Sprintf(
		"metamask signature is expired (signature timestamp: %d, current timestamp: %d)",
		e.SignatureTimestamp,
		e.CurrentTimestamp,
	)
}

/*
Used NewPayload constructor to ensure all necessary keys
are presented in payload. For now, it is method and path keys.
*/
type MissingPayloadKeyError struct {
	MissedKey string
}

func (e MissingPayloadKeyError) Error() string {
	return fmt.Sprintf("payload is missing key: %s", e.MissedKey)
}

/*
Used to notify that wallet doesn't have required Role.
*/
type NotValidSignerError struct {
	Msg          string
	Wallet       string
	Vault        string
	RequiredRole uint
}

func (e NotValidSignerError) Error() string {
	return fmt.Sprintf("trader: %s doesn't have role: %d on vault: %s error: %s", e.Wallet, e.RequiredRole, e.Vault, e.Msg)
}
