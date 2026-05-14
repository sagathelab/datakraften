package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	verbose   bool
	jsonOutput bool
)

type App struct {
	RootCmd *cobra.Command
}

func SetVersionInfo(v, c string) {
	version = v
	commit = c
}

func newApp() *App {
	a := &App{}

	rootCmd := &cobra.Command{
		Use:     "dk",
		Short:   "Datakraften — The WSL-first developer workstation platform",
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Name() != "dk" && cmd.Parent() != nil && cmd.Parent().Name() == "dk" {
				// Only set on root command
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "JSON output")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newProfileCmd())

	a.RootCmd = rootCmd
	return a
}

func Execute() error {
	a := newApp()
	return a.RootCmd.Execute()
}
