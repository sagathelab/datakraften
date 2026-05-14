package app

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/ai"
	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/doctor"
	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/shell"
	"github.com/sagathelab/datakraften/internal/system"
)

type ApplyReport struct {
	System   int
	Brew     int
	BrewPkgs int
	Node     bool
	Python   bool
	Dotnet   bool
	AITools  int
	Shell    bool
	Errors   []string
}

func RunApply(cfg *config.Config, dryRun bool) *ApplyReport {
	report := &ApplyReport{}
	pm := installers.DetectPackageManager()

	fmt.Println()
	fmt.Println("  System")
	fmt.Println("  ------")

	if !dryRun {
		n, err := installers.NativeEnsurePackages(installers.NativeBasePackages())
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", pm, err))
			fmt.Printf("    ✗ System dependencies: %s\n", err)
		} else {
			report.System = n
			if n == 0 {
				fmt.Printf("    ✓ System dependencies via %s (already satisfied)\n", pm)
			} else {
				fmt.Printf("    ✓ %d system packages installed via %s\n", n, pm)
			}
		}
	} else {
		fmt.Printf("    ~ Would install system dependencies via %s\n", pm)
	}

	fmt.Println()
	fmt.Println("  Tooling")
	fmt.Println("  -------")

	if pm != installers.BREW {
		if !dryRun {
			installed, err := installers.BrewEnsureInstalled()
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("Brew: %s", err))
				fmt.Printf("    ✗ Homebrew: %s\n", err)
			} else if installed {
				fmt.Println("    ✓ Homebrew installed")
				report.Brew = 1
			} else {
				fmt.Println("    ✓ Homebrew (already installed)")
			}
		} else {
			fmt.Println("    ~ Would ensure Homebrew is installed")
		}
	} else {
		fmt.Println("    ✓ Homebrew (native package manager)")
	}

	brewPrefix := installers.BrewPrefix()

	if !dryRun {
		n, err := installers.BrewEnsurePackages(installers.DefaultBrewPackages)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Brew packages: %s", err))
			fmt.Printf("    ✗ Brew packages: %s\n", err)
		} else {
			report.BrewPkgs = n
			if n == 0 {
				fmt.Println("    ✓ Brew packages (already installed)")
			} else {
				fmt.Printf("    ✓ %d Brew packages installed\n", n)
			}
		}
	} else {
		fmt.Println("    ~ Would install Brew packages: " + strings.Join(installers.DefaultBrewPackages, ", "))
	}

	fmt.Println()
	fmt.Println("  Runtimes")
	fmt.Println("  --------")

	if !dryRun {
		installed, err := runtimes.EnsureNode()
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Node: %s", err))
			fmt.Printf("    ✗ Node.js: %s\n", err)
		} else {
			report.Node = installed
			if installed {
				fmt.Printf("    ✓ Node.js %s installed\n", runtimes.NodeVersion())
			} else {
				fmt.Printf("    ✓ Node.js %s (already installed)\n", runtimes.NodeVersion())
			}
		}
	} else {
		fmt.Println("    ~ Would install Node.js LTS via fnm")
	}

	if !dryRun {
		installed, err := runtimes.EnsurePython()
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Python: %s", err))
			fmt.Printf("    ✗ Python: %s\n", err)
		} else {
			report.Python = installed
			pyVer := runtimes.PythonVersion()
			if installed {
				fmt.Printf("    ✓ Python %s installed\n", pyVer)
			} else if pyVer != "" {
				fmt.Printf("    ✓ Python %s (already installed)\n", pyVer)
			} else {
				fmt.Println("    ✓ Python configured")
			}
		}
	} else {
		fmt.Println("    ~ Would install Python via uv")
	}

	if !dryRun {
		installed, err := runtimes.EnsureDotnet()
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf(".NET: %s", err))
			fmt.Printf("    ✗ .NET SDK: %s\n", err)
		} else {
			report.Dotnet = installed
			dnVer := runtimes.DotnetVersion()
			if installed {
				fmt.Printf("    ✓ .NET SDK %s installed\n", dnVer)
			} else if dnVer != "" {
				fmt.Printf("    ✓ .NET SDK %s (already installed)\n", dnVer)
			} else {
				fmt.Println("    ✓ .NET SDK configured")
			}
		}
	} else {
		fmt.Println("    ~ Would install .NET SDK via Homebrew")
	}

	fmt.Println()
	fmt.Println("  Shell")
	fmt.Println("  -----")

	if !dryRun {
		if !shell.FishInstalled() {
			fmt.Println("    Installing fish via native package manager...")
			if err := installers.NativeInstall(installers.FishPackageName()); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("Fish: %s", err))
				fmt.Printf("    ✗ Fish install: %s\n", err)
			}
		}
		if shell.FishInstalled() {
			_, err := shell.FishEnsureSetup(brewPrefix)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("Shell config: %s", err))
				fmt.Printf("    ✗ Fish config: %s\n", err)
			} else {
				report.Shell = true
				fmt.Println("    ✓ Fish shell configured")
			}
		}
	} else {
		fmt.Println("    ~ Would install and configure Fish shell")
	}

	fmt.Println()
	fmt.Println("  Editors")
	fmt.Println("  -------")

	for _, ed := range editors.DetectAll() {
		if ed.Installed {
			status := "ok"
			if ed.WindowsSide {
				status = "Windows-side, consider Linux CLI install"
			}
			fmt.Printf("    ✓ %s (%s)\n", ed.Name, status)
		} else {
			fmt.Printf("    – %s not found\n", ed.Name)
		}
	}

	fmt.Println()
	fmt.Println("  AI Tools")
	fmt.Println("  --------")

	if !dryRun {
		installed, err := ai.EnsureAITools(cfg)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("AI tools: %s", err))
			fmt.Printf("    ✗ AI tools: %s\n", err)
		} else {
			report.AITools = installed
			if installed == 0 {
				fmt.Println("    ✓ AI tools (already installed)")
			} else {
				fmt.Printf("    ✓ %d AI tool(s) installed\n", installed)
			}
		}
	} else {
		fmt.Println("    ~ Would install AI tools")
	}

	fmt.Println()
	fmt.Println("  Docker")
	fmt.Println("  ------")

	dockerStatus := docker.Detect()
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		fmt.Println("    ✓ Docker running")
	} else {
		fmt.Printf("    – %s\n", dockerStatus.Message)
	}

	fmt.Println()
	return report
}

func RunDoctorSystem(r *doctor.Report) {
	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()
	systemd := system.HasSystemd()
	shell_ := system.CurrentShell()

	r.Add(doctor.Check{ID: "wsl", Title: "WSL detected", Category: "system", Severity: "info",
		Status: boolStatus(wsl && wslVer > 0), Message: fmt.Sprintf("WSL%d: %s %s", wslVer, distro, distroVer)})
	r.Add(doctor.Check{ID: "systemd", Title: "systemd available", Category: "system", Severity: "info",
		Status: boolStatus(systemd)})
	r.Add(doctor.Check{ID: "shell", Title: "Current shell", Category: "system", Severity: "info",
		Status: "pass", Message: shell_})

	fmt.Println("  System")
	if wsl {
		fmt.Printf("    ✓ WSL%d detected\n", wslVer)
	} else {
		fmt.Println("    – Not running in WSL")
	}
	if distro != "" {
		fmt.Printf("    ✓ %s %s\n", distro, distroVer)
	}
	if systemd {
		fmt.Println("    ✓ systemd available")
	}
	if shell_ != "" {
		fmt.Printf("    ✓ Shell: %s\n", shell_)
	}
}

func RunDoctorTools(r *doctor.Report) {
	tools := []struct{
		name string
		id   string
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
		r.Add(doctor.Check{ID: t.id, Title: t.name + " installed", Category: "tools",
			Severity: "warning", Status: boolStatus(t.installed)})
	}

	fmt.Println("  Tools")
	for _, t := range tools {
		if t.installed {
			fmt.Printf("    ✓ %s\n", t.name)
		} else {
			fmt.Printf("    – %s not installed\n", t.name)
		}
	}
}

func RunDoctorRuntimes(r *doctor.Report) {
	nv := runtimes.NodeVersion()
	pv := runtimes.PythonVersion()
	dv := runtimes.DotnetVersion()

	r.Add(doctor.Check{ID: "runtime.node", Title: "Node.js", Category: "runtimes",
		Severity: "warning", Status: boolStatus(nv != ""), Message: nv})
	r.Add(doctor.Check{ID: "runtime.python", Title: "Python", Category: "runtimes",
		Severity: "warning", Status: boolStatus(pv != ""), Message: pv})
	r.Add(doctor.Check{ID: "runtime.dotnet", Title: ".NET SDK", Category: "runtimes",
		Severity: "warning", Status: boolStatus(dv != ""), Message: dv})

	fmt.Println("  Runtimes")
	if nv != "" {
		fmt.Printf("    ✓ Node.js %s\n", nv)
	} else {
		fmt.Println("    – Node.js not installed")
	}
	if pv != "" {
		fmt.Printf("    ✓ Python %s\n", pv)
	} else {
		fmt.Println("    – Python not installed")
	}
	if dv != "" {
		fmt.Printf("    ✓ .NET SDK %s\n", dv)
	} else {
		fmt.Println("    – .NET SDK not installed")
	}
}

func RunDoctorEditors(r *doctor.Report) {
	fmt.Println("  Editors")
	for _, ed := range editors.DetectAll() {
		status := "pass"
		msg := ""
		if ed.Installed && !ed.WindowsSide {
			status = "pass"
		} else if ed.Installed {
			status = "fail"
			msg = "Windows-side path"
		} else {
			status = "fail"
			msg = "not found"
		}
		r.Add(doctor.Check{ID: "editor." + ed.Name, Title: ed.Name + " CLI", Category: "editors",
			Severity: "info", Status: status, Message: msg})

		if ed.Installed && !ed.WindowsSide {
			fmt.Printf("    ✓ %s\n", ed.Name)
		} else if ed.Installed {
			fmt.Printf("    ⚠ %s (Windows-side path)\n", ed.Name)
		} else {
			fmt.Printf("    – %s not found\n", ed.Name)
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
	r.Add(doctor.Check{ID: "docker", Title: "Docker", Category: "docker",
		Severity: "warning", Status: status, Message: msg})

	fmt.Println("  Docker")
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		fmt.Println("    ✓ Docker running")
	} else {
		fmt.Printf("    ⚠ %s\n", dockerStatus.Message)
	}
}

func boolStatus(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}
