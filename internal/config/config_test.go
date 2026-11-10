package config

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLegacyAISettings(t *testing.T) {
	const input = "ai_tools:\n  old: true\n  disabled: false\n  modern:\n    enabled: true\n    version: latest\nai_apps:\n  old: true\n"
	check := func(t *testing.T, cfg *Config) {
		t.Helper()
		if !cfg.AITools["old"].Enabled || cfg.AITools["disabled"].Enabled || !cfg.AITools["modern"].Enabled || cfg.AITools["modern"].Version != "latest" || !cfg.AIApps["old"].Enabled {
			t.Fatalf("unexpected settings: %#v %#v", cfg.AITools, cfg.AIApps)
		}
	}
	t.Run("local", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		path := DefaultConfigPath()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		check(t, cfg)
	})
	t.Run("remote", func(t *testing.T) {
		previous := http.DefaultTransport
		http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(input)), Header: make(http.Header)}, nil
		})
		t.Cleanup(func() { http.DefaultTransport = previous })
		cfg, err := FetchRemote("https://example.test/config.yaml")
		if err != nil {
			t.Fatal(err)
		}
		check(t, cfg)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
