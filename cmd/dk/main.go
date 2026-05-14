package main

import (
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
		os.Exit(1)
	}
}
