package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
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
				if profile == "team" {
					thinCfg, err := promptTeamURL()
					if err != nil {
						return err
					}
					os.WriteFile(configPath, []byte(thinCfg), 0644)

					state := LoadState()
					state.ActiveProfile = profile
					state.Save()

					fmt.Println()
					fmt.Println("  Config preview:")
					fmt.Printf("    profile: team\n")
					for _, line := range strings.Split(thinCfg, "\n") {
						if strings.TrimSpace(line) != "" {
							fmt.Printf("    %s\n", line)
						}
					}
					fmt.Println()
					fmt.Printf("  Writing config to: %s\n", configPath)
					fmt.Println()
					fmt.Printf("  ✓ Profile updated to: %s\n", profile)
					fmt.Println()
					fmt.Println("  Run 'dk apply' to apply this profile.")
					return nil
				}

				data, err := os.ReadFile(configPath)
				if err == nil {
					lines := strings.Split(string(data), "\n")
					for i, line := range lines {
						if strings.HasPrefix(strings.TrimSpace(line), "profile:") {
							lines[i] = fmt.Sprintf("profile: %s", profile)
							break
						}
					}
					updatedCfg := strings.Join(lines, "\n")

					os.WriteFile(configPath, []byte(updatedCfg), 0644)

					fmt.Println()
					fmt.Println("  Config preview:")
					for _, line := range strings.Split(updatedCfg, "\n") {
						fmt.Printf("    %s\n", line)
					}
					fmt.Println()
					fmt.Printf("  Writing config to: %s\n", configPath)
				}

				state := LoadState()
				state.ActiveProfile = profile
				state.Save()

				fmt.Println()
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

			nativePM := string(installers.DetectPackageManager())

			var cfg string
			switch profile {
			case "minimal":
				cfg = fmt.Sprintf(`version: 1
profile: minimal

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
    manager: fnm
    version: lts
  python:
    enabled: false
    manager: uv
    version: latest
  dotnet:
    enabled: false

tools: {}

editors: {}

ai_tools: {}
ai_apps: {}
`, nativePM)

				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			case "custom":
				cfg = fmt.Sprintf(`version: 1
profile: custom

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

ai_tools:
  codex: false
  opencode: false
  copilot: false
  claude: false
  gemini: false
ai_apps:
  codex: false
  claude: false
  copilot: false
`, nativePM)

				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			case "team":
				thinCfg, err := promptTeamURL()
				if err != nil {
					return err
				}
				if err := os.WriteFile(configPath, []byte(thinCfg), 0644); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			default:
				cfg = fmt.Sprintf(`version: 1
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
    enabled: true

editors:
  vscode: true
  zed: true
  cursor: optional

ai_tools:
  codex: true
  opencode: true
  copilot: true
  claude: false
  gemini: false
ai_apps:
  codex: true
  claude: false
  copilot: true
`, profile, nativePM)

				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			}

			state := LoadState()
			state.ActiveProfile = profile
			state.Save()

			fmt.Println()
			fmt.Println("  Config preview:")
			if data, err := os.ReadFile(configPath); err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					fmt.Printf("    %s\n", line)
				}
			}
			fmt.Println()
			fmt.Printf("  Writing config to: %s\n", configPath)
			fmt.Println()
			if profile == "custom" {
				fmt.Printf("  ✓ Config created at %s\n", configPath)
				fmt.Println()
				fmt.Println("  Next steps:")
				fmt.Println("    1. Edit the config file to customize your toolset")
				fmt.Println("    2. Run 'dk apply' to install tools")
				fmt.Println("    3. Run 'dk doctor' to verify")
			} else if profile == "team" {
				fmt.Printf("  ✓ Config created at %s\n", configPath)
				fmt.Println()
				fmt.Println("  Next steps:")
				fmt.Println("    1. Run 'dk apply' to install tools (fetches remote config)")
				fmt.Println("    2. Run 'dk doctor' to verify")
			} else {
				fmt.Println("  ✓ Config created")
				fmt.Println()
				fmt.Println("  Next steps:")
				fmt.Println("    1. Run 'dk apply' to install tools")
				fmt.Println("    2. Run 'dk doctor' to verify")
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to initialize (minimal, default, ai, custom, team)")

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

func promptTeamURL() (string, error) {
	var url string
	urlPrompt := &survey.Input{
		Message: "Team config URL:",
	}
	if err := survey.AskOne(urlPrompt, &url, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	fmt.Printf("    Fetching remote config from %s...\n", url)

	remoteCfg, err := config.FetchRemote(url)
	if err != nil {
		fmt.Printf("    ✗ %s\n", err)
		return "", fmt.Errorf("failed to load remote config")
	}

	_ = remoteCfg
	fmt.Println("    ✓ Remote config validated")

	thinCfg := fmt.Sprintf("profile: team\nteam:\n  url: %s\n", url)
	return thinCfg, nil
}
