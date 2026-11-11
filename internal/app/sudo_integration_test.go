package app

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sagathelab/datakraften/internal/installers"
)

func TestSudoProgressHelper(t *testing.T) {
	if os.Getenv("DK_SUDO_HELPER") != "1" {
		return
	}
	p := newProgressOutput(os.Stdout)
	restore := p.activate()
	defer restore()
	p.event(ProgressEvent{Status: ProgressStarted, Message: "sudo step"})
	err := installers.APTUpdate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	p.event(ProgressEvent{Status: ProgressCompleted, Step: "sudo step"})
	if err := finishProgress(p, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
}

func TestSudoNonInteractivePipe(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "args")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$DK_ARGS\"\n[ \"$1\" = -n ] || exit 9\n"
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := osexec.CommandContext(ctx, os.Args[0], "-test.run=^TestSudoProgressHelper$")
	cmd.Env = append(os.Environ(), "DK_SUDO_HELPER=1", "DK_ARGS="+marker, "PATH="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper: %v: %s", err, out)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "-n\napt-get\nupdate\n-qq\n" {
		t.Fatalf("sudo args: %q", data)
	}
	if strings.ContainsAny(string(out), "\x1b\r") {
		t.Fatalf("terminal controls in pipe: %q", out)
	}
}

func TestSudoInteractivePTY(t *testing.T) {
	python, err := osexec.LookPath("python3")
	if err != nil {
		t.Skip("python3 unavailable")
	}
	dir := t.TempDir()
	script := "#!/bin/sh\n[ \"$1\" != -n ] || exit 9\nprintf 'CHILD STDOUT\\n'\nprintf 'PASSWORD PROMPT\\n' >&2\nread answer\n[ \"$answer\" = yes ] || exit 8\n"
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	py := `import os,pty,fcntl,termios,struct,select,sys
fd,slave=pty.openpty()
fcntl.ioctl(slave,termios.TIOCSWINSZ,struct.pack('HHHH',24,80,0,0))
pid=os.fork()
if pid==0:
 os.close(fd)
 os.login_tty(slave)
 os.environ['DK_SUDO_HELPER']='1'
 os.environ['PATH']=sys.argv[2]
 os.environ['TERM']='xterm'
 os.environ.pop('NO_COLOR',None)
 os.execv(sys.argv[1],[sys.argv[1],'-test.run=^TestSudoProgressHelper$'])
os.close(slave)
data=b''
while True:
 try:
  chunk=os.read(fd,4096)
 except OSError: break
 if not chunk: break
 data+=chunk
 if b'PASSWORD PROMPT' in data and b'SENT ANSWER' not in data:
  os.write(fd,b'yes\n'); data+=b'SENT ANSWER'
_,status=os.waitpid(pid,0)
sys.stdout.buffer.write(data)
sys.exit(os.waitstatus_to_exitcode(status))`
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := osexec.CommandContext(ctx, python, "-c", py, os.Args[0], dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PTY helper: %v: %q", err, out)
	}
	got := string(out)
	if !strings.Contains(got, "CHILD STDOUT") || !strings.Contains(got, "PASSWORD PROMPT") || !strings.Contains(got, "[OK] sudo step") {
		t.Fatalf("unexpected PTY output: %q", got)
	}
	if strings.Count(got, "[...] sudo step") != 1 {
		t.Fatalf("duplicate status: %q", got)
	}
	if strings.Count(got, "PASSWORD PROMPT") != 1 {
		t.Fatalf("duplicate prompt: %q", got)
	}

	failureScript := "#!/bin/sh\nprintf 'PASSWORD PROMPT\\n' >&2\nread answer\n[ \"$answer\" = yes ] || exit 8\nprintf 'permission denied\\n' >&2\nexit 42\n"
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte(failureScript), 0755); err != nil {
		t.Fatal(err)
	}
	failure := osexec.CommandContext(ctx, python, "-c", py, os.Args[0], dir)
	out, err = failure.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "permission denied") || !strings.Contains(string(out), "sudo exited with status 42") {
		t.Fatalf("interactive failure: %v: %q", err, out)
	}
}

func TestSudoPipeFailurePreservesError(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nprintf 'authentication required\\n' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := osexec.CommandContext(ctx, os.Args[0], "-test.run=^TestSudoProgressHelper$")
	cmd.Env = append(os.Environ(), "DK_SUDO_HELPER=1", "PATH="+dir)
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "apt update failed: authentication required") {
		t.Fatalf("error = %v, output = %q", err, out)
	}
}
