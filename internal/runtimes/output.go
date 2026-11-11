package runtimes

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var outputState struct {
	sync.Mutex
	writer io.Writer
}

// SetOutput sets the destination for this package's status messages. Passing nil
// restores the default destination, os.Stdout. It does not redirect subprocess output.
func SetOutput(w io.Writer) {
	outputState.Lock()
	defer outputState.Unlock()
	outputState.writer = w
}

func outputPrintf(format string, args ...any) {
	outputState.Lock()
	defer outputState.Unlock()
	w := outputState.writer
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, format, args...)
}
