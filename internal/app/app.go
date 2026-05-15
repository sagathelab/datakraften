package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	commit     = "none"
	verbose    bool
	jsonOutput bool
)

func verbosePrintf(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "  [verbose] "+format, args...)
	}
}

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
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "JSON output")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newStatusCmd())
	// profile command removed in favor of dk init --custom and --team
	rootCmd.AddCommand(newUpgradeCmd())
	rootCmd.AddCommand(newUpdateCmd())

	a.RootCmd = rootCmd
	return a
}

func Execute() error {
	a := newApp()
	return a.RootCmd.Execute()
}
