package app

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/spf13/cobra"
)

func TestWorkstationCommandsRejectInvalidConfigBeforeEffects(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	marker := filepath.Join(bin, "ran")
	t.Setenv("DK_TEST_MARKER", marker)
	for _, name := range []string{"brew", "fnm", "uv", "npm", "sudo"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nprintf '%s\\n' ran > \"$DK_TEST_MARKER\"\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	old := http.DefaultTransport
	http.DefaultTransport = initTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("workstation command attempted to fetch a configuration")
		return nil, nil
	})
	t.Cleanup(func() { http.DefaultTransport = old })

	path := config.DefaultConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ input, want string }{
		{"source: team\nurl: https://example.test/config?token=hidden\n", "legacy team configuration"},
		{"profile: team\nteam: {url: 'https://example.test/config?token=hidden'}\n", "legacy team configuration"},
		{"editors: {zed: optional}\n", "editors.zed must be true or false"},
	} {
		if err := os.WriteFile(path, []byte(tc.input), 0600); err != nil {
			t.Fatal(err)
		}
		for _, command := range []struct {
			newCmd func() *cobra.Command
			args   []string
		}{
			{newApplyCmd, []string{"--yes"}},
			{newApplyCmd, []string{"--dry-run"}},
			{newUpdateCmd, nil},
			{newUpdateCmd, []string{"npm"}},
			{newUpdateCmd, []string{"brew"}},
		} {
			cmd := command.newCmd()
			cmd.SilenceErrors, cmd.SilenceUsage = true, true
			cmd.SetArgs(append([]string{}, command.args...))
			var err error
			captureStdout(t, func() { err = cmd.Execute() })
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "token=hidden") || strings.Contains(err.Error(), "no workstation configuration found") {
				t.Fatalf("%s %v: wanted %q, got %v", cmd.Name(), command.args, tc.want, err)
			}
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != tc.input {
			t.Fatalf("command changed workstation config: %s, %v", data, err)
		}
	}
	for _, path := range []string{marker, statePath()} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("unexpected side effect at %s: %v", path, err)
		}
	}
}

func TestInitRejectsRetiredFlags(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, flag := range []string{"--team", "--migrate-team"} {
		cmd := newInitCmd()
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		cmd.SetArgs([]string{flag})
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "unknown flag: "+flag) {
			t.Fatalf("%s: %v", flag, err)
		}
	}
	if _, err := os.Stat(config.DefaultConfigPath()); !os.IsNotExist(err) {
		t.Fatalf("unexpected config: %v", err)
	}
}

func TestInitCustomFileValidationAndMerge(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	run := func(args ...string) error {
		t.Helper()
		cmd := newInitCmd()
		cmd.SetArgs(append([]string{}, args...))
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		var err error
		captureStdout(t, func() { err = cmd.Execute() })
		return err
	}
	if err := run(); err != nil {
		t.Fatal(err)
	}
	path := config.DefaultConfigPath()
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "local.yaml")
	if err := run("--custom", input, "--yes"); err == nil {
		t.Fatal("missing local input accepted")
	}
	if err := os.WriteFile(input, []byte("runtimes: {node: {enabled: 'false'}}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run("--custom", input, "--yes"); err == nil || !strings.Contains(err.Error(), "runtimes.node.enabled") {
		t.Fatalf("invalid local input: %v", err)
	}
	if after, err := os.ReadFile(path); err != nil || string(after) != string(original) {
		t.Fatalf("failed merge changed original: %s, %v", after, err)
	}
	if err := os.WriteFile(input, []byte("source: default\nversion: 1\nruntimes: {node: {enabled: false}}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run("--custom", input, "--yes"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil || cfg.Source != "custom" || cfg.Runtimes.Node.Enabled || cfg.Runtimes.Node.Manager != "fnm" || !cfg.Runtimes.Python.Enabled {
		t.Fatalf("unexpected merged config: %+v, %v", cfg, err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("expected a private config: %v, %v", info, err)
	}
}
