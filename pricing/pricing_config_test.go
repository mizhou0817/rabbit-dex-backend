package pricing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

const (	
	configFile = "pricing.yaml"
	configDir = ".rabbit"
)

func TestPricingConfig(t *testing.T) {
	config := &Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}
	configPath := filepath.Join(homeDir, configDir, configFile)

	err = cleanenv.ReadConfig(configPath, config)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	_, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config to JSON: %v", err)
	}

	logrus.Printf("config: %s", configPath)
}
