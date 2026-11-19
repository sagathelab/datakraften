package ai

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	execpkg "github.com/sagathelab/datakraften/internal/exec"
)

func TestEnsureAIToolsInstallsStandaloneCopilot(t *testing.T) {
	var output bytes.Buffer
	SetOutput(&output)
	t.Cleanup(func() { SetOutput(nil) })
	oldExists, oldRun, oldBrew := commandExists, runCommand, brewInstalled
	t.Cleanup(func() { commandExists, runCommand, brewInstalled = oldExists, oldRun, oldBrew })
	commandExists = func(name string) bool { return name == "gh" }
	brewInstalled = func() bool { return true }

	var commands []string
	runCommand = func(name string, args ...string) execpkg.Result {
		command := name + " " + strings.Join(args, " ")
		commands = append(commands, command)
		switch command {
		case "brew list --cask copilot-cli":
			return execpkg.Result{Code: 1}
		case "brew install --cask copilot-cli":
			return execpkg.Result{Code: 0}
		default:
			t.Fatalf("unexpected command: %s %s", name, strings.Join(args, " "))
			return execpkg.Result{}
		}
	}

	cfg := &config.Config{
		AITools: map[string]config.RuntimeConfig{
			"copilot": {Enabled: true},
		},
	}

	report := EnsureAITools(cfg)

	if !strings.Contains(output.String(), "Installing GitHub Copilot CLI via Homebrew") {
		t.Fatalf("unexpected output: %q", output.String())
	}

	if report.Enabled != 1 {
		t.Fatalf("expected 1 enabled tool, got %d", report.Enabled)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(report.Installed) != 1 || report.Installed[0] != "GitHub Copilot CLI" {
		t.Fatalf("expected Copilot to be installed, got %+v", report)
	}
	if !slices.Equal(commands, []string{"brew list --cask copilot-cli", "brew install --cask copilot-cli"}) {
		t.Fatalf("unexpected Copilot install commands: %v", commands)
	}
}

func TestStandaloneCopilotOnPATHSkipsInstallation(t *testing.T) {
	oldExists, oldRun := commandExists, runCommand
	t.Cleanup(func() { commandExists, runCommand = oldExists, oldRun })
	commandExists = func(name string) bool { return name == "copilot" }
	runCommand = func(name string, args ...string) execpkg.Result {
		t.Fatalf("unexpected command: %s %v", name, args)
		return execpkg.Result{}
	}
	report := EnsureAITools(&config.Config{AITools: map[string]config.RuntimeConfig{"copilot": {Enabled: true}}})
	if len(report.AlreadyInstalled) != 1 || len(report.Errors) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestStandaloneCopilotMissingFromPATHIsNotReinstalled(t *testing.T) {
	oldExists, oldRun, oldBrew := commandExists, runCommand, brewInstalled
	t.Cleanup(func() { commandExists, runCommand, brewInstalled = oldExists, oldRun, oldBrew })
	commandExists = func(string) bool { return false }
	brewInstalled = func() bool { return true }
	runCommand = func(name string, args ...string) execpkg.Result {
		if name != "brew" || !slices.Equal(args, []string{"list", "--cask", "copilot-cli"}) {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return execpkg.Result{}
	}
	report := EnsureAITools(&config.Config{AITools: map[string]config.RuntimeConfig{"copilot": {Enabled: true}}})
	if len(report.Skipped) != 1 || !strings.Contains(report.Skipped[0], "not on PATH") || report.HasFailures() {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestStandaloneCopilotInstallFailureDoesNotFallBackToGHExtension(t *testing.T) {
	oldExists, oldRun, oldBrew := commandExists, runCommand, brewInstalled
	t.Cleanup(func() { commandExists, runCommand, brewInstalled = oldExists, oldRun, oldBrew })
	commandExists = func(name string) bool { return name == "gh" }
	brewInstalled = func() bool { return true }
	runCommand = func(name string, args ...string) execpkg.Result {
		switch name + " " + strings.Join(args, " ") {
		case "brew list --cask copilot-cli":
			return execpkg.Result{Code: 1}
		case "brew install --cask copilot-cli":
			return execpkg.Result{Code: 1, Stderr: "download failed"}
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return execpkg.Result{}
		}
	}
	report := EnsureAITools(&config.Config{AITools: map[string]config.RuntimeConfig{"copilot": {Enabled: true}}})
	if len(report.Errors) != 1 || !strings.Contains(report.Errors[0], "brew install --cask copilot-cli failed: download failed") ||
		len(report.Installed) != 0 || len(report.AlreadyInstalled) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestCopilotBuiltInConflictReportedAsSkipped(t *testing.T) {
	oldExists, oldRun, oldOS := commandExists, runCommand, currentOS
	t.Cleanup(func() { commandExists, runCommand, currentOS = oldExists, oldRun, oldOS })
	commandExists = func(name string) bool { return name == "code" }
	currentOS = func() string { return "linux" }
	runCommand = func(name string, args ...string) execpkg.Result {
		if args[0] == "--list-extensions" {
			return execpkg.Result{Code: 0}
		}
		return execpkg.Result{Code: 1, Stderr: "Extension 'github.copilot-chat' is a built-in extension with version '0.73.0' and cannot be downgraded to version '0.48.1'."}
	}
	report := EnsureAIApps(&config.Config{AIApps: map[string]config.RuntimeConfig{"copilot": {Enabled: true}}})
	if len(report.Errors) != 0 || len(report.Skipped) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestEnsureAIAppsSkipsMacOSOnlyCasksOnLinux(t *testing.T) {
	originalCommandExists := commandExists
	originalRunCommand := runCommand
	originalBrewInstalled := brewInstalled
	originalCurrentOS := currentOS
	defer func() {
		commandExists = originalCommandExists
		runCommand = originalRunCommand
		brewInstalled = originalBrewInstalled
		currentOS = originalCurrentOS
	}()

	commandExists = func(string) bool {
		return false
	}
	brewInstalled = func() bool {
		return true
	}
	currentOS = func() string {
		return "linux"
	}

	called := false
	runCommand = func(name string, args ...string) execpkg.Result {
		called = true
		return execpkg.Result{Code: 0}
	}

	cfg := &config.Config{
		AIApps: map[string]config.RuntimeConfig{
			"codex": {Enabled: true},
		},
	}

	report := EnsureAIApps(cfg)

	if report.Enabled != 1 {
		t.Fatalf("expected 1 enabled app, got %d", report.Enabled)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(report.Skipped) != 1 || report.Skipped[0] != "Codex Desktop (requires macOS)" {
		t.Fatalf("expected macOS skip, got %+v", report)
	}
	if called {
		t.Fatal("expected Linux skip to avoid invoking brew or code commands")
	}
}

func TestInstallErrorCallbackBeforeNextTool(t *testing.T) {
	oldExists, oldRun := commandExists, runCommand
	t.Cleanup(func() { commandExists, runCommand = oldExists, oldRun })
	commandExists = func(name string) bool { return name == "npm" }
	var events []string
	runCommand = func(name string, args ...string) execpkg.Result {
		events = append(events, args[2])
		return execpkg.Result{Code: 1, Stderr: "failure"}
	}
	report := EnsureAIToolsWithOptions(&config.Config{AITools: map[string]config.RuntimeConfig{
		"codex": {Enabled: true}, "claude": {Enabled: true},
	}}, InstallOptions{OnError: func(message string) { events = append(events, message) }})
	if len(events) != 4 || events[0] != "@openai/codex" || events[2] != "@anthropic-ai/claude-code" || events[1] != report.Errors[0] || events[3] != report.Errors[1] {
		t.Fatalf("unexpected events %v and report %+v", events, report)
	}
}

func TestNpmInstallRetriesPermissionFailure(t *testing.T) {
	oldExists, oldRun := commandExists, runCommand
	t.Cleanup(func() { commandExists, runCommand = oldExists, oldRun })
	commandExists = func(name string) bool { return name == "npm" }
	runCommand = func(name string, args ...string) execpkg.Result {
		return execpkg.Result{Code: 1, Stderr: "npm error code EACCES\nlong error details"}
	}
	var pkg, detail string
	report := EnsureAIToolsWithOptions(&config.Config{AITools: map[string]config.RuntimeConfig{
		"claude": {Enabled: true},
	}}, InstallOptions{RetryNPM: func(p, stderr string) error { pkg, detail = p, stderr; return nil }})
	if pkg != "@anthropic-ai/claude-code" || !strings.Contains(detail, "EACCES") || len(report.Installed) != 1 || report.HasFailures() {
		t.Fatalf("retry %q %q, report %+v", pkg, detail, report)
	}
}

func TestAppInstallErrorCallback(t *testing.T) {
	oldExists, oldRun := commandExists, runCommand
	t.Cleanup(func() { commandExists, runCommand = oldExists, oldRun })
	commandExists = func(name string) bool { return name == "code" }
	runCommand = func(name string, args ...string) execpkg.Result {
		if args[0] == "--list-extensions" {
			return execpkg.Result{}
		}
		return execpkg.Result{Code: 1, Stderr: "failure"}
	}
	var messages []string
	report := EnsureAIAppsWithOptions(&config.Config{AIApps: map[string]config.RuntimeConfig{
		"copilot": {Enabled: true},
	}}, InstallOptions{OnError: func(message string) { messages = append(messages, message) }})
	if len(messages) != 1 || len(report.Errors) != 1 || messages[0] != report.Errors[0] {
		t.Fatalf("unexpected messages %v and report %+v", messages, report)
	}
}
