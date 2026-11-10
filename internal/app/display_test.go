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

func TestDisplayStep(t *testing.T) {
	var out bytes.Buffer
	displayStep(&out, 1, 8, "System packages")
	if out.String() != "  01 / 08  ·  System packages\n" {
		t.Fatalf("unexpected step: %q", out.String())
	}
}
