package referrals

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "referrals.yaml"
)

type ServiceConfig struct {
	TimescaledbConnectionURI           string   `yaml:"timescaledb_connection_uri"`
	MigrationsTimescaledbConnectionURI string   `yaml:"migrations_timescaledb_connection_uri"`
	VolumesInterval                    int64    `yaml:"volumes_interval"`
	LeaderBoardInterval                int64    `yaml:"leaderboard_interval"`
	CreatePayoutsInterval              int64    `yaml:"create_payouts_interval"`
	ProcessPayoutsInterval             int64    `yaml:"process_payouts_interval"`
	ShardIds                           []string `yaml:"shard_ids"`
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
