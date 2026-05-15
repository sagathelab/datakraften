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
				enc.Encode(r)
				return nil
			}

			fmt.Println("  Datakraften status")
			fmt.Println()

			fmt.Println("  WSL")
			if wsl {
				fmt.Printf("    ✓ WSL%d detected\n", wslVer)
			} else {
				fmt.Println("    – Not running in WSL")
			}
			if distro != "" {
				fmt.Printf("    ✓ %s %s\n", distro, distroVer)
			}
			if system.HasSystemd() {
				fmt.Println("    ✓ systemd available")
			}
			fmt.Println()

			fmt.Println("  Tools")
			for _, name := range []string{"git", "gh", "az", "docker", "fnm", "uv"} {
				if exec.CommandExists(name) {
					fmt.Printf("    ✓ %s\n", name)
				} else {
					fmt.Printf("    – %s not found\n", name)
				}
			}
			fmt.Println()

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
			if v := runtimes.GoVersion(); v != "" {
				fmt.Printf("    ✓ Go %s\n", v)
			} else {
				fmt.Println("    – Go not installed")
			}
			if v := runtimes.DotnetVersion(); v != "" {
				fmt.Printf("    ✓ .NET SDK %s\n", v)
			} else {
				fmt.Println("    – .NET SDK not installed")
			}
			fmt.Println()

			fmt.Println("  Editors")
			for _, ed := range editors.DetectAll() {
				if ed.Installed {
					fmt.Printf("    ✓ %s\n", ed.Name)
				} else {
					fmt.Printf("    – %s not found\n", ed.Name)
				}
			}
			fmt.Println()

			fmt.Println("  Docker")
			dockerStatus := docker.Detect()
			if dockerStatus.DaemonRunning {
				fmt.Println("    ✓ Docker running")
			} else if dockerStatus.CliInstalled {
				fmt.Println("    – Docker daemon not running")
			} else {
				fmt.Println("    – Docker CLI not found")
			}
			fmt.Println()

			if state.LastApply != "" {
				fmt.Printf("  Last apply: %s\n", state.LastApply)
				fmt.Printf("  Source: %s\n", state.ActiveSource)
			}

			return nil
		},
	}

	return cmd
}
