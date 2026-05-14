package runtimes

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
)

func UvInstalled() bool {
	return exec.CommandExists("uv")
}

func PythonInstalled() bool {
	return exec.CommandExists("python3") || exec.CommandExists("python")
}

func PythonVersion() string {
	cmd := "python3"
	if !exec.CommandExists(cmd) {
		cmd = "python"
		if !exec.CommandExists(cmd) {
			return ""
		}
	}
	r := exec.Run(cmd, "--version")
	return r.Stdout
}

func EnsurePython() (bool, error) {
	if PythonInstalled() {
		return false, nil
	}

	if !UvInstalled() {
		return false, fmt.Errorf("uv is required but not installed")
	}

	fmt.Println("    Installing Python via uv...")
	r := exec.Run("uv", "python", "install")
	if r.Code != 0 {
		return false, fmt.Errorf("uv python install failed: %s", r.Stderr)
	}

	return true, nil
}
