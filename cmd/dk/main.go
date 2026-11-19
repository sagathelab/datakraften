package main

import (
	"errors"
	"os"

	"github.com/sagathelab/datakraften/internal/app"
)

var (
	Version = "dev"
	Commit  = "none"
)

func main() {
	app.SetVersionInfo(Version, Commit)
	if err := app.Execute(); err != nil {
		var exit interface{ ExitCode() int }
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
}
