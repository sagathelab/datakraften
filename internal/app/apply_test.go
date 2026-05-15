package app

import (
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/ai"
)

func TestInstallReportSummaryShowsFailureInsteadOfAlreadyInstalled(t *testing.T) {
	report := ai.InstallReport{
		Enabled: 1,
		Errors:  []string{"GitHub Copilot CLI: gh extension install failed: boom"},
	}

	got := installReportSummary(report)

	if got == "(already installed)" {
		t.Fatalf("expected failure summary, got %q", got)
	}
	if !strings.Contains(got, "failed: GitHub Copilot CLI: gh extension install failed: boom") {
		t.Fatalf("expected detailed failure summary, got %q", got)
	}
}

func TestDefaultConfigUsesGhManagerForCopilot(t *testing.T) {
	if !strings.Contains(defaultConfigYAML, "copilot:\n    enabled: true\n    manager: gh\n    version: latest") {
		t.Fatal("expected default config to declare Copilot with manager: gh")
	}
}
