package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateBrewScope(t *testing.T) {
	const installed = "unrelated\nthird-party/tools/fd\nhomebrew/core/fzf\ngh\n"
	const scopedPlan = "Would run: brew upgrade --formula gh homebrew/core/fzf (auto-update and cleanup disabled)"
	const list = "list --formula --full-name [1,,1]\n"
	const scopedUpgrade = "upgrade --formula gh homebrew/core/fzf [1,,1]\n"
	const fullUpdate = "update [0,1,0]\n"
	const fullUpgrade = "upgrade [0,1,0]\n"
	tests := []struct {
		name       string
		args       []string
		installed  string
		fail       string
		wantCalls  string
		wantOutput string
		wantError  string
	}{
		{name: "default", installed: installed, wantCalls: list + scopedUpgrade, wantOutput: "Homebrew may also change dependencies and dependents"},
		{name: "named", args: []string{"brew"}, installed: installed, wantCalls: list + scopedUpgrade, wantOutput: "Upgrading installed formulae from dk's default list: gh, homebrew/core/fzf"},
		{name: "default preview", args: []string{"--dry-run"}, installed: installed, wantCalls: list, wantOutput: scopedPlan},
		{name: "named preview", args: []string{"brew", "--dry-run"}, installed: installed, wantCalls: list, wantOutput: scopedPlan},
		{name: "empty", args: []string{"brew"}, installed: "unrelated\nthird-party/tools/fd\n", wantCalls: list, wantOutput: "No installed formulae match dk's curated default Homebrew list"},
		{name: "empty preview", args: []string{"brew", "--dry-run"}, wantCalls: list, wantOutput: "No installed formulae match dk's curated default Homebrew list"},
		{name: "full", args: []string{"brew-all"}, wantCalls: fullUpdate + fullUpgrade, wantOutput: "Upgrading all eligible Homebrew formulae and casks"},
		{name: "full preview", args: []string{"brew-all", "--dry-run"}, wantOutput: "Would run: brew update && brew upgrade (all eligible formulae and casks)"},
		{name: "list failure", args: []string{"brew", "--dry-run"}, installed: installed, fail: "list", wantCalls: list, wantError: "inventory unavailable"},
		{name: "scoped partial upgrade", installed: installed, fail: "upgrade", wantCalls: list + scopedUpgrade, wantOutput: "Upgrade failed after changing a package", wantError: "brew: brew upgrade --formula gh homebrew/core/fzf failed: exit status 1; some formulae may already have been upgraded"},
		{name: "full refresh failure", args: []string{"brew-all"}, fail: "update", wantCalls: fullUpdate, wantOutput: "Could not refresh metadata", wantError: "brew update failed: exit status 1"},
		{name: "full partial upgrade", args: []string{"brew-all"}, fail: "upgrade", wantCalls: fullUpdate + fullUpgrade, wantOutput: "Upgrade failed after changing a package", wantError: "brew upgrade failed: exit status 1; some packages may already have been upgraded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home, bin := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("PATH", bin)
			t.Setenv("NO_COLOR", "1")
			t.Setenv("HOMEBREW_NO_AUTO_UPDATE", "0")
			t.Setenv("HOMEBREW_FORCE_API_AUTO_UPDATE", "1")
			t.Setenv("HOMEBREW_NO_INSTALL_CLEANUP", "0")
			log := filepath.Join(home, "brew-calls")
			t.Setenv("TEST_BREW_LOG", log)
			t.Setenv("TEST_BREW_FORMULAE", tt.installed)
			t.Setenv("TEST_BREW_FAIL", tt.fail)

			configDir := filepath.Join(home, ".config", "datakraften")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("source: default\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(log, nil, 0600); err != nil {
				t.Fatal(err)
			}
			script := `#!/bin/sh
printf '%s [%s,%s,%s]\n' "$*" "$HOMEBREW_NO_AUTO_UPDATE" "$HOMEBREW_FORCE_API_AUTO_UPDATE" "$HOMEBREW_NO_INSTALL_CLEANUP" >> "$TEST_BREW_LOG"
case "$1" in
    list) printf '%s' "$TEST_BREW_FORMULAE"; message='inventory unavailable';;
    update) message='Could not refresh metadata';;
    upgrade) message='Upgrade failed after changing a package';;
    *) exit 2;;
esac
if [ "$TEST_BREW_FAIL" = "$1" ]; then
    printf '%s\n' "$message" >&2
    exit 1
fi
`
			if err := os.WriteFile(filepath.Join(bin, "brew"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}

			var err error
			output := captureStdout(t, func() {
				cmd := newUpdateCmd()
				cmd.SetErr(io.Discard)
				cmd.SetArgs(tt.args)
				err = cmd.Execute()
			})
			if tt.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
				if strings.Contains(output, "Update complete") || strings.Contains(output, "Preview complete") {
					t.Fatalf("failure reported as success: %q", output)
				}
			}
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output = %q, want %q", output, tt.wantOutput)
			}
			calls, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if string(calls) != tt.wantCalls {
				t.Errorf("brew calls = %q, want %q", calls, tt.wantCalls)
			}
		})
	}
}

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
		err := UpdateToolWithProgress("npm", false, nil)
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
	if err := UpdateToolWithProgress("npm", true, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("dry run invoked npm update")
	}
	if err := UpdateToolWithProgress("npm", false, nil); err != nil {
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
	if !strings.Contains(output, "[ERROR] Update finished with errors") {
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
	if !strings.Contains(output, "01 / 02  ·  brew") || !strings.Contains(output, "02 / 02  ·  uv") {
		t.Fatalf("expected ordered progress, got %q", output)
	}
	if strings.Contains(output, "[ERROR] Update finished with errors") {
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

func TestUpdateAllReportsStepsBeforeRunningThem(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{
		{name: "first", run: func(bool) error { fmt.Println("    first work"); return nil }},
		{name: "second", run: func(bool) error { fmt.Println("    second work"); return nil }},
	}
	defer func() { updateSteps = originalSteps }()
	output := captureStdout(t, func() {
		if err := updateAll(false); err != nil {
			t.Fatal(err)
		}
	})
	last := -1
	for _, part := range []string{"01 / 02", "first work", "02 / 02", "second work", "✓ Update complete"} {
		next := strings.Index(output, part)
		if next <= last {
			t.Fatalf("progress out of order at %q: %q", part, output)
		}
		last = next
	}
}

func TestListUpdatableOnlyShowsSupportedTargets(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := os.WriteFile(filepath.Join(dir, "code"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, func() {
		if err := listUpdatable(); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(output, "code —") {
		t.Fatalf("unsupported target listed: %q", output)
	}
	for _, target := range []string{"brew", "brew-all", "fnm", "uv", "npm"} {
		if !strings.Contains(output, target+" —") {
			t.Fatalf("missing %s: %q", target, output)
		}
	}
	if !strings.Contains(output, "brew-all — full Homebrew refresh and upgrade (explicit only; includes eligible casks)") {
		t.Fatalf("full upgrade scope not shown: %q", output)
	}
}

func TestUpdateProgressFailure(t *testing.T) {
	failure := errors.New("update failed")
	var events []ProgressEvent
	step := updateStep{name: "example", run: func(bool) error { return failure }}
	if err := runUpdateStep(step, false, func(event ProgressEvent) { events = append(events, event) }); !errors.Is(err, failure) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 || events[0].Status != ProgressStarted || events[1].Status != ProgressFailed || !errors.Is(events[1].Err, failure) {
		t.Fatalf("unexpected events: %+v", events)
	}
}
