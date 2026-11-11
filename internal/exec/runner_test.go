package exec

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("DK_EXEC_HELPER") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, "prompt\n")
	var input string
	fmt.Fscanln(os.Stdin, &input)
	fmt.Fprintln(os.Stdout, input)
	fmt.Fprintln(os.Stderr, "diagnostic")
	os.Exit(7)
}

func TestRunWithOptionsStreamsAndCaptures(t *testing.T) {
	t.Setenv("DK_EXEC_HELPER", "1")
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	prompt := make(chan struct{}, 1)
	var stdout, stderr bytes.Buffer
	out := &signalWriter{buffer: &stdout, signal: prompt}
	result := make(chan Result, 1)
	go func() {
		result <- RunWithOptions(Options{Stdin: reader, Stdout: out, Stderr: &stderr, CaptureStdout: true, CaptureStderr: true}, os.Args[0], "-test.run=^TestHelperProcess$")
	}()
	select {
	case <-prompt:
	case <-time.After(5 * time.Second):
		writer.Close()
		<-result
		t.Fatal("output was not streamed before input")
	}
	fmt.Fprintln(writer, "answer")
	writer.Close()
	r := <-result
	if r.Code != 7 || r.Stdout != "prompt\nanswer" || r.Stderr != "diagnostic" {
		t.Fatalf("unexpected result: %+v", r)
	}
	if stdout.String() != "prompt\nanswer\n" || stderr.String() != "diagnostic\n" {
		t.Fatalf("unexpected streams: %q, %q", stdout.String(), stderr.String())
	}
}

type signalWriter struct {
	buffer *bytes.Buffer
	signal chan struct{}
}

func (w *signalWriter) Write(p []byte) (int, error) {
	n, err := w.buffer.Write(p)
	select {
	case w.signal <- struct{}{}:
	default:
	}
	return n, err
}

func TestCaptureAndStartFailure(t *testing.T) {
	r := RunWithInput(" hello \n", "cat")
	if r.Code != 0 || r.Stdout != "hello" || r.Stderr != "" {
		t.Fatalf("unexpected captured result: %+v", r)
	}
	r = Run("dk-command-that-does-not-exist")
	if r.Code != -1 || strings.TrimSpace(r.Stdout+r.Stderr) != "" {
		t.Fatalf("unexpected start failure: %+v", r)
	}
}
