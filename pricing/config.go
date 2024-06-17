package pricing

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "pricing.yaml"
)

type ServiceConfig struct {
	UpdateInterval   string         `yaml:"update_interval"`
	DefaultMaxUseAge string         `yaml:"default_max_use_age"`
	ReferenceCoin    string         `yaml:"reference_coin"`
	CoinData         []CoinData     `yaml:"coin_data"`
	ExchangeData     []ExchangeData `yaml:"exchange_data"`
	MarketData       []MarketData   `yaml:"market_data"`
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

	logrus.Info("Reading config from ", configPath)
	return config, cleanenv.ReadConfig(configPath, config)
}
