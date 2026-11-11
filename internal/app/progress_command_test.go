package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
)

func TestApplyCommandDoesNotDuplicateStepHeadings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	dir := filepath.Join(home, ".config", "datakraften")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("source: default\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		cmd := newApplyCmd()
		cmd.SetArgs([]string{"--dry-run"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	for _, heading := range []string{"Checking system packages", "Checking developer tools", "Checking runtimes", "Configuring shell", "Checking editors", "Checking AI tools", "Checking AI apps", "Checking Docker"} {
		if strings.Count(out, "[...] "+heading) != 1 {
			t.Errorf("missing or duplicate %q: %q", heading, out)
		}
	}
	if strings.Contains(out, " / 08  · ") {
		t.Fatalf("duplicate numbered heading: %q", out)
	}
}

func TestApplyWithoutHookRetainsNumberedHeadings(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	out := captureStdout(t, func() { RunApply(&config.Config{}, true) })
	if !strings.Contains(out, "01 / 08  ·  System packages") || !strings.Contains(out, "08 / 08  ·  Docker") {
		t.Fatalf("missing headings: %q", out)
	}
}

func TestUpdateCommandDoesNotDuplicateStepHeadings(t *testing.T) {
	old := updateSteps
	updateSteps = []updateStep{{name: "first", run: func(bool) error { return nil }}, {name: "second", run: func(bool) error { return nil }}}
	t.Cleanup(func() { updateSteps = old })
	out := captureStdout(t, func() {
		cmd := newUpdateCmd()
		cmd.SetArgs([]string{"--dry-run"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	for _, name := range []string{"first", "second"} {
		if strings.Count(out, "[...] Updating "+name) != 1 {
			t.Errorf("missing or duplicate %s: %q", name, out)
		}
	}
	if strings.Contains(out, " / 02  · ") {
		t.Fatalf("duplicate numbered heading: %q", out)
	}
}
