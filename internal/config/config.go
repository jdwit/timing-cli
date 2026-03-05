package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey   string `yaml:"api-key"`
	Timezone string `yaml:"timezone"`
}

var configDir = filepath.Join(os.Getenv("HOME"), ".config", "timing-cli")
var configFile = filepath.Join(configDir, "config.yaml")

func Dir() string {
	return configDir
}

func File() string {
	return configFile
}

func Load() (*Config, error) {
	cfg := &Config{
		Timezone: "Europe/Amsterdam",
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "api-key":
		return cfg.APIKey, nil
	case "timezone":
		return cfg.Timezone, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "api-key":
		cfg.APIKey = value
	case "timezone":
		cfg.Timezone = value
	default:
		return fmt.Errorf("unknown config key: %s (valid keys: api-key, timezone)", key)
	}

	return Save(cfg)
}

func GetAPIKey() string {
	if key := os.Getenv("TIMING_API_KEY"); key != "" {
		return key
	}
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.APIKey
}

func GetTimezone() string {
	cfg, err := Load()
	if err != nil {
		return "Europe/Amsterdam"
	}
	if cfg.Timezone == "" {
		return "Europe/Amsterdam"
	}
	return cfg.Timezone
}
