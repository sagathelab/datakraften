package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadLocalSettings(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := DefaultConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	const input = `version: 1
profile: custom
ai_tools:
  codex: true
  claude: false
  gemini:
    enabled: true
    version: latest
ai_apps:
  codex: true
editors:
  zed: 'true'
  cursor: optional
`
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "custom" || cfg.Version != "1" || !cfg.AITools["codex"].Enabled || cfg.AITools["claude"].Enabled ||
		!cfg.AITools["gemini"].Enabled || cfg.AITools["gemini"].Version != "latest" || !cfg.AIApps["codex"].Enabled ||
		cfg.Editors["zed"] != "true" || cfg.Editors["cursor"] != "optional" {
		t.Fatalf("unexpected settings: %+v", cfg)
	}
}

func TestLoadRejectsInvalidWorkstationConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := DefaultConfigPath()
	if _, err := Load(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing config: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ input, want string }{
		{"", "empty"},
		{"<html>not YAML</html>", "mapping"},
		{"source: [", "YAML syntax"},
		{"source: custom\nsource: default\n", "duplicate field source"},
		{"{}\n---\n{}\n", "exactly one"},
		{"version: 2", "version must be 1"},
		{"source: other", "default or custom"},
		{"source: team\nurl: https://example.test/config?token=hidden", "legacy team configuration"},
		{"profile: team", "legacy team configuration"},
		{"source: custom\nteam: {url: https://example.test/config}", "legacy team configuration"},
		{"source: custom\nurl: https://example.test/config", "legacy team configuration"},
		{"runtimes: {python: {enabled: 'false'}}", "runtimes.python.enabled must be a boolean"},
		{"runtimes: {node: {enabled: null}}", "runtimes.node.enabled cannot be null"},
		{"runtimes: {node: {manager: 1}}", "runtimes.node.manager must be a string"},
		{"ai_tools: {codex: yes}", "ai_tools.codex must be a mapping"},
		{"ai_apps: {codex: {enabled: 'true'}}", "ai_apps.codex.enabled must be a boolean"},
		{"editors: {zed: optional}", "editors.zed must be true or false"},
		{"tools: {docker: 1}", "tools.docker must be a boolean"},
		{"shell: {unused: value}", "unknown field shell.unused"},
		{"system: {package_manager: &pm apt}\ntooling: {package_manager: *pm}", "YAML aliases"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tt.input), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), tt.want) || strings.Contains(err.Error(), "token=hidden") {
				t.Fatalf("wanted %q, got %v", tt.want, err)
			}
		})
	}
}
