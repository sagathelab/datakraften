package editors

import (
	"fmt"

	"github.com/sagathelab/datakraften/internal/exec"
)

func EnsureZed() error {
	if exec.CommandExists("zed") {
		return nil
	}

	fmt.Println("    Installing Zed via install script...")
	r := exec.Run("bash", "-c", "curl -fsSL https://zed.dev/install.sh | sh")
	if r.Code != 0 {
		return fmt.Errorf("zed install failed: %s", r.Stderr)
	}
	return nil
}
