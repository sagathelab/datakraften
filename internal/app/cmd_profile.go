package app

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/profiles"
	"github.com/spf13/cobra"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage Datakraften profiles",
		Long:  `List available profiles and switch between them.`,
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileUseCmd())

	return cmd
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			available := profiles.Available()
			fmt.Println("Available profiles:")
			fmt.Println()
			for _, p := range available {
				fmt.Printf("  %s\n", p)
			}
			fmt.Println()
			fmt.Println("  Use 'dk profile use <name>' to switch.")

			return nil
		},
	}
}

func newProfileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "use <profile>",
		Short:     "Switch to a profile",
		ValidArgs: profiles.Available(),
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			fmt.Printf("  Switched to profile: %s\n", name)
			fmt.Println()
			fmt.Println("  Run 'dk apply' to apply this profile.")
			return nil
		},
	}
}
