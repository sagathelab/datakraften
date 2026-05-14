package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Result struct {
	Stdout string
	Stderr string
	Code   int
}

func Run(name string, args ...string) Result {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}

	return Result{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
		Code:   code,
	}
}

func RunWithInput(input string, name string, args ...string) Result {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = -1
		}
	}

	return Result{
		Stdout: strings.TrimSpace(stdout.String()),
		Stderr: strings.TrimSpace(stderr.String()),
		Code:   code,
	}
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
