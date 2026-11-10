package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateNpmPermissionFailure(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "datakraften")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("ai_tools:\n  codex:\n    enabled: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "npm"), []byte("#!/bin/sh\nif [ \"$1\" = list ]; then echo '{\"dependencies\":{\"@openai/codex\":{}}}'; exit 0; fi\necho 'npm error code EACCES' >&2\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	captureStdout(t, func() {
		err := updateNpm(false)
		if err == nil || !strings.Contains(err.Error(), "user-writable npm prefix") || !strings.Contains(err.Error(), "EACCES") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUpdateNpmOnlyEnabledInstalledPackages(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	configDir := filepath.Join(home, ".config", "datakraften")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("ai_tools:\n  codex:\n    enabled: true\n  gemini:\n    enabled: true\n  claude:\n    enabled: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(home, "args")
	t.Setenv("TEST_NPM_LOG", log)
	script := "#!/bin/sh\nif [ \"$1\" = list ]; then echo '{\"dependencies\":{\"@openai/codex\":{},\"@anthropic-ai/claude-code\":{},\"unrelated\":{}}}'; else printf '%s\\n' \"$@\" > \"$TEST_NPM_LOG\"; fi\n"
	if err := os.WriteFile(filepath.Join(bin, "npm"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := updateNpm(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("dry run invoked npm update")
	}
	if err := updateNpm(false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "update\n-g\n@openai/codex\n" {
		t.Fatalf("unexpected npm arguments: %q", data)
	}
}

func TestUpdateAllReturnsErrorWhenAChildUpdateFails(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{
		{name: "brew", run: func(bool) error { return nil }},
		{name: "npm", run: func(bool) error { return errors.New("permission denied") }},
	}
	defer func() {
		updateSteps = originalSteps
	}()

	output := captureStdout(t, func() {
		err := updateAll(false)
		if err == nil {
			t.Fatal("expected updateAll to return an error")
		}
		if !strings.Contains(err.Error(), "npm: permission denied") {
			t.Fatalf("expected npm failure in returned error, got %q", err.Error())
		}
	})

	if strings.Contains(output, "✓ Update complete") {
		t.Fatalf("expected success summary to be omitted, got %q", output)
	}
	if !strings.Contains(output, "✗ Update finished with errors") {
		t.Fatalf("expected failure summary, got %q", output)
	}
}

func TestUpdateAllPrintsSuccessWhenNoFailuresOccur(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{
		{name: "brew", run: func(bool) error { return nil }},
		{name: "uv", run: func(bool) error { return nil }},
	}
	defer func() {
		updateSteps = originalSteps
	}()

	output := captureStdout(t, func() {
		if err := updateAll(false); err != nil {
			t.Fatalf("expected updateAll to succeed, got %v", err)
		}
	})

	if !strings.Contains(output, "✓ Update complete") {
		t.Fatalf("expected success summary, got %q", output)
	}
	if !strings.Contains(output, "[1/2] brew") || !strings.Contains(output, "[2/2] uv") {
		t.Fatalf("expected ordered progress, got %q", output)
	}
	if strings.Contains(output, "✗ Update finished with errors") {
		t.Fatalf("did not expect failure summary, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	return <-done
}

func TestUpdateAllDryRunIsPreview(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{{name: "example", run: func(dry bool) error {
		if !dry {
			t.Fatal("expected dry run")
		}
		return nil
	}}}
	defer func() { updateSteps = originalSteps }()
	output := captureStdout(t, func() {
		if err := updateAll(true); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "Preview complete; no updates made") || strings.Contains(output, "Update complete") {
		t.Fatalf("misleading preview: %q", output)
	}
}
