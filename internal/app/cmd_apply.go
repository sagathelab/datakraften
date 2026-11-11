package app

import (
	"bufio"
	"fmt"
	"io"
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
		RunE: func(cmd *cobra.Command, args []string) (runErr error) {
			state := LoadState()

			cfg, err := config.Load()
			if err != nil {
				fmt.Println("  No configuration found. Run 'dk init' first.")
				fmt.Println()
				return fmt.Errorf("cannot load configuration: %w", err)
			}

			subtitle := fmt.Sprintf("Source: %s", cfg.Source)
			if dryRun {
				subtitle += " · preview only"
			}
			displayTitle(os.Stdout, "apply", subtitle)

			if !yes && !dryRun {
				confirmed, err := confirmApply(os.Stdin, os.Stdout)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Println("  Cancelled.")
					return nil
				}
			}

			output := newLineStartProgressOutput(os.Stdout)
			restore := output.activate()
			defer restore()
			defer func() { runErr = finishProgress(output, runErr) }()
			report := RunApplyWithProgress(cfg, dryRun, output.event)

			if dryRun {
				if len(report.Errors) > 0 {
					fmt.Fprintln(progressWriter(), "  Preview stopped due to errors; no changes made")
					return fmt.Errorf("preview failed: %s", strings.Join(report.Errors, "; "))
				} else {
					fmt.Fprintln(progressWriter(), "  Preview complete; no changes made")
				}
				return nil
			}

			if len(report.Errors) == 0 {
				state.RecordApply(cfg.Source)
				WriteLog("apply", fmt.Sprintf("Applied source: %s\n", cfg.Source))
			}

			if len(report.Errors) > 0 {
				fmt.Fprintf(progressWriter(), "  Apply finished with %d error(s)\n", len(report.Errors))
			} else {
				fmt.Fprintln(progressWriter(), "  Apply complete")
			}
			fmt.Fprintln(progressWriter(), "  Summary")
			if len(report.System) > 0 {
				fmt.Fprintf(progressWriter(), "    System packages: %s\n", strings.Join(report.System, ", "))
			} else if report.SystemChecked {
				fmt.Fprintln(progressWriter(), "    System packages: (already satisfied)")
			} else {
				fmt.Fprintln(progressWriter(), "    System packages: not completed")
			}
			if len(report.BrewPkgs) > 0 {
				fmt.Fprintf(progressWriter(), "    Brew packages: %s\n", strings.Join(report.BrewPkgs, ", "))
			} else if report.BrewChecked {
				fmt.Fprintln(progressWriter(), "    Brew packages: (already installed)")
			} else {
				fmt.Fprintln(progressWriter(), "    Brew packages: not completed")
			}
			if report.NodeVer != "" {
				fmt.Fprintf(progressWriter(), "    Node.js %s (%s)\n", report.NodeVer, report.Node)
			}
			if report.PythonVer != "" {
				fmt.Fprintf(progressWriter(), "    Python %s (%s)\n", report.PythonVer, report.Python)
			}
			if report.GoVer != "" {
				fmt.Fprintf(progressWriter(), "    Go %s (%s)\n", report.GoVer, report.Go)
			}
			if report.DotnetVer != "" {
				fmt.Fprintf(progressWriter(), "    .NET SDK %s (%s)\n", report.DotnetVer, report.Dotnet)
			}
			if report.Started {
				fmt.Fprintf(progressWriter(), "    AI tools: %s\n", installReportSummary(report.AITools))
				fmt.Fprintf(progressWriter(), "    AI apps: %s\n", installReportSummary(report.AIApps))
			} else {
				fmt.Fprintln(progressWriter(), "    AI tools and apps: not checked")
			}
			if len(report.Errors) > 0 {
				fmt.Fprintln(progressWriter())
				fmt.Fprintln(progressWriter(), "  Errors:")
				for _, e := range report.Errors {
					displayPrintf("    [ERROR] %s\n", e)
				}
			}

			fmt.Fprintln(progressWriter())
			fmt.Fprintln(progressWriter(), "  Next steps:")
			if report.Shell {
				fmt.Fprintln(progressWriter(), "    Restart your shell or run: exec fish")
			}
			fmt.Fprintln(progressWriter(), "    Run 'dk doctor' to verify everything")
			fmt.Fprintln(progressWriter())

			if len(report.Errors) > 0 {
				return fmt.Errorf("apply failed: %s", strings.Join(report.Errors, "; "))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")

	return cmd
}

func confirmApply(input io.Reader, output io.Writer) (bool, error) {
	fmt.Fprint(output, "  Continue? [Y/n]: ")
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && !(err == io.EOF && strings.TrimSpace(line) != "") {
		return false, fmt.Errorf("cannot read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid confirmation; answer yes or no")
	}
}
