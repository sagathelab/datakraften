package terminal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type rawTerminal struct {
	mu    sync.Mutex
	tty   *os.File
	state *term.State
}

func (t *rawTerminal) restore() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return term.Restore(int(t.tty.Fd()), t.state)
}

func (t *rawTerminal) resume() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := term.MakeRaw(int(t.tty.Fd()))
	return err
}

func Run(cmd *exec.Cmd, tty *os.File, out io.Writer) (runErr error) {
	if cmd.Stdin != nil || cmd.Stdout != nil || cmd.Stderr != nil {
		return errors.New("terminal command must not have preconfigured streams")
	}
	if tty == nil || !term.IsTerminal(int(tty.Fd())) {
		return errors.New("interactive terminal unavailable")
	}
	signals := make(chan os.Signal, 8)
	signal.Notify(signals, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTSTP)
	defer signal.Stop(signals)
	continued := make(chan os.Signal, 1)
	signal.Notify(continued, syscall.SIGCONT)
	defer signal.Stop(continued)
	size, err := pty.GetsizeFull(tty)
	if err != nil {
		return fmt.Errorf("get terminal size: %w", err)
	}
	state, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return fmt.Errorf("set terminal mode: %w", err)
	}
	raw := &rawTerminal{tty: tty, state: state}
	defer func() {
		if err := raw.restore(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("restore terminal mode: %w", err))
		}
	}()
	master, err := pty.StartWithSize(cmd, size)
	if err != nil {
		return fmt.Errorf("start interactive command: %w", err)
	}
	_ = pty.InheritSize(tty, master)
	done := make(chan struct{})
	suspends := make(chan struct{}, 1)

	inputDone := make(chan error, 1)
	go func() { inputDone <- forwardInput(tty, master, suspends, done) }()
	outputDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(out, master)
		if errors.Is(err, syscall.EIO) {
			err = nil
		}
		outputDone <- err
	}()
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	var waitErr, outputErr, inputErr error
	var drainTimeout <-chan time.Time
	var timer *time.Timer
	var cutoff bool
	waited, outputClosed := false, false
	for !waited || !outputClosed {
		select {
		case sig := <-signals:
			if sig == syscall.SIGWINCH {
				_ = pty.InheritSize(tty, master)
			} else if sig == syscall.SIGTSTP {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTSTP)
				if err := suspend(raw, master, tty, cmd.Process.Pid, continued); err != nil {
					inputErr = err
					_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				}
			} else if s, ok := sig.(syscall.Signal); ok {
				_ = syscall.Kill(-cmd.Process.Pid, s)
			}
		case err := <-inputDone:
			inputErr = err
			inputDone = nil
			if err != nil {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
		case <-suspends:
			if err := suspend(raw, master, tty, cmd.Process.Pid, continued); err != nil {
				inputErr = err
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
		case err := <-outputDone:
			if cutoff && errors.Is(err, os.ErrClosed) {
				err = nil
			}
			outputErr = err
			outputDone = nil
			outputClosed = true
			if err != nil {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
		case waitErr = <-waitDone:
			waitDone = nil
			waited = true
			if !outputClosed {
				timer = time.NewTimer(2 * time.Second)
				drainTimeout = timer.C
			}
		case <-drainTimeout:
			cutoff = true
			drainTimeout = nil
			_ = master.Close()
		}
	}
	if timer != nil {
		timer.Stop()
	}
	close(done)
	_ = master.Close()
	if inputDone != nil {
		if err := <-inputDone; inputErr == nil {
			inputErr = err
		}
	}
	if inputErr != nil {
		return fmt.Errorf("read terminal input: %w", inputErr)
	}
	if outputErr != nil {
		return fmt.Errorf("write terminal output: %w", outputErr)
	}
	return waitErr
}

func suspend(raw *rawTerminal, master, tty *os.File, pid int, continued <-chan os.Signal) error {
	select {
	case <-continued:
	default:
	}
	if err := raw.restore(); err != nil {
		return err
	}
	if err := syscall.Kill(os.Getpid(), syscall.SIGSTOP); err != nil {
		return err
	}
	<-continued
	if err := raw.resume(); err != nil {
		return err
	}
	_ = pty.InheritSize(tty, master)
	_ = syscall.Kill(-pid, syscall.SIGCONT)
	return nil
}

func forwardInput(tty, master *os.File, suspends chan<- struct{}, done <-chan struct{}) error {
	buffer := make([]byte, 4096)
	fds := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
	for {
		select {
		case <-done:
			return nil
		default:
		}
		n, err := unix.Poll(fds, 100)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		if fds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return io.EOF
		}
		read, err := unix.Read(int(tty.Fd()), buffer)
		if errors.Is(err, syscall.EINTR) || errors.Is(err, syscall.EAGAIN) {
			continue
		}
		if err != nil {
			return err
		}
		if read == 0 {
			return io.EOF
		}
		suspendChar, ok := suspendKey(master)
		for _, b := range buffer[:read] {
			if _, err := master.Write([]byte{b}); err != nil {
				if errors.Is(err, os.ErrClosed) || errors.Is(err, syscall.EIO) {
					return nil
				}
				return err
			}
			if ok && b == suspendChar {
				select {
				case suspends <- struct{}{}:
				case <-done:
					return nil
				}
			}
		}
	}
}
