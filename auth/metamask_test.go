package auth

import (
	"crypto/ecdsa"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/strips-finance/rabbit-dex-backend/model"
	"github.com/strips-finance/rabbit-dex-backend/signer"
)

const (
	message = "Welcome to RabbitX!\n\nClick to sign in and on-board your wallet for trading perpetuals.\n\nThis request will not trigger a blockchain transaction or cost any gas fees. This signature only proves you are the true owner of this wallet.\n\nBy signing this message you agree to the terms and conditions of the exchange."
)

func pwd() {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Printf("Error: %s", err)
		return
	}

    logrus.Printf("Current Working Directory: %s", dir)
}


func TestHasRole(t *testing.T) {
	pwd()
	godotenv.Load()
	currentTimestamp := time.Now().Unix()
	signatureLifetime := int64(1)
	expirationTimestamp := currentTimestamp + signatureLifetime

	wallet := os.Getenv("VAULT_ADDRESS")
	logrus.Printf("vault wallet %s", wallet)
	privateKeyString := os.Getenv("VAULT_TRADER_PRIVATE_KEY")
	if privateKeyString[:2] == "0x" {
		privateKeyString = privateKeyString[2:]
	}

	// Convert the hex string to an ecdsa.PrivateKey
	privateKey, err := crypto.HexToECDSA(privateKeyString)
	require.NoError(t, err)

	// sign the welcome message
	signRequest := &MetamaskSignRequest{
		Message:   message,
		Timestamp: expirationTimestamp,
	}
	encoder := signer.NewEIP712Encoder(
		"RabbitXId",
		"1",
		"",
		big.NewInt(int64(31337)),
	)
	signature, err := EthSign(EIP_712, signRequest, privateKey, encoder)
	assert.NoError(t, err)

	// verify signing address has trader role
	verifyRequest := &MetamaskVerifyRequest{
		Wallet:    wallet,
		Timestamp: expirationTimestamp,
		Signature: signature,
		EIP712Encoder: encoder,
        ProfileType: model.PROFILE_TYPE_VAULT,
	}
	messages := []string{message}
	err = VerifyProfile(EIP_191, verifyRequest, TRADER_ROLE, messages)
	require.NoError(t, err)

}

func TestMetamaskSignVerify(t *testing.T) {
	currentTimestamp := time.Now().Unix()
	signatureLifetime := int64(1)
	expirationTimestamp := currentTimestamp + signatureLifetime

	// Generate arbitrary Private Key to sign message
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Derive Public Key from Private Key
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	assert.True(t, ok)

	// Next derive wallet address from Public Key
	wallet := crypto.PubkeyToAddress(*publicKey).String()

	// Perform a signing of our message
	signRequest := &MetamaskSignRequest{
		Message:   message,
		Timestamp: expirationTimestamp,
	}
	encoder := signer.NewEIP712Encoder(
		"RabbitXId",
		"1",
		"",
		big.NewInt(int64(31337)),
	)
	// signature, err := EthSign(EIP_712, signRequest, privateKey, encoder)
	signature, err := EthSign(EIP_191, signRequest, privateKey, encoder)
	assert.NoError(t, err)

	// Now try to verify obtained signature and message
	verifyRequest := &MetamaskVerifyRequest{
		Wallet:    wallet,
		Timestamp: expirationTimestamp,
		Signature: signature,
		EIP712Encoder: encoder,
	}
	messages := []string{message}
	// err = VerifyProfile(EIP_712, verifyRequest, TRADER_ROLE)
	err = VerifyProfile(EIP_191, verifyRequest, TRADER_ROLE, messages)
	require.NoError(t, err)

	// And now try to verify malicious signature
	evilSignature := "0xf942293eff01d56e981e371e3943b6a13936ecc825deb8c5efc2e972c43de66d6267dcf342f955beac2700e1a476642a4815f804217691fd7eb92b43def1880000"
	verifyRequest.Signature = evilSignature
	messages = []string{message}
	// err = VerifyProfile(EIP_712, verifyRequest, TRADER_ROLE)
	err = VerifyProfile(EIP_191, verifyRequest, TRADER_ROLE, messages)
	require.Error(t, err)

	// Ensure expired signature is invalid
	verifyRequest.Signature = signature
	time.Sleep(time.Second * time.Duration(signatureLifetime))
	messages = []string{message}
	// err = VerifyProfile(EIP_712, verifyRequest, TRADER_ROLE)
	err = VerifyProfile(EIP_191, verifyRequest, TRADER_ROLE, messages)
	currentTimestamp = time.Now().Unix()
	require.ErrorIs(t, err, ExpiredSignatureError{
		SignatureTimestamp: expirationTimestamp,
		CurrentTimestamp:   currentTimestamp,
	})
}
