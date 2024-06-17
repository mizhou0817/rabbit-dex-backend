package archiver

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "archiver.yaml"
)

type TarantoolSpace struct {
	TimescaledbTable string      `yaml:"timescaledb_table"`
	SyncInterval     uint64      `yaml:"sync_interval"`
	BatchSize        uint64      `yaml:"batch_size"`
	Instances        []string    `yaml:"instances"`
	Mode             ArchiveMode `yaml:"mode"`
	UniqueId         []string    `yaml:"unique_id"`
}

// Service config example can be found in configs-example/archiver.yaml
type ServiceConfig struct {
	TimescaledbConnectionURI string                    `yaml:"timescaledb_connection_uri"`
	TarantoolSpaces          map[string]TarantoolSpace `yaml:"tarantool_spaces"`

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
