package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

func TestNewPayload_error(t *testing.T) {
	// Check constructor ensures payload contains all necessary keys
	malformedData := map[string]string{"hello": "world"}
	payload, err := NewPayload(0, malformedData)
	require.ErrorIs(t, err, MissingPayloadKeyError{MissedKey: PayloadKeyMethod})
	require.Nil(t, payload)

	malformedData[PayloadKeyMethod] = http.MethodPost
	payload, err = NewPayload(0, malformedData)
	require.ErrorIs(t, err, MissingPayloadKeyError{MissedKey: PayloadKeyPath})
	require.Nil(t, payload)
}

func TestNewPayload(t *testing.T) {
	// Now data contains all required keys
	data := map[string]string{}
	data[PayloadKeyMethod] = http.MethodPost
	data[PayloadKeyPath] = "/"

	payload, err := NewPayload(0, data)
	require.NoError(t, err)
	require.NotNil(t, payload)
}

func TestPayload_Hash(t *testing.T) {
	data := map[string]string{}
	data[PayloadKeyMethod] = http.MethodPost
	data[PayloadKeyPath] = "/"

	// Ensure timestamp involved in payload hash calculation
	currentTimestamp := time.Now().Unix()
	p1, err := NewPayload(currentTimestamp, data)
	require.NoError(t, err)
	p2, err := NewPayload(currentTimestamp+1, data)
	require.NoError(t, err)
	require.NotEqual(t, p1.Hash(), p2.Hash())

	// Next check changed data will change entire hash
	data = map[string]string{}
	data[PayloadKeyMethod] = http.MethodPost
	data[PayloadKeyPath] = "/"
	data["hello"] = "world"
	p3, err := NewPayload(currentTimestamp, data)
	require.NoError(t, err)
	require.NotEqual(t, p1.Hash(), p3.Hash())
}

func TestPayload_Sign_Verify(t *testing.T) {
	secret := hexutil.Encode([]byte("test secret"))
	data := map[string]string{}
	data[PayloadKeyMethod] = http.MethodPost
	data[PayloadKeyPath] = "/"

	currentTimestamp := time.Now().Unix()
	signatureLifetime := int64(1)
	expiredTimestamp := currentTimestamp + signatureLifetime

	payload, err := NewPayload(expiredTimestamp, data)
	require.NoError(t, err)

	signature, err := payload.Sign(secret)
	require.NoError(t, err)

	// Check signature is valid
	err = payload.Verify(signature, secret, "testnet")
	require.NoError(t, err)

	// Check malicious signature verification
	evilSignature := "0x734ac43bf06cd3ff9a3ed58e02b1350c23abcdc05629abff6c896d5a6f63c992"
	err = payload.Verify(evilSignature, secret, "testnet")
	require.ErrorIs(t, err, InvalidPayloadSignatureError{
		Expected: signature,
		Actual:   evilSignature,
	})

	// Verify signature after expiration
	time.Sleep(time.Second * time.Duration(signatureLifetime))
	err = payload.Verify(signature, secret, "testnet")
	currentTimestamp = time.Now().Unix()
	require.ErrorIs(t, err, ExpiredSignatureError{
		SignatureTimestamp: payload.Timestamp,
		CurrentTimestamp:   currentTimestamp,
	})

	// Trying to change payloads timestamp with no luck
	payload.Timestamp += 10
	err = payload.Verify(signature, secret, "testnet")
	require.Error(t, err)
}
