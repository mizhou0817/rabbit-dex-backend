package api_client

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

type ClientCredentials struct {
	Wallet     string `yaml:"wallet"`
	PrivateKey string `yaml:"private_key"`
	APIKey     string `yaml:"api_key"`
	APISecret  string `yaml:"api_secret"`
	Jwt        string
	ProfileID  uint
}

func (c *ClientCredentials) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	if strings.HasPrefix(c.PrivateKey, "0x") {
		c.PrivateKey = c.PrivateKey[2:]
	}

	if len(c.PrivateKey) == 0 {
		return nil, errors.New("PrivateKey is required")
	}

	return crypto.HexToECDSA(c.PrivateKey)
}

func (c *ClientCredentials) GetWallet() (string, error) {
	if len(c.Wallet) > 0 {
		return c.Wallet, nil
	}

	privateKey, err := c.GetPrivateKey()
	if err != nil {
		return "", err
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("failed to convert to ecdsa.PublicKey")
	}

	return crypto.PubkeyToAddress(*publicKey).String(), nil
}
