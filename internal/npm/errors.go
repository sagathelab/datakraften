package npm

import (
	"fmt"
	"strings"
)

func IsPermissionError(stderr string) bool {
	return strings.Contains(stderr, "EACCES") || strings.Contains(stderr, "EPERM")
}

func CommandError(operation, stderr string) error {
	detail := strings.TrimSpace(stderr)
	if IsPermissionError(detail) {
		code := "EACCES"
		if !strings.Contains(detail, code) {
			code = "EPERM"
		}
		return fmt.Errorf("%s failed (%s): npm prefix is not writable; run 'dk npm setup' for a user-writable npm prefix or use a Node version manager", operation, code)
	}
	for _, line := range strings.Split(detail, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return fmt.Errorf("%s failed: %s", operation, line)
		}
	}
	return fmt.Errorf("%s failed", operation)
}
