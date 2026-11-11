package shell

import (
	"bytes"
	"io"
	"testing"
)

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	t.Cleanup(func() { SetOutput(nil) })
	outputPrintf("status %s\n", "message")
	if got := buf.String(); got != "status message\n" {
		t.Fatalf("output = %q", got)
	}
	SetOutput(io.Discard)
	outputPrintf("discarded\n")
	if got := buf.String(); got != "status message\n" {
		t.Fatalf("unexpected output = %q", got)
	}
	SetOutput(nil)
	outputState.Lock()
	defer outputState.Unlock()
	if outputState.writer != nil {
		t.Fatal("nil did not restore default output")
	}
}
