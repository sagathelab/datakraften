package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellConfigPreservesUserContent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("ZDOTDIR", "")
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

func TestShellConfigPathsHonorAbsoluteOverrides(t *testing.T) {
	home, dir := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("ZDOTDIR", dir)
	if path := FishConfigPath(); path != filepath.Join(dir, "fish", "config.fish") {
		t.Errorf("unexpected Fish path: %q", path)
	}
	if path := ZshConfigPath(); path != filepath.Join(dir, ".zshrc") {
		t.Errorf("unexpected Zsh path: %q", path)
	}
	if path := BashConfigPath(); path != filepath.Join(home, ".bashrc") {
		t.Errorf("unexpected Bash path: %q", path)
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

func TestWriteConfigTreatsInlineMarkerTextAsUserContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	before := "printf '%s' '" + managedStart + "'\r\n"
	after := "printf '%s' '" + managedEnd + "'\r\n"
	original := before + managedStart + "\r\nold\r\n" + managedEnd + "\r\n" + after
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeConfig(path, managedStart+"\nnew\n"+managedEnd); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if want := strings.Replace(original, "\r\nold\r\n", "\r\nnew\r\n", 1); err != nil || string(got) != want {
		t.Fatalf("expected %q, got %q: %v", want, got, err)
	}
}

func TestManagedBlockStatusDoesNotFollowDanglingSymlinkToCreateFile(t *testing.T) {
	dir := t.TempDir()
	path, target := filepath.Join(dir, "config"), filepath.Join(dir, "missing")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if got := ManagedBlockStatus(path); got != "unreadable" {
		t.Errorf("expected unreadable, got %q", got)
	}
	if err := writeConfig(path, managedStart+"\nnew\n"+managedEnd); err == nil {
		t.Fatal("expected write error")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dangling link target was changed: %v", err)
	}
}
