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
		Use:          "apply",
		Short:        "Apply Datakraften configuration",
		Long:         `Install missing tools, configure shell, runtimes, editors, and AI tooling based on your configuration.`,
		SilenceUsage: true,
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
					cmd.SilenceErrors = true
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
				displayEmphasis(progressWriter(), fmt.Sprintf("  ! Apply finished with %d error(s)", len(report.Errors)), "1;31")
			} else {
				displayComplete(progressWriter(), "Apply complete")
			}
			fmt.Fprintln(progressWriter())
			displayHeading(progressWriter(), "Summary")
			fmt.Fprintln(progressWriter(), "  -------")
			row := func(label, value string) {
				fmt.Fprintf(progressWriter(), "    %-16s %s\n", label, value)
			}
			if len(report.System) > 0 {
				row("System packages", strings.Join(report.System, ", "))
			} else if report.SystemChecked {
				row("System packages", "already satisfied")
			} else {
				row("System packages", "not completed")
			}
			if len(report.BrewPkgs) > 0 {
				row("Brew packages", strings.Join(report.BrewPkgs, ", "))
			} else if report.BrewChecked {
				row("Brew packages", "already installed")
			} else {
				row("Brew packages", "not completed")
			}
			if report.NodeVer != "" {
				row("Node.js", fmt.Sprintf("%s (%s)", report.NodeVer, report.Node))
			}
			if report.PythonVer != "" {
				row("Python", fmt.Sprintf("%s (%s)", report.PythonVer, report.Python))
			}
			if report.GoVer != "" {
				row("Go", fmt.Sprintf("%s (%s)", report.GoVer, report.Go))
			}
			if report.DotnetVer != "" {
				row(".NET SDK", fmt.Sprintf("%s (%s)", report.DotnetVer, report.Dotnet))
			}
			if report.Started {
				row("AI tools", installSummaryWithoutDetails(report.AITools))
				row("AI apps", installSummaryWithoutDetails(report.AIApps))
			} else {
				fmt.Fprintln(progressWriter(), "    AI tools and apps: not checked")
			}
			fmt.Fprintln(progressWriter())
			displayNextSteps(progressWriter())
			if report.Shell {
				fmt.Fprintln(progressWriter(), "    Restart your shell or run: exec fish")
			}
			fmt.Fprintln(progressWriter(), "    Run 'dk doctor' to verify everything")
			fmt.Fprintln(progressWriter())

			if len(report.Errors) > 0 {
				cmd.SilenceErrors = true
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
