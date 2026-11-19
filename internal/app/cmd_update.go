package app

import (
	"encoding/json"
	"errors"
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
	name       string
	run        func(bool) error
	configured func(*config.Config, bool) error
}

var updateSteps = []updateStep{
	{name: "brew", run: updateBrew},
	{name: "fnm", run: updateFnm},
	{name: "uv", run: updateUv},
	{name: "npm", configured: updateConfiguredNpm},
}

func newUpdateCmd() *cobra.Command {
	var list bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update [target]",
		Short: "Update developer tools",
		Long: `Update developer tools, or select a specific target with 'dk update <target>'.

The default Homebrew step and 'dk update brew' target only installed formulae
from dk's curated default list, regardless of who installed them. They use local
Homebrew metadata and skip automatic cleanup. Run 'brew update' separately
to refresh that metadata. Homebrew can still change dependencies and dependents.

'dk update brew-all' explicitly runs 'brew update' followed by 'brew upgrade'
for all eligible formulae and casks. It is not part of the default update.`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return listUpdatable()
			}

			if len(args) > 0 {
				subtitle := "Updating a selected target"
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
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview update commands without running them")

	return cmd
}

func listUpdatable() error {
	displayTitle(os.Stdout, "update", "Available update targets")

	fmt.Println("    brew — installed formulae from dk's curated default list")
	fmt.Println("    brew-all — full Homebrew refresh and upgrade (explicit only; includes eligible casks)")
	fmt.Println("    fnm — Node.js runtime")
	fmt.Println("    uv — Python runtime")
	fmt.Println("    npm — enabled Datakraften npm tools")

	fmt.Println()
	fmt.Println("  Run 'dk update <target>' to update a specific target.")

	return nil
}

func UpdateToolWithProgress(name string, dryRun bool, hook ProgressHook) error {
	name = strings.ToLower(name)
	switch name {
	case "node":
		name = "fnm"
	case "python":
		name = "uv"
	}
	steps := updateSteps
	if name == "brew-all" {
		steps = []updateStep{{name: "brew-all", run: updateBrewAll}}
	}
	for _, step := range steps {
		if step.name == name {
			cfg, err := loadUpdateConfig(step.configured != nil)
			if err != nil {
				return err
			}
			return runUpdateStep(step.withConfig(cfg), dryRun, hook)
		}
	}
	return fmt.Errorf("unknown tool: %s — run 'dk update --list' to see available tools", name)
}

func loadUpdateConfig(required bool) (*config.Config, error) {
	cfg, err := config.Load()
	if errors.Is(err, os.ErrNotExist) {
		if required {
			return nil, fmt.Errorf("no workstation configuration found; run 'dk init' first: %w", err)
		}
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot load configuration: %w", err)
	}
	return cfg, nil
}

func (s updateStep) withConfig(cfg *config.Config) updateStep {
	if run := s.configured; run != nil {
		s.run = func(dryRun bool) error { return run(cfg, dryRun) }
	}
	return s
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
	subtitle := "Updating developer tools"
	if dryRun {
		subtitle = "Preview only; no updates will be made"
	}
	displayTitle(progressWriter(), "update", subtitle)
}

func UpdateAllWithProgress(dryRun bool, hook ProgressHook) error {
	return updateAllWithProgress(dryRun, hook, true)
}

func updateAllWithProgress(dryRun bool, hook ProgressHook, showTitle bool) error {
	required := false
	for _, step := range updateSteps {
		required = required || step.configured != nil
	}
	cfg, err := loadUpdateConfig(required)
	if err != nil {
		return err
	}
	if showTitle {
		updateTitle(dryRun)
	}
	var failures []string

	for i, step := range updateSteps {
		if hook == nil {
			fmt.Fprintln(progressWriter())
			displayStep(progressWriter(), i+1, len(updateSteps), step.name)
		}
		if err := runUpdateStep(step.withConfig(cfg), dryRun, hook); err != nil {
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

	list := scopedBrewCommand("list", "--formula", "--full-name")
	out, err := list.Output()
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && strings.TrimSpace(string(exitErr.Stderr)) != "" {
			return fmt.Errorf("cannot list installed Homebrew formulae (brew list --formula --full-name): %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return fmt.Errorf("cannot list installed Homebrew formulae (brew list --formula --full-name): %w", err)
	}

	installed := make(map[string]bool)
	for _, name := range strings.Fields(string(out)) {
		installed[name] = true
	}
	var formulae []string
	for _, name := range installers.DefaultBrewPackages {
		if installed[name] {
			formulae = append(formulae, name)
		} else if coreName := "homebrew/core/" + name; installed[coreName] {
			formulae = append(formulae, coreName)
		}
	}
	if len(formulae) == 0 {
		displayPrintln("    [SKIP] No installed formulae match dk's curated default Homebrew list")
		return nil
	}

	displayPrintln("    Using local Homebrew metadata; run 'brew update' separately to refresh it.")
	displayPrintln("    Homebrew may also change dependencies and dependents.")
	args := append([]string{"upgrade", "--formula"}, formulae...)
	if dryRun {
		displayPrintf("    [PLAN] Would run: brew %s (auto-update and cleanup disabled)\n", strings.Join(args, " "))
		return nil
	}

	fmt.Fprintf(progressWriter(), "    Upgrading installed formulae from dk's default list: %s\n", strings.Join(formulae, ", "))
	if err := runBrewCommand(scopedBrewCommand(args...)); err != nil {
		return fmt.Errorf("%w; some formulae may already have been upgraded", err)
	}
	displayPrintln("    [OK] Command completed")
	return nil
}

func updateBrewAll(dryRun bool) error {
	if !installers.BrewInstalled() {
		displayPrintln("    [SKIP] Homebrew not installed, skipping")
		return nil
	}

	if dryRun {
		displayPrintln("    [PLAN] Would run: brew update && brew upgrade (all eligible formulae and casks)")
		return nil
	}

	fmt.Fprintln(progressWriter(), "    Refreshing Homebrew metadata...")
	if err := runBrewCommand(osexec.Command("brew", "update")); err != nil {
		return err
	}
	fmt.Fprintln(progressWriter(), "    Upgrading all eligible Homebrew formulae and casks...")
	if err := runBrewCommand(osexec.Command("brew", "upgrade")); err != nil {
		return fmt.Errorf("%w; some packages may already have been upgraded", err)
	}
	displayPrintln("    [OK] Command completed")
	return nil
}

func scopedBrewCommand(args ...string) *osexec.Cmd {
	cmd := osexec.Command("brew", args...)
	// A Homebrew metadata refresh can also migrate installed packages.
	cmd.Env = append(cmd.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_FORCE_API_AUTO_UPDATE=", "HOMEBREW_NO_INSTALL_CLEANUP=1")
	return cmd
}

func runBrewCommand(cmd *osexec.Cmd) error {
	cmd.Stdout = progressWriter()
	cmd.Stderr = cmd.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", strings.Join(cmd.Args, " "), err)
	}
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

func updateConfiguredNpm(cfg *config.Config, dryRun bool) error {
	if !exec.CommandExists("npm") {
		displayPrintln("    [SKIP] npm not installed, skipping")
		return nil
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
