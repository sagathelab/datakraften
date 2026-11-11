package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func displayColorEnabled(w io.Writer) bool {
	file, ok := w.(*os.File)
	_, noColor := os.LookupEnv("NO_COLOR")
	return ok && term.IsTerminal(int(file.Fd())) && !noColor && os.Getenv("TERM") != "dumb"
}

func colorStatus(w io.Writer, text string) string {
	return formatStatus(text, displayColorEnabled(w))
}

func formatStatus(text string, color bool) string {
	if !color {
		return text
	}
	for _, item := range []struct{ label, color string }{{"[OK]", "32"}, {"[PLAN]", "36"}, {"[SKIP]", "90"}, {"[WARN]", "33"}, {"[ERROR]", "31"}} {
		trimmed := strings.TrimLeft(text, " ")
		if strings.HasPrefix(trimmed, item.label) {
			indent := text[:len(text)-len(trimmed)]
			return indent + "\x1b[" + item.color + "m● " + item.label + "\x1b[0m" + strings.TrimPrefix(trimmed, item.label)
		}
	}
	return text
}

func displayPrintf(format string, args ...interface{}) {
	w := progressWriter()
	fmt.Fprint(w, colorStatus(w, fmt.Sprintf(format, args...)))
}
func displayPrintln(args ...interface{}) {
	w := progressWriter()
	fmt.Fprint(w, colorStatus(w, fmt.Sprintln(args...)))
}
func displayFprintf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprint(w, colorStatus(w, fmt.Sprintf(format, args...)))
}
func displayFprintln(w io.Writer, args ...interface{}) {
	fmt.Fprint(w, colorStatus(w, fmt.Sprintln(args...)))
}

func displayTitle(w io.Writer, title, subtitle string) {
	name := "Datakraften"
	if displayColorEnabled(w) {
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
