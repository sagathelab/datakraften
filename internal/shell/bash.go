package shell

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/system"
)

func BashConfigPath() string {
	return filepath.Join(system.HomeDir(), ".bashrc")
}

func BashGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString(managedStart + "\n")
	sb.WriteString("case \":$PATH:\" in *:\"$HOME/.local/bin\":*) ;; *) export PATH=\"$HOME/.local/bin:$PATH\" ;; esac\n")

	if brewPrefix != "" {
		brew := posixQuote(filepath.Join(brewPrefix, "bin", "brew"))
		fmt.Fprintf(&sb, "[ ! -x %s ] || eval \"$(%s shellenv bash)\"\n", brew, brew)
	}

	if fnmConfigured {
		sb.WriteString("! command -v fnm >/dev/null 2>&1 || eval \"$(fnm env --use-on-cd --shell bash)\"\n")
	}

	if uvConfigured {
		sb.WriteString("! command -v uv >/dev/null 2>&1 || eval \"$(uv generate-shell-completion bash)\"\n")
	}

	if enableStarship {
		sb.WriteString("! command -v starship >/dev/null 2>&1 || eval \"$(starship init bash)\"\n")
	}

	if enableAtuin {
		sb.WriteString("! command -v atuin >/dev/null 2>&1 || eval \"$(atuin init bash)\"\n")
	}

	if enableFzf {
		sb.WriteString("! command -v fzf >/dev/null 2>&1 || eval \"$(fzf --bash)\"\n")
	}

	sb.WriteString(managedEnd + "\n")

	return sb.String()
}

func posixQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func BashWriteConfig(config string) error {
	return writeConfig(BashConfigPath(), config)
}
