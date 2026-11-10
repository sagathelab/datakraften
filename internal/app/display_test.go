package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisplayTitle(t *testing.T) {
	var out bytes.Buffer
	displayTitle(&out, "status", "Environment overview")
	if out.String() != "\n  Datakraften / status\n  Environment overview\n\n" {
		t.Fatalf("unexpected title: %q", out.String())
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Fatal("unexpected terminal escape codes")
	}
}
