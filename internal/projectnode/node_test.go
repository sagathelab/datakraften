package projectnode

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeFnm(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	fnm := filepath.Join(bin, "fnm")
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$TEST_FNM_LOG"
[ "$1" = install ] || exit 2
[ "$2" = 22 ] || exit 3
[ "$PWD" -ef "$FNM_DIR" ] || exit 4
[ -z "$FNM_MULTISHELL_PATH$FNM_COREPACK_ENABLED" ] || exit 5
/bin/mkdir -p "$FNM_DIR/node-versions/v22.11.0/installation/bin"
printf '#!/bin/sh\nprintf "v22.11.0\\n"\n' > "$FNM_DIR/node-versions/v22.11.0/installation/bin/node"
/bin/chmod 755 "$FNM_DIR/node-versions/v22.11.0/installation/bin/node"
`
	if err := os.WriteFile(fnm, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(root, "fnm.log")
	t.Setenv("TEST_FNM_LOG", log)
	t.Setenv("PATH", bin)
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("FNM_DIR", filepath.Join(root, "user-fnm"))
	t.Setenv("FNM_MULTISHELL_PATH", filepath.Join(root, "current-node"))
	t.Setenv("FNM_COREPACK_ENABLED", "true")
	return root, log
}

func addVersion(t *testing.T, root, v string) string {
	t.Helper()
	bin := filepath.Join(root, "node-versions", "v"+v, "installation", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "node"), []byte("#!/bin/sh\necho v"+v+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestResolveMatchesHighestInstalledVersionWithoutRunningFnm(t *testing.T) {
	root, log := fakeFnm(t)
	store := filepath.Join(root, "user-fnm")
	addVersion(t, store, "22.9.0")
	best := addVersion(t, store, "22.10.1")
	addVersion(t, store, "24.1.0")
	inst, err := Resolve("22")
	if err != nil || inst.Version != "22.10.1" || inst.BinDir != best || inst.Backend != "fnm" {
		t.Fatalf("got %+v, %v", inst, err)
	}
	inst, err = Resolve("v22.9")
	if err != nil || inst.Version != "22.9.0" {
		t.Fatalf("got %+v, %v", inst, err)
	}
	if _, err := Resolve("23"); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("missing version error: %v", err)
	}
	if _, err := Resolve("--latest"); !errors.Is(err, ErrInvalidVersion) {
		t.Fatalf("invalid version error: %v", err)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("resolve executed fnm")
	}
}

func TestInstallUsesPrivateFnmStoreWithoutChangingUserDefault(t *testing.T) {
	root, log := fakeFnm(t)
	userStore := filepath.Join(root, "user-fnm")
	if err := os.MkdirAll(filepath.Join(userStore, "aliases"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userStore, "aliases", "default"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Install("22", &out, &out); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(userStore, "aliases", "default")); err != nil || string(got) != "original" {
		t.Fatalf("user default changed: %q, %v", got, err)
	}
	inst, err := Resolve("22")
	if err != nil || inst.StorePath != filepath.Join(root, "data", "datakraften", "fnm") {
		t.Fatalf("unexpected installation: %+v, %v", inst, err)
	}
	cmd := exec.Command("/bin/sh", "-c", "node --version")
	cmd.Env = Environment(inst, []string{"PATH=/somewhere", "OTHER=value"})
	got, err := cmd.Output()
	if err != nil || string(got) != "v22.11.0\n" {
		t.Fatalf("running installed version: %q, %v", got, err)
	}
	if err := Install("22", &out, &out); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(log)
	if err != nil || string(calls) != "install 22\n" {
		t.Fatalf("install should be idempotent: %q, %v", calls, err)
	}
}

func TestEnvironmentUsesSelectedNodeWithoutMutatingInput(t *testing.T) {
	root, _ := fakeFnm(t)
	bin := addVersion(t, filepath.Join(root, "user-fnm"), "24.2.1")
	inst, err := Resolve("24")
	if err != nil {
		t.Fatal(err)
	}
	input := []string{"PATH=/usr/bin:/bin", "A=B", "NPM_CONFIG_PREFIX=/npm"}
	result := Environment(inst, input)
	if input[0] != "PATH=/usr/bin:/bin" || strings.Join(result, "\n") != "A=B\nNPM_CONFIG_PREFIX=/npm\nPATH="+bin+":/usr/bin:/bin" {
		t.Fatalf("input=%v result=%v", input, result)
	}
}

func TestMissingFnmDoesNotInstallAnything(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := Resolve("22"); !errors.Is(err, ErrBackendNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := Install("22", nil, nil); !errors.Is(err, ErrBackendNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestFnmDarwinLegacyStore(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("legacy macOS fnm store")
	}
	root, _ := fakeFnm(t)
	t.Setenv("FNM_DIR", "")
	bin := addVersion(t, filepath.Join(root, "Library", "Application Support", "fnm"), "22.10.0")
	inst, err := Resolve("22.10.0")
	if err != nil || inst.BinDir != bin {
		t.Fatalf("resolve legacy store: %+v, %v", inst, err)
	}
}

func TestInstallRejectsSharedStore(t *testing.T) {
	root, log := fakeFnm(t)
	store := filepath.Join(root, "data", "datakraften", "fnm")
	if err := os.MkdirAll(filepath.Dir(store), 0700); err != nil {
		t.Fatal(err)
	}
	user := os.Getenv("FNM_DIR")
	if err := os.Mkdir(user, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(user, store); err != nil {
		t.Fatal(err)
	}
	if err := Install("22", io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), "also a user fnm store") {
		t.Fatalf("shared store: %v", err)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatalf("installer ran for shared store: %v", err)
	}
}

func TestNvmResolvesAndInstallsWithoutChangingUserDefault(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("NVM_DIR", filepath.Join(root, "user nvm"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	dir := os.Getenv("NVM_DIR")
	if err := os.MkdirAll(filepath.Join(dir, "alias"), 0755); err != nil {
		t.Fatal(err)
	}
	script := `[ "$1" = --no-use ] || return 2
nvm() {
  [ "$1" = install ] && [ "$2" = --no-progress ] && [ "$3" = 22 ] || return 3
  /bin/mkdir -p "$NVM_DIR/versions/node/v22.11.0/bin"
  printf '#!/bin/sh\necho v22.11.0\n' > "$NVM_DIR/versions/node/v22.11.0/bin/node"
  /bin/chmod 755 "$NVM_DIR/versions/node/v22.11.0/bin/node"
}
`
	if err := os.WriteFile(filepath.Join(dir, "nvm.sh"), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "alias", "default"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveFor("nvm", "22"); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("missing version: %v", err)
	}
	if err := InstallFor("nvm", "22", io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	inst, err := ResolveFor("nvm", "22")
	if err != nil || inst.Backend != "nvm" || inst.Version != "22.11.0" {
		t.Fatalf("resolve: %+v, %v", inst, err)
	}
	cmd := exec.Command("/bin/sh", "-c", "node --version")
	cmd.Env = Environment(inst, []string{"PATH=/usr/bin:/bin"})
	if got, err := cmd.Output(); err != nil || string(got) != "v22.11.0\n" {
		t.Fatalf("run node: %q, %v", got, err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "alias", "default")); err != nil || string(data) != "original" {
		t.Fatalf("user default changed: %q, %v", data, err)
	}
}

func TestNvmReusesExistingNode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("HOME", dir)
	t.Setenv("NVM_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	store := filepath.Join(dir, ".nvm")
	bin := filepath.Join(store, "versions", "node", "v22.3.1", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "nvm.sh"), []byte("exit 99\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "node"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	inst, err := ResolveFor("nvm", "22")
	if err != nil || inst.NodePath != filepath.Join(bin, "node") {
		t.Fatalf("resolve existing: %+v, %v", inst, err)
	}
	if err := InstallFor("nvm", "22", io.Discard, io.Discard); err != nil {
		t.Fatalf("existing installation should not source nvm: %v", err)
	}
	if _, err := ResolveFor("fnm", "22"); !errors.Is(err, ErrBackendNotFound) {
		t.Fatalf("missing fnm should not fall back to nvm: %v", err)
	}
}

func TestInstallVerifiesPrivateStoreAndNamesManagerErrors(t *testing.T) {
	root, _ := fakeFnm(t)
	nvm := filepath.Join(root, "user nvm")
	t.Setenv("NVM_DIR", nvm)
	if err := os.Mkdir(nvm, 0700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(nvm, "nvm.sh")
	for _, tc := range []struct{ script, want string }{
		{"nvm() { return 7; }\n", "nvm install 24 failed: exit status 7"},
		{"nvm() { return 0; }\n", "nvm install did not produce an executable Node.js 24"},
	} {
		if err := os.WriteFile(script, []byte(tc.script), 0644); err != nil {
			t.Fatal(err)
		}
		if err := InstallFor("nvm", "24", io.Discard, io.Discard); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("expected %q, got %v", tc.want, err)
		}
	}
	if err := os.WriteFile(filepath.Join(nvm, "versions"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveFor("nvm", "24"); err == nil || !strings.Contains(err.Error(), "read nvm installations") {
		t.Fatalf("expected nvm store error, got %v", err)
	}
}
