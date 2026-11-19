package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/sagathelab/datakraften/internal/config"
	"github.com/sagathelab/datakraften/internal/docker"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/exec"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/shell"
	"github.com/sagathelab/datakraften/internal/system"
	"github.com/spf13/cobra"
)

type workstationReport struct {
	Configuration workstationConfiguration `json:"configuration"`
	LastApply     string                   `json:"last_apply,omitempty"`
	LastSource    string                   `json:"last_apply_source,omitempty"`
}

type workstationConfiguration struct {
	Present  bool               `json:"present"`
	Path     string             `json:"path"`
	Source   string             `json:"source,omitempty"`
	Shell    config.ShellTarget `json:"shell,omitempty"`
	Runtimes map[string]bool    `json:"runtimes,omitempty"`
	Editors  map[string]bool    `json:"editors,omitempty"`
	AITools  map[string]bool    `json:"ai_tools,omitempty"`
	AIApps   map[string]bool    `json:"ai_apps,omitempty"`
	Warnings []string           `json:"warnings,omitempty"`
}

func loadWorkstationReport() (workstationReport, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return workstationReport{}, nil, fmt.Errorf("cannot load configuration: %w", err)
	}
	state := LoadState()
	result := workstationReport{
		Configuration: workstationConfiguration{Present: cfg != nil, Path: config.DefaultConfigPath()},
		LastApply:     state.LastApply,
	}
	if result.LastApply != "" {
		result.LastSource = state.ActiveSource
	}
	if cfg != nil {
		intent := &result.Configuration
		intent.Source, intent.Shell, intent.Warnings = cfg.Source, cfg.ShellTarget(), cfg.Warnings
		intent.Runtimes = make(map[string]bool)
		for _, name := range config.RuntimeNames() {
			intent.Runtimes[name] = cfg.RuntimeEnabled(name)
		}
		intent.Editors = map[string]bool{"zed": cfg.Editors["zed"] == "true"}
		intent.AITools = enabledAI(config.AIToolNames(), cfg.AITools)
		intent.AIApps = enabledAI(config.AIAppNames(), cfg.AIApps)
	}
	return result, cfg, nil
}

func enabledAI(names []string, entries map[string]config.RuntimeConfig) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = entries[name].Enabled
	}
	return result
}

func (r workstationReport) print(w io.Writer) {
	displayHeading(w, "Configured intent")
	intent := r.Configuration
	fmt.Fprintf(w, "    Workstation config: %s\n", intent.Path)
	if !intent.Present {
		fmt.Fprintln(w, "    No workstation configuration found; run 'dk init' to create one.")
	} else {
		fmt.Fprintf(w, "    Source: %s\n    Shell target: %s\n", intent.Source, intent.Shell)
		printOptions := func(label string, options map[string]bool) {
			values := make([]string, 0, len(options))
			for _, name := range slices.Sorted(maps.Keys(options)) {
				state := "disabled"
				if options[name] {
					state = "enabled"
				}
				values = append(values, name+"="+state)
			}
			fmt.Fprintf(w, "    %s: %s\n", label, strings.Join(values, ", "))
		}
		printOptions("Runtimes", intent.Runtimes)
		printOptions("Editors", intent.Editors)
		printOptions("AI tools", intent.AITools)
		printOptions("AI apps", intent.AIApps)
	}
	if r.LastApply != "" {
		fmt.Fprintln(w)
		displayHeading(w, "Apply history")
		fmt.Fprintf(w, "    Last apply: %s\n", r.LastApply)
		if r.LastSource != "" {
			fmt.Fprintf(w, "    Last apply source: %s\n", r.LastSource)
		}
	}
	fmt.Fprintln(w)
}

type detectedShells struct {
	Current string                             `json:"current"`
	Targets map[config.ShellTarget]shellStatus `json:"targets"`
}

type shellStatus struct {
	Installed    bool   `json:"installed"`
	ConfigPath   string `json:"config_path"`
	ManagedBlock string `json:"managed_block"`
}

func detectShells() detectedShells {
	r := detectedShells{Current: system.CurrentShell(), Targets: make(map[config.ShellTarget]shellStatus)}
	for target, rc := range managedShells {
		path := rc.path()
		r.Targets[target] = shellStatus{exec.CommandExists(string(target)), path, shell.ManagedBlockStatus(path)}
	}
	return r
}

func (r detectedShells) currentLabel() string {
	if r.Current == "" {
		return "unknown"
	}
	return r.Current
}

func (s shellStatus) description() string {
	return fmt.Sprintf("on PATH: %t; managed block: %s (%s)", s.Installed, s.ManagedBlock, s.ConfigPath)
}

type StatusReport struct {
	workstationReport
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
	Shell detectedShells `json:"shell"`
}

type editorStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "status",
		Short:        "Show configured intent and detected environment",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workstation, _, err := loadWorkstationReport()
			if err != nil {
				return err
			}
			printConfigWarnings(workstation.Configuration.Warnings)
			report := detectStatus(workstation)
			if jsonOutput {
				return printReportJSON(report)
			}
			report.print()
			return nil
		},
	}
}

func printReportJSON(report any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func detectStatus(workstation workstationReport) StatusReport {
	r := StatusReport{workstationReport: workstation, Tools: make(map[string]bool), Shell: detectShells()}
	r.WSL.Detected, r.WSL.Version = system.DetectWSL()
	r.WSL.Distro, r.WSL.DistroVer = system.Distro()
	r.WSL.Systemd = system.HasSystemd()
	for _, name := range []string{"git", "gh", "az", "docker", "brew", "fnm", "uv", "copilot"} {
		r.Tools[name] = exec.CommandExists(name)
	}
	r.Runtimes.Node = runtimes.NodeVersion()
	r.Runtimes.Python = runtimes.PythonVersion()
	r.Runtimes.Go = runtimes.GoVersion()
	r.Runtimes.Dotnet = runtimes.DotnetVersion()
	for _, ed := range editors.DetectAll() {
		r.Editors = append(r.Editors, editorStatus{Name: ed.Name, Installed: ed.Installed, Path: ed.Path})
	}
	dockerStatus := docker.Detect()
	r.Docker.CliInstalled = dockerStatus.CliInstalled
	r.Docker.DaemonRunning = dockerStatus.DaemonRunning
	return r
}

func (r StatusReport) print() {
	displayTitle(os.Stdout, "status", "Environment overview")
	r.workstationReport.print(os.Stdout)
	displayHeading(os.Stdout, "Detected system")
	fmt.Printf("    WSL: %t", r.WSL.Detected)
	if r.WSL.Detected {
		fmt.Printf(" (version %d)", r.WSL.Version)
	}
	fmt.Println()
	if r.WSL.Distro != "" {
		fmt.Printf("    Distribution: %s\n", strings.TrimSpace(r.WSL.Distro+" "+r.WSL.DistroVer))
	}
	fmt.Printf("    systemd available: %t\n\n", r.WSL.Systemd)

	displayHeading(os.Stdout, "Detected shell")
	fmt.Printf("    Current shell ($SHELL): %s\n", r.Shell.currentLabel())
	for _, target := range slices.Sorted(maps.Keys(r.Shell.Targets)) {
		fmt.Printf("    %s: %s\n", target, r.Shell.Targets[target].description())
	}
	fmt.Println()

	displayHeading(os.Stdout, "Detected tools")
	for _, name := range slices.Sorted(maps.Keys(r.Tools)) {
		fmt.Printf("    %s on PATH: %t\n", name, r.Tools[name])
	}
	fmt.Println()

	displayHeading(os.Stdout, "Detected runtimes")
	for _, item := range []struct{ name, version string }{
		{"Node.js", r.Runtimes.Node}, {"Python", r.Runtimes.Python}, {"Go", r.Runtimes.Go}, {".NET SDK", r.Runtimes.Dotnet},
	} {
		if item.version == "" {
			item.version = "not detected"
		}
		fmt.Printf("    %s: %s\n", item.name, item.version)
	}
	fmt.Println()

	displayHeading(os.Stdout, "Detected editors")
	for _, ed := range r.Editors {
		fmt.Printf("    %s installed: %t", ed.Name, ed.Installed)
		if ed.Path != "" {
			fmt.Printf(" (%s)", ed.Path)
		}
		fmt.Println()
	}
	fmt.Println()

	displayHeading(os.Stdout, "Detected Docker")
	fmt.Printf("    CLI installed: %t\n    Daemon running: %t\n\n", r.Docker.CliInstalled, r.Docker.DaemonRunning)
}
