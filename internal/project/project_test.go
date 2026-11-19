package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = "version: 1\nproject:\n  tools:\n    node: '22'\n  env:\n    NODE_ENV: development\n  tasks:\n    test: npm test\n"

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverGitBoundary(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "repo")
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(base, "datakraften.yaml"), validConfig)
	if got, err := Discover(sub); err != nil || got != nil {
		t.Fatalf("crossed repo boundary: project=%+v err=%v", got, err)
	}
	writeFile(t, filepath.Join(root, "datakraften.yaml"), validConfig)
	got, err := Discover(sub)
	if err != nil || got == nil || got.Root != root || got.Config.Project.Tools["node"].Version != "22" {
		t.Fatalf("root project: %+v, %v", got, err)
	}
	near := filepath.Join(root, "a", "datakraften.yaml")
	writeFile(t, near, strings.Replace(validConfig, "'22'", "'24'", 1))
	got, err = Discover(sub)
	if err != nil || got == nil || got.Path != near || got.Config.Project.Tools["node"].Version != "24" {
		t.Fatalf("nearest project: %+v, %v", got, err)
	}
	nested := filepath.Join(root, "nested")
	writeFile(t, filepath.Join(nested, ".git"), "gitdir: /tmp/worktree\n")
	got, err = Discover(nested)
	if err != nil || got != nil {
		t.Fatalf("crossed nested worktree: %+v, %v", got, err)
	}
}

func TestNodeManagerConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, tool, version, manager string
		invalid                      bool
	}{
		{"shorthand", `"22"`, "22", "fnm", false},
		{"numeric shorthand", `22`, "22", "fnm", false},
		{"nvm", "\n      version: '24'\n      manager: nvm", "24", "nvm", false},
		{"fnm", "\n      version: 22.10\n      manager: fnm", "22.10", "fnm", false},
		{"unknown", "\n      version: '22'\n      manager: other", "", "", true},
		{"missing version", "\n      manager: nvm", "", "", true},
		{"missing manager", "\n      version: '22'", "", "", true},
		{"numeric manager", "\n      version: '22'\n      manager: 1", "", "", true},
		{"duplicate manager", "\n      version: '22'\n      manager: fnm\n      manager: nvm", "", "", true},
		{"unknown option", "\n      version: '22'\n      option: value", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "datakraften.yaml"), "version: 1\nproject:\n  tools:\n    node: "+tc.tool+"\n")
			p, err := Discover(dir)
			if tc.invalid {
				if err == nil {
					t.Fatal("invalid configuration accepted")
				}
				return
			}
			if err != nil || p.Config.Project.Tools["node"].Version != tc.version || p.Config.Project.Tools["node"].Manager != tc.manager {
				t.Fatalf("got %+v, %v", p, err)
			}
		})
	}
}

func TestDiscoverOutsideGitAndSymlinks(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(base, "child")
	if err := os.Mkdir(child, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(base, "datakraften.yaml"), validConfig)
	t.Run("outside git", func(t *testing.T) {
		gitRoot, err := findGitRoot(child)
		if err != nil {
			t.Fatal(err)
		}
		if gitRoot != "" {
			t.Skip("test scratch directory is inside a Git repository")
		}
		if got, err := Discover(child); err != nil || got != nil {
			t.Fatalf("searched outside standalone directory: %+v, %v", got, err)
		}
	})
	if err := os.Symlink(filepath.Join(base, "datakraften.yaml"), filepath.Join(child, "datakraften.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(child); err == nil {
		t.Fatal("accepted symlink as project config")
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(base, alias); err != nil {
		t.Fatal(err)
	}
	got, err := Discover(alias)
	if err != nil || got == nil || got.Root != base {
		t.Fatalf("canonical project directory: %+v, %v", got, err)
	}
}

func TestInvalidProjectConfig(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"invalid version", strings.Replace(validConfig, "version: 1", "version: 2", 1)},
		{"missing project", "version: 1\n"},
		{"invalid project", "version: 1\nproject: []\n"},
		{"unknown field", validConfig + "  unexpected: true\n"},
		{"duplicate key", validConfig + "  tools: {}\n"},
		{"unsupported tool", strings.Replace(validConfig, "node: '22'", "python: '3.14'", 1)},
		{"invalid env", strings.Replace(validConfig, "NODE_ENV", "BAD-NAME", 1)},
		{"invalid task", strings.Replace(validConfig, "test: npm test", "test: ''", 1)},
		{"null", strings.Replace(validConfig, "development", "null", 1)},
		{"alias", "version: 1\nproject:\n  env: &value {FOO: bar}\n  tasks: *value\n"},
		{"multiple documents", validConfig + "---\nversion: 1\n"},
		{"oversized", strings.Repeat("x", maxConfigSize+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "datakraften.yaml"), tc.data)
			if _, err := Discover(dir); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}

func TestTrustTracksExactFileAndContent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "datakraften.yaml")
	writeFile(t, path, validConfig)
	p, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if trusted, err := Trusted(p); err != nil || trusted {
		t.Fatalf("initial trust: %v, %v", trusted, err)
	}
	if err := Trust(p); err != nil {
		t.Fatal(err)
	}
	if trusted, err := Trusted(p); err != nil || !trusted {
		t.Fatalf("approved trust: %v, %v", trusted, err)
	}
	writeFile(t, path, validConfig+"# changed\n")
	if _, err := Trusted(p); err == nil {
		t.Fatal("old project snapshot accepted after modification")
	}
	changed, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if trusted, err := Trusted(changed); err != nil || trusted {
		t.Fatalf("changed config trust: %v, %v", trusted, err)
	}
	otherDir := t.TempDir()
	writeFile(t, filepath.Join(otherDir, "datakraften.yaml"), validConfig)
	other, err := Discover(otherDir)
	if err != nil {
		t.Fatal(err)
	}
	if trusted, err := Trusted(other); err != nil || trusted {
		t.Fatalf("trust leaked to another path: %v, %v", trusted, err)
	}
}
