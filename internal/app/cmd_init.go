package app

import (
	"fmt"
	"io"
	"net/http"
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
					updatedCfg := strings.Join(lines, "\n")

					if profile == "team" {
						useRemote, remoteCfg := promptTeamURL()
						if useRemote {
							updatedCfg = remoteCfg
						}
					}

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
				useRemote, remoteCfg := promptTeamURL()
				if useRemote {
					cfg = remoteCfg
				} else {
					cfg = fmt.Sprintf(`version: 1
profile: team

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
`, nativePM)
				}

				if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
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
			if profile == "custom" || profile == "team" {
				fmt.Printf("  ✓ Config created at %s\n", configPath)
				fmt.Println()
				fmt.Println("  Next steps:")
				fmt.Println("    1. Edit the config file to customize your toolset")
				fmt.Println("    2. Run 'dk apply' to install tools")
				fmt.Println("    3. Run 'dk doctor' to verify")
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

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to initialize (minimal, default, custom, team)")

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

func promptTeamURL() (bool, string) {
	useRemote := false
	prompt := &survey.Confirm{
		Message: "Use team config from a remote YAML URL?",
		Default: false,
	}
	survey.AskOne(prompt, &useRemote)

	if !useRemote {
		return false, ""
	}

	var url string
	urlPrompt := &survey.Input{
		Message: "Remote config URL:",
	}
	survey.AskOne(urlPrompt, &url)

	if url == "" {
		fmt.Println("  No URL provided. Using local config as-is.")
		return false, ""
	}

	fmt.Printf("    Fetching remote config from %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("    ✗ Failed to fetch remote config: %s\n", err)
		return false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("    ✗ Remote config returned status %d\n", resp.StatusCode)
		fmt.Println("  Using local config as-is.")
		return false, ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("    ✗ Failed to read remote config: %s\n", err)
		return false, ""
	}

	remoteCfg := fmt.Sprintf("profile: team\nteam:\n  url: %s\n%s", url, string(body))
	fmt.Println("    ✓ Remote config applied")
	return true, remoteCfg
}
