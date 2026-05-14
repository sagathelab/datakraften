package app

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var dryRun bool
	var yes bool
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply Datakraften configuration",
		Long:  `Install missing tools, configure shell, runtimes, editors, and AI tooling based on the current profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := LoadState()

			cfg, err := config.Load()
			if err != nil {
				fmt.Println("  No configuration found. Run 'dk init' first.")
				fmt.Println()
				return nil
			}

			activeProfile := cfg.Profile
			if profileFlag != "" {
				activeProfile = profileFlag
			}

			if dryRun {
				fmt.Printf("  Datakraften apply (dry-run) — profile: %s\n", activeProfile)
			} else {
				fmt.Printf("  Datakraften apply — profile: %s\n", activeProfile)
				fmt.Println()
			}

			if !yes && !dryRun {
				fmt.Print("  Continue? [Y/n]: ")
				var input string
				fmt.Scanln(&input)
				if input == "n" || input == "N" || input == "no" {
					fmt.Println("  Cancelled.")
					return nil
				}
			}

			report := RunApply(cfg, dryRun)

			if dryRun {
				fmt.Println("  (dry-run — no changes made)")
				return nil
			}

			state.RecordApply(activeProfile)
			WriteLog("apply", fmt.Sprintf("Applied profile: %s\n", activeProfile))

			fmt.Println("  Summary")
			fmt.Println("  -------")
			fmt.Printf("    System packages: %d installed\n", report.System)
			fmt.Printf("    Brew packages: %d installed\n", report.BrewPkgs)
			if report.Node {
				fmt.Println("    Node.js: installed")
			}
			if report.Python {
				fmt.Println("    Python: installed")
			}
			if report.Dotnet {
				fmt.Println("    .NET: installed")
			}
			if report.AITools > 0 {
				fmt.Printf("    AI tools: %d installed\n", report.AITools)
			}
			if len(report.Errors) > 0 {
				fmt.Println()
				fmt.Println("  Errors:")
				for _, e := range report.Errors {
					fmt.Printf("    ✗ %s\n", e)
				}
			}

			fmt.Println()
			fmt.Println("  Next steps:")
			fmt.Println("    1. Restart your shell or run: exec fish")
			fmt.Println("    2. Run 'dk doctor' to verify everything")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to apply")

	return cmd
}
