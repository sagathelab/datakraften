package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/sagathelab/datakraften/internal/doctor"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var fix bool
	var category string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose your development environment",
		Long:  `Check WSL status, tools, runtimes, editors, Docker, AI tooling, and more.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := LoadState()
			report := &doctor.Report{}

			if jsonOutput {
				buildDoctorReport(report, category)
				report.PrintJSON()
				return nil
			}

			displayTitle(os.Stdout, "doctor", "Environment diagnostics")

			if state.LastApply != "" {
				fmt.Printf("  Last apply: %s\n", state.LastApply)
				fmt.Printf("  Source: %s", state.ActiveSource)
				fmt.Println()
				fmt.Println()
			}

			if category == "" || category == "system" {
				RunDoctorSystem(report)
				fmt.Println()
			}

			if category == "" || category == "tools" {
				RunDoctorTools(report)
				fmt.Println()
			}

			if category == "" || category == "runtimes" {
				RunDoctorRuntimes(report)
				fmt.Println()
			}

			if category == "" || category == "editors" {
				RunDoctorEditors(report)
				fmt.Println()
			}

			if category == "" || category == "docker" {
				RunDoctorDocker(report)
				fmt.Println()
			}

			if category == "" || category == "shell" {
				RunDoctorShell(report, state)
				fmt.Println()
			}

			displayHeading(os.Stdout, "Summary")
			fmt.Println("  -------")
			fmt.Println("    " + report.Summary())
			fmt.Println()

			if fix {
				displayHeading(os.Stdout, "Auto-fix")
				runAutoFix(report, category)
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&fix, "fix", "f", false, "Attempt to fix detected issues")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Only check a specific category (system, tools, runtimes, shell, docker, editors)")

	return cmd
}

func runAutoFix(r *doctor.Report, category string) {
	fixed := 0
	skipped := 0

	for _, c := range r.Checks {
		if category != "" && c.Category != category {
			continue
		}
		if c.Status == "pass" {
			continue
		}
		if c.Fix == "" {
			displayPrintf("    [SKIP] %s: no auto-fix available\n", c.Title)
			skipped++
			continue
		}

		if strings.HasPrefix(c.Fix, "run '") {
			displayPrintf("    [ACTION] %s: %s\n", c.Title, c.Fix)
			skipped++
			continue
		}

		if strings.HasPrefix(c.Fix, "Install") || strings.HasPrefix(c.Fix, "Start") || strings.HasPrefix(c.Fix, "Enable") || strings.HasPrefix(c.Fix, "If") {
			displayPrintf("    [ACTION] %s: %s\n", c.Title, c.Fix)
			skipped++
			continue
		}

		if strings.HasPrefix(c.Fix, "sudo ") || strings.Contains(c.Fix, "apt-get") || strings.Contains(c.Fix, "brew ") {
			parts := strings.Fields(c.Fix)
			if len(parts) >= 2 {
				cmd := parts[0]
				args := parts[1:]
				fmt.Printf("    Attempting: %s\n", c.Fix)
				r := exec.Run(cmd, args...)
				if r.Code == 0 {
					displayPrintf("    [OK] %s fixed\n", c.Title)
					fixed++
				} else {
					displayPrintf("    [ERROR] %s fix failed: %s\n", c.Title, r.Stderr)
					skipped++
				}
				continue
			}
		}

		displayPrintf("    [ACTION] %s: %s\n", c.Title, c.Fix)
		skipped++
	}

	if fixed == 0 && skipped == 0 {
		displayPrintln("    [OK] No issues to fix")
	} else {
		fmt.Printf("    Fixed: %d, Manual action needed: %d\n", fixed, skipped)
	}
}

func buildDoctorReport(r *doctor.Report, category string) {
	if category == "" || category == "system" {
		RunDoctorSystem(r)
	}
	if category == "" || category == "tools" {
		RunDoctorTools(r)
	}
	if category == "" || category == "runtimes" {
		RunDoctorRuntimes(r)
	}
	if category == "" || category == "editors" {
		RunDoctorEditors(r)
	}
	if category == "" || category == "docker" {
		RunDoctorDocker(r)
	}
	if category == "" || category == "shell" {
		RunDoctorShell(r, LoadState())
	}
}

func RunDoctorShell(r *doctor.Report, state *State) {
	current := system.CurrentShell()
	r.Add(doctor.Check{ID: "shell", Title: "Current shell", Category: "shell",
		Severity: "info", Status: skipStatus(current != ""), Message: current})
	r.Add(doctor.Check{ID: "shell.managed", Title: "Managed Fish config", Category: "shell",
		Severity: "info", Status: skipStatus(state.ManagedShell), Fix: "run 'dk apply' to configure shell"})
	displayHeading(doctorOutput(), "Shell")
	if current != "" {
		displayFprintf(doctorOutput(), "    [OK] Current shell: %s\n", current)
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Current shell unknown")
	}
	if state.ManagedShell {
		displayFprintln(doctorOutput(), "    [OK] Managed Fish config saved")
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Run 'dk apply' to configure shell")
	}
}
