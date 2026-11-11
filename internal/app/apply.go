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
	Shell         bool
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
		failed = true
		emitProgress(hook, ProgressEvent{Step: step, Status: ProgressFailed, Err: fmt.Errorf("%s", message)})
	}
	defer func() {
		if started && !failed {
			emitProgress(hook, ProgressEvent{Step: step, Status: ProgressCompleted})
		}
	}()
	pm := installers.DetectPackageManager()
	verbosePrintf("packageManager=%s source=%s\n", pm, cfg.Source)

	if cfg.Source == "team" {
		start("configuration", "Fetching team configuration")
		if cfg.URL == "" {
			fail("team source requires a URL — set url in config")
			displayPrintln("  [ERROR] Team source is active but no url is set.")
			fmt.Fprintln(progressWriter(), "  Run 'dk init --team <url>' to configure.")
			return report
		}

		fmt.Fprintf(progressWriter(), "  Fetching remote config from %s...\n", cfg.URL)

		remoteCfg, err := config.FetchRemote(cfg.URL)
		if err != nil {
			fail(fmt.Sprintf("remote config: %s", err))
			displayPrintf("    [ERROR] %s\n", err)
			return report
		}

		displayPrintln("    [OK] Remote config loaded")

		source := cfg.Source
		url := cfg.URL
		*cfg = *remoteCfg
		cfg.Source = source
		cfg.URL = url
	}

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
			installed, err := installers.BrewEnsureInstalled()
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

	brewPrefix := installers.BrewPrefix()

	verbosePrintf("brewPrefix=%s\n", brewPrefix)
	verbosePrintf("brewPkgs=%v\n", installers.DefaultBrewPackages)

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking Homebrew packages...")
		n, err := installers.BrewEnsurePackages(installers.DefaultBrewPackages)
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
		displayPrintln("    [PLAN] Would install Brew packages: " + strings.Join(installers.DefaultBrewPackages, ", "))
	}

	fmt.Fprintln(progressWriter())
	start("runtimes", "Checking runtimes")
	if hook == nil {
		displayStep(progressWriter(), 3, 8, "Runtimes")
	}

	if shouldInstallRuntime(cfg, "node") && !dryRun {
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
	} else if !dryRun && !shouldInstallRuntime(cfg, "node") {
		displayPrintln("    [SKIP] Node.js (disabled in config)")
	} else {
		if shouldInstallRuntime(cfg, "node") {
			displayPrintln("    [PLAN] Would install Node.js LTS via fnm")
		} else {
			displayPrintln("    [SKIP] Node.js (disabled in config)")
		}
	}

	if shouldInstallRuntime(cfg, "python") && !dryRun {
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
	} else if !dryRun && !shouldInstallRuntime(cfg, "python") {
		displayPrintln("    [SKIP] Python (disabled in config)")
	} else {
		if shouldInstallRuntime(cfg, "python") {
			displayPrintln("    [PLAN] Would install Python via uv")
		} else {
			displayPrintln("    [SKIP] Python (disabled in config)")
		}
	}

	if shouldInstallRuntime(cfg, "go") && !dryRun {
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
	} else if !dryRun && !shouldInstallRuntime(cfg, "go") {
		displayPrintln("    [SKIP] Go (disabled in config)")
	} else {
		if shouldInstallRuntime(cfg, "go") {
			displayPrintln("    [PLAN] Would install Go via Homebrew")
		} else {
			displayPrintln("    [SKIP] Go (disabled in config)")
		}
	}

	if shouldInstallRuntime(cfg, "dotnet") && !dryRun {
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
	} else if !dryRun && !shouldInstallRuntime(cfg, "dotnet") {
		displayPrintln("    [SKIP] .NET SDK (disabled in config)")
	} else {
		if shouldInstallRuntime(cfg, "dotnet") {
			displayPrintln("    [PLAN] Would install .NET SDK via Homebrew")
		} else {
			displayPrintln("    [SKIP] .NET SDK (disabled in config)")
		}
	}

	fmt.Fprintln(progressWriter())
	start("shell", "Configuring shell")
	if hook == nil {
		displayStep(progressWriter(), 4, 8, "Shell")
	}

	if !dryRun {
		if !shell.FishInstalled() {
			fmt.Fprintln(progressWriter(), "    Installing fish via native package manager...")
			if err := installers.NativeInstall(installers.FishPackageName()); err != nil {
				fail(fmt.Sprintf("Fish: %s", err))
				displayPrintf("    [ERROR] Fish install: %s\n", err)
			}
		}
		if shell.FishInstalled() {
			_, err := shell.FishEnsureSetup(brewPrefix)
			if err != nil {
				fail(fmt.Sprintf("Shell config: %s", err))
				displayPrintf("    [ERROR] Fish config: %s\n", err)
			} else {
				report.Shell = true
				s := LoadState()
				s.ManagedShell = true
				s.Save()
				displayPrintln("    [OK] Fish shell configured")
			}
		}
	} else {
		displayPrintln("    [PLAN] Would install and configure Fish shell")
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

	fmt.Fprintln(progressWriter())
	start("ai-tools", "Checking AI tools")
	if hook == nil {
		displayStep(progressWriter(), 6, 8, "AI tools")
	}

	if !dryRun {
		fmt.Fprintln(progressWriter(), "    Checking AI tools...")
		report.AITools = ai.EnsureAIToolsWithOptions(cfg, ai.InstallOptions{OnError: func(message string) { fail("AI tools: " + message) }})
		if report.AITools.HasFailures() {
			displayPrintf("    [ERROR] AI tools: %s\n", installReportSummary(report.AITools))
		} else if report.AITools.Enabled == 0 {
			displayPrintln("    [SKIP] AI tools: (disabled)")
		} else {
			displayPrintf("    [OK] AI tools: %s\n", installReportSummary(report.AITools))
		}
	} else {
		displayPrintln("    [PLAN] Would install AI tools")
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
		displayPrintln("    [PLAN] Would install AI apps")
	}

	fmt.Fprintln(progressWriter())
	start("docker", "Checking Docker")
	if hook == nil {
		displayStep(progressWriter(), 8, 8, "Docker")
	}

	dockerStatus := docker.Detect()
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		displayPrintln("    [OK] Docker running")
	} else {
		displayPrintf("    [SKIP] %s\n", dockerStatus.Message)
	}

	fmt.Fprintln(progressWriter())
	return report
}

func shouldInstallRuntime(cfg *config.Config, name string) bool {
	switch name {
	case "node":
		return cfg.Runtimes.Node.Enabled
	case "python":
		return cfg.Runtimes.Python.Enabled
	case "go":
		return cfg.Runtimes.Go.Enabled
	case "dotnet":
		return cfg.Runtimes.Dotnet.Enabled
	}
	return true
}

func RunDoctorSystem(r *doctor.Report) {
	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()
	systemd := system.HasSystemd()
	shell_ := system.CurrentShell()

	r.Add(doctor.Check{ID: "wsl", Title: "WSL detected", Category: "system", Severity: "info",
		Status: boolStatus(wsl && wslVer > 0), Message: fmt.Sprintf("WSL%d: %s %s", wslVer, distro, distroVer),
		Fix: "If WSL is not detected, ensure you are running inside WSL: 'wsl --install' from PowerShell"})
	r.Add(doctor.Check{ID: "systemd", Title: "systemd available", Category: "system", Severity: "info",
		Status: boolStatus(systemd),
		Fix:    "Enable systemd in /etc/wsl.conf: add [boot]\nsystemd=true, then 'wsl --terminate' from PowerShell"})
	r.Add(doctor.Check{ID: "shell", Title: "Current shell", Category: "system", Severity: "info",
		Status: "pass", Message: shell_})

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
	}
	if shell_ != "" {
		displayFprintf(doctorOutput(), "    [OK] Shell: %s\n", shell_)
	}
}

func RunDoctorTools(r *doctor.Report) {
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
		{"docker", "tool.docker", exec.CommandExists("docker")},
	}

	for _, t := range tools {
		fix := ""
		if !t.installed {
			fix = "run 'dk apply' to install missing tools"
		}
		r.Add(doctor.Check{ID: t.id, Title: t.name + " installed", Category: "tools",
			Severity: "warning", Status: boolStatus(t.installed), Fix: fix})
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
		fmt.Fprintf(doctorOutput(), "    %s\n", message)
	}

	for _, t := range tools {
		if t.installed {
			displayFprintf(doctorOutput(), "    [OK] %s\n", t.name)
		} else {
			displayFprintf(doctorOutput(), "    [SKIP] %s not installed\n", t.name)
		}
	}
}

func RunDoctorRuntimes(r *doctor.Report) {
	nv := runtimes.NodeVersion()
	pv := runtimes.PythonVersion()
	gv := runtimes.GoVersion()
	dv := runtimes.DotnetVersion()

	fixRuntime := func(missing string) string {
		if missing != "" {
			return ""
		}
		return "run 'dk apply' to install missing runtime"
	}
	r.Add(doctor.Check{ID: "runtime.node", Title: "Node.js", Category: "runtimes",
		Severity: "warning", Status: boolStatus(nv != ""), Message: nv, Fix: fixRuntime(nv)})
	r.Add(doctor.Check{ID: "runtime.python", Title: "Python", Category: "runtimes",
		Severity: "warning", Status: boolStatus(pv != ""), Message: pv, Fix: fixRuntime(pv)})
	r.Add(doctor.Check{ID: "runtime.go", Title: "Go", Category: "runtimes",
		Severity: "warning", Status: boolStatus(gv != ""), Message: gv, Fix: fixRuntime(gv)})
	r.Add(doctor.Check{ID: "runtime.dotnet", Title: ".NET SDK", Category: "runtimes",
		Severity: "warning", Status: boolStatus(dv != ""), Message: dv, Fix: fixRuntime(dv)})

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
			msg = ""
		} else if ed.Installed {
			status = "fail"
			msg = "Windows-side path"
		} else {
			status = "fail"
			msg = "not found"
		}
		fix := ""
		if !ed.Installed {
			fix = fmt.Sprintf("Install %s manually from the official website or run 'dk apply' if supported", ed.Name)
		} else if ed.WindowsSide {
			fix = fmt.Sprintf("Install the Linux version of %s inside WSL: 'brew install --cask %s' if available, or download from the official site", ed.Name, strings.ToLower(ed.Name))
		}
		r.Add(doctor.Check{ID: "editor." + ed.Name, Title: ed.Name + " CLI", Category: "editors",
			Severity: "info", Status: status, Message: msg, Fix: fix})

		if ed.Installed && !ed.WindowsSide {
			displayFprintf(doctorOutput(), "    [OK] %s\n", ed.Name)
		} else if ed.Installed {
			displayFprintf(doctorOutput(), "    [WARN] %s (Windows-side path)\n", ed.Name)
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

func boolStatus(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}

func doctorOutput() io.Writer {
	if jsonOutput {
		return io.Discard
	}
	return os.Stdout
}
