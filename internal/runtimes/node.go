package runtimes

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
)

func FnmInstalled() bool {
	return exec.CommandExists("fnm")
}

func NodeInstalled() bool {
	return exec.CommandExists("node")
}

func NodeVersion() string {
	if !NodeInstalled() {
		return ""
	}
	r := exec.Run("node", "--version")
	return r.Stdout
}

func EnsureNode() (bool, error) {
	if NodeInstalled() {
		return false, nil
	}

	if !FnmInstalled() {
		return false, fmt.Errorf("fnm is required but not installed")
	}

	fmt.Println("    Installing Node.js LTS via fnm...")
	r := exec.Run("fnm", "install", "--lts")
	if r.Code != 0 {
		return false, fmt.Errorf("fnm install lts failed: %s", r.Stderr)
	}

	r2 := exec.Run("fnm", "default", "lts-latest")
	if r2.Code != 0 {
		return false, fmt.Errorf("fnm default failed: %s", r2.Stderr)
	}

	return true, nil
}
