package app

import (
	"fmt"

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

			fmt.Println("  Datakraften Doctor")
			fmt.Println()

			if state.LastApply != "" {
				fmt.Printf("  Last apply: %s\n", state.LastApply)
				fmt.Printf("  Profile: %s", state.ActiveProfile)
				fmt.Println()
				fmt.Println()
			}

			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()

			if category == "" || category == "system" {
				RunDoctorSystem()

				if wsl {
					fmt.Println()
					fmt.Println("  WSL Details")
					fmt.Printf("    ✓ WSL%d\n", wslVer)
					fmt.Printf("    ✓ %s %s\n", distro, distroVer)
				}
				fmt.Println()
			}

			if category == "" || category == "tools" {
				RunDoctorTools()
				fmt.Println()
			}

			if category == "" || category == "runtimes" {
				RunDoctorRuntimes()
				fmt.Println()
			}

			if category == "" || category == "editors" {
				RunDoctorEditors()
				fmt.Println()
			}

			if category == "" || category == "docker" {
				RunDoctorDocker()
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
				fmt.Println("  Auto-fix mode (not yet implemented)")
				fmt.Println("  Run 'dk apply' to apply the full configuration.")
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&fix, "fix", "f", false, "Attempt to fix detected issues")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Only check a specific category (system, tools, runtimes, shell, docker, editors, ai, auth)")

	return cmd
}
