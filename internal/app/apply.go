package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sagathelab/datakraften/internal/ai"
	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/doctor"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/shell"
	"github.com/sagathelab/datakraften/internal/system"
)

type ProgressStatus string

const (
	ProgressStarted   ProgressStatus = "started"
	ProgressFailed    ProgressStatus = "failed"
	ProgressCompleted ProgressStatus = "completed"
)

type ProgressEvent struct {
	Step    string
	Message string
	Status  ProgressStatus
	Err     error
}

type ProgressHook func(ProgressEvent)

func emitProgress(hook ProgressHook, event ProgressEvent) {
	if hook != nil {
		hook(event)
	}
}

type RuntimeStatus int

const (
	RuntimeMissing RuntimeStatus = iota
	RuntimeAlready
	RuntimeNew
)

func (s RuntimeStatus) String() string {
	switch s {
	case RuntimeNew:
		return "installed"
	case RuntimeAlready:
		return "already installed"
	default:
		return "not installed"
	}
}

type ApplyReport struct {
	Started       bool
	System        []string
	SystemChecked bool
	BrewChecked   bool
	BrewPkgs      []string
	Node          RuntimeStatus
	NodeVer       string
	Python        RuntimeStatus
	PythonVer     string
	Go            RuntimeStatus
	GoVer         string
	Dotnet        RuntimeStatus
	DotnetVer     string
	AITools       ai.InstallReport
	AIApps        ai.InstallReport
	Shell         config.ShellTarget
	Errors        []string
}

func installReportSummary(report ai.InstallReport) string {
	if report.Enabled == 0 {
		return "(disabled)"
	}

	parts := make([]string, 0, 4)
	if len(report.Installed) > 0 {
		parts = append(parts, fmt.Sprintf("installed: %s", strings.Join(report.Installed, ", ")))
	}
	if len(report.AlreadyInstalled) > 0 {
		parts = append(parts, fmt.Sprintf("already installed: %s", strings.Join(report.AlreadyInstalled, ", ")))
	}
	if len(report.Skipped) > 0 {
		parts = append(parts, fmt.Sprintf("skipped: %s", strings.Join(report.Skipped, ", ")))
	}
	if len(report.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("failed: %s", strings.Join(report.Errors, "; ")))
	}
	if len(parts) == 0 {
		return "(no results reported)"
	}
	return strings.Join(parts, "; ")
}

func installSummaryWithoutDetails(report ai.InstallReport) string {
	for i, message := range report.Errors {
		name, _, _ := strings.Cut(message, ":")
		if i == 0 {
			report.Errors = append([]string(nil), report.Errors...)
		}
		report.Errors[i] = name
	}
	return installReportSummary(report)
}

func RunApply(cfg *config.Config, dryRun bool) *ApplyReport {
	return RunApplyWithProgress(cfg, dryRun, nil)
}

func RunApplyWithProgress(cfg *config.Config, dryRun bool, hook ProgressHook) *ApplyReport {
	report := &ApplyReport{}
	step := "configuration"
	failed := false
	started := false
	start := func(name, message string) {
		if started && !failed {
			emitProgress(hook, ProgressEvent{Step: step, Status: ProgressCompleted})
		}
		step, failed, started = name, false, true
		emitProgress(hook, ProgressEvent{Step: step, Message: message, Status: ProgressStarted})
	}
	fail := func(message string) {
		report.Errors = append(report.Errors, message)
		if !failed {
			failed = true
			emitProgress(hook, ProgressEvent{Step: step, Status: ProgressFailed, Err: fmt.Errorf("%s", message)})
		}
	}
	defer func() {
		if started && !failed {
			emitProgress(hook, ProgressEvent{Step: step, Status: ProgressCompleted})
		}
	}()
	if cfg.Source != "" {
		if err := config.ValidateSource(cfg.Source); err != nil {
			start("configuration", "Validating workstation configuration")
			fail(err.Error())
			displayPrintf("  [ERROR] %s\n", err)
			return report
		}
	}
	target := cfg.ShellTarget()
	if err := config.ValidateShellTarget(target); err != nil {
		start("configuration", "Validating workstation configuration")
		fail(err.Error())
		return report
	}
	pm := installers.DetectPackageManager()
	verbosePrintf("packageManager=%s source=%s\n", pm, cfg.Source)

	if hook == nil {
		fmt.Fprintln(progressWriter())
	}
	report.Started = true
	start("system", "Checking system packages")
	if hook == nil {
		displayStep(progressWriter(), 1, 8, "System packages")
	}

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking system dependencies...")
		n, err := installers.NativeEnsurePackages(installers.NativeBasePackages())
		if err != nil {
			fail(fmt.Sprintf("%s: %s", pm, err))
			displayPrintf("    [ERROR] System dependencies: %s\n", err)
		} else {
			report.SystemChecked = true
			report.System = n
			if len(n) == 0 {
				displayPrintf("    [OK] System dependencies via %s (already satisfied)\n", pm)
			} else {
				displayPrintf("    [OK] %d system packages installed via %s: %s\n", len(n), pm, strings.Join(n, ", "))
			}
		}
	} else {
		displayPrintf("    [PLAN] Would install system dependencies via %s\n", pm)
	}

	fmt.Fprintln(progressWriter())
	start("developer-tools", "Checking developer tools")
	if hook == nil {
		displayStep(progressWriter(), 2, 8, "Developer tools")
	}

	if pm != installers.BREW {
		if !dryRun {
			fmt.Fprintln(progressWriter(), "    Checking Homebrew...")
			installed, err := ensureApplyBrew()
			if err != nil {
				fail(fmt.Sprintf("Brew: %s", err))
				displayPrintf("    [ERROR] Homebrew: %s\n", err)
			} else if installed {
				displayPrintln("    [OK] Homebrew installed")
			} else {
				displayPrintln("    [OK] Homebrew (already installed)")
			}
		} else {
			displayPrintln("    [PLAN] Would ensure Homebrew is installed")
		}
	} else {
		displayPrintln("    [OK] Homebrew (native package manager)")
	}

	brewPackages := applyBrewPackages(cfg)
	verbosePrintf("brewPkgs=%v\n", brewPackages)

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking Homebrew packages...")
		n, err := installers.BrewEnsurePackages(brewPackages)
		if err != nil {
			fail(fmt.Sprintf("Brew packages: %s", err))
			displayPrintf("    [ERROR] Brew packages: %s\n", err)
		} else {
			report.BrewChecked = true
			report.BrewPkgs = n
			if len(n) == 0 {
				displayPrintln("    [OK] Brew packages (already installed)")
			} else {
				displayPrintf("    [OK] %d Brew packages installed: %s\n", len(n), strings.Join(n, ", "))
			}
		}
	} else {
		displayPrintln("    [PLAN] Would ensure Brew packages: " + strings.Join(brewPackages, ", "))
	}

	fmt.Fprintln(progressWriter())
	start("runtimes", "Checking runtimes")
	if hook == nil {
		displayStep(progressWriter(), 3, 8, "Runtimes")
	}

	if !cfg.RuntimeEnabled("node") {
		displayPrintln("    [SKIP] Node.js (disabled in config)")
	} else if dryRun {
		displayPrintln("    [PLAN] Would install Node.js LTS via fnm")
	} else {
		fmt.Fprintln(progressWriter(), "    Checking Node.js...")
		installed, err := runtimes.EnsureNode()
		if err != nil {
			fail(fmt.Sprintf("Node: %s", err))
			displayPrintf("    [ERROR] Node.js: %s\n", err)
		} else {
			report.NodeVer = runtimes.NodeVersion()
			if installed {
				report.Node = RuntimeNew
				displayPrintf("    [OK] Node.js %s installed\n", report.NodeVer)
			} else if report.NodeVer != "" {
				report.Node = RuntimeAlready
				displayPrintf("    [OK] Node.js %s (already installed)\n", report.NodeVer)
			}
		}
	}

	if !cfg.RuntimeEnabled("python") {
		displayPrintln("    [SKIP] Python (disabled in config)")
	} else if dryRun {
		displayPrintln("    [PLAN] Would install Python via uv")
	} else {
		fmt.Fprintln(progressWriter(), "    Checking Python...")
		installed, err := runtimes.EnsurePython()
		if err != nil {
			fail(fmt.Sprintf("Python: %s", err))
			displayPrintf("    [ERROR] Python: %s\n", err)
		} else {
			report.PythonVer = runtimes.PythonVersion()
			if installed {
				report.Python = RuntimeNew
				displayPrintf("    [OK] Python %s installed\n", report.PythonVer)
			} else if report.PythonVer != "" {
				report.Python = RuntimeAlready
				displayPrintf("    [OK] Python %s (already installed)\n", report.PythonVer)
			}
		}
	}

	if !cfg.RuntimeEnabled("go") {
		displayPrintln("    [SKIP] Go (disabled in config)")
	} else if dryRun {
		displayPrintln("    [PLAN] Would install Go via Homebrew")
	} else {
		fmt.Fprintln(progressWriter(), "    Checking Go...")
		installed, err := runtimes.EnsureGo()
		if err != nil {
			fail(fmt.Sprintf("Go: %s", err))
			displayPrintf("    [ERROR] Go: %s\n", err)
		} else {
			report.GoVer = runtimes.GoVersion()
			if installed {
				report.Go = RuntimeNew
				displayPrintf("    [OK] Go %s installed\n", report.GoVer)
			} else if report.GoVer != "" {
				report.Go = RuntimeAlready
				displayPrintf("    [OK] Go %s (already installed)\n", report.GoVer)
			}
		}
	}

	if !cfg.RuntimeEnabled("dotnet") {
		displayPrintln("    [SKIP] .NET SDK (disabled in config)")
	} else if dryRun {
		displayPrintln("    [PLAN] Would install .NET SDK via Homebrew")
	} else {
		fmt.Fprintln(progressWriter(), "    Checking .NET SDK...")
		installed, err := runtimes.EnsureDotnet()
		if err != nil {
			fail(fmt.Sprintf(".NET: %s", err))
			displayPrintf("    [ERROR] .NET SDK: %s\n", err)
		} else {
			report.DotnetVer = runtimes.DotnetVersion()
			if installed {
				report.Dotnet = RuntimeNew
				displayPrintf("    [OK] .NET SDK %s installed\n", report.DotnetVer)
			} else if report.DotnetVer != "" {
				report.Dotnet = RuntimeAlready
				displayPrintf("    [OK] .NET SDK %s (already installed)\n", report.DotnetVer)
			}
		}
	}

	fmt.Fprintln(progressWriter())
	start("shell", "Configuring shell")
	if hook == nil {
		displayStep(progressWriter(), 4, 8, "Shell")
	}

	configured, err := applyShell(cfg, dryRun)
	if err != nil {
		fail(fmt.Sprintf("Shell config: %s", err))
		displayPrintf("    [ERROR] Shell config: %s\n", err)
	} else if configured {
		report.Shell = target
	}

	fmt.Fprintln(progressWriter())
	start("editors", "Checking editors")
	if hook == nil {
		displayStep(progressWriter(), 5, 8, "Editors")
	}

	if cfg.Editors != nil && cfg.Editors["zed"] == "true" {
		if !dryRun {
			fmt.Fprintln(progressWriter(), "    Checking Zed...")
			if err := editors.EnsureZed(); err != nil {
				fail(fmt.Sprintf("Zed: %s", err))
				displayPrintf("    [ERROR] Zed: %s\n", err)
			}
		} else {
			displayPrintln("    [PLAN] Would install Zed")
		}
	}

	if !dryRun {
		for _, ed := range editors.DetectAll() {
			if ed.Installed {
				status := "ok"
				if ed.WindowsSide {
					status = "Windows-side, consider Linux CLI install"
				}
				displayPrintf("    [OK] %s (%s)\n", ed.Name, status)
			} else {
				displayPrintf("    [SKIP] %s not found\n", ed.Name)
			}
		}
	}

	fmt.Fprintln(progressWriter())
	start("ai-tools", "Checking AI tools")
	if hook == nil {
		displayStep(progressWriter(), 6, 8, "AI tools")
	}

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking AI tools...")
		report.AITools = ai.EnsureAIToolsWithOptions(cfg, ai.InstallOptions{
			OnError: func(message string) { fail("AI tools: " + message) },
			OnNPMError: func(stderr string) {
				verbosePrintf("npm install output: %s\n", strings.TrimSpace(stderr))
			},
			RetryNPM: func(pkg, stderr string) error {
				return retryNpmOperationWithSudo("install", []string{pkg}, stderr)
			},
		})
		if report.AITools.HasFailures() {
			displayPrintf("    [ERROR] AI tools: %s\n", installReportSummary(report.AITools))
		} else if report.AITools.Enabled == 0 {
			displayPrintln("    [SKIP] AI tools: (disabled)")
		} else {
			displayPrintf("    [OK] AI tools: %s\n", installReportSummary(report.AITools))
		}
	} else {
		previewAI("AI tools", config.AIToolNames(), cfg.AITools)
	}

	fmt.Fprintln(progressWriter())
	start("ai-apps", "Checking AI apps")
	if hook == nil {
		displayStep(progressWriter(), 7, 8, "AI apps")
	}

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking AI apps...")
		report.AIApps = ai.EnsureAIAppsWithOptions(cfg, ai.InstallOptions{OnError: func(message string) { fail("AI apps: " + message) }})
		if report.AIApps.HasFailures() {
			displayPrintf("    [ERROR] AI apps: %s\n", installReportSummary(report.AIApps))
		} else if report.AIApps.Enabled == 0 {
			displayPrintln("    [SKIP] AI apps: (disabled)")
		} else {
			displayPrintf("    [OK] AI apps: %s\n", installReportSummary(report.AIApps))
		}
	} else {
		previewAI("AI apps", config.AIAppNames(), cfg.AIApps)
	}

	fmt.Fprintln(progressWriter())
	start("docker", "Checking Docker")
	if hook == nil {
		displayStep(progressWriter(), 8, 8, "Docker")
	}

	if dryRun {
		displayPrintln("    [PLAN] Would check Docker availability (no installation)")
	} else {
		dockerStatus := docker.Detect()
		if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
			displayPrintln("    [OK] Docker running")
		} else {
			displayPrintf("    [SKIP] %s\n", dockerStatus.Message)
		}
	}

	fmt.Fprintln(progressWriter())
	return report
}

func previewAI(label string, names []string, entries map[string]config.RuntimeConfig) {
	var enabled []string
	for _, name := range names {
		if entries[name].Enabled {
			enabled = append(enabled, name)
		}
	}
	if len(enabled) == 0 {
		displayPrintf("    [SKIP] %s: (disabled)\n", label)
		return
	}
	displayPrintf("    [PLAN] Would check enabled %s: %s\n", label, strings.Join(enabled, ", "))
}

func ensureApplyBrew() (bool, error) {
	if installers.BrewInstalled() {
		return false, nil
	}
	path := doctorBrewOutsidePath()
	installed := path == ""
	if installed {
		r := exec.Run("env", "NONINTERACTIVE=1", "/bin/bash", "-c", installers.BrewInstallScript())
		if r.Code != 0 {
			return false, fmt.Errorf("homebrew install failed: %s", r.Stderr)
		}
		path = doctorBrewOutsidePath()
		if path == "" {
			return false, fmt.Errorf("Homebrew was installed but brew could not be found")
		}
	}
	// The package installer needs brew on PATH, but only applyShell may edit startup files.
	return installed, os.Setenv("PATH", filepath.Dir(path)+":"+os.Getenv("PATH"))
}

func applyBrewPackages(cfg *config.Config) []string {
	var packages []string
	for _, pkg := range installers.DefaultBrewPackages {
		switch pkg {
		case "fnm":
			if !cfg.RuntimeEnabled("node") {
				continue
			}
		case "uv":
			if !cfg.RuntimeEnabled("python") {
				continue
			}
		case "atuin", "fzf", "starship":
			if cfg.ShellTarget() == config.ShellNone {
				continue
			}
		}
		packages = append(packages, pkg)
	}
	return packages
}

type shellRC struct {
	path     func() string
	generate func(string, bool, bool, bool, bool, bool) string
	write    func(string) error
}

var managedShells = map[config.ShellTarget]shellRC{
	config.ShellFish: {shell.FishConfigPath, shell.FishGenerateConfig, shell.FishWriteConfig},
	config.ShellBash: {shell.BashConfigPath, shell.BashGenerateConfig, shell.BashWriteConfig},
	config.ShellZsh:  {shell.ZshConfigPath, shell.ZshGenerateConfig, shell.ZshWriteConfig},
}

func applyShell(cfg *config.Config, dryRun bool) (bool, error) {
	target := cfg.ShellTarget()
	if target == config.ShellNone {
		displayPrintln("    [SKIP] Shell configuration disabled")
		return false, nil
	}
	rc, ok := managedShells[target]
	if !ok {
		return false, config.ValidateShellTarget(target)
	}
	path := rc.path()
	switch status := shell.ManagedBlockStatus(path); status {
	case "invalid":
		return false, fmt.Errorf("invalid managed block in %s", path)
	case "unreadable":
		return false, fmt.Errorf("cannot read shell config %s", path)
	}
	installFish := target == config.ShellFish && !shell.FishInstalled()
	if dryRun {
		if installFish {
			displayPrintln("    [PLAN] Would install fish via native package manager")
		}
		displayPrintf("    [PLAN] Would manage %s configuration in %s (login shell unchanged)\n", target, path)
		return false, nil
	}
	if installFish {
		if err := installers.NativeInstall(installers.FishPackageName()); err != nil {
			return false, err
		}
		if !shell.FishInstalled() {
			return false, fmt.Errorf("fish is not available on PATH after installation")
		}
	}
	prefix := ""
	if installers.BrewInstalled() {
		prefix = installers.BrewPrefix()
	}
	if err := rc.write(rc.generate(prefix, true, true, true, cfg.RuntimeEnabled("node"), cfg.RuntimeEnabled("python"))); err != nil {
		return false, err
	}
	displayPrintf("    [OK] %s managed block configured in %s\n", target, path)
	return true, nil
}

func RunDoctorSystem(r *doctor.Report) {
	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()
	systemd := system.HasSystemd()
	shell_ := system.CurrentShell()

	r.Add(doctor.Check{ID: "wsl", Title: "WSL detected", Category: "system", Severity: "info",
		Status: skipStatus(wsl), Message: fmt.Sprintf("WSL%d: %s %s", wslVer, distro, distroVer),
		Fix: "If WSL is not detected, ensure you are running inside WSL: 'wsl --install' from PowerShell"})
	if distro != "" {
		r.Add(doctor.Check{ID: "system.distro", Title: "Distribution", Category: "system", Severity: "info",
			Status: "pass", Message: strings.TrimSpace(distro + " " + distroVer)})
	}
	r.Add(doctor.Check{ID: "systemd", Title: "systemd available", Category: "system", Severity: "info",
		Status: skipStatus(systemd),
		Fix:    "Enable systemd in /etc/wsl.conf: add [boot]\nsystemd=true, then 'wsl --terminate' from PowerShell"})
	if shell_ != "" {
		r.Add(doctor.Check{ID: "system.shell", Title: "Current shell", Category: "system", Severity: "info",
			Status: "pass", Message: shell_})
	}

	displayHeading(doctorOutput(), "System")
	if wsl {
		displayFprintf(doctorOutput(), "    [OK] WSL%d detected\n", wslVer)
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Not running in WSL")
	}
	if distro != "" {
		displayFprintf(doctorOutput(), "    [OK] %s %s\n", distro, distroVer)
	}
	if systemd {
		displayFprintln(doctorOutput(), "    [OK] systemd available")
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] systemd not available")
	}
	if shell_ != "" {
		displayFprintf(doctorOutput(), "    [OK] Shell: %s\n", shell_)
	}
}

func RunDoctorTools(r *doctor.Report, cfg *config.Config) {
	brewPath := ""
	if !exec.CommandExists("brew") {
		brewPath = doctorBrewOutsidePath()
	}
	tools := []struct {
		name      string
		id        string
		installed bool
	}{
		{"git", "tool.git", exec.CommandExists("git")},
		{"gh", "tool.gh", exec.CommandExists("gh")},
		{"az", "tool.az", exec.CommandExists("az")},
		{"fnm", "tool.fnm", exec.CommandExists("fnm")},
		{"uv", "tool.uv", exec.CommandExists("uv")},
		{"brew", "tool.brew", exec.CommandExists("brew")},
		{"copilot", "tool.copilot", exec.CommandExists("copilot")},
		{"docker", "tool.docker", exec.CommandExists("docker")},
	}

	for _, t := range tools {
		fix := ""
		message := ""
		status := skipStatus(t.installed)
		if !t.installed {
			if cfg != nil && t.name != "docker" &&
				(t.name != "fnm" || cfg.RuntimeEnabled("node")) &&
				(t.name != "uv" || cfg.RuntimeEnabled("python")) &&
				(t.name != "copilot" || cfg.AITools["copilot"].Enabled) {
				fix = "run 'dk apply' to install missing tools"
			}
			if t.name == "brew" {
				if brewPath != "" {
					status = "fail"
					message = "Homebrew found at " + brewPath + " but not in PATH"
					fix = "Add Homebrew to your shell PATH: " + doctorBrewPathCommand(brewPath)
				}
			}
		}
		r.Add(doctor.Check{ID: t.id, Title: t.name + " installed", Category: "tools",
			Severity: "warning", Status: status, Message: message, Fix: fix})
	}

	displayHeading(doctorOutput(), "Tools")
	if exec.CommandExists("npm") {
		result := exec.Run("npm", "config", "get", "prefix")
		prefix := strings.TrimSpace(result.Stdout)
		status, message, fix := "fail", "Could not determine npm prefix", "Check npm configuration"
		if result.Code == 0 && filepath.IsAbs(prefix) {
			path := filepath.Join(prefix, "lib", "node_modules")
			for {
				_, err := os.Stat(path)
				if err == nil || !os.IsNotExist(err) || filepath.Dir(path) == path {
					break
				}
				path = filepath.Dir(path)
			}
			if syscall.Access(path, 2) == nil {
				status, message, fix = "pass", "npm prefix: "+prefix, ""
			} else {
				status, message, fix = "fail", "npm prefix may not be writable: "+prefix, "run 'dk npm setup' or use a Node version manager"
			}
		}
		r.Add(doctor.Check{ID: "tool.npm.prefix", Title: "npm global prefix", Category: "tools", Severity: "warning", Status: status, Message: message, Fix: fix})
		if status == "pass" {
			displayFprintf(doctorOutput(), "    [OK] %s\n", message)
		} else {
			displayFprintf(doctorOutput(), "    [WARN] %s\n", message)
		}
	}

	for _, t := range tools {
		if t.installed {
			displayFprintf(doctorOutput(), "    [OK] %s\n", t.name)
		} else if t.name == "brew" && brewPath != "" {
			displayFprintf(doctorOutput(), "    [WARN] brew found at %s but not in PATH\n", brewPath)
			displayFprintf(doctorOutput(), "           Run: %s\n", doctorBrewPathCommand(brewPath))
		} else {
			displayFprintf(doctorOutput(), "    [SKIP] %s not installed\n", t.name)
		}
	}
}

func RunDoctorRuntimes(r *doctor.Report, cfg *config.Config) {
	nv := runtimes.NodeVersion()
	pv := runtimes.PythonVersion()
	gv := runtimes.GoVersion()
	dv := runtimes.DotnetVersion()
	goPath := ""
	if gv == "" && !exec.CommandExists("go") {
		goPath = doctorGoOutsidePath()
	}

	fixRuntime := func(name, detected string) string {
		if detected != "" || cfg == nil || !cfg.RuntimeEnabled(name) {
			return ""
		}
		return "run 'dk apply' to install missing runtime"
	}
	r.Add(doctor.Check{ID: "runtime.node", Title: "Node.js", Category: "runtimes",
		Severity: "warning", Status: skipStatus(nv != ""), Message: nv, Fix: fixRuntime("node", nv)})
	r.Add(doctor.Check{ID: "runtime.python", Title: "Python", Category: "runtimes",
		Severity: "warning", Status: skipStatus(pv != ""), Message: pv, Fix: fixRuntime("python", pv)})
	goMessage, goFix := gv, fixRuntime("go", gv)
	goStatus := skipStatus(gv != "")
	if goPath != "" {
		goStatus = "fail"
		goMessage = "Go found at " + goPath + " but not in PATH"
		goFix = "Add Go to your shell PATH: " + doctorGoPathCommand(goPath)
	}
	r.Add(doctor.Check{ID: "runtime.go", Title: "Go", Category: "runtimes",
		Severity: "warning", Status: goStatus, Message: goMessage, Fix: goFix})
	r.Add(doctor.Check{ID: "runtime.dotnet", Title: ".NET SDK", Category: "runtimes",
		Severity: "warning", Status: skipStatus(dv != ""), Message: dv, Fix: fixRuntime("dotnet", dv)})

	displayHeading(doctorOutput(), "Runtimes")
	if nv != "" {
		displayFprintf(doctorOutput(), "    [OK] Node.js %s\n", nv)
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Node.js not installed")
	}
	if pv != "" {
		displayFprintf(doctorOutput(), "    [OK] Python %s\n", pv)
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Python not installed")
	}
	if gv != "" {
		displayFprintf(doctorOutput(), "    [OK] Go %s\n", gv)
	} else if goPath != "" {
		displayFprintf(doctorOutput(), "    [WARN] Go found at %s but not in PATH\n", goPath)
		displayFprintf(doctorOutput(), "           Run: %s\n", doctorGoPathCommand(goPath))
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] Go not installed")
	}
	if dv != "" {
		displayFprintf(doctorOutput(), "    [OK] .NET SDK %s\n", dv)
	} else {
		displayFprintln(doctorOutput(), "    [SKIP] .NET SDK not installed")
	}
}

func RunDoctorEditors(r *doctor.Report) {
	displayHeading(doctorOutput(), "Editors")
	for _, ed := range editors.DetectAll() {
		status := "pass"
		msg := ""
		if ed.Installed && !ed.WindowsSide {
			msg = ed.Path
		} else if ed.Installed {
			status = "fail"
			msg = "Windows-side path: " + ed.Path
		} else {
			status = "skip"
			msg = "not found"
		}
		fix := ""
		if !ed.Installed {
			fix = fmt.Sprintf("Install %s manually from the official website", ed.Name)
		} else if ed.WindowsSide {
			fix = fmt.Sprintf("Install the Linux version of %s inside WSL: 'brew install --cask %s' if available, or download from the official site", ed.Name, strings.ToLower(ed.Name))
		}
		r.Add(doctor.Check{ID: "editor." + ed.Name, Title: ed.Name + " CLI", Category: "editors",
			Severity: "warning", Status: status, Message: msg, Fix: fix})

		if ed.Installed && !ed.WindowsSide {
			displayFprintf(doctorOutput(), "    [OK] %s: %s\n", ed.Name, msg)
		} else if ed.Installed {
			displayFprintf(doctorOutput(), "    [WARN] %s: %s\n", ed.Name, msg)
		} else {
			displayFprintf(doctorOutput(), "    [SKIP] %s not found\n", ed.Name)
		}
	}
}

func RunDoctorDocker(r *doctor.Report) {
	dockerStatus := docker.Detect()

	status := "fail"
	msg := dockerStatus.Message
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		status = "pass"
		msg = "running"
	}
	fix := ""
	if !dockerStatus.CliInstalled {
		fix = "Install Docker Desktop from https://www.docker.com/products/docker-desktop/ and enable WSL integration"
	} else if !dockerStatus.DaemonRunning {
		fix = "Start Docker Desktop from the Windows Start menu, then ensure WSL integration is enabled in Docker Desktop settings"
	}
	r.Add(doctor.Check{ID: "docker", Title: "Docker", Category: "docker",
		Severity: "warning", Status: status, Message: msg, Fix: fix})

	displayHeading(doctorOutput(), "Docker")
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		displayFprintln(doctorOutput(), "    [OK] Docker running")
	} else {
		displayFprintf(doctorOutput(), "    [WARN] %s\n", dockerStatus.Message)
	}
}

func skipStatus(ok bool) string {
	if ok {
		return "pass"
	}
	return "skip"
}

func doctorOutput() io.Writer {
	if jsonOutput {
		return io.Discard
	}
	return os.Stdout
}
