package dashboards

import (
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "dashboards.yaml"
)

// Service config example can be found in configs-example/dashboards.yaml
type ServiceConfig struct {
	MigrationsTimescaledbConnectionURI string `yaml:"migrations_timescaledb_connection_uri"`
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
