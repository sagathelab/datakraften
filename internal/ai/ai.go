package ai

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
)

type installStatus int

const (
	installStatusInstalled installStatus = iota
	installStatusSkipped
)

type InstallReport struct {
	Enabled          int
	Installed        []string
	AlreadyInstalled []string
	Skipped          []string
	Errors           []string
}

func (r InstallReport) HasFailures() bool {
	return len(r.Errors) > 0
}

type cliTool struct {
	name        string
	key         string
	brewPkg     string
	npmPkg      string
	cmd         string
	ghExtension string
}

var cliTools = []cliTool{
	{name: "Codex CLI", key: "codex", npmPkg: "@openai/codex", cmd: "codex"},
	{name: "Claude Code", key: "claude", npmPkg: "@anthropic-ai/claude-code", cmd: "claude"},
	{name: "Gemini CLI", key: "gemini", npmPkg: "@google-gemini/gemini-cli", cmd: "gemini"},
	{name: "OpenCode", key: "opencode", brewPkg: "opencode", cmd: "opencode"},
	{name: "GitHub Copilot CLI", key: "copilot", ghExtension: "github/gh-copilot"},
}

type desktopApp struct {
	name      string
	key       string
	brewCask  string
	extID     string
	macOSOnly bool
}

var desktopApps = []desktopApp{
	{name: "Codex Desktop", key: "codex", brewCask: "codex-app", macOSOnly: true},
	{name: "Claude Desktop", key: "claude", brewCask: "claude", macOSOnly: true},
	{name: "Copilot (VS Code)", key: "copilot", extID: "GitHub.copilot"},
}

var (
	commandExists = exec.CommandExists
	runCommand    = exec.Run
	brewInstalled = installers.BrewInstalled
	currentOS     = func() string { return runtime.GOOS }
)

func npmInstalled() bool {
	return commandExists("npm")
}

func EnsureAITools(cfg *config.Config) InstallReport {
	var report InstallReport

	for _, tool := range cliTools {
		if !cfg.AITools[tool.key].Enabled {
			continue
		}

		report.Enabled++

		if toolInstalled(tool) {
			report.AlreadyInstalled = append(report.AlreadyInstalled, tool.name)
			continue
		}

		status, detail, err := installCLI(tool)
		switch {
		case err != nil:
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", tool.name, err))
		case status == installStatusInstalled:
			report.Installed = append(report.Installed, tool.name)
		default:
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s (%s)", tool.name, detail))
		}
	}

	return report
}

func toolInstalled(tool cliTool) bool {
	switch {
	case tool.ghExtension != "":
		return ghExtensionInstalled(tool.ghExtension)
	case tool.cmd != "":
		return commandExists(tool.cmd)
	default:
		return false
	}
}

func ghExtensionInstalled(repo string) bool {
	if !commandExists("gh") {
		return false
	}

	r := runCommand("gh", "extension", "list")
	if r.Code != 0 {
		return false
	}

	name := path.Base(repo)
	return strings.Contains(r.Stdout, repo) || strings.Contains(r.Stdout, name)
}

func installCLI(tool cliTool) (installStatus, string, error) {
	switch {
	case tool.ghExtension != "":
		if !commandExists("gh") {
			fmt.Printf("    – %s: GitHub CLI not available\n", tool.name)
			return installStatusSkipped, "GitHub CLI not available", nil
		}
		fmt.Printf("    Installing %s via GitHub CLI extension...\n", tool.name)
		r := runCommand("gh", "extension", "install", tool.ghExtension)
		if r.Code != 0 {
			return installStatusSkipped, "", fmt.Errorf("gh extension install failed: %s", r.Stderr)
		}
		return installStatusInstalled, "", nil
	case tool.brewPkg != "" && brewInstalled():
		fmt.Printf("    Installing %s via Homebrew...\n", tool.name)
		if err := installers.BrewInstall(tool.brewPkg); err != nil {
			return installStatusSkipped, "", err
		}
		return installStatusInstalled, "", nil
	case tool.npmPkg != "" && npmInstalled():
		fmt.Printf("    Installing %s via npm...\n", tool.name)
		r := runCommand("npm", "install", "-g", tool.npmPkg)
		if r.Code != 0 {
			return installStatusSkipped, "", fmt.Errorf("npm install failed: %s", r.Stderr)
		}
		return installStatusInstalled, "", nil
	default:
		fmt.Printf("    – %s: no suitable install method found\n", tool.name)
		return installStatusSkipped, "no suitable install method found", nil
	}
}

func EnsureAIApps(cfg *config.Config) InstallReport {
	var report InstallReport

	for _, app := range desktopApps {
		if !cfg.AIApps[app.key].Enabled {
			continue
		}

		report.Enabled++

		if reason := unsupportedAppReason(app); reason != "" {
			fmt.Printf("    – %s: %s\n", app.name, reason)
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s (%s)", app.name, reason))
			continue
		}

		if appInstalled(app) {
			report.AlreadyInstalled = append(report.AlreadyInstalled, app.name)
			continue
		}

		status, detail, err := installApp(app)
		switch {
		case err != nil:
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", app.name, err))
		case status == installStatusInstalled:
			report.Installed = append(report.Installed, app.name)
		default:
			report.Skipped = append(report.Skipped, fmt.Sprintf("%s (%s)", app.name, detail))
		}
	}

	return report
}

func unsupportedAppReason(app desktopApp) string {
	if app.macOSOnly && currentOS() != "darwin" {
		return "requires macOS"
	}
	return ""
}

func appInstalled(app desktopApp) bool {
	switch {
	case app.brewCask != "":
		r := runCommand("brew", "list", "--cask", app.brewCask)
		return r.Code == 0
	case app.extID != "":
		r := runCommand("code", "--list-extensions")
		if r.Code != 0 {
			return false
		}
		return strings.Contains(r.Stdout, app.extID)
	default:
		return false
	}
}

func installApp(app desktopApp) (installStatus, string, error) {
	switch {
	case app.brewCask != "":
		if reason := unsupportedAppReason(app); reason != "" {
			fmt.Printf("    – %s: %s\n", app.name, reason)
			return installStatusSkipped, reason, nil
		}
		if !brewInstalled() {
			fmt.Printf("    – %s: Homebrew not available\n", app.name)
			return installStatusSkipped, "Homebrew not available", nil
		}
		fmt.Printf("    Installing %s via Homebrew...\n", app.name)
		r := runCommand("brew", "install", "--cask", app.brewCask)
		if r.Code != 0 {
			return installStatusSkipped, "", fmt.Errorf("brew install --cask failed: %s", r.Stderr)
		}
		return installStatusInstalled, "", nil
	case app.extID != "":
		if !commandExists("code") {
			fmt.Printf("    – %s: VS Code CLI not available\n", app.name)
			return installStatusSkipped, "VS Code CLI not available", nil
		}
		fmt.Printf("    Installing %s VS Code extension...\n", app.name)
		r := runCommand("code", "--install-extension", app.extID)
		if r.Code != 0 {
			return installStatusSkipped, "", fmt.Errorf("code --install-extension failed: %s", r.Stderr)
		}
		return installStatusInstalled, "", nil
	default:
		fmt.Printf("    – %s: no install method defined\n", app.name)
		return installStatusSkipped, "no install method defined", nil
	}
}
