package runtimes

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
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

	fmt.Println("    Installing .NET SDK via Homebrew...")
	r := exec.Run("brew", "install", "dotnet")
	if r.Code != 0 {
		return false, fmt.Errorf("dotnet install failed: %s", r.Stderr)
	}

	return true, nil
}
