package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/system"
)

func ZshConfigPath() string {
	if dir := os.Getenv("ZDOTDIR"); filepath.IsAbs(dir) {
		return filepath.Join(dir, ".zshrc")
	}
	return filepath.Join(system.HomeDir(), ".zshrc")
}

func ZshGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString(managedStart + "\n")
	sb.WriteString("case \":$PATH:\" in *:\"$HOME/.local/bin\":*) ;; *) export PATH=\"$HOME/.local/bin:$PATH\" ;; esac\n")

	if brewPrefix != "" {
		brew := posixQuote(filepath.Join(brewPrefix, "bin", "brew"))
		fmt.Fprintf(&sb, "[ ! -x %s ] || eval \"$(%s shellenv zsh)\"\n", brew, brew)
	}

	if fnmConfigured {
		sb.WriteString("! command -v fnm >/dev/null 2>&1 || eval \"$(fnm env --use-on-cd --shell zsh)\"\n")
	}

	if uvConfigured {
		sb.WriteString("! command -v uv >/dev/null 2>&1 || eval \"$(uv generate-shell-completion zsh)\"\n")
	}

	if enableStarship {
		sb.WriteString("! command -v starship >/dev/null 2>&1 || eval \"$(starship init zsh)\"\n")
	}

	if enableAtuin {
		sb.WriteString("! command -v atuin >/dev/null 2>&1 || eval \"$(atuin init zsh)\"\n")
	}

	if enableFzf {
		sb.WriteString("! command -v fzf >/dev/null 2>&1 || source <(fzf --zsh)\n")
	}

	sb.WriteString(managedEnd + "\n")

	return sb.String()
}

func ZshWriteConfig(config string) error {
	return writeConfig(ZshConfigPath(), config)
}
