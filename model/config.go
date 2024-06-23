package model

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "broker.yaml"
)

const (
	TEST_MARKET_TICKER = "TEST-MARKET"
	API_INSTANCE       = "api-gateway"
	AUTH_INSTANCE      = "auth"
	PROFILE_INSTANCE   = "profile"
	MARKET_INSTANCE    = "market-instance"
)

type InstanceConfig struct {
	Title    string   `yaml:"title"`
	Host     string   `yaml:"host"`
	Port     uint32   `yaml:"port"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	Tags     []string `yaml:"tags"`
}

// TODO: move risk manager params here
type BrokerConfig struct {
	Instances []InstanceConfig `yaml:"instances"`
}

func ReadClusterConfig() (*BrokerConfig, error) {
	config := &BrokerConfig{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := path.Join(homeDir, DefaultConfigPath, DefaultConfigFile)
	return config, cleanenv.ReadConfig(configPath, config)
}
