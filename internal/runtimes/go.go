package runtimes

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
)

func GoInstalled() bool {
	return exec.CommandExists("go")
}

func GoVersion() string {
	if !GoInstalled() {
		return ""
	}
	r := exec.Run("go", "version")
	return r.Stdout
}

func EnsureGo() (bool, error) {
	if GoInstalled() {
		return false, nil
	}

	if err := installers.BrewInstall("go"); err != nil {
		return false, fmt.Errorf("go install failed: %w", err)
	}

	return true, nil
}
