package shell

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func managedBlock(content string) (start, end int, err error) {
	start, end = -1, -1
	offset := 0
	for _, line := range strings.Split(content, "\n") {
		marker := strings.TrimSuffix(line, "\r")
		switch marker {
		case managedStart:
			if start >= 0 {
				return -1, -1, fmt.Errorf("duplicate managed block")
			}
			start = offset
		case managedEnd:
			if start < 0 || end >= 0 {
				return -1, -1, fmt.Errorf("unexpected managed block end")
			}
			end = offset + len(marker)
		}
		offset += len(line) + 1
	}
	if start >= 0 && end < 0 {
		return -1, -1, fmt.Errorf("unclosed managed block")
	}
	return start, end, nil
}

func readConfig(path string) (string, error) {
	link, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	info := link
	if link.Mode()&fs.ModeSymlink != 0 {
		info, err = os.Stat(path)
		if err != nil {
			return "", err
		}
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file")
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func ManagedBlockStatus(path string) string {
	content, err := readConfig(path)
	if err != nil {
		return "unreadable"
	}
	start, _, err := managedBlock(content)
	if err != nil {
		return "invalid"
	}
	if start >= 0 {
		return "present"
	}
	return "absent"
}

func writeConfig(path, config string) error {
	start, end, err := managedBlock(config)
	if err != nil || start < 0 {
		return fmt.Errorf("invalid managed shell configuration")
	}
	block := config[start:end]
	existing, err := readConfig(path)
	if err != nil {
		return fmt.Errorf("failed to read shell config %s: %w", path, err)
	}
	start, end, err = managedBlock(existing)
	if err != nil {
		return fmt.Errorf("invalid managed block in %s: %w", path, err)
	}
	newline := "\n"
	if first := strings.IndexByte(existing, '\n'); first > 0 && existing[first-1] == '\r' {
		newline = "\r\n"
		block = strings.ReplaceAll(block, "\n", newline)
	}
	var updated string
	if start < 0 {
		updated = existing
		if updated != "" && !strings.HasSuffix(updated, "\n") {
			updated += newline
		}
		updated += block + newline
	} else {
		updated = existing[:start] + block + existing[end:]
	}
	if updated == existing {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create shell config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write shell config: %w", err)
	}
	return nil
}
