package config

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type RuntimeConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Manager string `mapstructure:"manager"`
	Version string `mapstructure:"version"`
}

type CustomConfig struct{}

type TeamConfig struct {
	URL string `mapstructure:"url"`
}

type Config struct {
	Version string `mapstructure:"version"`
	Source  string `mapstructure:"source"`
	URL     string `mapstructure:"url"`
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
	Runtimes struct {
		Node   RuntimeConfig `mapstructure:"node"`
		Python RuntimeConfig `mapstructure:"python"`
		Go     RuntimeConfig `mapstructure:"go"`
		Dotnet RuntimeConfig `mapstructure:"dotnet"`
	} `mapstructure:"runtimes"`
	Tools   map[string]bool          `mapstructure:"tools"`
	Editors map[string]string        `mapstructure:"editors"`
	AITools map[string]RuntimeConfig `mapstructure:"ai_tools"`
	AIApps  map[string]RuntimeConfig `mapstructure:"ai_apps"`
	Custom  CustomConfig             `mapstructure:"custom"`
	Team    TeamConfig               `mapstructure:"team"`
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

	// Backward compat: map old profile: field to source:
	if cfg.Source == "" {
		profile := v.GetString("profile")
		switch profile {
		case "custom":
			cfg.Source = "custom"
		case "team":
			cfg.Source = "team"
		default:
			cfg.Source = "default"
		}
	}

	// Backward compat: map old team.url: to flat url:
	if cfg.URL == "" && cfg.Team.URL != "" {
		cfg.URL = cfg.Team.URL
	}

	// Backward compat: migrate old bool format for ai_tools/ai_apps to struct format
	raw := v.AllSettings()
	if aiToolsRaw, ok := raw["ai_tools"].(map[string]interface{}); ok && cfg.AITools == nil {
		cfg.AITools = make(map[string]RuntimeConfig)
		for name, val := range aiToolsRaw {
			if b, ok := val.(bool); ok {
				cfg.AITools[name] = RuntimeConfig{Enabled: b}
			}
		}
	}
	if aiAppsRaw, ok := raw["ai_apps"].(map[string]interface{}); ok && cfg.AIApps == nil {
		cfg.AIApps = make(map[string]RuntimeConfig)
		for name, val := range aiAppsRaw {
			if b, ok := val.(bool); ok {
				cfg.AIApps[name] = RuntimeConfig{Enabled: b}
			}
		}
	}

	return &cfg, nil
}

func FetchRemote(url string) (*Config, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote config returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote config: %w", err)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewReader(body)); err != nil {
		return nil, fmt.Errorf("failed to parse remote config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote config: %w", err)
	}

	return &cfg, nil
}
