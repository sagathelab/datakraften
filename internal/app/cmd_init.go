package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultConfigYAML = `version: 1
source: default

system:
  package_manager: %s

tooling:
  package_manager: brew

shell:
  default: fish
  prompt: starship
  history: atuin
  fuzzy_finder: fzf

tools:
  github_cli: true
  azure_cli: true
  docker: true

runtimes:
  node:
    enabled: true
    manager: fnm
    version: lts
  python:
    enabled: true
    manager: uv
    version: latest
  go:
    enabled: true
    manager: brew
    version: latest
  dotnet:
    enabled: true
    manager: brew
    version: latest

editors:
  vscode: true
  zed: true
  cursor: optional

ai_tools:
  codex:
    enabled: true
    manager: npm
    version: latest
  opencode:
    enabled: true
    manager: brew
    version: latest
  copilot:
    enabled: true
    manager: gh
    version: latest
  claude:
    enabled: false
    manager: npm
    version: latest
  gemini:
    enabled: false
    manager: npm
    version: latest
ai_apps:
  codex:
    enabled: true
    manager: brew
    version: latest
  claude:
    enabled: false
    manager: brew
    version: latest
  copilot:
    enabled: true
    manager: vscode
    version: latest
`

var skeletonConfigYAML = `version: 1
source: custom

system:
  package_manager: %s

tooling:
  package_manager: ""

shell:
  default: bash
  prompt: ""
  history: ""
  fuzzy_finder: ""

runtimes:
  node:
    enabled: false
  python:
    enabled: false
  go:
    enabled: false
  dotnet:
    enabled: false
    manager: brew
    version: latest

tools: {}

editors: {}

ai_tools: {}
ai_apps: {}
`

var teamConfigYAML = `source: team
url: %s
`

func newInitCmd() *cobra.Command {
	var customFlag string
	var teamURLFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Datakraften configuration",
		Long: `Initialize Datakraften by creating a configuration file.

No arguments creates a default config (source: default).
Use --custom to create a custom config (source: custom) that you own.
Use --team to create a thin config pointing to a remote YAML (source: team).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			configDir := filepath.Join(home, ".config", "datakraften")
			configPath := filepath.Join(configDir, "config.yaml")

			hasCustom := cmd.Flags().Changed("custom")
			hasTeam := cmd.Flags().Changed("team")

			if hasCustom && hasTeam {
				fmt.Println("  Specify only one of --custom or --team.")
				return nil
			}

			existing := false
			if _, err := os.Stat(configPath); err == nil {
				existing = true
			}

			if !hasCustom && !hasTeam {
				return initDefault(configPath, configDir, existing)
			}

			if hasCustom {
				return initCustom(configPath, configDir, existing, customFlag)
			}

			return initTeam(configPath, configDir, existing, teamURLFlag)
		},
	}

	cmd.Flags().StringVarP(&customFlag, "custom", "c", "", "Path to a YAML file with custom tool configuration")
	cmd.Flags().StringVarP(&teamURLFlag, "team", "t", "", "Team config URL (creates a thin config pointing to a remote YAML)")

	return cmd
}

func initDefault(configPath, configDir string, existing bool) error {
	if existing {
		var overwrite bool
		prompt := &survey.Confirm{
			Message: "Config already exists. Overwrite with default config?",
			Default: false,
		}
		survey.AskOne(prompt, &overwrite)
		if !overwrite {
			fmt.Println()
			fmt.Println("  Run 'dk apply' to apply the configuration.")
			return nil
		}
	}

	pm := installers.DetectPackageManager()
	cfg := fmt.Sprintf(defaultConfigYAML, string(pm))

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	state := LoadState()
	state.ActiveSource = "default"
	state.Save()

	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()

	fmt.Println("  Datakraften init")
	fmt.Println()
	fmt.Println("  Detected:")
	if wsl {
		fmt.Printf("    ✓ WSL%d\n", wslVer)
	} else {
		fmt.Println("    – Not in WSL")
	}
	if distro != "" {
		fmt.Printf("    ✓ %s %s\n", distro, distroVer)
	}
	fmt.Printf("    ✓ Source: default\n")
	fmt.Println()

	fmt.Println("  Config preview:")
	for _, line := range strings.Split(cfg, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()
	fmt.Printf("  Writing config to: %s\n", configPath)
	fmt.Println()
	fmt.Println("  ✓ Default config created")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    1. Run 'dk apply' to install tools")
	fmt.Println("    2. Run 'dk doctor' to verify")

	return nil
}

func initCustom(configPath, configDir string, existing bool, customFile string) error {
	if customFile == "" {
		return initCustomPrompt(configPath, configDir, existing)
	}
	return initCustomMerge(configPath, configDir, existing, customFile)
}

func initCustomPrompt(configPath, configDir string, existing bool) error {
	if existing {
		var overwrite bool
		prompt := &survey.Confirm{
			Message: "Config already exists. Overwrite with custom config?",
			Default: false,
		}
		survey.AskOne(prompt, &overwrite)
		if !overwrite {
			fmt.Println()
			fmt.Println("  Run 'dk apply' to apply the configuration.")
			return nil
		}
	}

	var templateChoice int
	templatePrompt := &survey.Select{
		Message: "Create config as:",
		Options: []string{
			"Empty skeleton — start from scratch with minimal fields",
			"Pre-filled with defaults — edit what you need",
		},
	}
	survey.AskOne(templatePrompt, &templateChoice)

	pm := installers.DetectPackageManager()
	var cfg string
	if templateChoice == 0 {
		cfg = fmt.Sprintf(skeletonConfigYAML, string(pm))
	} else {
		cfg = fmt.Sprintf(defaultConfigYAML, string(pm))
		cfg = strings.Replace(cfg, "source: default", "source: custom", 1)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	state := LoadState()
	state.ActiveSource = "custom"
	state.Save()

	fmt.Println()
	fmt.Println("  Config preview:")
	for _, line := range strings.Split(cfg, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()
	fmt.Printf("  Writing config to: %s\n", configPath)
	fmt.Println()
	fmt.Printf("  ✓ Custom config created at %s\n", configPath)
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    1. Edit the config file to customize your toolset")
	fmt.Println("    2. Run 'dk apply' to install tools")
	fmt.Println("    3. Run 'dk doctor' to verify")

	return nil
}

func initCustomMerge(configPath, configDir string, existing bool, customFile string) error {
	data, err := os.ReadFile(customFile)
	if err != nil {
		fmt.Printf("  ✗ Cannot read file: %s\n", customFile)
		return nil
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println("  ✗ File is empty.")
		return nil
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		fmt.Printf("  ✗ Invalid YAML: %s\n", err)
		return nil
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	var finalCfg string
	if existing {
		existingData, err := os.ReadFile(configPath)
		if err == nil {
			ev := viper.New()
			ev.SetConfigType("yaml")
			ev.ReadConfig(strings.NewReader(string(existingData)))

			keys := []string{"system", "tooling", "shell", "runtimes", "tools", "editors", "ai_tools", "ai_apps"}
			for _, key := range keys {
				if v.IsSet(key) {
					ev.Set(key, v.Get(key))
				}
			}
			ev.Set("source", "custom")

			out := yamlFromViper(ev)
			finalCfg = string(out)
		}
	}

	if finalCfg == "" {
		finalCfg = fmt.Sprintf("source: custom\nversion: 1\n%s\n", string(data))
	}

	if err := os.WriteFile(configPath, []byte(finalCfg), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	state := LoadState()
	state.ActiveSource = "custom"
	state.Save()

	fmt.Println()
	fmt.Println("  Config preview:")
	for _, line := range strings.Split(finalCfg, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()
	fmt.Printf("  Writing config to: %s\n", configPath)
	fmt.Println()
	fmt.Printf("  ✓ Custom config created from %s\n", customFile)
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    1. Edit the config file to customize your toolset")
	fmt.Println("    2. Run 'dk apply' to install tools")
	fmt.Println("    3. Run 'dk doctor' to verify")

	return nil
}

func yamlFromViper(v *viper.Viper) []byte {
	allSettings := v.AllSettings()
	if len(allSettings) == 0 {
		return []byte{}
	}
	var b strings.Builder
	b.WriteString("version: 1\n")
	if s, ok := allSettings["source"]; ok {
		b.WriteString(fmt.Sprintf("source: %v\n", s))
		delete(allSettings, "source")
		delete(allSettings, "version")
	}
	for key, val := range allSettings {
		if key == "version" || key == "source" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s:\n", key))
		writeYAMLValue(&b, val, 2)
	}
	return []byte(b.String())
}

func writeYAMLValue(b *strings.Builder, val interface{}, indent int) {
	prefix := strings.Repeat(" ", indent)
	switch v := val.(type) {
	case map[string]interface{}:
		for k, vv := range v {
			switch inner := vv.(type) {
			case map[string]interface{}:
				b.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
				writeYAMLValue(b, inner, indent+2)
			default:
				b.WriteString(fmt.Sprintf("%s%s: ", prefix, k))
				writeYAMLValue(b, vv, indent)
			}
		}
	case bool:
		b.WriteString(fmt.Sprintf("%v\n", v))
	case string:
		b.WriteString(fmt.Sprintf("%s\n", v))
	case int:
		b.WriteString(fmt.Sprintf("%d\n", v))
	case nil:
		b.WriteString("\n")
	default:
		b.WriteString(fmt.Sprintf("%v\n", v))
	}
}

func initTeam(configPath, configDir string, existing bool, url string) error {
	if url == "" {
		var input string
		urlPrompt := &survey.Input{
			Message: "Team config URL:",
		}
		if err := survey.AskOne(urlPrompt, &input, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		url = input
	}

	fmt.Printf("    Fetching remote config from %s...\n", url)

	_, err := config.FetchRemote(url)
	if err != nil {
		fmt.Printf("    ✗ %s\n", err)
		return fmt.Errorf("failed to load remote config")
	}
	fmt.Println("    ✓ Remote config validated")

	thinCfg := fmt.Sprintf(teamConfigYAML, url)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(thinCfg), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	state := LoadState()
	state.ActiveSource = "team"
	state.Save()

	fmt.Println()
	fmt.Println("  Config preview:")
	for _, line := range strings.Split(thinCfg, "\n") {
		if strings.TrimSpace(line) != "" {
			fmt.Printf("    %s\n", line)
		}
	}
	fmt.Println()
	fmt.Printf("  Writing config to: %s\n", configPath)
	fmt.Println()
	fmt.Printf("  ✓ Team config created\n")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    1. Run 'dk apply' to install tools (fetches remote config)")
	fmt.Println("    2. Run 'dk doctor' to verify")

	return nil
}
