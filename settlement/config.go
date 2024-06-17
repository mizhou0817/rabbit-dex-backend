package settlement

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "settlement.yaml"
)

/*
Adhoc solution for monitoring multiple contracts on different blockchains.
Inside service there will be multiple handlers created
Need to refactor the whole Settlement cuz it's too coupled now.
*/

type EthHandlerCfg struct {
	ExchangeId           string   `yaml:"exchange_id"`
	ChainId              uint     `yaml:"chain_id"`
	ExchangeAddress      string   `yaml:"exchange_address"`
	DepositAddress       string   `yaml:"deposit_address"`
	ProviderUrl          string   `yaml:"provider_url"`
	DefaultFromBlock     string   `yaml:"default_from_block"`
	DepositInterval      string   `yaml:"deposit_interval" env-default:"15"`
	WithdrawalInterval   string   `yaml:"withdrawal_interval" env-default:"15"`
	WithdrawalBlockDelay string   `yaml:"withdrawal_block_delay" env-default:"1800"`
	ConfirmationBlocks   string   `yaml:"confirmation_blocks" env-default:"12"`
	CancelInterval       string   `yaml:"cancel_interval" env-default:"60"`
	ProcessYield         bool     `yaml:"process_yield" env-default:"false"`
	ProcessYieldInterval string   `yaml:"process_yield_interval"`
	ClaimerPk            string   `yaml:"claimer_pk" env-default:""`
	Vaults               []string `yaml:"vaults"`
	Decimals             int32    `yaml:"decimals"`
}

type Config struct {
	Handlers map[string]EthHandlerCfg `yaml:"service"`
}

func ReadConfig() (*Config, error) {
	config := &Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := path.Join(homeDir, DefaultConfigPath, DefaultConfigFile)

	return config, cleanenv.ReadConfig(configPath, config)
}
