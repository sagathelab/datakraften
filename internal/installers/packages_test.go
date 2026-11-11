package installers

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDNFInstalledPackagesAndGroups(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	for name, script := range map[string]string{
		"rpm": "#!/bin/sh\nif [ \"$1\" = -q ] && [ \"$2\" = --whatprovides ] && [ \"$3\" = wget ]; then exit 0; fi\nexit 1\n",
		"dnf": "#!/bin/sh\nprintf 'ID Name Installed\\ndevelopment-tools Development Tools yes\\n'\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if !DNF.CheckInstalled("wget") || !DNF.CheckInstalled("@development-tools") {
		t.Fatal("installed provider or group not detected")
	}
	if DNF.CheckInstalled("missing") || DNF.CheckInstalled("@missing") {
		t.Fatal("missing package or group detected as installed")
	}
}

func TestAPTUpdateOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	SetOutput(&output)
	t.Cleanup(func() { SetOutput(nil) })
	if err := APTUpdate(); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "    Updating APT cache...\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}
