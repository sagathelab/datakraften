package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/profiles"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
	"github.com/AlecAivazis/survey/v2"
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

			alreadyInit := false
			if _, err := os.Stat(configPath); err == nil {
				alreadyInit = true

				state := LoadState()
				currentProfile := state.ActiveProfile
				if currentProfile == "" {
					currentProfile = "default"
				}

				fmt.Println("  Datakraften is already initialized.")
				fmt.Printf("  Config: %s\n", configPath)
				if currentProfile != "" {
					fmt.Printf("  Current profile: %s\n", currentProfile)
				}
				fmt.Println()

				switchProfile := false
				prompt := &survey.Confirm{
					Message: "Switch to a different profile?",
					Default: false,
				}
				survey.AskOne(prompt, &switchProfile)

				if !switchProfile {
					fmt.Println()
					fmt.Println("  Run 'dk apply' to apply the configuration.")
					fmt.Println("  Run 'dk doctor' to check your environment.")
					return nil
				}
			}

			profile := profileFlag
			if profile == "" {
				selected, err := selectProfile()
				if err != nil {
					return err
				}
				profile = selected
			}

			if alreadyInit {
				data, err := os.ReadFile(configPath)
				if err == nil {
					lines := strings.Split(string(data), "\n")
					for i, line := range lines {
						if strings.HasPrefix(strings.TrimSpace(line), "profile:") {
							lines[i] = fmt.Sprintf("profile: %s", profile)
							break
						}
					}
					os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
				}
				state := LoadState()
				state.ActiveProfile = profile
				state.Save()

				fmt.Printf("  ✓ Profile updated to: %s\n", profile)
				fmt.Println()
				fmt.Println("  Run 'dk apply' to apply this profile.")
				return nil
			}

			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}

			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()

			verbosePrintf("home=%s configDir=%s\n", home, configDir)
			verbosePrintf("wsl=%v wslVer=%d distro=%s distroVer=%s\n", wsl, wslVer, distro, distroVer)
			verbosePrintf("profile=%s nativePM=%s\n", profile, installers.DetectPackageManager())

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

func selectProfile() (string, error) {
	allProfiles := profiles.All()
	opts := make([]string, len(allProfiles))
	for i, p := range allProfiles {
		opts[i] = fmt.Sprintf("%s  — %s", p.Name, p.Description)
	}

	var selected int
	prompt := &survey.Select{
		Message: "Select a profile:",
		Options: opts,
		Default: opts[1],
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	return allProfiles[selected].Name, nil
}
