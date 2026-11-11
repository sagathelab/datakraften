package installers

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() { SetOutput(nil) })
	t.Setenv("PATH", t.TempDir())
	_ = APTUpdate()
	if got := buf.String(); got != "    Updating APT cache...\n" {
		t.Fatalf("unexpected output: %q", got)
	}
	SetOutput(io.Discard)
	_ = APTUpdate()
	if strings.Count(buf.String(), "Updating") != 1 {
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
	_ = APTUpdate()
	w.Close()
	got, err := io.ReadAll(r)
	if err != nil || string(got) != "    Updating APT cache...\n" {
		t.Fatalf("default output: %q, %v", got, err)
	}
}
