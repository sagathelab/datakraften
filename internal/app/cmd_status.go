package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show a friendly overview of your environment",
		Long:  `Display a summary of installed tools, runtimes, AI tools, and editors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Datakraften status")
			fmt.Println()
			fmt.Println("  WSL")
			fmt.Println("    ✓ WSL2 detected")
			fmt.Println("    ✓ Ubuntu 24.04")
			fmt.Println("    ✓ systemd available")
			fmt.Println()
			fmt.Println("  Tools")
			fmt.Println("    ✓ git")
			fmt.Println("    – gh not found")
			fmt.Println("    – az not found")
			fmt.Println("    – docker not found")
			fmt.Println()
			fmt.Println("  Runtimes")
			fmt.Println("    – Node.js not installed")
			fmt.Println("    – Python not installed")
			fmt.Println("    – .NET not installed")
			fmt.Println()
			fmt.Println("  AI")
			fmt.Println("    – codex not installed")
			fmt.Println("    – opencode not installed")
			fmt.Println()
			fmt.Println("  Editors")
			fmt.Println("    – code not found")
			fmt.Println("    – zed not found")
			fmt.Println()
			fmt.Println("  (full status checks not yet implemented)")
			fmt.Println()
			fmt.Println("  Run 'dk doctor' for detailed diagnostics.")

			return nil
		},
	}

	return cmd
}
