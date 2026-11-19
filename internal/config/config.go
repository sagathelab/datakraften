package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type RuntimeConfig struct {
	Enabled bool   `yaml:"enabled"`
	Manager string `yaml:"manager"`
	Version string `yaml:"version"`
}

type Config struct {
	Version string `yaml:"version"`
	Source  string `yaml:"source"`
	System  struct {
		PackageManager string `yaml:"package_manager"`
	} `yaml:"system"`
	Tooling struct {
		PackageManager string `yaml:"package_manager"`
	} `yaml:"tooling"`
	Shell struct {
		Default     string `yaml:"default"`
		Prompt      string `yaml:"prompt"`
		History     string `yaml:"history"`
		FuzzyFinder string `yaml:"fuzzy_finder"`
	} `yaml:"shell"`
	Runtimes struct {
		Node   RuntimeConfig `yaml:"node"`
		Python RuntimeConfig `yaml:"python"`
		Go     RuntimeConfig `yaml:"go"`
		Dotnet RuntimeConfig `yaml:"dotnet"`
	} `yaml:"runtimes"`
	Tools   map[string]bool          `yaml:"tools"`
	Editors map[string]string        `yaml:"editors"`
	AITools map[string]RuntimeConfig `yaml:"ai_tools"`
	AIApps  map[string]RuntimeConfig `yaml:"ai_apps"`
}

func DefaultConfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "datakraften", "config.yaml")
	}
	return ""
}

func Load() (*Config, error) {
	path := DefaultConfigPath()
	if path == "" {
		return nil, fmt.Errorf("cannot determine home directory for workstation configuration")
	}
	settings, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := node.Encode(settings); err != nil {
		return nil, fmt.Errorf("cannot encode workstation configuration: %w", err)
	}
	cfg := Config{Source: "default"}
	if err := node.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("cannot decode workstation configuration: %w", err)
	}
	return &cfg, nil
}

func ReadFile(path string) (map[string]any, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("read workstation configuration: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		info, err = os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot follow workstation configuration symlink: %v", err)
		}
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("workstation configuration %s must be a regular file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workstation configuration: %w", err)
	}
	settings, err := Decode(data)
	if err != nil {
		return nil, fmt.Errorf("invalid workstation configuration %s: %w", path, err)
	}
	return settings, nil
}

func ValidateSource(source string) error {
	switch source {
	case "default", "custom":
		return nil
	case "team":
		return retiredTeamError()
	default:
		return fmt.Errorf("source must be default or custom")
	}
}

func retiredTeamError() error {
	return fmt.Errorf("legacy team configuration is no longer supported; use 'dk init' or explicitly import a raw YAML template with 'dk init --from <HTTPS URL>'")
}
