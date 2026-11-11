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
			if !strings.Contains(message, strings.TrimPrefix(tc.stderr, "npm error code ")) || strings.Contains(message, "user-writable npm prefix") != tc.wantPermission {
				t.Fatalf("unexpected error: %s", message)
			}
		})
	}
}

func TestCommandErrorDoesNotRepeatNpmStackTrace(t *testing.T) {
	message := CommandError("npm install -g", "npm error code EACCES\nnpm error path /usr/local/lib/node_modules\nnpm error stack trace").Error()
	if strings.Contains(message, "\n") || strings.Contains(message, "stack trace") || !strings.Contains(message, "dk npm setup") {
		t.Fatalf("unexpected error: %q", message)
	}
}
