package airdrop

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "airdrop.yaml"
)

type ServiceConfig struct {
	ChainId          uint   `yaml:"chain_id"`
	L1AirdropAddress string `yaml:"l1_airdrop_address"`
	SignerKeyId      string `yaml:"key_id"` //TODO: migrate to alias
}

type Config struct {
	Service ServiceConfig `yaml:"service"`
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
