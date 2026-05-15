package ai

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
)

type cliTool struct {
	name    string
	key     string
	brewPkg string
	npmPkg  string
	cmd     string
}

var cliTools = []cliTool{
	{name: "Codex CLI", key: "codex", npmPkg: "@openai/codex", cmd: "codex"},
	{name: "Claude Code", key: "claude", npmPkg: "@anthropic-ai/claude-code", cmd: "claude"},
	{name: "Gemini CLI", key: "gemini", npmPkg: "@google-gemini/gemini-cli", cmd: "gemini"},
	{name: "OpenCode", key: "opencode", brewPkg: "opencode", cmd: "opencode"},
	{name: "GitHub Copilot CLI", key: "copilot", brewPkg: "github/copilot-cli/copilot", cmd: "gh-copilot"},
}

type desktopApp struct {
	name string
	key  string
	brewCask string
	extID string
	checkCmd string
}

var desktopApps = []desktopApp{
	{name: "Codex Desktop", key: "codex", brewCask: "codex-app", checkCmd: "brew list --cask codex-app"},
	{name: "Claude Desktop", key: "claude", brewCask: "claude", checkCmd: "brew list --cask claude"},
	{name: "Copilot (VS Code)", key: "copilot", extID: "GitHub.copilot", checkCmd: "code --list-extensions"},
}

func npmInstalled() bool {
	return exec.CommandExists("npm")
}

func EnsureAITools(cfg *config.Config) ([]string, error) {
	var installed []string

	for _, tool := range cliTools {
		if !cfg.AITools[tool.key] {
			continue
		}
		if exec.CommandExists(tool.cmd) {
			continue
		}
		if err := installCLI(tool); err != nil {
			return installed, fmt.Errorf("%s: %w", tool.name, err)
		}
		installed = append(installed, tool.name)
	}

	return installed, nil
}

func installCLI(tool cliTool) error {
	if tool.brewPkg != "" && installers.BrewInstalled() {
		fmt.Printf("    Installing %s via Homebrew...\n", tool.name)
		return installers.BrewInstall(tool.brewPkg)
	}
	if tool.npmPkg != "" && npmInstalled() {
		fmt.Printf("    Installing %s via npm...\n", tool.name)
		r := exec.Run("npm", "install", "-g", tool.npmPkg)
		if r.Code != 0 {
			return fmt.Errorf("npm install failed: %s", r.Stderr)
		}
		return nil
	}
	fmt.Printf("    – %s: no suitable install method found\n", tool.name)
	return nil
}

func EnsureAIApps(cfg *config.Config) ([]string, error) {
	var installed []string

	for _, app := range desktopApps {
		if !cfg.AIApps[app.key] {
			continue
		}
		if appInstalled(app) {
			continue
		}
		if err := installApp(app); err != nil {
			return installed, fmt.Errorf("%s: %w", app.name, err)
		}
		installed = append(installed, app.name)
	}

	return installed, nil
}

func appInstalled(app desktopApp) bool {
	switch {
	case app.brewCask != "":
		r := exec.Run("brew", "list", "--cask", app.brewCask)
		return r.Code == 0
	case app.extID != "":
		r := exec.Run("code", "--list-extensions")
		if r.Code != 0 {
			return false
		}
		return strings.Contains(r.Stdout, app.extID)
	default:
		return false
	}
}

func installApp(app desktopApp) error {
	switch {
	case app.brewCask != "":
		if !installers.BrewInstalled() {
			fmt.Printf("    – %s: Homebrew not available\n", app.name)
			return nil
		}
		fmt.Printf("    Installing %s via Homebrew...\n", app.name)
		r := exec.Run("brew", "install", "--cask", app.brewCask)
		if r.Code != 0 {
			return fmt.Errorf("brew install --cask failed: %s", r.Stderr)
		}
		return nil
	case app.extID != "":
		if !exec.CommandExists("code") {
			fmt.Printf("    – %s: VS Code CLI not available\n", app.name)
			return nil
		}
		fmt.Printf("    Installing %s VS Code extension...\n", app.name)
		r := exec.Run("code", "--install-extension", app.extID)
		if r.Code != 0 {
			return fmt.Errorf("code --install-extension failed: %s", r.Stderr)
		}
		return nil
	default:
		fmt.Printf("    – %s: no install method defined\n", app.name)
		return nil
	}
}
