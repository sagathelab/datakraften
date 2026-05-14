package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Version string `mapstructure:"version"`
	Profile string `mapstructure:"profile"`
	System  struct {
		PackageManager string `mapstructure:"package_manager"`
	} `mapstructure:"system"`
	Tooling struct {
		PackageManager string `mapstructure:"package_manager"`
	} `mapstructure:"tooling"`
	Shell struct {
		Default     string `mapstructure:"default"`
		Prompt      string `mapstructure:"prompt"`
		History     string `mapstructure:"history"`
		FuzzyFinder string `mapstructure:"fuzzy_finder"`
	} `mapstructure:"shell"`
	Tools    map[string]bool `mapstructure:"tools"`
	Editors  map[string]string `mapstructure:"editors"`
	AI       map[string]string `mapstructure:"ai"`
}

func DefaultConfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "datakraften", "config.yaml")
	}
	return ""
}

func ConfigPaths() []string {
	return []string{
		"datakraften.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "datakraften", "config.yaml"),
	}
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	for _, p := range ConfigPaths() {
		v.AddConfigPath(filepath.Dir(p))
	}

	v.SetConfigFile(DefaultConfigPath())

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("no config file found (%s)", DefaultConfigPath())
		}
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &cfg, nil
}
