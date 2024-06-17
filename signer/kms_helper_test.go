package signer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	KEY             = "arn:aws:kms:ap-northeast-1:618528691313:key/d64e0d8b-ed1f-47c3-8112-fae6f8204d63"
	TESTNET_KEY     = "arn:aws:kms:ap-northeast-1:763292132769:key/d9237bcf-a1d0-480b-bff3-d440cf27b91e"
	BFX_MAINNET_KEY = "arn:aws:kms:ap-northeast-1:618528691313:key/c9b5d291-3ee5-4eae-8a0f-f132965bbe72"
)

func TestCreateKey(t *testing.T) {
	createKeyOutput, err := CreateEncryptDecryptKey("test_key")
	if err != nil {
		t.Error(err)
	}
	keyID := *createKeyOutput.KeyMetadata.KeyId
	keyARN := *createKeyOutput.KeyMetadata.Arn
	fmt.Println("Key Id:", keyID)
	fmt.Println("Key ARN:", keyARN)
}

func TestEncryptDecrypt(t *testing.T) {
	ciphertext, err := Encrypt("testing 123", KEY)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Encrypted data:", ciphertext)
	decryptedPlaintext, err := Decrypt(ciphertext)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Decrypted data:", decryptedPlaintext)
}

func TestGetAddresses(t *testing.T) {
	// address1, err := GetEthAddress(KEY)
	// assert.NoError(t, err)
	address2, err := GetEthAddress(TESTNET_KEY)
	assert.NoError(t, err)
	// fmt.Printf("Orig Address: %v\n", address1)
	fmt.Printf("Testnet Address: %v\n", address2)
}

func TestGetRbxAddresses(t *testing.T) {
	// address1, err := GetEthAddress(KEY)
	// assert.NoError(t, err)
	address2, err := GetEthAddress(KEY)
	assert.NoError(t, err)
	// fmt.Printf("Orig Address: %v\n", address1)
	fmt.Printf("Testnet Address: %v\n", address2)
}

func TestGetBfxAddresses(t *testing.T) {
	// address1, err := GetEthAddress(KEY)
	// assert.NoError(t, err)
	address2, err := GetEthAddress(BFX_MAINNET_KEY)
	assert.NoError(t, err)
	// fmt.Printf("Orig Address: %v\n", address1)
	fmt.Printf("Bfx key Address: %v\n", address2)
}
