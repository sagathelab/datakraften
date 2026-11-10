package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeConfig(path, config string) error {
	start := strings.Index(config, managedStart)
	end := strings.Index(config, managedEnd)
	if start < 0 || end < start || strings.Count(config, managedStart) != 1 || strings.Count(config, managedEnd) != 1 {
		return fmt.Errorf("invalid managed shell configuration")
	}
	block := config[start : end+len(managedEnd)]
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read shell config: %w", err)
	}
	existing := string(data)
	start = strings.Index(existing, managedStart)
	end = strings.Index(existing, managedEnd)
	if start < 0 && end < 0 {
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		existing += block + "\n"
	} else if start < 0 || end < start || strings.Count(existing, managedStart) != 1 || strings.Count(existing, managedEnd) != 1 {
		return fmt.Errorf("invalid managed block in %s", path)
	} else {
		existing = existing[:start] + block + existing[end+len(managedEnd):]
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create shell config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		return fmt.Errorf("failed to write shell config: %w", err)
	}
	return nil
}
