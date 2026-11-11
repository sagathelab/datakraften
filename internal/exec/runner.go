package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Result struct {
	Stdout string
	Stderr string
	Code   int
}

type Options struct {
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	CaptureStdout bool
	CaptureStderr bool
}

func RunWithOptions(opts Options, name string, args ...string) Result {
	cmd := exec.Command(name, args...)
	cmd.Stdin = opts.Stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	if opts.CaptureStdout {
		cmd.Stdout = &stdout
		if opts.Stdout != nil {
			cmd.Stdout = io.MultiWriter(opts.Stdout, &stdout)
		}
	}
	if opts.CaptureStderr {
		cmd.Stderr = &stderr
		if opts.Stderr != nil {
			cmd.Stderr = io.MultiWriter(opts.Stderr, &stderr)
		}
	}

	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}
	return Result{Stdout: strings.TrimSpace(stdout.String()), Stderr: strings.TrimSpace(stderr.String()), Code: code}
}

func Run(name string, args ...string) Result {
	return RunWithOptions(Options{CaptureStdout: true, CaptureStderr: true}, name, args...)
}

func RunWithInput(input string, name string, args ...string) Result {
	return RunWithOptions(Options{Stdin: strings.NewReader(input), CaptureStdout: true, CaptureStderr: true}, name, args...)
}

func RunVerbose(name string, args ...string) Result {
	fmt.Fprintf(os.Stderr, "  running: %s %s\n", name, strings.Join(args, " "))
	return Run(name, args...)
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func CommandPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func IsWindowsBacked(path string) bool {
	return strings.HasPrefix(path, "/mnt/")
}
