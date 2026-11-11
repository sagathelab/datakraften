package app

import (
	"strings"
	"testing"

	"github.com/sagathelab/datakraften/internal/ai"
	"github.com/sagathelab/datakraften/internal/config"
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

func TestInstallReportSummaryDoesNotAssumeUnreportedResults(t *testing.T) {
	got := installReportSummary(ai.InstallReport{Enabled: 1})
	if strings.Contains(got, "already installed") {
		t.Fatalf("unreported result presented as installed: %q", got)
	}
}

func TestInstallSummaryOmitsFailureDetails(t *testing.T) {
	report := ai.InstallReport{Enabled: 1, Errors: []string{"Claude Code: npm failed: EACCES"}}
	got := installSummaryWithoutDetails(report)
	if got != "failed: Claude Code" || report.Errors[0] != "Claude Code: npm failed: EACCES" {
		t.Fatalf("summary %q report %+v", got, report)
	}
}

func TestApplyProgressConfigurationFailure(t *testing.T) {
	var events []ProgressEvent
	report := RunApplyWithProgress(&config.Config{Source: "team"}, false, func(event ProgressEvent) {
		events = append(events, event)
	})
	if len(report.Errors) != 1 || len(events) != 2 || events[0].Status != ProgressStarted || events[1].Status != ProgressFailed || events[1].Err == nil {
		t.Fatalf("unexpected report or events: %+v, %+v", report, events)
	}
}
