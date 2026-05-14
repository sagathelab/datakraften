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
			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()
			systemd := system.HasSystemd()
			shell := system.CurrentShell()

			fmt.Println("Datakraften Doctor")
			fmt.Println()

			fmt.Println("System")
			if wsl {
				fmt.Printf("  ✓ WSL%d detected\n", wslVer)
			} else {
				fmt.Println("  – Not running in WSL")
			}
			if distro != "" {
				fmt.Printf("  ✓ %s %s\n", distro, distroVer)
			}
			if systemd {
				fmt.Println("  ✓ systemd available")
			} else {
				fmt.Println("  – systemd not detected")
			}
			if shell != "" {
				fmt.Printf("  ✓ Shell: %s\n", shell)
			}
			fmt.Println()

			if category != "" {
				fmt.Printf("  Category filter: %s\n", category)
				fmt.Printf("  (filtered checks not yet implemented)\n")
				fmt.Println()
			}

			fmt.Println("  (full doctor checks not yet implemented)")
			fmt.Println()
			fmt.Println("  Run 'dk doctor --fix' to attempt automatic fixes.")
			fmt.Println("  Run 'dk doctor --json' for machine-readable output.")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&fix, "fix", "f", false, "Attempt to fix detected issues")
	cmd.Flags().StringVarP(&category, "category", "c", "", "Only check a specific category (system, tools, runtimes, shell, docker, editors, ai, auth)")

	return cmd
}
