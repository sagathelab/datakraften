package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/ai"
	"github.com/sagathelab/datakraften/internal/config"
)

func TestInstallReportSummaryShowsFailureInsteadOfAlreadyInstalled(t *testing.T) {
	report := ai.InstallReport{
		Enabled: 1,
		Errors:  []string{"GitHub Copilot CLI: brew install --cask copilot-cli failed: boom"},
	}

	got := installReportSummary(report)

	if got == "(already installed)" {
		t.Fatalf("expected failure summary, got %q", got)
	}
	if !strings.Contains(got, "failed: GitHub Copilot CLI: brew install --cask copilot-cli failed: boom") {
		t.Fatalf("expected detailed failure summary, got %q", got)
	}
}

func TestApplyOnlyManagesSelectedShell(t *testing.T) {
	for _, target := range []config.ShellTarget{config.ShellFish, config.ShellBash, config.ShellZsh, config.ShellNone, ""} {
		t.Run(string(target), func(t *testing.T) {
			home, bin := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("PATH", bin)
			t.Setenv("SHELL", "/bin/ksh")
			t.Setenv("XDG_CONFIG_HOME", "")
			t.Setenv("ZDOTDIR", "")
			t.Setenv("DK_TEST_BREW_PREFIX", home)
			marker := filepath.Join(bin, "ran")
			t.Setenv("DK_TEST_MARKER", marker)
			writeWorkstationTestFile(t, filepath.Join(bin, "brew"), `#!/bin/sh
printf '%s\n' "$*" >> "$DK_TEST_MARKER"
case "$1" in
    --prefix) printf '%s\n' "$DK_TEST_BREW_PREFIX";;
    list) exit 0;;
    *) exit 99;;
esac
`, 0700)
			for _, name := range []string{"fish", "chsh"} {
				writeWorkstationTestFile(t, filepath.Join(bin, name), "#!/bin/sh\nprintf '%s\\n' unexpected >> \"$DK_TEST_MARKER\"\n", 0700)
			}
			const untouched = "user prelude\n# >>> datakraften >>>\nold setting\n# <<< datakraften <<<\nuser epilogue\n"
			for _, rc := range managedShells {
				writeWorkstationTestFile(t, rc.path(), untouched, 0600)
			}
			profile := filepath.Join(home, ".profile")
			writeWorkstationTestFile(t, profile, "user profile\n", 0600)
			settings := "source: custom\n"
			if target != "" {
				settings += "shell: {default: " + string(target) + ", prompt: ignored-value}\n"
			}
			writeWorkstationTestFile(t, config.DefaultConfigPath(), settings, 0600)
			run := func(arg string) (string, string) {
				t.Helper()
				var output string
				warning := captureWorkstationStderr(t, func() {
					output = captureStdout(t, func() {
						cmd := newApplyCmd()
						cmd.SetArgs([]string{arg})
						if err := cmd.Execute(); err != nil {
							t.Fatal(err)
						}
					})
				})
				return output, warning
			}
			assertUnchanged := func(path, expected string) {
				t.Helper()
				data, err := os.ReadFile(path)
				if err != nil || string(data) != expected {
					t.Fatalf("%s changed: %q, %v", path, data, err)
				}
			}
			preview, warning := run("--dry-run")
			selected, manage := managedShells[target]
			if manage {
				if !strings.Contains(preview, "Would manage "+string(target)+" configuration in "+selected.path()) || !strings.Contains(warning, "shell.prompt is ignored by dk apply") {
					t.Errorf("missing shell plan or legacy warning: %s %s", preview, warning)
				}
			} else if !strings.Contains(preview, "Shell configuration disabled") {
				t.Errorf("preview did not disable shell work: %s", preview)
			}
			for _, rc := range managedShells {
				assertUnchanged(rc.path(), untouched)
			}
			for _, path := range []string{marker, statePath()} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("preview caused a side effect at %s: %v", path, err)
				}
			}
			output, _ := run("--yes")
			for name, rc := range managedShells {
				if name != target {
					assertUnchanged(rc.path(), untouched)
					continue
				}
				got, err := os.ReadFile(rc.path())
				if err != nil || !strings.HasPrefix(string(got), "user prelude\n# >>> datakraften >>>\n") ||
					!strings.HasSuffix(string(got), "# <<< datakraften <<<\nuser epilogue\n") ||
					!strings.Contains(string(got), "starship init "+string(target)) ||
					strings.Contains(string(got), "fnm env") || strings.Contains(string(got), "uv generate") || strings.Contains(string(got), "old setting") {
					t.Fatalf("unexpected managed block: %s, %v", got, err)
				}
				if !strings.Contains(output, "Start a new "+string(target)+" session") {
					t.Errorf("incorrect next step: %s", output)
				}
				run("--yes")
				assertUnchanged(rc.path(), string(got))
			}
			assertUnchanged(profile, "user profile\n")
			assertUnchanged(config.DefaultConfigPath(), settings)
			if data, err := os.ReadFile(marker); err != nil || strings.Contains(string(data), "unexpected") {
				t.Fatalf("unexpected command execution: %q, %v", data, err)
			}
		})
	}
}

func TestApplyBrewPackagesAndShellHooksFollowConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Setenv("ZDOTDIR", "")
	cfg := &config.Config{}
	packages := applyBrewPackages(cfg)
	for _, pkg := range []string{"fnm", "uv", "starship", "atuin", "fzf"} {
		if slices.Contains(packages, pkg) {
			t.Fatalf("disabled components added %s: %v", pkg, packages)
		}
	}
	cfg.Runtimes.Node.Enabled, cfg.Runtimes.Python.Enabled = true, true
	cfg.Shell.Default = config.ShellZsh
	packages = applyBrewPackages(cfg)
	for _, pkg := range []string{"fnm", "uv", "starship", "atuin", "fzf", "gh"} {
		if !slices.Contains(packages, pkg) {
			t.Fatalf("missing %s: %v", pkg, packages)
		}
	}
	if configured, err := applyShell(cfg, false); err != nil || !configured {
		t.Fatalf("could not configure zsh without an installer: %t, %v", configured, err)
	}
	data, err := os.ReadFile(managedShells[config.ShellZsh].path())
	for _, hook := range []string{"fnm env --use-on-cd --shell zsh", "uv generate-shell-completion zsh"} {
		if err != nil || !strings.Contains(string(data), hook) {
			t.Errorf("missing %s: %s, %v", hook, data, err)
		}
	}
}

func TestApplyFishInstallationIsOnlyPlannedDuringDryRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	cfg := &config.Config{}
	cfg.Shell.Default = config.ShellFish
	output := captureStdout(t, func() {
		configured, err := applyShell(cfg, true)
		if err != nil || configured {
			t.Fatalf("unexpected preview result: %t, %v", configured, err)
		}
	})
	for _, expected := range []string{"Would install fish", "Would manage fish", "login shell unchanged"} {
		if !strings.Contains(output, expected) {
			t.Errorf("missing %q: %s", expected, output)
		}
	}
	if _, err := os.Stat(managedShells[config.ShellFish].path()); !os.IsNotExist(err) {
		t.Fatalf("preview changed the rc file: %v", err)
	}
}

func TestApplyCanInstallMissingFishWithoutChangingLoginShell(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	t.Setenv("SHELL", "/bin/ksh")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("DK_TEST_BIN", bin)
	writeWorkstationTestFile(t, filepath.Join(bin, "brew"), `#!/bin/sh
case "$*" in
    '--prefix') printf '%s\n' "$HOME";;
    'install fish') printf '#!/bin/sh\nexit 0\n' > "$DK_TEST_BIN/fish"; /bin/chmod +x "$DK_TEST_BIN/fish";;
    *) exit 99;;
esac
`, 0700)
	cfg := &config.Config{}
	cfg.Shell.Default = config.ShellFish
	if configured, err := applyShell(cfg, false); err != nil || !configured {
		t.Fatalf("could not configure Fish: %t, %v", configured, err)
	}
	if os.Getenv("SHELL") != "/bin/ksh" {
		t.Fatal("login shell changed")
	}
	data, err := os.ReadFile(managedShells[config.ShellFish].path())
	if err != nil || !strings.Contains(string(data), "starship init fish") {
		t.Fatalf("missing Fish configuration: %q, %v", data, err)
	}
}

func TestApplyBrewOutsidePathDoesNotWriteProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", t.TempDir())
	brewBin := filepath.Join(home, ".linuxbrew", "bin")
	writeWorkstationTestFile(t, filepath.Join(brewBin, "brew"), "#!/bin/sh\nexit 0\n", 0700)
	if installed, err := ensureApplyBrew(); err != nil || installed {
		t.Fatalf("existing Homebrew not reused: %t, %v", installed, err)
	}
	if !strings.HasPrefix(os.Getenv("PATH"), brewBin+":") {
		t.Fatal("existing Homebrew not added to apply process PATH")
	}
	if _, err := os.Stat(filepath.Join(home, ".profile")); !os.IsNotExist(err) {
		t.Fatalf("unexpected profile modification: %v", err)
	}
}

func TestDefaultConfigEnablesCopilotWithoutLegacyFields(t *testing.T) {
	if strings.Count(defaultConfigYAML, "copilot:\n    enabled: true\n") != 2 || strings.Contains(defaultConfigYAML, "    manager:") || strings.Contains(defaultConfigYAML, "    version:") {
		t.Fatal("expected both Copilot options enabled without unused manager or version fields")
	}
}

func TestInstallReportSummaryDoesNotAssumeUnreportedResults(t *testing.T) {
	got := installReportSummary(ai.InstallReport{Enabled: 1})
	if strings.Contains(got, "already installed") {
		t.Fatalf("unreported result presented as installed: %q", got)
	}
}

func TestInstallSummaryOmitsFailureDetails(t *testing.T) {
	report := ai.InstallReport{Enabled: 1, Errors: []string{"Claude Code: npm failed: EACCES"}}
	got := installSummaryWithoutDetails(report)
	if got != "failed: Claude Code" || report.Errors[0] != "Claude Code: npm failed: EACCES" {
		t.Fatalf("summary %q report %+v", got, report)
	}
}

func TestApplyProgressConfigurationFailure(t *testing.T) {
	var events []ProgressEvent
	report := RunApplyWithProgress(&config.Config{Source: "team"}, false, func(event ProgressEvent) {
		events = append(events, event)
	})
	if len(report.Errors) != 1 || len(events) != 2 || events[0].Status != ProgressStarted || events[1].Status != ProgressFailed || events[1].Err == nil {
		t.Fatalf("unexpected report or events: %+v, %+v", report, events)
	}
}
