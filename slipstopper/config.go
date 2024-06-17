package slipstopper

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "slipstopper.yaml"
)

type ServiceConfig struct {
	Markets                   []string `yaml:"markets"`
	CentrifugoHMACSecretToken string   `yaml:"centrifugo_hmac_secret_token"`
	WebsocketURI              string   `yaml:"websocket_uri"`
}

type Config struct {
	Service ServiceConfig `yaml:"service"`
}

func ReadConfig(fullPath ...string) (*Config, error) {
	config := &Config{}
	var homeDir string
	var configPath string
	var err error

	if len(fullPath) > 0 {
		configPath = fullPath[0]
	} else {
		homeDir, err = os.UserHomeDir()
		configPath = path.Join(homeDir, DefaultConfigPath, DefaultConfigFile)
	}

	if err != nil {
		return nil, err
	}

	return config, cleanenv.ReadConfig(configPath, config)
}
