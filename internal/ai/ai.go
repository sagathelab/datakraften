package ai

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
)

type toolDef struct {
	name    string
	brewPkg string
	npmPkg  string
}

var tools = []toolDef{
	{name: "Codex CLI", npmPkg: "@openai/codex"},
	{name: "Claude Code", npmPkg: "@anthropic-ai/claude-code"},
	{name: "Gemini CLI", npmPkg: "@google-gemini/gemini-cli"},
	{name: "OpenCode", brewPkg: "opencode"},
	{name: "GitHub Copilot CLI", brewPkg: "github/copilot-cli/copilot"},
}

var aiConfigKeys = map[string]string{
	"Codex CLI":          "codex",
	"Claude Code":        "claude_code",
	"Gemini CLI":         "gemini_cli",
	"OpenCode":           "opencode",
	"GitHub Copilot CLI": "github_copilot",
}

var toolCommands = map[string]string{
	"Codex CLI":          "codex",
	"Claude Code":        "claude",
	"Gemini CLI":         "gemini",
	"OpenCode":           "opencode",
	"GitHub Copilot CLI": "gh-copilot",
}

func npmInstalled() bool {
	return exec.CommandExists("npm")
}

func EnsureAITools(cfg *config.Config) ([]string, error) {
	if cfg.AI == nil {
		return nil, nil
	}

	var installed []string

	for _, tool := range tools {
		if !isEnabled(cfg.AI, tool.name) {
			continue
		}
		if alreadyInstalled(tool) {
			continue
		}
		if err := installTool(tool); err != nil {
			return installed, fmt.Errorf("%s: %w", tool.name, err)
		}
		installed = append(installed, tool.name)
	}

	return installed, nil
}

func isEnabled(aiCfg map[string]string, name string) bool {
	key, ok := aiConfigKeys[name]
	if !ok {
		return false
	}
	return aiCfg[key] == "true"
}

func alreadyInstalled(tool toolDef) bool {
	cmd, ok := toolCommands[tool.name]
	if !ok {
		return false
	}
	return exec.CommandExists(cmd)
}

func installTool(tool toolDef) error {
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
