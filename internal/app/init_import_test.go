package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	"go.yaml.in/yaml/v3"
)

type initTransport func(*http.Request) (*http.Response, error)

func (f initTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func withTemplate(t *testing.T, input string, status int) {
	t.Helper()
	old := http.DefaultTransport
	http.DefaultTransport = initTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(input)), Header: make(http.Header)}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = old })
}

func TestImportConfigMergesAndValidates(t *testing.T) {
	defaults := []byte(fmt.Sprintf(defaultConfigYAML, "dnf"))
	out, paths, err := importConfig(defaults, []byte("runtimes:\n  node:\n    enabled: false\n  python: {}\nshell:\n  prompt: ''\nai_tools:\n  codex: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]interface{}
	if err := yaml.Unmarshal(out, &data); err != nil {
		t.Fatal(err)
	}
	runtimes := data["runtimes"].(map[string]interface{})
	node := runtimes["node"].(map[string]interface{})
	python := runtimes["python"].(map[string]interface{})
	if node["enabled"] != false || node["manager"] != "fnm" || node["version"] != "lts" || python["enabled"] != true || data["source"] != "custom" {
		t.Fatalf("invalid merge: %s", out)
	}
	if data["shell"].(map[string]interface{})["prompt"] != "" || data["ai_tools"].(map[string]interface{})["codex"].(map[string]interface{})["enabled"] != false {
		t.Fatalf("explicit zero lost: %s", out)
	}
	if strings.Join(paths, ",") != "ai_tools.codex.enabled,runtimes.node.enabled,shell.prompt" {
		t.Fatalf("paths: %q", paths)
	}
	for _, input := range []string{"<html>bad</html>", "runtimes: [", "runtimes:\n  node:\n    enabled: null", "runtimes:\n  node:\n    enabled: maybe", "runtimes:\n  node:\n    unknown: 1", "unknown: true", "version: 2", "source: team", "shell: null", "{}\n---\nunknown: true", ""} {
		t.Run(input, func(t *testing.T) {
			if _, _, err := importConfig(defaults, []byte(input)); err == nil {
				t.Fatalf("accepted %q", input)
			}
		})
	}
}

func TestInitFromURLIsLocalSnapshot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	if err := os.WriteFile(filepath.Join(bin, "npm"), []byte("#!/bin/sh\n[ \"$1\" = list ] || exit 1\nprintf '%s\\n' '{\"dependencies\":{\"@openai/codex\":{}}}'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	withTemplate(t, "runtimes:\n  node:\n    enabled: false\n", http.StatusOK)
	path := config.DefaultConfigPath()
	if err := initFromURL(path, "https://example.com/config.yaml?token=secret", false, false); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "custom" || cfg.Runtimes.Node.Enabled || !cfg.Runtimes.Python.Enabled {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(original), "secret") || strings.Contains(string(original), "example.com") {
		t.Fatalf("URL persisted: %s", original)
	}
	http.DefaultTransport = initTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("apply must not fetch template")
		return nil, nil
	})
	applyPreview := func() string {
		return captureStdout(t, func() {
			cmd := newApplyCmd()
			cmd.SetArgs([]string{"--dry-run"})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
	if got := applyPreview(); !strings.Contains(got, "Would install Python via uv") {
		t.Fatalf("original config was not used: %s", got)
	}
	updatePreview := func() string {
		return captureStdout(t, func() {
			if err := UpdateToolWithProgress("npm", true, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
	if got := updatePreview(); !strings.Contains(got, "Would update npm packages: @openai/codex") {
		t.Fatalf("original config was not used: %s", got)
	}
	var local map[string]interface{}
	if err := yaml.Unmarshal(original, &local); err != nil {
		t.Fatal(err)
	}
	local["runtimes"].(map[string]interface{})["python"].(map[string]interface{})["enabled"] = false
	local["ai_tools"].(map[string]interface{})["codex"] = false
	modified, err := yaml.Marshal(local)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, modified, 0600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if got := applyPreview(); !strings.Contains(got, "Python (disabled in config)") {
			t.Fatalf("local edit was not used: %s", got)
		}
	}
	if got := updatePreview(); !strings.Contains(got, "No enabled npm tools installed") {
		t.Fatalf("local npm edit was not used: %s", got)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != string(modified) {
		t.Fatalf("apply changed local config: %s, %v", got, err)
	}
}

func TestInitFailurePreservesExistingConfigAndState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := config.DefaultConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := "source: custom\nruntimes: {node: {enabled: false}}\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if err := (&State{ActiveSource: "custom", LastApply: "previous"}).Save(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		data string
		code int
		want string
	}{
		{"<html>not yaml</html>", http.StatusOK, "YAML mapping"},
		{"runtimes: [", http.StatusOK, "invalid YAML"},
		{"", http.StatusForbidden, "HTTP 403"},
		{"url: https://other.example/template?token=payload-secret", http.StatusOK, "legacy team configuration"},
	} {
		withTemplate(t, tc.data, tc.code)
		err := initFromURL(path, "https://example.com/config.yaml?token=request-secret", true, true)
		if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "-secret") {
			t.Fatalf("wanted safe %q error, got %v", tc.want, err)
		}
	}
	if err := writeInitAtomic(path, []byte("{}"), false); err == nil {
		t.Fatal("new config unexpectedly overwrote existing config")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != original {
		t.Fatalf("existing config was modified: %s, %v", data, err)
	}
	if state := LoadState(); state.ActiveSource != "custom" || state.LastApply != "previous" {
		t.Fatalf("state was modified: %+v", state)
	}
}

func TestMergeImportListsAndZeros(t *testing.T) {
	dst := map[string]interface{}{"nested": map[string]interface{}{"keep": true, "off": true, "number": 4, "list": []interface{}{1}}}
	src := map[string]interface{}{"nested": map[string]interface{}{"off": false, "number": 0, "list": []interface{}{}}}
	var paths []string
	mergeImport(dst, src, "", &paths)
	got := dst["nested"].(map[string]interface{})
	if got["keep"] != true || got["off"] != false || got["number"] != 0 || len(got["list"].([]interface{})) != 0 {
		t.Fatalf("merge: %#v", got)
	}
}

func TestImportPreviewMasksValuesAndWarnsOnUnsupportedSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withTemplate(t, "shell:\n  prompt: secret-value\nruntimes:\n  node:\n    enabled: false\n    version: '22'\n", http.StatusOK)
	path := config.DefaultConfigPath()
	output := captureStdout(t, func() {
		if err := initFromURL(path, "https://example.com/config?token=secret-token", false, false); err != nil {
			t.Fatal(err)
		}
	})
	for _, forbidden := range []string{"secret-value", "secret-token"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("secret leaked: %s", output)
		}
	}
	for _, expected := range []string{"runtimes.node.enabled: true -> false", "current dk apply does not use them", "runtimes.node.version", "shell.prompt"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("missing %q: %s", expected, output)
		}
	}
}

func TestInitFromFlagCreatesLocalConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	withTemplate(t, "runtimes:\n  node:\n    enabled: false\n", http.StatusOK)
	output := captureStdout(t, func() {
		cmd := newInitCmd()
		cmd.SetArgs([]string{"--from", "https://example.com/settings.yaml"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "not fetched again") {
		t.Fatalf("missing snapshot explanation: %s", output)
	}
	cfg, err := config.Load()
	if err != nil || cfg.Source != "custom" || cfg.Runtimes.Node.Enabled || !cfg.Runtimes.Python.Enabled {
		t.Fatalf("configuration: %+v, err: %v", cfg, err)
	}
}
