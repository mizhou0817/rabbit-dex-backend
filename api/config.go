package api

import (
	"os"
	"path"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = ".rabbit"
	DefaultConfigFile = "rest.yaml"
)

type PlatformConfig struct {
	CurrentVersion string `yaml:"current_version"`
	ForceUpdate    bool   `yaml:"force_update"`
	AppUrl         string `yaml:"app_url"`
}

type AnalyticsConfig struct {
	IgnoreProfileIds []uint              `yaml:"ignore_profile_ids"`
	AllowedURLPaths  map[string][]string `yaml:"allowed_url_paths"`
}

type ExchangeConfig struct {
	DomainNameEncoder  string   `yaml:"domain_name_encoder"`
	DomainNameWithdraw string   `yaml:"domain_name_withdraw"`
	ProviderUrl        string   `yaml:"provider_url"`
	OnboardingMessages []string `yaml:"onboarding_messages"`
	ChainId            uint     `yaml:"chain_id"`
	ExchangeAddress    string   `yaml:"exchange_address"`
	SignerKeyId        string   `yaml:"key_id"` //TODO: migrate to alias
	JwtPublic          string   `yaml:"jwt_public"`
}

type CompressionConfig struct {
	Enabled         bool   `yaml:"enabled"`
	MimeTypes       string `yaml:"mime_types"`
	MinResponseSize int    `yaml:"min_response_size"`
}

type ServiceConfig struct {
	Host                               string                    `yaml:"host"`
	Port                               uint32                    `yaml:"port"`
	Compression                        CompressionConfig         `yaml:"compression"`
	HMACSecret                         string                    `yaml:"hmac_secret"`
	JwtLifetime                        uint64                    `yaml:"jwt_lifetime"`
	RefreshTokenLifetime               uint64                    `yaml:"refresh_token_lifetime"`
	EnvMode                            string                    `yaml:"env_mode"`
	Markets                            []string                  `yaml:"markets"`
	AdminWallets                       []string                  `yaml:"admins"`
	SuperAdminWallets                  []string                  `yaml:"super_admins"`
	TimescaledbConnectionURI           string                    `yaml:"timescaledb_connection_uri"`
	Platforms                          map[string]PlatformConfig `yaml:"platforms"`
	AnalyticsConfig                    AnalyticsConfig           `yaml:"analytics_config"`
	Exchanges                          map[string]ExchangeConfig `yaml:"exchanges"`
	MigrationsTimescaledbConnectionURI string                    `yaml:"migrations_timescaledb_connection_uri"`
}

type Config struct {
	Service ServiceConfig `yaml:"service"`
}

var mutexCfg sync.Mutex
var cfgInstance *Config = nil

func ReadConfig() (*Config, error) {
	mutexCfg.Lock()
	defer mutexCfg.Unlock()

	if cfgInstance != nil {
		return cfgInstance, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var cfg Config
	configPath := path.Join(homeDir, DefaultConfigPath, DefaultConfigFile)
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, err
	}
	cfgInstance = &cfg

	return cfgInstance, nil
}

/*
Returns config with default parameters. For not it's mainly used in tests.
*/
func ReadDefaultConfig() *Config {
	return &Config{
		Service: ServiceConfig{
			Host:                 "127.0.0.1",
			Port:                 8888,
			HMACSecret:           "f6b15b49cd77fbb243ab118407dee3786dacfb8d6bdd3b281706199202dddec0",
			JwtLifetime:          172800,
			RefreshTokenLifetime: 1209600,
		},
	}
}
