package app

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
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
		Use:   "update",
		Short: "Update managed tools to their latest versions",
		Long:  `Update all managed developer tools (brew packages, runtimes, AI tools) or a specific tool by name.`,
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

	cmd.Flags().BoolVarP(&list, "list", "l", false, "List tools that have updates available")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be updated without making changes")

	return cmd
}

func listUpdatable() error {
	fmt.Println("  Updatable tools")
	fmt.Println("  ---------------")

	fmt.Println("    brew — all Homebrew packages")
	fmt.Println("    fnm — Node.js runtime")
	fmt.Println("    uv — Python runtime")
	fmt.Println("    npm — global npm packages")

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
	fmt.Println("  Updating all managed tools")
	fmt.Println("  -------------------------")

	var failures []string

	for _, step := range updateSteps {
		if err := step.run(dryRun); err != nil {
			fmt.Printf("    ✗ %s: %s\n", step.name, err)
			failures = append(failures, fmt.Sprintf("%s: %s", step.name, err))
		}
	}

	fmt.Println()
	if len(failures) > 0 {
		fmt.Println("  ✗ Update finished with errors")
		return fmt.Errorf("update failed: %s", strings.Join(failures, "; "))
	}

	fmt.Println("  ✓ Update complete")
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

	fmt.Print("    Updating Homebrew...")
	r := exec.Run("brew", "update")
	if r.Code != 0 {
		return fmt.Errorf("brew update failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println(" done")

	fmt.Print("    Upgrading packages...")
	r = exec.Run("brew", "upgrade")
	if r.Code != 0 {
		return fmt.Errorf("brew upgrade failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println(" done")

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

	fmt.Print("    Updating Node.js LTS via fnm...")
	r := exec.Run("fnm", "install", "--lts")
	if r.Code != 0 {
		return fmt.Errorf("fnm install lts failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println(" done")

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

	fmt.Print("    Updating uv...")
	r := exec.Run("uv", "self", "update")
	if r.Code != 0 {
		return fmt.Errorf("uv self update failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println(" done")

	return nil
}

func updateNpm(dryRun bool) error {
	if !exec.CommandExists("npm") {
		fmt.Println("    – npm not installed, skipping")
		return nil
	}

	if dryRun {
		fmt.Println("    ~ Would run: npm update -g")
		return nil
	}

	fmt.Print("    Updating global npm packages...")
	r := exec.Run("npm", "update", "-g")
	if r.Code != 0 {
		return fmt.Errorf("npm update -g failed: %s", strings.TrimSpace(r.Stderr))
	}
	fmt.Println(" done")

	return nil
}
