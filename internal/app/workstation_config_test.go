package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/spf13/cobra"
)

func captureWorkstationStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	saved := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = saved; w.Close() }()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeWorkstationTestFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func TestWorkstationReportsDistinguishIntentFromDetection(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("SHELL", "/not-on-path/bash")
	t.Setenv("ZDOTDIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	writeWorkstationTestFile(t, filepath.Join(bin, "bash"), "#!/bin/sh\nexit 0\n", 0700)
	writeWorkstationTestFile(t, filepath.Join(bin, "node"), "#!/bin/sh\nprintf '%s\\n' v99.1.0\n", 0700)
	bashRC := filepath.Join(home, ".bashrc")
	writeWorkstationTestFile(t, bashRC, "# >>> datakraften >>>\n# <<< datakraften <<<\n", 0600)
	zshRC := filepath.Join(home, ".zshrc")
	writeWorkstationTestFile(t, zshRC, "# >>> datakraften >>>\n", 0600)
	if err := (&State{LastApply: "previous", ActiveSource: "default", ManagedShell: true}).Save(); err != nil {
		t.Fatal(err)
	}
	writeWorkstationTestFile(t, config.DefaultConfigPath(), `source: custom
shell: {default: zsh, prompt: hidden-value}
runtimes: {node: {enabled: false}, python: {enabled: true}}
editors: {zed: true}
ai_tools: {codex: true}
ai_apps: {claude: true}
`, 0600)
	for _, args := range [][]string{{"status"}, {"doctor"}, {"doctor", "--category", "shell"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			run := func(arguments []string) (string, string) {
				t.Helper()
				cmd := newApp().RootCmd
				t.Cleanup(func() { jsonOutput = false })
				cmd.SetArgs(arguments)
				var output string
				warning := captureWorkstationStderr(t, func() {
					output = captureStdout(t, func() {
						if err := cmd.Execute(); err != nil {
							t.Fatal(err)
						}
					})
				})
				return output, warning
			}
			text, stderr := run(args)
			for _, line := range []string{"Configured intent", "Source: custom", "Shell target: zsh", "node=disabled", "python=enabled", "zed=enabled",
				"AI tools: ", "codex=enabled", "AI apps: ", "claude=enabled", "Last apply source: default", "Current shell ($SHELL): /not-on-path/bash",
				"bash: on PATH: true; managed block: present", "fish: on PATH: false; managed block: absent", "zsh: on PATH: false; managed block: invalid"} {
				if !strings.Contains(text, line) {
					t.Errorf("missing %q in %s", line, text)
				}
			}
			if len(args) == 1 && !strings.Contains(text, "v99.1.0") {
				t.Errorf("detected Node should be reported even when disabled: %s", text)
			}
			raw, jsonStderr := run(append([]string{"--json"}, args...))
			var data struct {
				workstationReport
				Shell detectedShells `json:"shell"`
			}
			if err := json.Unmarshal([]byte(raw), &data); err != nil {
				t.Fatalf("invalid JSON: %v: %s", err, raw)
			}
			cfg := data.Configuration
			if !cfg.Present || cfg.Path != config.DefaultConfigPath() || cfg.Source != "custom" || cfg.Shell != config.ShellZsh ||
				cfg.Runtimes["node"] || !cfg.Runtimes["python"] || !cfg.Editors["zed"] || !cfg.AITools["codex"] || !cfg.AIApps["claude"] ||
				len(cfg.Warnings) != 1 || data.LastSource != "default" || data.LastApply != "previous" {
				t.Errorf("incorrect intent: %+v", data.workstationReport)
			}
			if data.Shell.Current != "/not-on-path/bash" || data.Shell.Targets[config.ShellBash] != (shellStatus{true, bashRC, "present"}) ||
				data.Shell.Targets[config.ShellFish].ManagedBlock != "absent" ||
				data.Shell.Targets[config.ShellZsh] != (shellStatus{false, zshRC, "invalid"}) {
				t.Errorf("incorrect detection: %+v", data.Shell)
			}
			if stderr != jsonStderr || !strings.Contains(stderr, "shell.prompt is ignored by dk apply") || strings.Contains(raw+stderr+text, "hidden-value") {
				t.Errorf("unexpected warning or leak: %q, %q, %q", stderr, jsonStderr, raw)
			}
		})
	}
}

func TestWorkstationReportsWithNoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	for _, name := range []string{"status", "doctor"} {
		for _, jsonMode := range []bool{false, true} {
			cmd := newApp().RootCmd
			args := []string{name}
			if jsonMode {
				args = append(args, "--json")
			}
			cmd.SetArgs(args)
			output := captureStdout(t, func() {
				if err := cmd.Execute(); err != nil {
					t.Fatal(err)
				}
			})
			if !jsonMode {
				if !strings.Contains(output, "No workstation configuration found") || !strings.Contains(output, "Detected shell") {
					t.Errorf("%s: no intent or detection: %s", name, output)
				}
				continue
			}
			var parsed map[string]json.RawMessage
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Fatal(err)
			}
			var intent map[string]any
			if err := json.Unmarshal(parsed["configuration"], &intent); err != nil || len(intent) != 2 || intent["present"] != false || intent["path"] != config.DefaultConfigPath() {
				t.Errorf("%s: unexpected missing config intent: %+v, %v", name, intent, err)
			}
			if _, ok := parsed["shell"]; !ok {
				t.Errorf("%s: detection missing: %s", name, output)
			}
		}
	}
	jsonOutput = false
}

func TestWorkstationCommandsRejectInvalidConfigBeforeEffects(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	marker := filepath.Join(bin, "ran")
	t.Setenv("DK_TEST_MARKER", marker)
	for _, name := range []string{"brew", "fnm", "uv", "npm", "sudo"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nprintf '%s\\n' ran > \"$DK_TEST_MARKER\"\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	old := http.DefaultTransport
	http.DefaultTransport = initTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("workstation command attempted to fetch a configuration")
		return nil, nil
	})
	t.Cleanup(func() { http.DefaultTransport = old })

	path := config.DefaultConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ input, want string }{
		{"source: team\nurl: https://example.test/config?token=hidden\n", "legacy team configuration"},
		{"profile: team\nteam: {url: 'https://example.test/config?token=hidden'}\n", "legacy team configuration"},
		{"editors: {zed: optional}\n", "editors.zed must be true or false"},
		{"shell: {default: tcsh}\n", "shell.default must be fish, bash, zsh or none"},
	} {
		if err := os.WriteFile(path, []byte(tc.input), 0600); err != nil {
			t.Fatal(err)
		}
		for _, command := range []struct {
			newCmd func() *cobra.Command
			args   []string
		}{
			{newApplyCmd, []string{"--yes"}},
			{newApplyCmd, []string{"--dry-run"}},
			{newStatusCmd, nil},
			{newDoctorCmd, []string{"--category", "shell"}},
			{newUpdateCmd, nil},
			{newUpdateCmd, []string{"npm"}},
			{newUpdateCmd, []string{"brew"}},
		} {
			cmd := command.newCmd()
			cmd.SilenceErrors, cmd.SilenceUsage = true, true
			cmd.SetArgs(append([]string{}, command.args...))
			var err error
			captureStdout(t, func() { err = cmd.Execute() })
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "token=hidden") || strings.Contains(err.Error(), "no workstation configuration found") {
				t.Fatalf("%s %v: wanted %q, got %v", cmd.Name(), command.args, tc.want, err)
			}
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != tc.input {
			t.Fatalf("command changed workstation config: %s, %v", data, err)
		}
	}
	for _, path := range []string{marker, statePath()} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("unexpected side effect at %s: %v", path, err)
		}
	}
}

func TestInitRejectsRetiredFlags(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, flag := range []string{"--team", "--migrate-team"} {
		cmd := newInitCmd()
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		cmd.SetArgs([]string{flag})
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "unknown flag: "+flag) {
			t.Fatalf("%s: %v", flag, err)
		}
	}
	if _, err := os.Stat(config.DefaultConfigPath()); !os.IsNotExist(err) {
		t.Fatalf("unexpected config: %v", err)
	}
}

func TestInitCustomFileValidationAndMerge(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	run := func(args ...string) error {
		t.Helper()
		cmd := newInitCmd()
		cmd.SetArgs(append([]string{}, args...))
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		var err error
		captureStdout(t, func() { err = cmd.Execute() })
		return err
	}
	if err := run(); err != nil {
		t.Fatal(err)
	}
	path := config.DefaultConfigPath()
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "local.yaml")
	if err := run("--custom", input, "--yes"); err == nil {
		t.Fatal("missing local input accepted")
	}
	if err := os.WriteFile(input, []byte("runtimes: {node: {enabled: 'false'}}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run("--custom", input, "--yes"); err == nil || !strings.Contains(err.Error(), "runtimes.node.enabled") {
		t.Fatalf("invalid local input: %v", err)
	}
	if after, err := os.ReadFile(path); err != nil || string(after) != string(original) {
		t.Fatalf("failed merge changed original: %s, %v", after, err)
	}
	if err := os.WriteFile(input, []byte("source: default\nversion: 1\nruntimes: {node: {enabled: false}}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run("--custom", input, "--yes"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil || cfg.Source != "custom" || cfg.Runtimes.Node.Enabled || !cfg.Runtimes.Python.Enabled || cfg.ShellTarget() != config.ShellFish {
		t.Fatalf("unexpected merged config: %+v, %v", cfg, err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("expected a private config: %v, %v", info, err)
	}
}
