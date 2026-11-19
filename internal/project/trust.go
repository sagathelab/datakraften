package project

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Verify(p *Project) error {
	if p == nil || p.Path == "" || p.Digest == "" {
		return errors.New("invalid project")
	}
	info, err := os.Lstat(p.Path)
	if err != nil {
		return fmt.Errorf("inspect project config: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("project config must be a regular file")
	}
	data, err := readConfig(p.Path)
	if err != nil {
		return fmt.Errorf("read project config: %w", err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != p.Digest {
		return errors.New("project config changed; inspect and approve it again")
	}
	return nil
}

func trustFile(p *Project) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find config directory: %w", err)
	}
	sum := sha256.Sum256([]byte(p.Path))
	return filepath.Join(configDir, "datakraften", "trust", hex.EncodeToString(sum[:])), nil
}

func Trusted(p *Project) (bool, error) {
	if err := Verify(p); err != nil {
		return false, err
	}
	path, err := trustFile(p)
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read project trust: %w", err)
	}
	return strings.TrimSpace(string(data)) == p.Digest, nil
}

func Trust(p *Project) error {
	if err := Verify(p); err != nil {
		return err
	}
	path, err := trustFile(p)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create trust directory: %w", err)
	}
	f, err := os.CreateTemp(dir, ".trust-*")
	if err != nil {
		return fmt.Errorf("create trust record: %w", err)
	}
	defer os.Remove(f.Name())
	_, writeErr := f.WriteString(p.Digest + "\n")
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("write trust record: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close trust record: %w", closeErr)
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("save trust record: %w", err)
	}
	return nil
}
