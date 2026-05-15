package ai

import (
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	execpkg "github.com/sagathelab/datakraften/internal/exec"
)

func TestEnsureAIToolsInstallsCopilotViaGHExtension(t *testing.T) {
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

	if report.Enabled != 1 {
		t.Fatalf("expected 1 enabled tool, got %d", report.Enabled)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", report.Errors)
	}
	if len(report.Installed) != 1 || report.Installed[0] != "GitHub Copilot CLI" {
		t.Fatalf("expected Copilot to be installed, got %+v", report)
	}
	if len(commands) != 2 {
		t.Fatalf("expected two gh extension commands, got %v", commands)
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
