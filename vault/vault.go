package vault

import (
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/strips-finance/rabbit-dex-backend/vault/vault"
)

var mutex sync.Mutex
var vault_instance *vault.Vault = nil

func dial(vault_l1_address common.Address) (*vault.Vault, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if vault_instance != nil {
		return vault_instance, nil
	}

	providerUrl := os.Getenv("ALCHEMY_URL")
	if providerUrl == "" {
		return nil, fmt.Errorf("no env ALCHEMY_URL = %s", providerUrl)
	}

	var err error
	var client *ethclient.Client
	var sleepFor = 4 * time.Second

	for i := 0; i < 5; i++ {
		client, err = ethclient.Dial(providerUrl)
		if err != nil {
			return nil, err
		}

		vault_instance, err = vault.NewVault(vault_l1_address, client)
		if err == nil {
			break
		}

		time.Sleep(sleepFor)
		sleepFor = sleepFor * 2
	}
	if err != nil {
		return nil, fmt.Errorf("error dialing eth client: %s", err.Error())
	}

	return vault_instance, nil
}

func IsValidSigner(vaultContract, traderWallet common.Address, requireRole uint) error {

	logrus.Infof("MOCK env = %s", os.Getenv("MOCK"))
	// We use some exact word like "mocked" to never set by mistake
	if os.Getenv("MOCK") == "mocked" {
		return MockedIsValidSigner(vaultContract, traderWallet, requireRole)
	}

	instance, err := dial(vaultContract)
	if err != nil {
		return err
	}

	isValid, err := instance.IsValidSigner(nil, traderWallet, big.NewInt(int64(requireRole)))

	if err != nil {
		return err
	}

	if !isValid {
		return fmt.Errorf("requiredRole=%d not allowed for trader=%s", requireRole, traderWallet.String())
	}

	return nil
}

func MockedIsValidSigner(vaultContract, traderWallet common.Address, requireRole uint) error {
	// We use some exact word like "mocked" to never set by mistake
	if os.Getenv("MOCK") != "mocked" {
		return fmt.Errorf("WRONG_ENV_CALL")
	}

	if os.Getenv("WRONG_SIGNER") != "" {
		return fmt.Errorf("MOCKED_WRONG_SIGNER")
	}

	return nil
}
