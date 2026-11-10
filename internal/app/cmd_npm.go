package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/spf13/cobra"
)

func newNpmCmd() *cobra.Command {
	root := &cobra.Command{Use: "npm", Short: "Configure npm for your user"}
	var dryRun bool
	setup := &cobra.Command{Use: "setup", Short: "Set a user-owned npm prefix", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		displayTitle(os.Stdout, "npm setup", "Configure a user-owned npm prefix")
		if !exec.CommandExists("npm") {
			return fmt.Errorf("npm is not installed")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		prefix := filepath.Join(home, ".local", "share", "datakraften", "npm")
		shellName := filepath.Base(os.Getenv("SHELL"))
		var path, line string
		switch shellName {
		case "bash":
			path = filepath.Join(home, ".bashrc")
			line = `case ":$PATH:" in *:"$HOME/.local/share/datakraften/npm/bin":*) ;; *) export PATH="$HOME/.local/share/datakraften/npm/bin:$PATH" ;; esac`
		case "zsh":
			path = filepath.Join(home, ".zshrc")
			line = `case ":$PATH:" in *:"$HOME/.local/share/datakraften/npm/bin":*) ;; *) export PATH="$HOME/.local/share/datakraften/npm/bin:$PATH" ;; esac`
		case "fish":
			path = filepath.Join(home, ".config", "fish", "config.fish")
			line = `fish_add_path --path $HOME/.local/share/datakraften/npm/bin`
		default:
			return fmt.Errorf("unsupported shell %q; configure npm and PATH manually", shellName)
		}
		fmt.Println("  Configuration")
		fmt.Printf("    npm prefix: %s\n", prefix)
		fmt.Printf("    Shell config: %s\n", path)
		fmt.Println()
		fmt.Println("  Note: Existing global packages may need reinstalling; Node version managers may conflict with a custom prefix.")
		fmt.Println()
		if dryRun {
			fmt.Println("  [SKIP] Dry run: no changes made.")
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		const start = "# >>> datakraften npm >>>"
		const end = "# <<< datakraften npm <<<"
		content := string(data)
		a, b := strings.Index(content, start), strings.Index(content, end)
		if (a < 0) != (b < 0) || (a >= 0 && (b < a || strings.Count(content, start) != 1 || strings.Count(content, end) != 1)) {
			return fmt.Errorf("invalid npm managed block in %s", path)
		}
		block := start + "\n" + line + "\n" + end
		if a >= 0 {
			content = content[:a] + block + content[b+len(end):]
		} else {
			if content != "" && !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += block + "\n"
		}
		if err := os.MkdirAll(filepath.Join(prefix, "bin"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		r := exec.Run("npm", "config", "set", "prefix", prefix, "--location=user")
		if r.Code != 0 {
			return fmt.Errorf("npm config failed: %s", strings.TrimSpace(r.Stderr))
		}
		fmt.Println("  [OK] npm prefix and shell PATH configured")
		fmt.Println()
		fmt.Println("  Next steps")
		fmt.Println("    Restart your shell to use the new PATH.")
		return nil
	}}
	setup.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show changes without applying them")
	root.AddCommand(setup)
	return root
}
