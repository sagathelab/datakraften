package app

import (
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"sync"

	"github.com/sagathelab/datakraften/internal/exec"
	"golang.org/x/term"

	"github.com/sagathelab/datakraften/internal/ai"
	"github.com/sagathelab/datakraften/internal/editors"
	"github.com/sagathelab/datakraften/internal/installers"
	"github.com/sagathelab/datakraften/internal/progress"
	"github.com/sagathelab/datakraften/internal/runtimes"
	"github.com/sagathelab/datakraften/internal/shell"
	"github.com/sagathelab/datakraften/internal/terminal"
)

type progressOutput struct {
	renderer *progress.Renderer
	message  string
	mu       sync.Mutex
	err      error
}

var activeProgress struct {
	sync.RWMutex
	output *progressOutput
}

func progressWriter() io.Writer {
	activeProgress.RLock()
	p := activeProgress.output
	activeProgress.RUnlock()
	if p != nil {
		return p
	}
	return os.Stdout
}

func newProgressOutput(out io.Writer) *progressOutput {
	return &progressOutput{renderer: progress.New(out)}
}

func newLineStartProgressOutput(out io.Writer) *progressOutput {
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor || verbose {
		return newProgressOutput(out)
	}
	return &progressOutput{renderer: progress.NewAtLineStart(out)}
}

func (p *progressOutput) activate() func() {
	activeProgress.Lock()
	defer activeProgress.Unlock()
	previous := activeProgress.output
	activeProgress.output = p
	ai.SetOutput(p)
	installers.SetOutput(p)
	installers.SetSudoRunner(p.runSudo)
	runtimes.SetOutput(p)
	editors.SetOutput(p)
	shell.SetOutput(p)
	return func() {
		activeProgress.Lock()
		defer activeProgress.Unlock()
		activeProgress.output = previous
		var w io.Writer
		if previous != nil {
			w = previous
		}
		ai.SetOutput(w)
		installers.SetOutput(w)
		if previous != nil {
			installers.SetSudoRunner(previous.runSudo)
		} else {
			installers.SetSudoRunner(nil)
		}
		runtimes.SetOutput(w)
		editors.SetOutput(w)
		shell.SetOutput(w)
	}
}

func (p *progressOutput) runSudo(args ...string) exec.Result {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return exec.Run("sudo", append([]string{"-n"}, args...)...)
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return exec.Result{Code: -1, Stderr: fmt.Sprintf("cannot open terminal: %s", err)}
	}
	defer tty.Close()
	cmd := osexec.Command("sudo", args...)
	if err := terminal.Run(cmd, tty, p.renderer.TerminalWriter()); err != nil {
		code := -1
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			code = cmd.ProcessState.ExitCode()
		}
		if code != -1 {
			return exec.Result{Code: code, Stderr: fmt.Sprintf("sudo exited with status %d (see terminal output)", code)}
		}
		return exec.Result{Code: code, Stderr: fmt.Sprintf("sudo failed: %s (see terminal output)", err)}
	}
	return exec.Result{}
}

func progressTerminalWriter() io.Writer {
	activeProgress.RLock()
	p := activeProgress.output
	activeProgress.RUnlock()
	if p != nil {
		return p.renderer.TerminalWriter()
	}
	return os.Stdout
}

func progressHandoff(fn func() error) error {
	activeProgress.RLock()
	p := activeProgress.output
	activeProgress.RUnlock()
	if p != nil {
		return p.renderer.Handoff(fn)
	}
	return fn()
}

func (p *progressOutput) Write(b []byte) (int, error) {
	n, err := p.renderer.Write(b)
	if err != nil {
		p.mu.Lock()
		if p.err == nil {
			p.err = err
		}
		p.mu.Unlock()
	}
	return n, err
}

func (p *progressOutput) writeStatus(text string) {
	_, err := p.renderer.WriteFormatted([]byte(text), func(b []byte) []byte {
		return []byte(formatStatus(string(b), p.renderer.ColorEnabled()))
	})
	if err != nil {
		p.mu.Lock()
		if p.err == nil {
			p.err = err
		}
		p.mu.Unlock()
	}
}

func (p *progressOutput) event(event ProgressEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return
	}
	switch event.Status {
	case ProgressStarted:
		p.message = event.Message
		p.err = p.renderer.Running(p.message)
	case ProgressCompleted:
		p.err = p.renderer.Done(p.message)
	case ProgressFailed:
		p.err = p.renderer.Fail(p.message)
	}
}

func (p *progressOutput) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.renderer.Close()
	if p.err != nil {
		return p.err
	}
	return err
}

func finishProgress(output *progressOutput, runErr error) error {
	if err := output.close(); err != nil {
		if runErr != nil {
			return fmt.Errorf("%w; write progress: %v", runErr, err)
		}
		return fmt.Errorf("write progress: %w", err)
	}
	return runErr
}
