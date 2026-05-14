package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/profiles"
	"github.com/spf13/cobra"
	"github.com/AlecAivazis/survey/v2"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage Datakraften profiles",
		Long:  `List available profiles and switch between them.`,
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileUseCmd())

	return cmd
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			all := profiles.All()
			fmt.Println("  Available profiles:")
			fmt.Println()
			for _, p := range all {
				fmt.Printf("    %s  — %s\n", p.Name, p.Description)
			}
			fmt.Println()
			fmt.Println("  Use 'dk profile use <name>' to switch.")

			return nil
		},
	}
}

func newProfileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "use [profile]",
		Short:     "Switch to a profile",
		ValidArgs: profiles.Available(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			} else {
				selected, err := selectProfile()
				if err != nil {
					return err
				}
				name = selected
			}

			home, _ := os.UserHomeDir()
			configPath := filepath.Join(home, ".config", "datakraften", "config.yaml")

			data, err := os.ReadFile(configPath)
			if err != nil {
				fmt.Println("  No config found. Run 'dk init' first.")
				return nil
			}

			lines := strings.Split(string(data), "\n")
			replaced := false
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "profile:") {
					lines[i] = fmt.Sprintf("profile: %s", name)
					replaced = true
					break
				}
			}

			if !replaced {
				lines = append([]string{fmt.Sprintf("profile: %s", name)}, lines...)
			}

			if err := os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			if name == "custom" {
				cfg, err := config.Load()
				if err == nil && cfg.Custom.URL == "" {
					useRemote := false
					prompt := &survey.Confirm{
						Message: "Add a remote team config (YAML URL)?",
						Default: false,
					}
					survey.AskOne(prompt, &useRemote)

					if useRemote {
						var url string
						urlPrompt := &survey.Input{
							Message: "Remote config URL:",
						}
						survey.AskOne(urlPrompt, &url)

						if url != "" {
							fmt.Printf("    Fetching remote config from %s...\n", url)
							resp, fetchErr := http.Get(url)
							if fetchErr == nil && resp.StatusCode == http.StatusOK {
								body, readErr := io.ReadAll(resp.Body)
								resp.Body.Close()
								if readErr == nil {
									remoteCfg := fmt.Sprintf("profile: custom\ncustom:\n  url: %s\n%s", url, string(body))
									os.WriteFile(configPath, []byte(remoteCfg), 0644)
									fmt.Println("    ✓ Remote config applied")
								}
							}
						}
					}
				}
			}

			state := LoadState()
			state.ActiveProfile = name
			state.Save()

			fmt.Printf("  Switched to profile: %s\n", name)
			if desc := profiles.Describe(name); desc != "" {
				fmt.Printf("  %s\n", desc)
			}
			fmt.Println()
			fmt.Println("  Run 'dk apply' to apply this profile.")

			return nil
		},
	}
}
