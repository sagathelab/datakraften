package app

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/doctor"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var fix bool
	var category string

	cmd := &cobra.Command{
		Use:          "doctor",
		Short:        "Diagnose your development environment",
		Long:         `Report configured intent and check WSL status, tools, runtimes, editors, Docker, and shell setup.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch category {
			case "", "system", "tools", "runtimes", "editors", "docker", "shell":
			default:
				return fmt.Errorf("unknown doctor category; choose system, tools, runtimes, editors, docker or shell")
			}
			workstation, cfg, err := loadWorkstationReport()
			if err != nil {
				return err
			}
			printConfigWarnings(workstation.Configuration.Warnings)
			report := &doctorReport{workstationReport: workstation}
			if !jsonOutput {
				displayTitle(os.Stdout, "doctor", "Environment diagnostics")
				workstation.print(os.Stdout)
			}
			buildDoctorReport(report, category, cfg)
			report.SummaryText = report.Summary()
			if jsonOutput {
				return printReportJSON(report)
			}

			displayHeading(os.Stdout, "Summary")
			fmt.Println("  -------")
			fmt.Println("    " + report.SummaryText)
			fmt.Println()

			if fix {
				displayHeading(os.Stdout, "Auto-fix")
				runAutoFix(&report.Report, category)
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

type doctorReport struct {
	workstationReport
	doctor.Report
	Shell       *detectedShells `json:"shell,omitempty"`
	SummaryText string          `json:"summary"`
}

func buildDoctorReport(r *doctorReport, category string, cfg *config.Config) {
	for _, section := range []struct {
		name string
		run  func()
	}{
		{"system", func() { RunDoctorSystem(&r.Report) }},
		{"tools", func() { RunDoctorTools(&r.Report, cfg) }},
		{"runtimes", func() { RunDoctorRuntimes(&r.Report, cfg) }},
		{"editors", func() { RunDoctorEditors(&r.Report) }},
		{"docker", func() { RunDoctorDocker(&r.Report) }},
		{"shell", func() {
			detected := detectShells()
			r.Shell = &detected
			runDoctorShell(&r.Report, detected, r.Configuration.Shell)
		}},
	} {
		if category == "" || category == section.name {
			section.run()
			fmt.Fprintln(doctorOutput())
		}
	}
}

func runDoctorShell(r *doctor.Report, detected detectedShells, target config.ShellTarget) {
	displayHeading(doctorOutput(), "Detected shell")
	addCheck := func(check doctor.Check) {
		check.Category, check.Severity = "shell", "warning"
		r.Add(check)
		label := map[string]string{"pass": "OK", "skip": "SKIP", "fail": "WARN"}[check.Status]
		displayFprintf(doctorOutput(), "    [%s] %s: %s\n", label, check.Title, check.Message)
		if check.Fix != "" {
			displayFprintf(doctorOutput(), "           %s\n", check.Fix)
		}
	}
	addCheck(doctor.Check{ID: "shell", Title: "Current shell ($SHELL)",
		Status: skipStatus(detected.Current != ""), Message: detected.currentLabel()})
	for _, name := range slices.Sorted(maps.Keys(detected.Targets)) {
		item := detected.Targets[name]
		status, fix := skipStatus(item.ManagedBlock == "present"), ""
		if name == target {
			switch item.ManagedBlock {
			case "absent":
				fix = "run 'dk apply' to configure this shell"
			case "invalid", "unreadable":
				status, fix = "fail", "Check the startup file and its managed block markers before running 'dk apply'"
			}
			if !item.Installed && item.ManagedBlock != "invalid" && item.ManagedBlock != "unreadable" {
				status = "fail"
				if name == config.ShellFish {
					fix = "run 'dk apply' to ensure Fish is installed and configured"
				} else if item.ManagedBlock == "absent" {
					fix = "Install " + string(name) + " separately and run 'dk apply' to configure it"
				} else {
					fix = "Install " + string(name) + " separately; dk apply only configures it"
				}
			}
		}
		addCheck(doctor.Check{ID: "shell." + string(name), Title: string(name),
			Status: status, Message: item.description(), Fix: fix})
	}
}
