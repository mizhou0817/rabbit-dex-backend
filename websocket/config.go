package websocket

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "websocket.yaml"
)

type Config struct {
	Service struct {
		CentrifugoUrl    string `yaml:"centrifugo_url"`
		CentrifugoApiKey string `yaml:"centrifugo_api_key"`
		HMACSecret       string `yaml:"hmac_secret"`
	} `yaml:"service"`
}

type ConfigReader = func() (*Config, error)

func ReadConfig() (*Config, error) {
	config := &Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := path.Join(homeDir, DefaultConfigPath, DefaultConfigFile)

	return config, cleanenv.ReadConfig(configPath, config)
}
