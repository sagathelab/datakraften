package app

import (
	"encoding/json"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/npm"
	"github.com/sagathelab/datakraften/internal/terminal"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return listUpdatable()
			}

			if len(args) > 0 {
				subtitle := "Updating a managed tool"
				if dryRun {
					subtitle = "Preview only; no updates will be made"
				}
				displayTitle(os.Stdout, "update / "+args[0], subtitle)
				output := newLineStartProgressOutput(os.Stdout)
				defer output.activate()()
				err := UpdateToolWithProgress(args[0], dryRun, output.event)
				if err == nil {
					fmt.Fprintln(progressWriter())
					if dryRun {
						fmt.Fprintln(progressWriter(), "  Preview complete; no updates made")
					} else {
						displayComplete(progressWriter(), "Update complete")
					}
					fmt.Fprintln(progressWriter())
				}
				return finishProgress(output, err)
			}

			updateTitle(dryRun)
			output := newLineStartProgressOutput(os.Stdout)
			defer output.activate()()
			err := updateAllWithProgress(dryRun, output.event, false)
			return finishProgress(output, err)
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

	fmt.Println()
	fmt.Println("  Run 'dk update <tool>' to update a specific tool.")

	return nil
}

func UpdateToolWithProgress(name string, dryRun bool, hook ProgressHook) error {
	name = strings.ToLower(name)
	var step updateStep
	switch name {
	case "brew":
		step = updateStep{name: "brew", run: updateBrew}
	case "fnm", "node":
		step = updateStep{name: "fnm", run: updateFnm}
	case "uv", "python":
		step = updateStep{name: "uv", run: updateUv}
	case "npm":
		step = updateStep{name: "npm", run: updateNpm}
	default:
		return fmt.Errorf("unknown tool: %s — run 'dk update --list' to see available tools", name)
	}
	return runUpdateStep(step, dryRun, hook)
}

func runUpdateStep(step updateStep, dryRun bool, hook ProgressHook) error {
	emitProgress(hook, ProgressEvent{Step: step.name, Message: "Updating " + step.name, Status: ProgressStarted})
	err := step.run(dryRun)
	if err != nil {
		emitProgress(hook, ProgressEvent{Step: step.name, Status: ProgressFailed, Err: err})
	} else {
		emitProgress(hook, ProgressEvent{Step: step.name, Status: ProgressCompleted})
	}
	return err
}

func updateAll(dryRun bool) error {
	return UpdateAllWithProgress(dryRun, nil)
}

func updateTitle(dryRun bool) {
	subtitle := "Updating managed tools"
	if dryRun {
		subtitle = "Preview only; no updates will be made"
	}
	displayTitle(progressWriter(), "update", subtitle)
}

func UpdateAllWithProgress(dryRun bool, hook ProgressHook) error {
	return updateAllWithProgress(dryRun, hook, true)
}

func updateAllWithProgress(dryRun bool, hook ProgressHook, showTitle bool) error {
	if showTitle {
		updateTitle(dryRun)
	}
	var failures []string

	for i, step := range updateSteps {
		if hook == nil {
			fmt.Fprintln(progressWriter())
			displayStep(progressWriter(), i+1, len(updateSteps), step.name)
		}
		if err := runUpdateStep(step, dryRun, hook); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", step.name, err))
			displayPrintf("    [ERROR] %s failed; details below\n", step.name)
		}
	}

	fmt.Fprintln(progressWriter())
	if len(failures) > 0 {
		displayPrintln("  [ERROR] Update finished with errors")
		return fmt.Errorf("update failed: %s", strings.Join(failures, "; "))
	}

	if dryRun {
		fmt.Fprintln(progressWriter(), "  Preview complete; no updates made")
	} else {
		displayComplete(progressWriter(), "Update complete")
	}
	fmt.Fprintln(progressWriter())
	return nil
}

func updateBrew(dryRun bool) error {
	if !installers.BrewInstalled() {
		displayPrintln("    [SKIP] Homebrew not installed, skipping")
		return nil
	}

	if dryRun {
		displayPrintln("    [PLAN] Would run: brew update && brew upgrade")
		return nil
	}

	fmt.Fprintln(progressWriter(), "    Updating Homebrew...")
	r := exec.Run("brew", "update")
	if r.Code != 0 {
		return fmt.Errorf("brew update failed: %s", strings.TrimSpace(r.Stderr))
	}
	displayPrintln("    [OK] Command completed")

	fmt.Fprintln(progressWriter(), "    Upgrading packages...")
	r = exec.Run("brew", "upgrade")
	if r.Code != 0 {
		return fmt.Errorf("brew upgrade failed: %s", strings.TrimSpace(r.Stderr))
	}
	displayPrintln("    [OK] Command completed")

	return nil
}

func updateFnm(dryRun bool) error {
	if !exec.CommandExists("fnm") {
		displayPrintln("    [SKIP] fnm not installed, skipping")
		return nil
	}

	if dryRun {
		displayPrintln("    [PLAN] Would run: fnm install --lts")
		return nil
	}

	fmt.Fprintln(progressWriter(), "    Updating Node.js LTS via fnm...")
	r := exec.Run("fnm", "install", "--lts")
	if r.Code != 0 {
		return fmt.Errorf("fnm install lts failed: %s", strings.TrimSpace(r.Stderr))
	}
	displayPrintln("    [OK] Command completed")

	fmt.Fprintln(progressWriter(), "    Setting fnm default to latest LTS...")
	r = exec.Run("fnm", "default", "lts-latest")
	if r.Code != 0 {
		return fmt.Errorf("fnm default failed: %s", strings.TrimSpace(r.Stderr))
	}

	return nil
}

func updateUv(dryRun bool) error {
	if !exec.CommandExists("uv") {
		displayPrintln("    [SKIP] uv not installed, skipping")
		return nil
	}

	if dryRun {
		displayPrintln("    [PLAN] Would run: uv self update")
		return nil
	}

	fmt.Fprintln(progressWriter(), "    Updating uv...")
	r := exec.Run("uv", "self", "update")
	if r.Code != 0 {
		return fmt.Errorf("uv self update failed: %s", strings.TrimSpace(r.Stderr))
	}
	displayPrintln("    [OK] Command completed")

	return nil
}

func updateNpm(dryRun bool) error {
	if !exec.CommandExists("npm") {
		displayPrintln("    [SKIP] npm not installed, skipping")
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
		displayPrintln("    [SKIP] No enabled npm tools installed, skipping")
		return nil
	}
	if dryRun {
		displayPrintf("    [PLAN] Would update npm packages: %s\n", strings.Join(packages, ", "))
		return nil
	}

	fmt.Fprintf(progressWriter(), "    Updating npm packages (%s)...\n", strings.Join(packages, ", "))
	args := append([]string{"update", "-g"}, packages...)
	r = exec.Run("npm", args...)
	if r.Code != 0 {
		if npm.IsPermissionError(r.Stderr) {
			displayPrintln("    [WARN] npm update needs administrator access")
			return retryNpmWithSudo(packages, r.Stderr)
		}
		displayPrintln("    [ERROR] npm update failed")
		return npm.CommandError("npm update -g", r.Stderr)
	}
	displayPrintln("    [OK] Command completed")

	return nil
}

func retryNpmWithSudo(packages []string, originalError string) error {
	return retryNpmOperationWithSudo("update", packages, originalError)
}

func retryNpmOperationWithSudo(operation string, packages []string, originalError string) error {
	command := "npm " + operation + " -g"
	if operation != "install" {
		verbosePrintf("%s output: %s\n", command, strings.TrimSpace(originalError))
	}
	if !exec.CommandExists("sudo") {
		return npm.CommandError(command, originalError)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("npm needs administrator access but no interactive terminal is available: %w", npm.CommandError(command, originalError))
	}
	prefixResult := exec.Run("npm", "config", "get", "prefix")
	prefix := strings.TrimSpace(prefixResult.Stdout)
	if prefixResult.Code != 0 || !strings.HasPrefix(prefix, "/") || strings.ContainsAny(prefix, "\r\n") {
		return fmt.Errorf("cannot determine npm prefix for sudo retry: %w", npm.CommandError(command, originalError))
	}
	prompt := fmt.Sprintf("  npm cannot write to %s. %s %s with sudo? [y/N]: ", prefix, strings.ToUpper(operation[:1])+operation[1:], strings.Join(packages, ", "))
	confirm := osexec.Command("/bin/sh", "-c", `printf '%s' "$1"; read -r answer || exit 1; case "$answer" in [yY]|[yY][eE][sS]) exit 0;; *) exit 1;; esac`, "sh", prompt)
	if err := terminal.Run(confirm, os.Stdin, progressTerminalWriter()); err != nil {
		return fmt.Errorf("sudo retry declined: %w", npm.CommandError(command, originalError))
	}
	npmPath := exec.CommandPath("npm")
	args := append([]string{npmPath, operation, "-g", "--prefix", prefix}, packages...)
	cmd := osexec.Command("sudo", args...)
	if err := terminal.Run(cmd, os.Stdin, progressTerminalWriter()); err != nil {
		return fmt.Errorf("sudo npm %s failed: %w", operation, err)
	}
	displayPrintf("    [OK] npm %s completed with sudo\n", operation)
	return nil
}
