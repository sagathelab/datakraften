package projectnode

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
/bin/mkdir -p "$FNM_DIR/node-versions/v22.11.0/installation/bin"
printf '#!/bin/sh\nprintf "v22.11.0\\n"\n' > "$FNM_DIR/node-versions/v22.11.0/installation/bin/node"
/bin/chmod 755 "$FNM_DIR/node-versions/v22.11.0/installation/bin/node"
`
	if err := os.WriteFile(fnm, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(root, "fnm.log")
	t.Setenv("TEST_FNM_LOG", log)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", root)
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("FNM_DIR", filepath.Join(root, "user-fnm"))
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
	if err != nil || !strings.HasPrefix(inst.NodePath, filepath.Join(root, "data", "datakraften", "fnm")) {
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
	input := []string{"PATH=/usr/bin:/bin", "A=B"}
	result := Environment(inst, input)
	if input[0] != "PATH=/usr/bin:/bin" || !strings.Contains(strings.Join(result, "\n"), "PATH="+bin+":/usr/bin:/bin") {
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
