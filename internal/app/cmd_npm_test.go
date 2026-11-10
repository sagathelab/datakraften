package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNpmSetupIdempotent(t *testing.T) {
	home := t.TempDir()
	bin := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("SHELL", "/bin/bash")
	if err := os.WriteFile(filepath.Join(bin, "npm"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(path, []byte("existing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		cmd := newNpmCmd()
		cmd.SetArgs([]string{"setup"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "existing\n") || strings.Count(string(data), "# >>> datakraften npm >>>") != 1 {
		t.Fatalf("unexpected shell config: %s", data)
	}
}

func TestNpmSetupDryRun(t *testing.T) {
	home := t.TempDir()
	bin := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("SHELL", "/bin/zsh")
	if err := os.WriteFile(filepath.Join(bin, "npm"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := newNpmCmd()
	cmd.SetArgs([]string{"setup", "--dry-run"})
	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	for _, want := range []string{"Datakraften / npm setup", "Configuration", "Dry run: no changes made"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q: %s", want, output)
		}
	}
	if strings.Contains(output, "configured") {
		t.Fatalf("dry run reported a completed change: %s", output)
	}
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote config: %v", err)
	}
}
