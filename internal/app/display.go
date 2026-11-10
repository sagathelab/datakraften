package app

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func displayTitle(w io.Writer, title, subtitle string) {
	name := "Datakraften"
	if file, ok := w.(*os.File); ok && term.IsTerminal(int(file.Fd())) && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" {
		name = "\x1b[1;32m" + name + "\x1b[0m"
	}
	fmt.Fprintf(w, "\n  %s / %s\n", name, title)
	if subtitle != "" {
		fmt.Fprintf(w, "  %s\n", subtitle)
	}
	fmt.Fprintln(w)
}

func displayStep(w io.Writer, current, total int, label string) {
	fmt.Fprintf(w, "  %02d / %02d  ·  %s\n", current, total, label)
}
