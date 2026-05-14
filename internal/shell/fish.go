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
	return filepath.Join(system.HomeDir(), ".config", "fish", "config.fish")
}

func FishDetectExistingConfig() string {
	path := FishConfigPath()
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func FishHasManagedBlock() bool {
	content := FishDetectExistingConfig()
	return strings.Contains(content, managedStart)
}

func FishGenerateConfig(brewPrefix string, enableStarship bool, enableAtuin bool, enableFzf bool, fnmConfigured bool, uvConfigured bool) string {
	var sb strings.Builder

	sb.WriteString("# Datakraften managed configuration\n")
	sb.WriteString("# Edit with caution — managed blocks are regenerated on 'dk apply'\n\n")

	sb.WriteString(managedStart + "\n")

	if brewPrefix != "" {
		sb.WriteString(fmt.Sprintf("eval (%s/bin/brew shellenv)\n", brewPrefix))
	}

	if fnmConfigured {
		sb.WriteString("fnm env --use-on-cd --shell fish | source\n")
	}

	if uvConfigured {
		sb.WriteString("uv generate-shell-completion fish | source\n")
	}

	if enableStarship {
		sb.WriteString("starship init fish | source\n")
	}

	if enableAtuin {
		sb.WriteString("atuin init fish | source\n")
	}

	if enableFzf {
		sb.WriteString("fzf --fish | source\n")
	}

	sb.WriteString("set -gx EDITOR \"code --wait\"\n")

	sb.WriteString(managedEnd + "\n")

	if !FishHasManagedBlock() {
		sb.WriteString("\n# User configuration below\n# Add your custom config here\n")
	}

	return sb.String()
}

func FishWriteConfig(config string) error {
	configDir := filepath.Dir(FishConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create fish config dir: %w", err)
	}

	existing := FishDetectExistingConfig()

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

	if err := os.WriteFile(FishConfigPath(), []byte(existing), 0644); err != nil {
		return fmt.Errorf("failed to write fish config: %w", err)
	}
	return nil
}

func FishSetDefault() error {
	if !FishInstalled() {
		return fmt.Errorf("fish is not installed")
	}

	fishPath := "/usr/bin/fish"
	if p := exec.CommandPath("fish"); p != "" {
		fishPath = p
	}

	r := exec.Run("sudo", "chsh", "-s", fishPath, system.UserName())
	if r.Code != 0 {
		return fmt.Errorf("failed to set fish as default shell: %s", r.Stderr)
	}
	return nil
}

func FishEnsureSetup(brewPrefix string) (bool, error) {
	if !FishInstalled() {
		return false, fmt.Errorf("fish is not installed")
	}

	config := FishGenerateConfig(brewPrefix, true, true, true, true, true)
	if err := FishWriteConfig(config); err != nil {
		return false, err
	}

	return true, nil
}
