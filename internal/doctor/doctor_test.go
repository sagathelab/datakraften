package doctor

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintSummarySeparatesSeverity(t *testing.T) {
	r := &Report{Checks: []Check{
		{Status: "pass", Severity: "warning"},
		{Status: "fail", Severity: "critical"},
		{Status: "fail", Severity: "warning"},
		{Status: "fail", Severity: "info"},
	}}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	r.Print()
	os.Stdout = original
	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(read)
	if err != nil {
		t.Fatal(err)
	}
	if err := read.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "1 passed  ·  1 critical  ·  1 warnings  ·  1 informational") {
		t.Fatalf("unexpected summary: %q", out)
	}
}
