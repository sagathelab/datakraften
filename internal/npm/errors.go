package npm

import (
	"fmt"
	"strings"
)

func CommandError(operation, stderr string) error {
	detail := strings.TrimSpace(stderr)
	if strings.Contains(detail, "EACCES") || strings.Contains(detail, "EPERM") {
		return fmt.Errorf("%s failed due to insufficient permissions; use a Node version manager or run 'dk npm setup' to configure a user-writable npm prefix (https://docs.npmjs.com/resolving-eacces-permissions-errors-when-installing-packages-globally). Existing global packages may need to be reinstalled in the new location. Details: %s", operation, detail)
	}
	return fmt.Errorf("%s failed: %s", operation, detail)
}
