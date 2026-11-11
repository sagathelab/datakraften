package ai

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/sagathelab/datakraften/internal/config"
)

func TestSetOutput(t *testing.T) {
	oldOS := currentOS
	currentOS = func() string { return "linux" }
	t.Cleanup(func() { currentOS = oldOS; SetOutput(nil) })
	cfg := &config.Config{AIApps: map[string]config.RuntimeConfig{"codex": {Enabled: true}}}
	var buf bytes.Buffer
	SetOutput(&buf)
	EnsureAIApps(cfg)
	want := "    – Codex Desktop: requires macOS\n"
	if got := buf.String(); got != want {
		t.Fatalf("unexpected output: %q", got)
	}
	SetOutput(io.Discard)
	EnsureAIApps(cfg)
	if buf.String() != want {
		t.Fatal("output was not redirected")
	}
	SetOutput(nil)
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old; r.Close(); w.Close() })
	EnsureAIApps(cfg)
	w.Close()
	got, err := io.ReadAll(r)
	if err != nil || string(got) != want {
		t.Fatalf("default output: %q, %v", got, err)
	}
}
