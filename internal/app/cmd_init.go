package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Datakraften configuration",
		Long: `Initialize Datakraften by creating a configuration file
and detecting the current system state.

Profiles: minimal, default, ai, dotnet, frontend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Datakraften init")
			fmt.Println()
			fmt.Println("  Initializing Datakraften configuration...")
			fmt.Println()
			fmt.Println("  Detected:")
			fmt.Println("    ✓ WSL2")
			fmt.Println("    ✓ Ubuntu 24.04")
			fmt.Println()
			if profileFlag != "" {
				fmt.Printf("  Profile: %s\n", profileFlag)
			} else {
				fmt.Println("  No profile specified. Use --profile or run 'dk profile use'")
			}
			fmt.Println()
			fmt.Println("  Run 'dk apply' to apply the configuration.")
			fmt.Println("  Run 'dk doctor' to check your environment.")
			fmt.Println()
			fmt.Println("  (not yet implemented)")
			return nil
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to initialize (minimal, default, ai, dotnet, frontend)")

	return cmd
}
