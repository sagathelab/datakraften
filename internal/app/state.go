package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	Version      string   `json:"version"`
	LastApply    string   `json:"last_apply,omitempty"`
	ActiveProfile string  `json:"active_profile"`
	InstalledTools []string `json:"installed_tools,omitempty"`
	ManagedShell  bool     `json:"managed_shell"`
}

func stateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "datakraften")
}

func statePath() string {
	return filepath.Join(stateDir(), "state.json")
}

func logDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "datakraften", "logs")
}

func LoadState() *State {
	s := &State{}
	data, err := os.ReadFile(statePath())
	if err != nil {
		return s
	}
	json.Unmarshal(data, s)
	return s
}

func (s *State) Save() error {
	dir := stateDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	if err := os.WriteFile(statePath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}
	return nil
}

func (s *State) RecordApply(profile string) {
	s.LastApply = time.Now().Format(time.RFC3339)
	s.ActiveProfile = profile
	s.Save()
}

func WriteLog(name string, content string) error {
	dir := logDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.log", ts, name))
	return os.WriteFile(path, []byte(content), 0644)
}
