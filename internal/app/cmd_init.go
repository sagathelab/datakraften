package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/profiles"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Datakraften configuration",
		Long:  `Initialize Datakraften by creating a configuration file and detecting the current system state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			configDir := filepath.Join(home, ".config", "datakraften")
			configPath := filepath.Join(configDir, "config.yaml")

			if _, err := os.Stat(configPath); err == nil {
				fmt.Println("  Datakraften is already initialized.")
				fmt.Printf("  Config: %s\n", configPath)
				fmt.Println()
				fmt.Println("  Run 'dk apply' to apply the configuration.")
				fmt.Println("  Run 'dk doctor' to check your environment.")
				return nil
			}

			profile := profileFlag
			if profile == "" {
				profile = promptProfile(profiles.Available())
			}

			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}

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
			fmt.Printf("    ✓ Profile: %s\n", profile)
			fmt.Println()
			fmt.Printf("  Writing config to: %s\n", configPath)
			fmt.Println()

			nativePM := string(installers.DetectPackageManager())

			cfg := fmt.Sprintf(`version: 1
profile: %s

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
  dotnet:
    enabled: false

editors:
  vscode: true
  zed: true
  cursor: optional

ai:
  codex: false
  opencode: false
  github_copilot: false
`, profile, nativePM)

			if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			state := LoadState()
			state.ActiveProfile = profile
			state.Save()

			fmt.Println("  ✓ Config created")
			fmt.Println()
			fmt.Println("  Next steps:")
			fmt.Println("    1. Run 'dk apply' to install tools")
			fmt.Println("    2. Run 'dk doctor' to verify")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to initialize (minimal, default, ai, dotnet, frontend, platform)")

	return cmd
}

func promptProfile(profiles []string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  Select a profile:")
	for i, p := range profiles {
		fmt.Printf("    [%d] %s\n", i+1, p)
	}
	fmt.Print("  Enter number [3]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return "default"
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil || idx < 1 || idx > len(profiles) {
		fmt.Println("  Invalid selection, using default.")
		return "default"
	}

	return profiles[idx-1]
}
