// Package progress serializes terminal progress, logs, and interactive prompts.
package progress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/term"
)

type Renderer struct {
	mu                           sync.Mutex
	out                          io.Writer
	color, partial               bool
	terminalHeadings             bool
	err                          error
	size                         func() (int, int, error)
	known, active                bool
	width, height, distance, col int
	wrap                         bool
}

// New creates a conservative append-only renderer.
func New(out io.Writer) *Renderer {
	r := &Renderer{out: out}
	f, ok := out.(interface{ Fd() uintptr })
	_, noColor := os.LookupEnv("NO_COLOR")
	if ok && term.IsTerminal(int(f.Fd())) && !noColor && os.Getenv("TERM") != "dumb" {
		r.color = true
	}
	return r
}

// NewAtLineStart enables anchoring when out is a suitable terminal. The caller
// must guarantee that the cursor is at the start of an empty line and all terminal output is
// serialized through this renderer or Handoff, including output on other file
// descriptors referring to the same terminal.
func NewAtLineStart(out io.Writer) *Renderer {
	r := New(out)
	_, noColor := os.LookupEnv("NO_COLOR")
	if f, ok := out.(interface{ Fd() uintptr }); ok && !noColor && term.IsTerminal(int(f.Fd())) && os.Getenv("TERM") != "dumb" {
		r.size = func() (int, int, error) { return term.GetSize(int(f.Fd())) }
		r.known = true
		r.terminalHeadings = true
	}
	return r
}
func (r *Renderer) Running(message string) error { return r.event("...", "90", message, true) }
func (r *Renderer) Done(message string) error    { return r.event("OK", "32", message, false) }
func (r *Renderer) Fail(message string) error    { return r.event("FAIL", "31", message, false) }
func clean(s string) string {
	return strings.Map(func(c rune) rune {
		if unicode.IsControl(c) || c == '\u2028' || c == '\u2029' {
			return ' '
		}
		return c
	}, s)
}
func (r *Renderer) write(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	n, err := r.out.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	if err != nil {
		r.err = err
		r.active = false
		r.known = false
	}
	return n, err
}
func (r *Renderer) endPartial() error {
	if !r.partial {
		return nil
	}
	r.known = false
	_, err := r.write([]byte("\n"))
	if err == nil {
		r.partial = false
	}
	return err
}
func (r *Renderer) dimensions() (int, int, bool) {
	if r.size == nil {
		return 0, 0, false
	}
	w, h, e := r.size()
	return w, h, e == nil && w >= 4 && h >= 2
}
func (r *Renderer) valid() bool {
	if !r.active {
		return false
	}
	w, h, ok := r.dimensions()
	if !ok || w != r.width || h != r.height {
		r.active = false
		r.known = false
		return false
	}
	return true
}
func (r *Renderer) formatAnchor(message, color string, width int) string {
	message = strings.Map(func(c rune) rune {
		if c < 32 || c > 126 {
			return '?'
		}
		return c
	}, message)
	if len(message) > width-4 {
		message = message[:width-4]
	}
	dot := "●"
	if r.color {
		dot = "\x1b[" + color + "m" + dot + "\x1b[0m"
	}
	return dot + " " + message
}
func (r *Renderer) event(status, color, message string, running bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	line := fmt.Sprintf("[%s] %s", status, clean(message))
	if r.valid() {
		if !r.wrap {
			seq := fmt.Sprintf("\r\x1b[%dA\r\x1b[2K%s\r\x1b[%dB\r", r.distance, r.formatAnchor(clean(message), color, r.width), r.distance)
			if r.col > 0 {
				seq += fmt.Sprintf("\x1b[%dC", r.col)
			}
			if _, err := r.write([]byte(seq)); err != nil {
				return err
			}
			if !running {
				r.active = false
			}
			return nil
		}
		r.active = false
		r.known = false
	}
	if err := r.endPartial(); err != nil {
		return err
	}
	if r.terminalHeadings && !r.known {
		if running {
			_, err := r.write([]byte(r.formatAnchor(clean(message), color, 10000) + "\n"))
			return err
		}
		if status == "OK" {
			return nil
		}
	}
	if running && r.known {
		w, h, ok := r.dimensions()
		if ok {
			if _, err := r.write([]byte(r.formatAnchor(clean(message), color, w) + "\r\n")); err != nil {
				return err
			}
			r.width, r.height, r.distance, r.col = w, h, 1, 0
			r.wrap = false
			r.active = true
			return nil
		}
		r.known = false
	}
	_, err := r.write([]byte(line + "\n"))
	return err
}

// Write serializes logs with progress. Anchored ASCII newlines are emitted as
// CRLF so cursor tracking does not depend on the terminal's output mode.
func (r *Renderer) Write(p []byte) (int, error) {
	return r.writeLog(p, nil)
}

func (r *Renderer) WriteFormatted(p []byte, format func([]byte) []byte) (int, error) {
	return r.writeLog(p, format)
}

func (r *Renderer) ColorEnabled() bool { return r.color }

func (r *Renderer) writeLog(p []byte, format func([]byte) []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	anchored := r.valid()
	if anchored {
		for _, c := range p {
			if c != '\n' && (c < 32 || c > 126) {
				r.active = false
				r.known = false
				anchored = false
				break
			}
		}
	}
	if anchored {
		for _, c := range p {
			if c == '\n' {
				r.distance++
				r.col = 0
				r.wrap = false
			} else {
				if r.wrap {
					r.distance++
					r.col = 0
					r.wrap = false
				}
				if r.col == r.width-1 {
					r.wrap = true
				} else {
					r.col++
				}
			}
			if r.distance >= r.height {
				r.active = false
				r.known = false
				break
			}
		}
	} else if len(p) > 0 {
		r.known = false
	}
	output := p
	if format != nil {
		output = format(p)
	}
	if anchored {
		output = bytes.ReplaceAll(output, []byte("\n"), []byte("\r\n"))
	}
	n, err := r.write(output)
	if format != nil {
		if err == nil {
			n = len(p)
		} else {
			n = 0
		}
	} else if anchored {
		written, consumed := 0, 0
		for consumed < len(p) {
			size := 1
			if p[consumed] == '\n' {
				size = 2
			}
			if written+size > n {
				break
			}
			written += size
			consumed++
		}
		n = consumed
	}
	if n > 0 {
		r.partial = p[n-1] != '\n'
	}
	return n, err
}

// Handoff suspends progress while fn owns the terminal. The callback must not call the renderer.
func (r *Renderer) Handoff(fn func() error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.active = false
	r.known = false
	if err := r.endPartial(); err != nil {
		return err
	}
	return fn()
}

// Close ends progress without closing the underlying writer.
func (r *Renderer) Close() error { r.mu.Lock(); defer r.mu.Unlock(); r.active = false; return r.err }
