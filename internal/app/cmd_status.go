package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
)

type StatusReport struct {
	WSL struct {
		Detected  bool   `json:"detected"`
		Version   int    `json:"version,omitempty"`
		Distro    string `json:"distro,omitempty"`
		DistroVer string `json:"distro_version,omitempty"`
		Systemd   bool   `json:"systemd"`
	} `json:"wsl"`
	Tools    map[string]bool `json:"tools"`
	Runtimes struct {
		Node   string `json:"node,omitempty"`
		Python string `json:"python,omitempty"`
		Go     string `json:"go,omitempty"`
		Dotnet string `json:"dotnet,omitempty"`
	} `json:"runtimes"`
	Editors []editorStatus `json:"editors"`
	Docker  struct {
		CliInstalled  bool `json:"cli_installed"`
		DaemonRunning bool `json:"daemon_running"`
	} `json:"docker"`
	LastApply string `json:"last_apply,omitempty"`
	Source    string `json:"source,omitempty"`
}

type editorStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show a friendly overview of your environment",
		Long:  `Display a summary of installed tools, runtimes, AI tools, and editors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := LoadState()
			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()

			if jsonOutput {
				r := StatusReport{}
				r.WSL.Detected = wsl
				if wsl {
					r.WSL.Version = wslVer
				}
				r.WSL.Distro = distro
				r.WSL.DistroVer = distroVer
				r.WSL.Systemd = system.HasSystemd()

				r.Tools = map[string]bool{}
				for _, name := range []string{"git", "gh", "az", "docker", "fnm", "uv"} {
					r.Tools[name] = exec.CommandExists(name)
				}

				r.Runtimes.Node = runtimes.NodeVersion()
				r.Runtimes.Python = runtimes.PythonVersion()
				r.Runtimes.Go = runtimes.GoVersion()
				r.Runtimes.Dotnet = runtimes.DotnetVersion()

				for _, ed := range editors.DetectAll() {
					r.Editors = append(r.Editors, editorStatus{
						Name:      ed.Name,
						Installed: ed.Installed,
						Path:      ed.Path,
					})
				}

				dockerStatus := docker.Detect()
				r.Docker.CliInstalled = dockerStatus.CliInstalled
				r.Docker.DaemonRunning = dockerStatus.DaemonRunning

				r.LastApply = state.LastApply
				r.Source = state.ActiveSource

				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(r)
			}

			displayTitle(os.Stdout, "status", "Environment overview")

			displayHeading(os.Stdout, "System")
			if wsl {
				displayPrintf("    [OK] WSL%d detected\n", wslVer)
			} else {
				displayPrintln("    [SKIP] Not running in WSL")
			}
			if distro != "" {
				displayPrintf("    [OK] %s %s\n", distro, distroVer)
			}
			if system.HasSystemd() {
				displayPrintln("    [OK] systemd available")
			}
			fmt.Println()

			displayHeading(os.Stdout, "Tools")
			for _, name := range []string{"git", "gh", "az", "docker", "fnm", "uv"} {
				if exec.CommandExists(name) {
					displayPrintf("    [OK] %s\n", name)
				} else {
					displayPrintf("    [SKIP] %s not found\n", name)
				}
			}
			fmt.Println()

			displayHeading(os.Stdout, "Runtimes")
			if v := runtimes.NodeVersion(); v != "" {
				displayPrintf("    [OK] Node.js %s\n", v)
			} else {
				displayPrintln("    [SKIP] Node.js not installed")
			}
			if v := runtimes.PythonVersion(); v != "" {
				displayPrintf("    [OK] Python %s\n", v)
			} else {
				displayPrintln("    [SKIP] Python not installed")
			}
			if v := runtimes.GoVersion(); v != "" {
				displayPrintf("    [OK] Go %s\n", v)
			} else {
				displayPrintln("    [SKIP] Go not installed")
			}
			if v := runtimes.DotnetVersion(); v != "" {
				displayPrintf("    [OK] .NET SDK %s\n", v)
			} else {
				displayPrintln("    [SKIP] .NET SDK not installed")
			}
			fmt.Println()

			displayHeading(os.Stdout, "Editors")
			for _, ed := range editors.DetectAll() {
				if ed.Installed {
					displayPrintf("    [OK] %s\n", ed.Name)
				} else {
					displayPrintf("    [SKIP] %s not found\n", ed.Name)
				}
			}
			fmt.Println()

			displayHeading(os.Stdout, "Docker")
			dockerStatus := docker.Detect()
			if dockerStatus.DaemonRunning {
				displayPrintln("    [OK] Docker running")
			} else if dockerStatus.CliInstalled {
				displayPrintln("    [SKIP] Docker daemon not running")
			} else {
				displayPrintln("    [SKIP] Docker CLI not found")
			}
			fmt.Println()

			if state.LastApply != "" {
				displayHeading(os.Stdout, "Configuration")
				fmt.Printf("    Last apply  %s\n", state.LastApply)
				fmt.Printf("    Source      %s\n", state.ActiveSource)
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}
