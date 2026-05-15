package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagathelab/datakraften/internal/system"
)

func BashConfigPath() string {
	return filepath.Join(system.HomeDir(), ".bashrc")
}

func BashDetectExistingConfig() string {
	path := BashConfigPath()
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func BashHasManagedBlock() bool {
	content := BashDetectExistingConfig()
	return strings.Contains(content, managedStart)
}

func BashGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString("# Datakraften managed configuration\n")
	sb.WriteString("# Edit with caution — managed blocks are regenerated on 'dk apply'\n\n")

	sb.WriteString(managedStart + "\n")

	if brewPrefix != "" {
		sb.WriteString(fmt.Sprintf("eval \"$(%s/bin/brew shellenv)\"\n", brewPrefix))
	}

	if fnmConfigured {
		sb.WriteString("eval \"$(fnm env --use-on-cd --shell bash)\"\n")
	}

	if uvConfigured {
		sb.WriteString("eval \"$(uv generate-shell-completion bash)\"\n")
	}

	if enableStarship {
		sb.WriteString("eval \"$(starship init bash)\"\n")
	}

	if enableAtuin {
		sb.WriteString("eval \"$(atuin init bash)\"\n")
	}

	if enableFzf {
		sb.WriteString("eval \"$(fzf --bash)\"\n")
	}

	sb.WriteString("export EDITOR=\"code --wait\"\n")

	sb.WriteString(managedEnd + "\n")

	if !BashHasManagedBlock() {
		sb.WriteString("\n# User configuration below\n# Add your custom config here\n")
	}

	return sb.String()
}

func BashWriteConfig(config string) error {
	configDir := filepath.Dir(BashConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create bash config dir: %w", err)
	}

	existing := BashDetectExistingConfig()

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

	if err := os.WriteFile(BashConfigPath(), []byte(existing), 0644); err != nil {
		return fmt.Errorf("failed to write bash config: %w", err)
	}
	return nil
}

func BashEnsureSetup(brewPrefix string) (bool, error) {
	config := BashGenerateConfig(brewPrefix, true, true, true, true, true)
	if err := BashWriteConfig(config); err != nil {
		return false, err
	}
	return true, nil
}
