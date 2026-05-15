package app

import (
	"fmt"
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

			fmt.Println("  Datakraften Doctor")
			fmt.Println()

			if state.LastApply != "" {
				fmt.Printf("  Last apply: %s\n", state.LastApply)
				fmt.Printf("  Source: %s", state.ActiveSource)
				fmt.Println()
				fmt.Println()
			}

			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()

			if category == "" || category == "system" {
				RunDoctorSystem(report)

				if wsl {
					fmt.Println()
					fmt.Println("  WSL Details")
					fmt.Printf("    ✓ WSL%d\n", wslVer)
					fmt.Printf("    ✓ %s %s\n", distro, distroVer)
				}
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
				fmt.Println("  Shell")
				fmt.Printf("    ✓ Current shell: %s\n", system.CurrentShell())
				if state.ManagedShell {
					fmt.Println("    ✓ Managed config active")
				} else {
					fmt.Println("    – Run 'dk apply' to configure shell")
				}
				fmt.Println()
			}

			if fix {
				fmt.Println("  Auto-fix mode")
				fmt.Println("  -------------")
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
			fmt.Printf("    – %s: no auto-fix available\n", c.Title)
			skipped++
			continue
		}

		if strings.HasPrefix(c.Fix, "run '") {
			fmt.Printf("    → %s: %s\n", c.Title, c.Fix)
			skipped++
			continue
		}

		if strings.HasPrefix(c.Fix, "Install") || strings.HasPrefix(c.Fix, "Start") || strings.HasPrefix(c.Fix, "Enable") || strings.HasPrefix(c.Fix, "If") {
			fmt.Printf("    → %s: %s\n", c.Title, c.Fix)
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
					fmt.Printf("    ✓ %s fixed\n", c.Title)
					fixed++
				} else {
					fmt.Printf("    ✗ %s fix failed: %s\n", c.Title, r.Stderr)
					skipped++
				}
				continue
			}
		}

		fmt.Printf("    → %s: %s\n", c.Title, c.Fix)
		skipped++
	}

	if fixed == 0 && skipped == 0 {
		fmt.Println("    ✓ No issues to fix")
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
		r.Add(doctor.Check{ID: "shell", Title: "Current shell", Category: "shell",
			Severity: "info", Status: "pass", Message: system.CurrentShell()})
	}
}
