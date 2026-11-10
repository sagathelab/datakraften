package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var dryRun bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply Datakraften configuration",
		Long:  `Install missing tools, configure shell, runtimes, editors, and AI tooling based on your configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := LoadState()

			cfg, err := config.Load()
			if err != nil {
				fmt.Println("  No configuration found. Run 'dk init' first.")
				fmt.Println()
				return nil
			}

			subtitle := fmt.Sprintf("Source: %s", cfg.Source)
			if dryRun {
				subtitle += " · preview only"
			}
			displayTitle(os.Stdout, "apply", subtitle)

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
				if len(report.Errors) > 0 {
					fmt.Println("  Preview stopped due to errors; no changes made")
				} else {
					fmt.Println("  Preview complete; no changes made")
				}
				return nil
			}

			state.RecordApply(cfg.Source)
			WriteLog("apply", fmt.Sprintf("Applied source: %s\n", cfg.Source))

			if len(report.Errors) > 0 {
				fmt.Printf("  Apply finished with %d error(s)\n", len(report.Errors))
			} else {
				fmt.Println("  Apply complete")
			}
			fmt.Println("  Summary")
			if len(report.System) > 0 {
				fmt.Printf("    System packages: %s\n", strings.Join(report.System, ", "))
			} else if report.SystemChecked {
				fmt.Println("    System packages: (already satisfied)")
			} else {
				fmt.Println("    System packages: not completed")
			}
			if len(report.BrewPkgs) > 0 {
				fmt.Printf("    Brew packages: %s\n", strings.Join(report.BrewPkgs, ", "))
			} else if report.BrewChecked {
				fmt.Println("    Brew packages: (already installed)")
			} else {
				fmt.Println("    Brew packages: not completed")
			}
			if report.NodeVer != "" {
				fmt.Printf("    Node.js %s (%s)\n", report.NodeVer, report.Node)
			}
			if report.PythonVer != "" {
				fmt.Printf("    Python %s (%s)\n", report.PythonVer, report.Python)
			}
			if report.GoVer != "" {
				fmt.Printf("    Go %s (%s)\n", report.GoVer, report.Go)
			}
			if report.DotnetVer != "" {
				fmt.Printf("    .NET SDK %s (%s)\n", report.DotnetVer, report.Dotnet)
			}
			if report.Started {
				fmt.Printf("    AI tools: %s\n", installReportSummary(report.AITools))
				fmt.Printf("    AI apps: %s\n", installReportSummary(report.AIApps))
			} else {
				fmt.Println("    AI tools and apps: not checked")
			}
			if len(report.Errors) > 0 {
				fmt.Println()
				fmt.Println("  Errors:")
				for _, e := range report.Errors {
					displayPrintf("    [ERROR] %s\n", e)
				}
			}

			fmt.Println()
			fmt.Println("  Next steps:")
			if report.Shell {
				fmt.Println("    Restart your shell or run: exec fish")
			}
			fmt.Println("    Run 'dk doctor' to verify everything")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")

	return cmd
}
