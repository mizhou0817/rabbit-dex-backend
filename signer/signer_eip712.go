package signer

import (
	"bytes"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type SignerEIP712 struct {
	keyID   string
	encoder EIP712Encoder
}

const (
	_vOffset = 27
)

// REFERENCE: https://github.com/welthee/go-ethereum-aws-kms-tx-signer/blob/main/signer.go

var (
	_secp256k1N     = crypto.S256().Params().N
	_secp256k1HalfN = new(big.Int).Div(_secp256k1N, big.NewInt(2))
)

type (
	asn1EcPublicKey struct {
		EcPublicKeyInfo asn1EcPublicKeyInfo
		PublicKey       asn1.BitString
	}

	asn1EcPublicKeyInfo struct {
		Algorithm  asn1.ObjectIdentifier
		Parameters asn1.ObjectIdentifier
	}

	asn1EcSig struct {
		R asn1.RawValue
		S asn1.RawValue
	}
)

// Signer can sign for EIP 712 with KMS using keyID
func NewEIP712Signer(domainName, version, verifyingContract string, chainId *big.Int, keyID string) (*SignerEIP712, error) {

	// Check that address is correct
	if !common.IsHexAddress(verifyingContract) {
		return nil, fmt.Errorf("invalid verifyingContract: %s", verifyingContract)
	}

	encoder := NewEIP712Encoder(domainName, version, verifyingContract, chainId)

	return &SignerEIP712{
		keyID:   keyID,
		encoder: *encoder,
	}, nil
}

// Equal to etherjs  signer._signTypedData
func (sn *SignerEIP712) KmsSignTypedData(primaryType string, types apitypes.Types, message apitypes.TypedDataMessage) (r []byte, s []byte, v uint, e error) {

	encodedData, err := sn.encoder.EncodeData(primaryType, types, message)
	if err != nil {
		return nil, nil, 0, err
	}

	kmsSignature, err := SignData(encodedData, sn.keyID, kms.MessageTypeDigest, kms.SigningAlgorithmSpecEcdsaSha256)
	if err != nil {
		return nil, nil, 0, err
	}

	//Transform signature to ETH compatible
	var sigAsn1 asn1EcSig
	_, err = asn1.Unmarshal(kmsSignature, &sigAsn1)
	if err != nil {
		return nil, nil, 0, err
	}

	// Adjust S value from signature according to Ethereum standard
	sBytes := sigAsn1.S.Bytes
	sBigInt := new(big.Int).SetBytes(sBytes)
	if sBigInt.Cmp(_secp256k1HalfN) > 0 {
		sBytes = new(big.Int).Sub(_secp256k1N, sBigInt).Bytes()
	}

	pubkeyBytes, err := getPublicKeyDerBytesFromKMS(sn.keyID)
	if err != nil {
		return nil, nil, 0, err
	}

	// fit signature to ETH standard
	rsSignature := append(adjustSignatureLength(sigAsn1.R.Bytes), adjustSignatureLength(sBytes)...)
	ethSignature := append(rsSignature, []byte{0}...)

	recoveredPublicKeyBytes, err := crypto.Ecrecover(encodedData, ethSignature)
	if err != nil {
		return nil, nil, 0, err
	}

	if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(pubkeyBytes) {
		ethSignature = append(rsSignature, []byte{1}...)
		recoveredPublicKeyBytes, err = crypto.Ecrecover(encodedData, ethSignature)
		if err != nil {
			return nil, nil, 0, err
		}

		if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(pubkeyBytes) {
			return nil, nil, 0, errors.New("can not reconstruct public key from sig")

		}
	}

	r = ethSignature[:32]
	s = ethSignature[32:64]
	v = uint(ethSignature[64]) + _vOffset

	return
}

func adjustSignatureLength(buffer []byte) []byte {
	buffer = bytes.TrimLeft(buffer, "\x00")
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}
	return buffer
}

func getPublicKeyDerBytesFromKMS(keyID string) ([]byte, error) {
	getPubKeyOutput, err := GetPublicKey(keyID)
	if err != nil {
		return nil, err
	}

	var asn1pubk asn1EcPublicKey
	_, err = asn1.Unmarshal(getPubKeyOutput.PublicKey, &asn1pubk)
	if err != nil {
		return nil, err
	}

	return asn1pubk.PublicKey.Bytes, nil
}
