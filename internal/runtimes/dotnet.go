package runtimes

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
)

func DotnetInstalled() bool {
	return exec.CommandExists("dotnet")
}

func DotnetVersion() string {
	if !DotnetInstalled() {
		return ""
	}
	r := exec.Run("dotnet", "--version")
	return r.Stdout
}

func EnsureDotnet() (bool, error) {
	if DotnetInstalled() {
		return false, nil
	}

	if err := installers.BrewInstall("dotnet"); err != nil {
		return false, fmt.Errorf("dotnet install failed: %w", err)
	}

	return true, nil
}
