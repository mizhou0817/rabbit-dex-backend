package config

import (
	"path"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/shopspring/decimal"

	"github.com/strips-finance/rabbit-dex-backend/profile"
)

type Config struct {
	Service struct {
		TimescaledbConnectionURI string `yaml:"timescaledb_connection_uri"`

		Cache struct {
			PeriodicsInterval time.Duration `yaml:"periodics_interval"`
			ParallelWorkers   uint          `yaml:"parallel_workers"`
			BatchSize         uint          `yaml:"batch_size"`
			Liquidate         struct {
				MarginV1 struct {
					ForcedMargin decimal.Decimal `yaml:"forced_margin"`
				}
			} `yaml:"liquidate"`
		} `yaml:"cache"`

		OnboardTier struct {
			PeriodicsInterval time.Duration `yaml:"periodics_interval"`
			BatchSize         uint          `yaml:"batch_size"`
		} `yaml:"onboard_tier"`

		Tier struct {
			PeriodicsInterval time.Duration `yaml:"periodics_interval"`
			BatchSize         uint          `yaml:"batch_size"`
		} `yaml:"tier"`

		Notify struct {
			Retries    uint          `yaml:"retries"`
			RetryDelay time.Duration `yaml:"retry_delay"`
		} `yaml:"notify"`

		Markets struct {
			Workers    uint               `yaml:"workers"`
			Retries    uint               `yaml:"retries"`
			RetryDelay time.Duration      `yaml:"retry_delay"`
			Ids        []profile.MarketId `yaml:"ids"`
		} `yaml:"markets"`

		Profile struct {
			Retries    uint          `yaml:"retries"`
			RetryDelay time.Duration `yaml:"retry_delay"`
		} `yaml:"profile"`

		VolumeCache struct {
			RefreshInterval time.Duration `yaml:"refresh_interval"`
		} `yaml:"volume_cache"`

		TierCalc struct {
			BatchSize uint `yaml:"batch_size"`
		} `yaml:"tier_calc"`
	} `yaml:"service"`
}

func ReadConfig(paths ...string) (*Config, error) {
	config := &Config{}
	return config, cleanenv.ReadConfig(path.Join(paths...), config)
}
