package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellConfigPreservesUserContent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		path     func() string
		generate func(string, bool, bool, bool, bool, bool) string
		write    func(string) error
	}{
		{"bash", BashConfigPath, BashGenerateConfig, BashWriteConfig},
		{"zsh", ZshConfigPath, ZshGenerateConfig, ZshWriteConfig},
		{"fish", FishConfigPath, FishGenerateConfig, FishWriteConfig},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			path := tc.path()
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("user setting"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := tc.write(tc.generate("", true, false, false, false, false)); err != nil {
				t.Fatal(err)
			}
			first, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(first), "user setting\n"+managedStart) {
				t.Fatalf("user content lost: %q", first)
			}
			if err := tc.write(tc.generate("", true, false, false, false, false)); err != nil {
				t.Fatal(err)
			}
			second, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(first) != string(second) {
				t.Fatalf("not idempotent: %q != %q", first, second)
			}
			if err := tc.write(tc.generate("", false, false, false, false, false)); err != nil {
				t.Fatal(err)
			}
			updated, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(updated), "starship") || !strings.HasPrefix(string(updated), "user setting\n") {
				t.Fatalf("incorrect update: %q", updated)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0600 {
				t.Fatalf("permissions changed: %v", info.Mode())
			}
		})
	}
}

func TestWriteConfigRejectsMalformedExistingBlock(t *testing.T) {
	for _, existing := range []string{managedStart + "\nuser data\n", managedEnd + "\n", managedStart + "\n" + managedEnd + "\n" + managedStart + "\n" + managedEnd} {
		t.Run(existing, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			if err := os.WriteFile(path, []byte(existing), 0600); err != nil {
				t.Fatal(err)
			}
			if err := writeConfig(path, managedStart+"\nnew\n"+managedEnd); err == nil {
				t.Fatal("expected error")
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != existing {
				t.Fatalf("changed malformed config: %q", got)
			}
		})
	}
}
