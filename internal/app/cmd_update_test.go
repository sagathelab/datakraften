package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestUpdateAllReturnsErrorWhenAChildUpdateFails(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{
		{name: "brew", run: func(bool) error { return nil }},
		{name: "npm", run: func(bool) error { return errors.New("permission denied") }},
	}
	defer func() {
		updateSteps = originalSteps
	}()

	output := captureStdout(t, func() {
		err := updateAll(false)
		if err == nil {
			t.Fatal("expected updateAll to return an error")
		}
		if !strings.Contains(err.Error(), "npm: permission denied") {
			t.Fatalf("expected npm failure in returned error, got %q", err.Error())
		}
	})

	if strings.Contains(output, "✓ Update complete") {
		t.Fatalf("expected success summary to be omitted, got %q", output)
	}
	if !strings.Contains(output, "✗ Update finished with errors") {
		t.Fatalf("expected failure summary, got %q", output)
	}
	if !strings.Contains(output, "✗ npm: permission denied") {
		t.Fatalf("expected npm failure output, got %q", output)
	}
}

func TestUpdateAllPrintsSuccessWhenNoFailuresOccur(t *testing.T) {
	originalSteps := updateSteps
	updateSteps = []updateStep{
		{name: "brew", run: func(bool) error { return nil }},
		{name: "uv", run: func(bool) error { return nil }},
	}
	defer func() {
		updateSteps = originalSteps
	}()

	output := captureStdout(t, func() {
		if err := updateAll(false); err != nil {
			t.Fatalf("expected updateAll to succeed, got %v", err)
		}
	})

	if !strings.Contains(output, "✓ Update complete") {
		t.Fatalf("expected success summary, got %q", output)
	}
	if strings.Contains(output, "✗ Update finished with errors") {
		t.Fatalf("did not expect failure summary, got %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
	}()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	return <-done
}
