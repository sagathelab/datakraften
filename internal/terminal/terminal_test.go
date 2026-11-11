package terminal

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type capturedOutput struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	update chan struct{}
}

type brokenOutput struct{}

func (brokenOutput) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func (c *capturedOutput) Write(p []byte) (int, error) {
	c.mu.Lock()
	n, err := c.buffer.Write(p)
	c.mu.Unlock()
	select {
	case c.update <- struct{}{}:
	default:
	}
	return n, err
}

func (c *capturedOutput) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buffer.String()
}

func TestRunForwardsPasswordWithoutEchoAndRestoresTerminal(t *testing.T) {
	outer, tty, err := pty.Open()
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	defer outer.Close()
	defer tty.Close()
	if err := pty.Setsize(tty, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	before, err := term.GetState(int(tty.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	out := &capturedOutput{update: make(chan struct{}, 1)}
	cmd := exec.Command("sh", "-c", `stty -echo; printf 'Password: '; IFS= read -r answer; stty echo; [ "$answer" = secret ] || exit 4; printf '\naccepted\n'`)
	done := make(chan error, 1)
	go func() { done <- Run(cmd, tty, out) }()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for !strings.Contains(out.String(), "Password: ") {
		select {
		case err := <-done:
			t.Fatalf("command exited before prompt: %v", err)
		case <-out.update:
		case <-timer.C:
			t.Fatalf("prompt timed out; output %q", out.String())
		}
	}
	if _, err := io.WriteString(outer, "secret\n"); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}
	case <-timer.C:
		t.Fatal("command timed out")
	}
	if got := out.String(); strings.Contains(got, "secret") || !strings.Contains(got, "accepted") {
		t.Fatalf("unexpected output: %q", got)
	}
	after, err := term.GetState(int(tty.Fd()))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("terminal state not restored: %v", err)
	}
}

func TestRunRestoresTerminalAfterStartFailure(t *testing.T) {
	outer, tty, err := pty.Open()
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	defer outer.Close()
	defer tty.Close()
	before, err := term.GetState(int(tty.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	if err := Run(exec.Command("/nonexistent/dk-test-command"), tty, io.Discard); err == nil {
		t.Fatal("expected start error")
	}
	after, err := term.GetState(int(tty.Fd()))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("terminal state not restored: %v", err)
	}
}

func TestRunStopsChildAndRestoresTerminalOnOutputFailure(t *testing.T) {
	outer, tty, err := pty.Open()
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	defer outer.Close()
	defer tty.Close()
	before, err := term.GetState(int(tty.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- Run(exec.Command("sh", "-c", "echo output; sleep 60"), tty, brokenOutput{})
	}()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "output unavailable") {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("output failure did not stop child")
	}
	after, err := term.GetState(int(tty.Fd()))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("terminal state not restored: %v", err)
	}
}
