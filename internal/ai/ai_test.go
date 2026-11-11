package ai

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	execpkg "github.com/sagathelab/datakraften/internal/exec"
)

func TestEnsureAIToolsInstallsCopilotViaGHExtension(t *testing.T) {
	var output bytes.Buffer
	SetOutput(&output)
	t.Cleanup(func() { SetOutput(nil) })
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

	commandExists = func(name string) bool {
		return name == "gh"
	}
	brewInstalled = func() bool {
		return false
	}
	currentOS = func() string {
		return "linux"
	}

	var commands []string
	runCommand = func(name string, args ...string) execpkg.Result {
		commands = append(commands, name+" "+strings.Join(args, " "))
		switch {
		case name == "gh" && len(args) == 2 && args[0] == "copilot" && args[1] == "--help":
			return execpkg.Result{Code: 1}
		case name == "gh" && len(args) == 2 && args[0] == "extension" && args[1] == "list":
			return execpkg.Result{Code: 0}
		case name == "gh" && len(args) == 3 && args[0] == "extension" && args[1] == "install" && args[2] == "github/gh-copilot":
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

	if !strings.Contains(output.String(), "Installing GitHub Copilot CLI via GitHub CLI extension") {
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
	if len(commands) != 3 {
		t.Fatalf("expected Copilot detection and extension commands, got %v", commands)
	}
}

func TestBuiltInGHCopilotAvoidsExtensionInstall(t *testing.T) {
	oldExists, oldRun := commandExists, runCommand
	t.Cleanup(func() { commandExists, runCommand = oldExists, oldRun })
	commandExists = func(name string) bool { return name == "gh" }
	runCommand = func(name string, args ...string) execpkg.Result {
		if name != "gh" || strings.Join(args, " ") != "copilot --help" {
			t.Fatalf("unexpected command: %s %v", name, args)
		}
		return execpkg.Result{Code: 0}
	}
	report := EnsureAITools(&config.Config{AITools: map[string]config.RuntimeConfig{"copilot": {Enabled: true}}})
	if len(report.AlreadyInstalled) != 1 || len(report.Errors) != 0 {
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
