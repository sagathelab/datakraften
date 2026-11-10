package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/npm"
	"github.com/spf13/cobra"
)

type updateStep struct {
	name string
	run  func(bool) error
}

var updateSteps = []updateStep{
	{name: "brew", run: updateBrew},
	{name: "fnm", run: updateFnm},
	{name: "uv", run: updateUv},
	{name: "npm", run: updateNpm},
}

func newUpdateCmd() *cobra.Command {
	var list bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update managed tools to their latest versions",
		Long:         `Update all managed developer tools (brew packages, runtimes, AI tools) or a specific tool by name.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return listUpdatable()
			}

			if len(args) > 0 {
				return updateTool(args[0], dryRun)
			}

			return updateAll(dryRun)
		},
	}

	cmd.Flags().BoolVarP(&list, "list", "l", false, "List supported update targets")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be updated without making changes")

	return cmd
}

func listUpdatable() error {
	displayTitle(os.Stdout, "update", "Available update targets")

	fmt.Println("    brew — all Homebrew packages")
	fmt.Println("    fnm — Node.js runtime")
	fmt.Println("    uv — Python runtime")
	fmt.Println("    npm — enabled Datakraften npm tools")

	if exec.CommandExists("code") {
		fmt.Println("    code — VS Code extensions")
	}

	fmt.Println()
	fmt.Println("  Run 'dk update <tool>' to update a specific tool.")

	return nil
}

func updateTool(name string, dryRun bool) error {
	switch strings.ToLower(name) {
	case "brew":
		return updateBrew(dryRun)
	case "fnm", "node":
		return updateFnm(dryRun)
	case "uv", "python":
		return updateUv(dryRun)
	case "npm":
		return updateNpm(dryRun)
	default:
		return fmt.Errorf("unknown tool: %s — run 'dk update --list' to see available tools", name)
	}
}

func updateAll(dryRun bool) error {
	subtitle := "Updating managed tools"
	if dryRun {
		subtitle = "Preview only; no updates will be made"
	}
	displayTitle(os.Stdout, "update", subtitle)

	var failures []string

	for i, step := range updateSteps {
		fmt.Printf("\n  [%d/%d] %s\n", i+1, len(updateSteps), step.name)
		if err := step.run(dryRun); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", step.name, err))
			fmt.Printf("    ✗ %s failed; details below\n", step.name)
		}
	}

	fmt.Println()
	if len(failures) > 0 {
		fmt.Println("  ✗ Update finished with errors")
		return fmt.Errorf("update failed: %s", strings.Join(failures, "; "))
	}

	if dryRun {
		fmt.Println("  Preview complete; no updates made")
	} else {
		fmt.Println("  ✓ Update complete")
	}
	return nil
}

func updateBrew(dryRun bool) error {
	if !installers.BrewInstalled() {
		fmt.Println("    – Homebrew not installed, skipping")
		return nil
	}

	if dryRun {
		fmt.Println("    ~ Would run: brew update && brew upgrade")
		return nil
	}

	fmt.Println("    Updating Homebrew...")
	r := exec.Run("brew", "update")
	if r.Code != 0 {
		return fmt.Errorf("brew update failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println("    ✓ Command completed")

	fmt.Println("    Upgrading packages...")
	r = exec.Run("brew", "upgrade")
	if r.Code != 0 {
		return fmt.Errorf("brew upgrade failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println("    ✓ Command completed")

	return nil
}

func updateFnm(dryRun bool) error {
	if !exec.CommandExists("fnm") {
		fmt.Println("    – fnm not installed, skipping")
		return nil
	}

	if dryRun {
		fmt.Println("    ~ Would run: fnm install --lts")
		return nil
	}

	fmt.Println("    Updating Node.js LTS via fnm...")
	r := exec.Run("fnm", "install", "--lts")
	if r.Code != 0 {
		return fmt.Errorf("fnm install lts failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println("    ✓ Command completed")

	fmt.Println("    Setting fnm default to latest LTS...")
	r = exec.Run("fnm", "default", "lts-latest")
	if r.Code != 0 {
		return fmt.Errorf("fnm default failed: %s", strings.TrimSpace(r.Stderr))
	}

	return nil
}

func updateUv(dryRun bool) error {
	if !exec.CommandExists("uv") {
		fmt.Println("    – uv not installed, skipping")
		return nil
	}

	if dryRun {
		fmt.Println("    ~ Would run: uv self update")
		return nil
	}

	fmt.Println("    Updating uv...")
	r := exec.Run("uv", "self", "update")
	if r.Code != 0 {
		return fmt.Errorf("uv self update failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println("    ✓ Command completed")

	return nil
}

func updateNpm(dryRun bool) error {
	if !exec.CommandExists("npm") {
		fmt.Println("    – npm not installed, skipping")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("cannot determine managed npm tools: %w", err)
	}
	if cfg.Source == "team" {
		cfg, err = config.FetchRemote(cfg.URL)
		if err != nil {
			return fmt.Errorf("cannot load team configuration: %w", err)
		}
	}
	r := exec.Run("npm", "list", "-g", "--depth=0", "--json")
	var listed struct {
		Dependencies map[string]json.RawMessage `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(r.Stdout), &listed); err != nil || r.Code != 0 {
		return fmt.Errorf("cannot list global npm packages: %s", strings.TrimSpace(r.Stderr))
	}
	var packages []string
	for _, tool := range []struct{ key, pkg string }{{"codex", "@openai/codex"}, {"claude", "@anthropic-ai/claude-code"}, {"gemini", "@google-gemini/gemini-cli"}} {
		if _, installed := listed.Dependencies[tool.pkg]; cfg.AITools[tool.key].Enabled && installed {
			packages = append(packages, tool.pkg)
		}
	}
	if len(packages) == 0 {
		fmt.Println("    – No enabled npm tools installed, skipping")
		return nil
	}
	if dryRun {
		fmt.Printf("    ~ Would update npm packages: %s\n", strings.Join(packages, ", "))
		return nil
	}

	fmt.Printf("    Updating npm packages (%s)...\n", strings.Join(packages, ", "))
	args := append([]string{"update", "-g"}, packages...)
	r = exec.Run("npm", args...)
	if r.Code != 0 {
		fmt.Println("    ✗ npm update failed")
		if strings.Contains(r.Stderr, "EACCES") || strings.Contains(r.Stderr, "EPERM") {
			return retryNpmWithSudo(packages, r.Stderr)
		}
		return npm.CommandError("npm update -g", r.Stderr)
	}
	fmt.Println("    ✓ Command completed")

	return nil
}

func retryNpmWithSudo(packages []string, originalError string) error {
	if !exec.CommandExists("sudo") {
		return npm.CommandError("npm update -g", originalError)
	}
	prefixResult := exec.Run("npm", "config", "get", "prefix")
	prefix := strings.TrimSpace(prefixResult.Stdout)
	if prefixResult.Code != 0 || !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "\r\n") {
		return fmt.Errorf("cannot determine npm prefix for sudo retry; original error: %s", strings.TrimSpace(originalError))
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("npm needs administrator access but no interactive terminal is available: %w", npm.CommandError("npm update -g", originalError))
	}
	defer tty.Close()
	fmt.Fprintf(tty, "  npm cannot write to %s. Update %s with sudo? [y/N] ", prefix, strings.Join(packages, ", "))
	answer, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil || (strings.TrimSpace(strings.ToLower(answer)) != "y" && strings.TrimSpace(strings.ToLower(answer)) != "yes") {
		return fmt.Errorf("sudo retry declined: %w", npm.CommandError("npm update -g", originalError))
	}
	npmPath := exec.CommandPath("npm")
	args := append([]string{npmPath, "update", "-g", "--prefix", prefix}, packages...)
	cmd := osexec.Command("sudo", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo npm update failed: %w", err)
	}
	fmt.Println("    ✓ npm packages updated with sudo")
	return nil
}
