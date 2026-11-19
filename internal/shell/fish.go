package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/system"
)

const managedStart = "# >>> datakraften >>>"
const managedEnd = "# <<< datakraften <<<"

func FishInstalled() bool {
	return exec.CommandExists("fish")
}

func FishConfigPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "fish", "config.fish")
	}
	return filepath.Join(system.HomeDir(), ".config", "fish", "config.fish")
}

func FishGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString(managedStart + "\n")
	sb.WriteString("fish_add_path --path $HOME/.local/bin\n")

	if brewPrefix != "" {
		brew := filepath.Join(brewPrefix, "bin", "brew")
		brew = "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(brew) + "'"
		fmt.Fprintf(&sb, "if test -x %s; %s shellenv fish | source; end\n", brew, brew)
	}

	if fnmConfigured {
		sb.WriteString("if command -q fnm; fnm env --use-on-cd --shell fish | source; end\n")
	}

	if uvConfigured {
		sb.WriteString("if command -q uv; uv generate-shell-completion fish | source; end\n")
	}

	if enableStarship {
		sb.WriteString("if command -q starship; starship init fish | source; end\n")
	}

	if enableAtuin {
		sb.WriteString("if command -q atuin; atuin init fish | source; end\n")
	}

	if enableFzf {
		sb.WriteString("if command -q fzf; fzf --fish | source; end\n")
	}

	sb.WriteString(managedEnd + "\n")

	return sb.String()
}

func FishWriteConfig(config string) error {
	return writeConfig(FishConfigPath(), config)
}
