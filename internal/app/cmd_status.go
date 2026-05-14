package app

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show a friendly overview of your environment",
		Long:  `Display a summary of installed tools, runtimes, AI tools, and editors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := LoadState()
			wsl, wslVer := system.DetectWSL()
			distro, distroVer := system.Distro()

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
				fmt.Printf("  Profile: %s\n", state.ActiveProfile)
			}

			return nil
		},
	}

	return cmd
}
