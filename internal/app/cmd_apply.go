package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	var dryRun bool
	var yes bool
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply Datakraften configuration",
		Long:  `Install missing tools, configure shell, runtimes, editors, and AI tooling based on the current profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				fmt.Println("Datakraften apply (dry-run)")
				fmt.Println()
				fmt.Println("  Would install and configure tools based on current profile.")
				fmt.Println("  (not yet implemented)")
				return nil
			}

			fmt.Println("Datakraften apply")
			fmt.Println()
			fmt.Println("  Applying configuration...")
			if profileFlag != "" {
				fmt.Printf("  Profile: %s\n", profileFlag)
			}
			fmt.Println()
			fmt.Println("  (not yet implemented)")
			fmt.Println()
			fmt.Println("  Run 'dk doctor' to check the status.")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile to apply")

	return cmd
}
