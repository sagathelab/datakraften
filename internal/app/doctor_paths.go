package app

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sagathelab/datakraften/internal/system"
)

func doctorExecutable(paths ...string) string {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && info.Mode().IsRegular() && syscall.Access(path, 1) == nil {
			return path
		}
	}
	return ""
}

func doctorBrewOutsidePath() string {
	return doctorExecutable(
		filepath.Join(system.HomeDir(), ".linuxbrew/bin/brew"),
		"/home/linuxbrew/.linuxbrew/bin/brew",
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
	)
}

func doctorGoOutsidePath() string {
	return doctorExecutable(
		filepath.Join(system.HomeDir(), ".linuxbrew/bin/go"),
		"/usr/local/go/bin/go",
		"/home/linuxbrew/.linuxbrew/bin/go",
		"/opt/homebrew/bin/go",
		"/usr/local/bin/go",
	)
}

func doctorShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func doctorBrewPathCommand(path string) string {
	if filepath.Base(system.CurrentShell()) == "fish" {
		return doctorShellQuote(path) + " shellenv | source"
	}
	return "eval \"$(" + doctorShellQuote(path) + " shellenv)\""
}

func doctorGoPathCommand(path string) string {
	dir := doctorShellQuote(filepath.Dir(path))
	if filepath.Base(system.CurrentShell()) == "fish" {
		return "fish_add_path " + dir
	}
	return "export PATH=" + dir + ":\"$PATH\""
}
