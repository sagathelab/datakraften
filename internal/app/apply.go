package app

import (
	"fmt"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/shell"
	"github.com/sagathelab/datakraften/internal/system"
)

type ApplyReport struct {
	System  int
	Brew    int
	BrewPkgs int
	Node    bool
	Python  bool
	Shell   bool
	Errors  []string
}

func RunApply(cfg *config.Config, dryRun bool) *ApplyReport {
	report := &ApplyReport{}

	fmt.Println()
	fmt.Println("  System")
	fmt.Println("  ------")

	if !dryRun {
		n, err := installers.APTEnsurePackages(installers.APTBasePackages)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("APT: %s", err))
			fmt.Printf("    ✗ APT dependencies: %s\n", err)
		} else {
			report.System = n
			if n == 0 {
				fmt.Println("    ✓ APT dependencies (already satisfied)")
			} else {
				fmt.Printf("    ✓ %d APT packages installed\n", n)
			}
		}
	} else {
		fmt.Println("    ~ Would install APT base dependencies")
	}

	fmt.Println()
	fmt.Println("  Tooling")
	fmt.Println("  -------")

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

	fmt.Println()
	fmt.Println("  Shell")
	fmt.Println("  -----")

	if !dryRun {
		if shell.FishInstalled() {
			_, err := shell.FishEnsureSetup(brewPrefix)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("Shell: %s", err))
				fmt.Printf("    ✗ Fish config: %s\n", err)
			} else {
				report.Shell = true
				fmt.Println("    ✓ Fish config written")
			}
		} else {
			fmt.Println("    – Fish not installed (install via Brew above)")
		}
	} else {
		fmt.Println("    ~ Would configure Fish shell with managed blocks")
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

func RunDoctorSystem() {
	wsl, wslVer := system.DetectWSL()
	distro, distroVer := system.Distro()
	systemd := system.HasSystemd()
	shell_ := system.CurrentShell()

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

func RunDoctorTools() {
	tools := []struct{
		name string
		installed bool
	}{
		{"git", exec.CommandExists("git")},
		{"gh", exec.CommandExists("gh")},
		{"az", exec.CommandExists("az")},
		{"fnm", exec.CommandExists("fnm")},
		{"uv", exec.CommandExists("uv")},
		{"brew", exec.CommandExists("brew")},
		{"docker", exec.CommandExists("docker")},
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

func RunDoctorRuntimes() {
	fmt.Println("  Runtimes")
	if v := runtimes.NodeVersion(); v != "" {
		fmt.Printf("    ✓ Node.js %s\n", v)
	} else {
		fmt.Println("    – Node.js not installed")
	}
	if v := runtimes.PythonVersion(); v != "" {
		fmt.Printf("    ✓ Python %s\n", v)
	} else {
		fmt.Println("    – Python not installed")
	}
}

func RunDoctorEditors() {
	fmt.Println("  Editors")
	for _, ed := range editors.DetectAll() {
		if ed.Installed && !ed.WindowsSide {
			fmt.Printf("    ✓ %s\n", ed.Name)
		} else if ed.Installed {
			fmt.Printf("    ⚠ %s (Windows-side path)\n", ed.Name)
		} else {
			fmt.Printf("    – %s not found\n", ed.Name)
		}
	}
}

func RunDoctorDocker() {
	fmt.Println("  Docker")
	dockerStatus := docker.Detect()
	if dockerStatus.CliInstalled && dockerStatus.DaemonRunning {
		fmt.Println("    ✓ Docker running")
	} else {
		fmt.Printf("    ⚠ %s\n", dockerStatus.Message)
	}
}
