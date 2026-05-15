package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/system"
)

func ZshConfigPath() string {
	return filepath.Join(system.HomeDir(), ".zshrc")
}

func ZshDetectExistingConfig() string {
	path := ZshConfigPath()
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func ZshHasManagedBlock() bool {
	content := ZshDetectExistingConfig()
	return strings.Contains(content, managedStart)
}

func ZshGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString("# Datakraften managed configuration\n")
	sb.WriteString("# Edit with caution — managed blocks are regenerated on 'dk apply'\n\n")

	sb.WriteString(managedStart + "\n")

	if brewPrefix != "" {
		sb.WriteString(fmt.Sprintf("eval \"$(%s/bin/brew shellenv)\"\n", brewPrefix))
	}

	if fnmConfigured {
		sb.WriteString("eval \"$(fnm env --use-on-cd --shell zsh)\"\n")
	}

	if uvConfigured {
		sb.WriteString("eval \"$(uv generate-shell-completion zsh)\"\n")
	}

	if enableStarship {
		sb.WriteString("eval \"$(starship init zsh)\"\n")
	}

	if enableAtuin {
		sb.WriteString("eval \"$(atuin init zsh)\"\n")
	}

	if enableFzf {
		sb.WriteString("source <(fzf --zsh)\n")
	}

	sb.WriteString("export EDITOR=\"code --wait\"\n")

	sb.WriteString(managedEnd + "\n")

	if !ZshHasManagedBlock() {
		sb.WriteString("\n# User configuration below\n# Add your custom config here\n")
	}

	return sb.String()
}

func ZshWriteConfig(config string) error {
	configDir := filepath.Dir(ZshConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create zsh config dir: %w", err)
	}

	existing := ZshDetectExistingConfig()

	if strings.Contains(existing, managedStart) {
		startIdx := strings.Index(existing, managedStart)
		endIdx := strings.Index(existing, managedEnd)
		if endIdx > startIdx {
			before := existing[:startIdx]
			after := existing[endIdx+len(managedEnd):]
			existing = before + config + after
		}
	} else {
		existing = config
	}

	if err := os.WriteFile(ZshConfigPath(), []byte(existing), 0644); err != nil {
		return fmt.Errorf("failed to write zsh config: %w", err)
	}
	return nil
}

func ZshEnsureSetup(brewPrefix string) (bool, error) {
	config := ZshGenerateConfig(brewPrefix, true, true, true, true, true)
	if err := ZshWriteConfig(config); err != nil {
		return false, err
	}
	return true, nil
}
