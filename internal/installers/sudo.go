package installers

import (
	"sync"

	"github.com/sagathelab/datakraften/internal/exec"
)

var sudoState struct {
	sync.RWMutex
	run func(...string) exec.Result
}

// SetSudoRunner sets the package-wide sudo runner. A nil runner restores the
// default captured command execution. Calls already in progress are unaffected.
func SetSudoRunner(run func(...string) exec.Result) {
	sudoState.Lock()
	defer sudoState.Unlock()
	sudoState.run = run
}

func runInstallerCommand(name string, args ...string) exec.Result {
	if name != "sudo" {
		return exec.Run(name, args...)
	}
	sudoState.RLock()
	run := sudoState.run
	sudoState.RUnlock()
	if run != nil {
		return run(args...)
	}
	return exec.Run(name, args...)
}
