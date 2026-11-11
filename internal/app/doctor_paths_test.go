package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/doctor"
)

func TestDoctorExecutable(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "program")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := doctorExecutable(dir, file); got != "" {
		t.Fatalf("non-executable candidate found: %q", got)
	}
	if err := os.Chmod(file, 0755); err != nil {
		t.Fatal(err)
	}
	if got := doctorExecutable(filepath.Join(dir, "missing"), file); got != file {
		t.Fatalf("got %q, want %q", got, file)
	}
}

func TestDoctorFindsCommandsOnPath(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"brew", "go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\necho go-version\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	report := &doctor.Report{}
	captureStdout(t, func() {
		RunDoctorTools(report)
		RunDoctorRuntimes(report)
	})
	for _, id := range []string{"tool.brew", "runtime.go"} {
		found := false
		for _, check := range report.Checks {
			if check.ID == id {
				found = true
				if check.Status != "pass" {
					t.Errorf("%s: got status %q", id, check.Status)
				}
			}
		}
		if !found {
			t.Errorf("missing check %s", id)
		}
	}
}

func TestDoctorReportsInstalledCommandsOutsidePath(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".linuxbrew", "bin")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"brew", "go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	t.Setenv("SHELL", "/bin/bash")
	report := &doctor.Report{}
	output := captureStdout(t, func() {
		RunDoctorTools(report)
		RunDoctorRuntimes(report)
	})
	for _, item := range []struct{ id, name string }{{"tool.brew", "brew"}, {"runtime.go", "go"}} {
		path := filepath.Join(dir, item.name)
		if !strings.Contains(output, path+" but not in PATH") {
			t.Errorf("missing PATH warning for %s: %s", item.name, output)
		}
		found := false
		for _, check := range report.Checks {
			if check.ID == item.id {
				found = true
				if check.Status != "fail" || !strings.Contains(check.Message, path) || !strings.Contains(check.Fix, dir) {
					t.Errorf("incorrect check: %+v", check)
				}
			}
		}
		if !found {
			t.Errorf("missing check %s", item.id)
		}
	}
}

func TestDoctorPathCommandsQuotePaths(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	path := "/tmp/user's tools/go"
	if got := doctorGoPathCommand(path); !strings.Contains(got, `'/tmp/user'\''s tools'`) {
		t.Fatalf("path not safely quoted: %q", got)
	}
	t.Setenv("SHELL", "/usr/bin/fish")
	if got := doctorBrewPathCommand(path); !strings.HasSuffix(got, " shellenv | source") {
		t.Fatalf("unexpected fish command: %q", got)
	}
}
