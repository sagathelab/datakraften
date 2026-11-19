package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/sagathelab/datakraften/internal/project"
	"github.com/sagathelab/datakraften/internal/projectnode"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func currentProject() (*project.Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p, err := project.Discover(cwd)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("no datakraften.yaml found in the current project")
	}
	return p, nil
}

func projectTasks(p *project.Project) []string {
	names := make([]string, 0, len(p.Config.Project.Tasks))
	for name := range p.Config.Project.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func printProjectApproval(p *project.Project) {
	fmt.Printf("  Project: %s\n  Configuration: %s\n", p.Root, p.Path)
	if version := p.Config.Project.Tools["node"]; version != "" {
		fmt.Printf("  Node.js: %s (fnm)\n", version)
	}
	for _, name := range projectTasks(p) {
		fmt.Printf("  Task %q: %q\n", name, p.Config.Project.Tasks[name])
	}
	keys := make([]string, 0, len(p.Config.Project.Env))
	for key := range p.Config.Project.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		fmt.Printf("  Environment variables: %s (values not shown)\n", strings.Join(keys, ", "))
		fmt.Println("  Project environment variables can affect command execution; review their values in the configuration.")
	}
	fmt.Println("  Trust allows these project commands to run. Review the configuration first.")
}

func approveProject(p *project.Project, explicit bool) error {
	trusted, err := project.Trusted(p)
	if err != nil {
		return err
	}
	if trusted {
		return project.Verify(p)
	}
	printProjectApproval(p)
	if !explicit {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("project is not trusted; review the file and run 'dk project trust --yes' explicitly")
		}
		fmt.Print("  Trust this configuration? [y/N]: ")
		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return fmt.Errorf("cannot read confirmation: %w", err)
		}
		answer = strings.TrimSpace(answer)
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			return fmt.Errorf("project trust declined")
		}
	}
	if err := project.Verify(p); err != nil {
		return err
	}
	return project.Trust(p)
}

func newProjectCmd() *cobra.Command {
	root := &cobra.Command{Use: "project", Short: "Inspect and prepare the current project"}
	root.AddCommand(&cobra.Command{Use: "status", Short: "Inspect project requirements without changing anything", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		p, err := currentProject()
		if err != nil {
			return err
		}
		trusted, err := project.Trusted(p)
		if err != nil {
			return err
		}
		result := struct {
			Root          string   `json:"root"`
			Config        string   `json:"config"`
			Trusted       bool     `json:"trusted"`
			NodeRequested string   `json:"node_requested,omitempty"`
			NodeVersion   string   `json:"node_version,omitempty"`
			NodePath      string   `json:"node_path,omitempty"`
			NodeStatus    string   `json:"node_status,omitempty"`
			Tasks         []string `json:"tasks"`
		}{Root: p.Root, Config: p.Path, Trusted: trusted, Tasks: projectTasks(p)}
		if requested := p.Config.Project.Tools["node"]; requested != "" {
			result.NodeRequested = requested
			inst, err := projectnode.Resolve(requested)
			if err != nil {
				result.NodeStatus = err.Error()
			} else {
				result.NodeVersion, result.NodePath, result.NodeStatus = inst.Version, inst.NodePath, "available"
			}
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}
		fmt.Printf("  Project: %s\n  Configuration: %s\n  Trusted: %t\n", result.Root, result.Config, result.Trusted)
		if result.NodeRequested != "" {
			fmt.Printf("  Node.js requested: %s\n", result.NodeRequested)
			if result.NodePath != "" {
				fmt.Printf("  Node.js available: %s (%s)\n", result.NodeVersion, result.NodePath)
			} else {
				fmt.Printf("  Node.js: %s\n", result.NodeStatus)
			}
		}
		if len(result.Tasks) > 0 {
			fmt.Printf("  Tasks: %s\n", strings.Join(result.Tasks, ", "))
		}
		return nil
	}})
	var trustYes bool
	trust := &cobra.Command{Use: "trust", Short: "Review and trust the current project configuration", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return fmt.Errorf("--json is not supported for project trust")
		}
		p, err := currentProject()
		if err != nil {
			return err
		}
		if err := approveProject(p, trustYes); err != nil {
			return err
		}
		fmt.Println("  Project configuration trusted.")
		return nil
	}}
	trust.Flags().BoolVarP(&trustYes, "yes", "y", false, "Explicitly trust this configuration without prompting")
	root.AddCommand(trust)
	var setupYes, setupDry bool
	setup := &cobra.Command{Use: "setup", Short: "Install missing project tools", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return fmt.Errorf("--json is not supported for project setup")
		}
		p, err := currentProject()
		if err != nil {
			return err
		}
		requested := p.Config.Project.Tools["node"]
		if requested == "" {
			fmt.Println("  No project tools to install.")
			return nil
		}
		inst, err := projectnode.Resolve(requested)
		if err == nil {
			fmt.Printf("  Node.js %s is available at %s\n", inst.Version, inst.NodePath)
			return nil
		}
		if !errors.Is(err, projectnode.ErrNotInstalled) {
			return err
		}
		fmt.Printf("  Install Node.js %s using fnm into Datakraften's project tools directory.\n", requested)
		if setupDry {
			fmt.Println("  Dry run: no changes made.")
			return nil
		}
		if err := approveProject(p, false); err != nil {
			return err
		}
		if !setupYes {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("interactive confirmation required; use --yes to approve installation")
			}
			fmt.Print("  Install this Node.js version? [y/N]: ")
			answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("cannot read confirmation: %w", err)
			}
			if answer = strings.TrimSpace(answer); !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
				return fmt.Errorf("installation declined")
			}
		}
		if err := project.Verify(p); err != nil {
			return err
		}
		if err := projectnode.Install(requested, os.Stdout, os.Stderr); err != nil {
			return err
		}
		inst, err = projectnode.Resolve(requested)
		if err != nil {
			return fmt.Errorf("Node.js installation could not be verified: %w", err)
		}
		fmt.Printf("  Node.js %s available at %s\n", inst.Version, inst.NodePath)
		return nil
	}}
	setup.Flags().BoolVarP(&setupYes, "yes", "y", false, "Approve installation without prompting")
	setup.Flags().BoolVarP(&setupDry, "dry-run", "n", false, "Preview installation without changes")
	root.AddCommand(setup)
	return root
}

func projectEnvironment(base []string, vars map[string]string) []string {
	result := make([]string, 0, len(base)+len(vars))
	for _, item := range base {
		key, _, _ := strings.Cut(item, "=")
		if _, ok := vars[key]; !ok {
			result = append(result, item)
		}
	}
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, key+"="+vars[key])
	}
	return result
}

func newRunCmd() *cobra.Command {
	var dryRun bool
	command := &cobra.Command{Use: "run <task> [-- arguments...]", Short: "Run a trusted project task with its project environment", Args: cobra.MinimumNArgs(1), SilenceUsage: true, RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return fmt.Errorf("--json is not supported for project tasks")
		}
		p, err := currentProject()
		if err != nil {
			return err
		}
		task, ok := p.Config.Project.Tasks[args[0]]
		if !ok {
			return fmt.Errorf("unknown project task %q; available: %s", args[0], strings.Join(projectTasks(p), ", "))
		}
		env := projectEnvironment(os.Environ(), p.Config.Project.Env)
		nodePath := ""
		if requested := p.Config.Project.Tools["node"]; requested != "" {
			inst, err := projectnode.Resolve(requested)
			if err != nil {
				if !dryRun {
					return fmt.Errorf("project Node.js %s: %w; run 'dk project setup'", requested, err)
				}
				fmt.Fprintf(os.Stderr, "  Node.js %s: %v\n", requested, err)
			} else {
				env = projectnode.Environment(inst, env)
				nodePath = inst.NodePath
			}
		}
		fmt.Fprintf(os.Stderr, "  Project: %s\n  Task %q: %q\n", p.Root, args[0], task)
		if nodePath != "" {
			fmt.Fprintf(os.Stderr, "  Node.js: %s\n", nodePath)
		}
		if dryRun {
			fmt.Fprintln(os.Stderr, "  Dry run: no task executed and no trust stored.")
			return nil
		}
		if err := approveProject(p, false); err != nil {
			return err
		}
		if err := project.Verify(p); err != nil {
			return err
		}
		script := task
		if len(args) > 1 {
			script += ` "$@"`
		}
		shellArgs := append([]string{"-c", script, "dk"}, args[1:]...)
		process := osexec.Command("/bin/sh", shellArgs...)
		process.Dir = p.Root
		process.Env = env
		process.Stdin, process.Stdout, process.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = runProjectProcess(process)
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) {
			cmd.SilenceErrors = true
			code := exitErr.ExitCode()
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				code = 128 + int(status.Signal())
			}
			return &taskExitError{code: code}
		}
		return err
	}}
	command.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview the task without executing or trusting it")
	return command
}

func runProjectProcess(process *osexec.Cmd) error {
	interactive := term.IsTerminal(int(os.Stdin.Fd())) || term.IsTerminal(int(os.Stdout.Fd())) || term.IsTerminal(int(os.Stderr.Fd()))
	if !interactive {
		process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	signals := make(chan os.Signal, 8)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(signals)
	if err := process.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	forwarded := make(chan struct{})
	go func() {
		defer close(forwarded)
		for {
			select {
			case s := <-signals:
				sig, ok := s.(syscall.Signal)
				if !ok {
					continue
				}
				pid := process.Process.Pid
				if !interactive {
					pid = -pid
				}
				_ = syscall.Kill(pid, sig)
			case <-done:
				return
			}
		}
	}()
	err := process.Wait()
	close(done)
	<-forwarded
	return err
}

type taskExitError struct{ code int }

func (e *taskExitError) Error() string {
	return fmt.Sprintf("project task exited with status %d", e.code)
}
func (e *taskExitError) ExitCode() int { return e.code }
