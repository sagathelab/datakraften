package npm

import (
	"strings"
	"testing"
)

func TestCommandError(t *testing.T) {
	for _, tc := range []struct {
		stderr         string
		wantPermission bool
	}{
		{"npm error code EACCES", true},
		{"npm error code EPERM", true},
		{"npm error code ENETUNREACH", false},
	} {
		t.Run(tc.stderr, func(t *testing.T) {
			message := CommandError("npm install -g", tc.stderr).Error()
			if !strings.Contains(message, tc.stderr) || strings.Contains(message, "user-writable npm prefix") != tc.wantPermission {
				t.Fatalf("unexpected error: %s", message)
			}
		})
	}
}
