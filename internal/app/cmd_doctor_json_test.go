package app

import (
	"encoding/json"
	"testing"
)

func TestDoctorJSONIsMachineReadable(t *testing.T) {
	old := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = old })
	output := captureStdout(t, func() {
		cmd := newDoctorCmd()
		cmd.SetArgs([]string{"--category", "runtimes"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("invalid JSON: %v; output: %q", err, output)
	}
	if _, ok := data["checks"]; !ok {
		t.Fatalf("missing checks: %q", output)
	}
}
