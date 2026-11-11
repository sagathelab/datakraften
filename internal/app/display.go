package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func displayColorEnabled(w io.Writer) bool {
	if p, ok := w.(*progressOutput); ok {
		return p.renderer.ColorEnabled()
	}
	file, ok := w.(*os.File)
	_, noColor := os.LookupEnv("NO_COLOR")
	return ok && term.IsTerminal(int(file.Fd())) && !noColor && os.Getenv("TERM") != "dumb"
}

func displayEmphasis(w io.Writer, text, color string) {
	if displayColorEnabled(w) {
		fmt.Fprintf(w, "\x1b[%sm%s\x1b[0m\n", color, text)
	} else {
		fmt.Fprintln(w, text)
	}
}

func displayHeading(w io.Writer, title string) {
	displayEmphasis(w, "  "+title, "1")
}

func displayNextSteps(w io.Writer) {
	displayHeading(w, "Next steps")
	fmt.Fprintln(w)
}

func displayComplete(w io.Writer, message string) {
	displayEmphasis(w, "  ✓ "+message, "1;32")
}

func colorStatus(w io.Writer, text string) string {
	return formatStatus(text, displayColorEnabled(w))
}

func formatStatus(text string, color bool) string {
	if !color {
		return text
	}
	for _, item := range []struct{ label, color string }{{"[OK]", "32"}, {"[PLAN]", "36"}, {"[SKIP]", "90"}, {"[WARN]", "33"}, {"[ACTION]", "33"}, {"[ERROR]", "31"}} {
		trimmed := strings.TrimLeft(text, " ")
		if strings.HasPrefix(trimmed, item.label) {
			indent := text[:len(text)-len(trimmed)]
			return indent + "\x1b[" + item.color + "m" + item.label + "\x1b[0m" + strings.TrimPrefix(trimmed, item.label)
		}
	}
	return text
}

func displayPrintf(format string, args ...interface{}) {
	w := progressWriter()
	displayText(w, fmt.Sprintf(format, args...))
}
func displayPrintln(args ...interface{}) {
	w := progressWriter()
	displayText(w, fmt.Sprintln(args...))
}
func displayFprintf(w io.Writer, format string, args ...interface{}) {
	displayText(w, fmt.Sprintf(format, args...))
}
func displayFprintln(w io.Writer, args ...interface{}) {
	displayText(w, fmt.Sprintln(args...))
}

func displayText(w io.Writer, text string) {
	if p, ok := w.(*progressOutput); ok {
		p.writeStatus(text)
		return
	}
	fmt.Fprint(w, colorStatus(w, text))
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
