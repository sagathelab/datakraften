package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/doctor"
)

func TestDoctorSummaryMatchesVisibleAndJSONChecks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/bash")
	bin := t.TempDir()
	for name, script := range map[string]string{
		"uname": "echo Linux", "cat": "printf 'ID=fedora\\nVERSION_ID=43\\n'", "systemctl": "exit 0",
		"git": "exit 0", "gh": "exit 0", "uv": "exit 0", "node": "echo v22.0.0",
		"python3": "echo Python 3.14.0", "dotnet": "echo 10.0.0", "npm": "echo /not-a-real-prefix",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	if err := (&State{ManagedShell: true}).Save(); err != nil {
		t.Fatal(err)
	}
	oldJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = oldJSON })
	for _, category := range []string{"", "system", "tools", "runtimes", "editors", "docker", "shell"} {
		t.Run(category, func(t *testing.T) {
			args := []string{}
			if category != "" {
				args = []string{"--category", category}
			}
			jsonOutput = false
			output := captureStdout(t, func() {
				cmd := newDoctorCmd()
				cmd.SetArgs(args)
				if err := cmd.Execute(); err != nil {
					t.Fatal(err)
				}
			})
			passed := strings.Count(output, "[OK]")
			warnings := strings.Count(output, "[WARN]")
			skipped := strings.Count(output, "[SKIP]")
			want := fmt.Sprintf("%d passed  ·  0 critical  ·  %d warnings  ·  0 informational  ·  %d skipped", passed, warnings, skipped)
			if !strings.Contains(output, want) {
				t.Fatalf("summary does not match displayed checks; want %q in:\n%s", want, output)
			}
			jsonOutput = true
			data := captureStdout(t, func() {
				cmd := newDoctorCmd()
				cmd.SetArgs(args)
				if err := cmd.Execute(); err != nil {
					t.Fatal(err)
				}
			})
			var report doctor.Report
			if err := json.Unmarshal([]byte(data), &report); err != nil {
				t.Fatal(err)
			}
			if got := report.Summary(); got != want {
				t.Errorf("JSON summary = %q, want %q", got, want)
			}
			ids := map[string]bool{}
			for _, check := range report.Checks {
				if ids[check.ID] {
					t.Errorf("duplicate check ID %q", check.ID)
				}
				ids[check.ID] = true
			}
		})
	}
}
