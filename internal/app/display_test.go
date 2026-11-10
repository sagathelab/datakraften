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

func TestStatusColorsAndPlainOutput(t *testing.T) {
	for _, tc := range []struct{ label, code string }{{"OK", "32"}, {"PLAN", "36"}, {"SKIP", "90"}, {"WARN", "33"}, {"ERROR", "31"}} {
		plain := "    [" + tc.label + "] message\n"
		if got := formatStatus(plain, false); got != plain {
			t.Fatalf("plain output changed: %q", got)
		}
		if got := formatStatus(plain, true); !strings.Contains(got, "\x1b["+tc.code+"m● ["+tc.label+"]\x1b[0m") {
			t.Fatalf("missing color: %q", got)
		}
	}
	if got := formatStatus("message [ERROR] is data", true); got != "message [ERROR] is data" {
		t.Fatalf("colored message body: %q", got)
	}
}
