package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
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

func newInitCmd() *cobra.Command {
	var customFlag, fromURL string
	var yes bool
	const customPrompt = "-"

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Datakraften configuration",
		Long: `Initialize Datakraften by creating a configuration file.

No arguments creates a default config (source: default).
Use --custom to choose a built-in starting point, or --custom <file> to merge a local YAML file.
Use --from <HTTPS URL> to import a raw YAML template once and merge it with defaults.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && cmd.Flags().Changed("custom") && customFlag == customPrompt {
				return nil
			}
			return cobra.NoArgs(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			hasCustom := cmd.Flags().Changed("custom")
			hasFrom := cmd.Flags().Changed("from")
			if hasCustom && hasFrom {
				return fmt.Errorf("specify only one of --custom or --from")
			}
			configPath := config.DefaultConfigPath()
			if configPath == "" {
				return fmt.Errorf("cannot determine home directory for workstation configuration")
			}
			info, err := os.Lstat(configPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("inspect workstation configuration: %w", err)
			}
			existing := err == nil
			if existing && !info.Mode().IsRegular() {
				return fmt.Errorf("cannot replace workstation configuration %s: not a regular file", configPath)
			}
			displayTitle(os.Stdout, "init", "Configure your workstation")
			switch {
			case hasFrom:
				return initFromURL(configPath, fromURL, existing, yes)
			case !hasCustom:
				return initDefault(configPath, existing, yes)
			case len(args) == 1:
				return initCustomMerge(configPath, args[0], existing, yes)
			case customFlag != customPrompt && customFlag != "":
				return initCustomMerge(configPath, customFlag, existing, yes)
			default:
				return initCustomPrompt(configPath, existing, yes)
			}
		},
	}

	cmd.Flags().StringVarP(&customFlag, "custom", "c", "", "Choose a built-in custom config or merge a local YAML file")
	cmd.Flags().Lookup("custom").NoOptDefVal = customPrompt
	cmd.Flags().StringVar(&fromURL, "from", "", "Import an HTTPS YAML template once and merge it with defaults")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Confirm overwriting the existing workstation configuration")
	return cmd
}

func builtInConfig(template string) string {
	return fmt.Sprintf(template, installers.DetectPackageManager())
}

func saveInitConfig(path string, data []byte, existing, yes bool) (bool, error) {
	if _, err := config.Decode(data); err != nil {
		return false, fmt.Errorf("invalid configuration: %w", err)
	}
	if existing && !yes {
		var confirmed bool
		prompt := &survey.Confirm{Message: "Overwrite the existing workstation configuration?", Default: false}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return false, fmt.Errorf("confirmation required (use --yes to overwrite non-interactively): %w", err)
		}
		if !confirmed {
			fmt.Println("  Cancelled; existing configuration preserved.")
			return false, nil
		}
	}
	if err := writeInitAtomic(path, data, existing); err != nil {
		return false, err
	}
	return true, nil
}

func initDefault(configPath string, existing, yes bool) error {
	cfg := builtInConfig(defaultConfigYAML)
	printConfigPreview(cfg)
	saved, err := saveInitConfig(configPath, []byte(cfg), existing, yes)
	if err != nil || !saved {
		return err
	}
	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()

	displayHeading(os.Stdout, "Environment")
	if wsl {
		displayPrintf("    [OK] WSL%d\n", wslVer)
	} else {
		displayPrintln("    [SKIP] Not in WSL")
	}
	if distro != "" {
		displayPrintf("    [OK] %s %s\n", distro, distroVer)
	}
	fmt.Println("    Source: default")
	fmt.Println()
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Println()
	displayComplete(os.Stdout, "Default config created")
	fmt.Println()
	displayNextSteps(os.Stdout)
	fmt.Println("    1. Run 'dk apply' to install tools")
	fmt.Println("    2. Run 'dk doctor' to verify")
	fmt.Println()
	return nil
}

func initCustomPrompt(configPath string, existing, yes bool) error {
	var templateChoice int
	templatePrompt := &survey.Select{
		Message: "Create config as:",
		Options: []string{
			"Empty skeleton — start from scratch with minimal fields",
			"Pre-filled with defaults — edit what you need",
		},
	}
	if err := survey.AskOne(templatePrompt, &templateChoice); err != nil {
		return fmt.Errorf("choose a custom configuration: %w", err)
	}
	cfg := builtInConfig(skeletonConfigYAML)
	if templateChoice == 1 {
		cfg = strings.Replace(builtInConfig(defaultConfigYAML), "source: default", "source: custom", 1)
	}
	printConfigPreview(cfg)
	saved, err := saveInitConfig(configPath, []byte(cfg), existing, yes)
	if err != nil || !saved {
		return err
	}
	printCustomInitResult(configPath)
	return nil
}

func initCustomMerge(configPath, customFile string, existing, yes bool) error {
	partial, err := config.ReadFile(customFile)
	if err != nil {
		return fmt.Errorf("cannot load custom input: %w", err)
	}
	base := make(map[string]any)
	if existing {
		base, err = config.ReadFile(configPath)
		if err != nil {
			return err
		}
	}
	result, paths, err := mergeConfig(base, partial)
	if err != nil {
		return err
	}
	fmt.Printf("  Merge local YAML %s into %s (source: custom).\n", customFile, configPath)
	if len(paths) > 0 {
		fmt.Printf("  Settings supplied: %s\n", strings.Join(paths, ", "))
	}
	warnUnusedConfig(paths)
	saved, err := saveInitConfig(configPath, result, existing, yes)
	if err != nil || !saved {
		return err
	}
	printCustomInitResult(configPath)
	return nil
}

func printConfigPreview(cfg string) {
	displayHeading(os.Stdout, "Config preview")
	for _, line := range strings.Split(cfg, "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()
}

func printCustomInitResult(configPath string) {
	displayComplete(os.Stdout, fmt.Sprintf("Custom config saved: %s", configPath))
	fmt.Println()
	displayNextSteps(os.Stdout)
	fmt.Println("    1. Edit the config file to customize your toolset")
	fmt.Println("    2. Run 'dk apply' to install tools")
	fmt.Println("    3. Run 'dk doctor' to verify")
	fmt.Println()
}
