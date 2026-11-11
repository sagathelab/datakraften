package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sagathelab/datakraften/internal/installers"
	"strings"
	"testing"
	"time"
)

type failingProgressWriter struct{}

func (failingProgressWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func TestProgressUpdateSequence(t *testing.T) {
	var out bytes.Buffer
	output := newProgressOutput(&out)
	failure := errors.New("failed to update")
	err := runUpdateStep(updateStep{name: "example", run: func(bool) error {
		_, err := output.renderer.Write([]byte("log line\n"))
		if err != nil {
			t.Fatal(err)
		}
		return failure
	}}, false, output.event)
	if !errors.Is(finishProgress(output, err), failure) {
		t.Fatal("lost step error")
	}
	got := out.String()
	if got != "[...] Updating example\nlog line\n[FAIL] Updating example\n" {
		t.Fatalf("unexpected output: %q", got)
	}
	if strings.ContainsAny(got, "\x1b\r") {
		t.Fatalf("terminal controls in output: %q", got)
	}
}

func TestProgressWriteFailurePreservesRunError(t *testing.T) {
	output := newProgressOutput(failingProgressWriter{})
	output.event(ProgressEvent{Status: ProgressStarted, Message: "starting"})
	failure := errors.New("operation failed")
	err := finishProgress(output, failure)
	if !errors.Is(err, failure) || !strings.Contains(err.Error(), "output unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProgressRoutesPackageOutputAndRestores(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	var outer, inner bytes.Buffer
	first := newProgressOutput(&outer)
	restoreFirst := first.activate()
	defer restoreFirst()
	first.event(ProgressEvent{Status: ProgressStarted, Message: "starting"})
	if err := installers.APTUpdate(); err != nil {
		t.Fatal(err)
	}
	second := newProgressOutput(&inner)
	restoreSecond := second.activate()
	if err := installers.APTUpdate(); err != nil {
		t.Fatal(err)
	}
	restoreSecond()
	if err := installers.APTUpdate(); err != nil {
		t.Fatal(err)
	}
	if strings.Count(outer.String(), "Updating APT cache") != 2 || strings.Count(inner.String(), "Updating APT cache") != 1 {
		t.Fatalf("outer = %q, inner = %q", outer.String(), inner.String())
	}
	if err := finishProgress(first, nil); err != nil {
		t.Fatal(err)
	}
}

func TestProgressLogWriteFailure(t *testing.T) {
	output := newProgressOutput(failingProgressWriter{})
	restore := output.activate()
	defer restore()
	displayPrintln("log")
	if err := finishProgress(output, nil); err == nil || !strings.Contains(err.Error(), "output unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProgressHandoffPropagatesError(t *testing.T) {
	var out bytes.Buffer
	output := newProgressOutput(&out)
	restore := output.activate()
	defer restore()
	output.event(ProgressEvent{Status: ProgressStarted, Message: "starting"})
	want := errors.New("prompt failed")
	err := progressHandoff(func() error { out.WriteString("prompt\n"); return want })
	if !errors.Is(err, want) || !strings.Contains(out.String(), "prompt\n") {
		t.Fatalf("error = %v, output = %q", err, out.String())
	}
}

func TestActiveProgressConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = progressWriter()
				_ = progressHandoff(func() error { return nil })
			}
		}()
	}
	for i := 0; i < 100; i++ {
		p := newProgressOutput(io.Discard)
		restore := p.activate()
		restore()
	}
	wg.Wait()
}

func TestHandoffAllowsWaitingWriterToResume(t *testing.T) {
	var out bytes.Buffer
	p := newProgressOutput(&out)
	restore := p.activate()
	defer restore()
	entered, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(finished)
		_ = progressHandoff(func() error { close(entered); <-release; return nil })
	}()
	<-entered
	written := make(chan struct{})
	go func() { defer close(written); _, _ = p.Write([]byte("after prompt\n")) }()
	close(release)
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("handoff deadlocked")
	}
	select {
	case <-written:
	case <-time.After(time.Second):
		t.Fatal("writer deadlocked")
	}
	if out.String() != "after prompt\n" {
		t.Fatalf("output = %q", out.String())
	}
}
