package terminal

import (
	"os"

	"golang.org/x/sys/unix"
)

func suspendKey(f *os.File) (byte, bool) {
	t, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	if err != nil || t.Lflag&unix.ISIG == 0 {
		return 0, false
	}
	return t.Cc[unix.VSUSP], t.Cc[unix.VSUSP] != 0
}
