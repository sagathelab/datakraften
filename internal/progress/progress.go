// Package progress serializes terminal progress, logs, and interactive prompts.
package progress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

type Renderer struct {
	mu                           sync.Mutex
	out                          io.Writer
	color, partial               bool
	terminalHeadings             bool
	handoffHeading               bool
	message                      string
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
	if running {
		r.message = clean(message)
		r.handoffHeading = false
	} else if r.handoffHeading {
		r.handoffHeading = false
		if status == "OK" {
			return nil
		}
		if err := r.endPartial(); err != nil {
			return err
		}
		_, err := r.write([]byte(fmt.Sprintf("[FAIL] %s\n", clean(message))))
		return err
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
		if _, err := r.write([]byte(r.formatAnchor(clean(message), color, 10000) + "\n")); err != nil {
			return err
		}
		r.known = true
		return nil
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

type terminalWriter struct {
	renderer *Renderer
	pending  []byte
}

func (w *terminalWriter) Write(p []byte) (int, error) {
	w.pending = append(w.pending, p...)
	end := completeTerminalPrefix(w.pending)
	if end > 0 {
		if _, err := w.renderer.writeTerminal(w.pending[:end]); err != nil {
			return 0, err
		}
		w.pending = append(w.pending[:0], w.pending[end:]...)
	}
	if len(w.pending) > 4096 {
		if err := w.Flush(); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (w *terminalWriter) Flush() error {
	if len(w.pending) == 0 {
		return nil
	}
	_, err := w.renderer.writeTerminal(w.pending)
	w.pending = nil
	return err
}

func completeTerminalPrefix(p []byte) int {
	for i := 0; i < len(p); {
		start := i
		if p[i] == '\x1b' {
			if i+1 == len(p) {
				return start
			}
			if p[i+1] == ']' {
				end, _, complete := terminalOSC(p, i+2)
				if !complete {
					return start
				}
				i = end
				continue
			}
			if p[i+1] == '[' {
				i += 2
				for i < len(p) && (p[i] < 0x40 || p[i] > 0x7e) {
					i++
				}
				if i == len(p) {
					return start
				}
				i++
				continue
			}
			i += 2
			continue
		}
		if !utf8.FullRune(p[i:]) {
			return start
		}
		_, size := utf8.DecodeRune(p[i:])
		i += size
	}
	return len(p)
}

func terminalOSC(p []byte, start int) (int, bool, bool) {
	valid := true
	for i := start; i < len(p); i++ {
		if p[i] == '\a' || (p[i] == '\x1b' && i+1 < len(p) && p[i+1] == '\\') {
			end := i + 1
			if p[i] == '\x1b' {
				end++
			}
			command, _, hasParams := strings.Cut(string(p[start:i]), ";")
			switch command {
			case "0", "1", "2", "4", "7", "8", "9", "10", "11", "12", "17", "19", "52", "104", "110", "111", "112", "133", "134":
				return end, valid && hasParams, true
			default:
				return end, false, true
			}
		}
		if p[i] == '\x1b' && i+1 == len(p) {
			return 0, false, false
		}
		if p[i] < 0x20 || p[i] == 0x7f {
			valid = false
		}
	}
	return 0, false, false
}

func (r *Renderer) TerminalWriter() io.Writer { return &terminalWriter{renderer: r} }

func (r *Renderer) writeTerminal(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	wasActive := r.active
	if r.valid() {
		distance, col, wrap, ok := r.trackTerminal(p)
		if ok {
			r.distance, r.col, r.wrap = distance, col, wrap
		} else {
			if r.terminalHeadings && !r.wrap {
				seq := fmt.Sprintf("\r\x1b[%dA\r\x1b[2K%s\r\x1b[%dB\r", r.distance, r.formatAnchor(r.message, "37", r.width), r.distance)
				if r.col > 0 {
					seq += fmt.Sprintf("\x1b[%dC", r.col)
				}
				if _, err := r.write([]byte(seq)); err != nil {
					return 0, err
				}
				r.handoffHeading = true
			}
			r.active = false
			r.known = false
		}
	} else {
		if wasActive && r.terminalHeadings {
			r.handoffHeading = true
		}
		r.known = false
	}
	n, err := r.write(p)
	if n > 0 {
		r.partial = p[n-1] != '\n'
	}
	return n, err
}

func (r *Renderer) trackTerminal(p []byte) (int, int, bool, bool) {
	distance, col, wrap := r.distance, r.col, r.wrap
	for i := 0; i < len(p); {
		c, size := utf8.DecodeRune(p[i:])
		if c == utf8.RuneError && size == 1 {
			return 0, 0, false, false
		}
		i += size
		switch {
		case c == '\x1b':
			if i < len(p) && p[i] == ']' {
				end, valid, complete := terminalOSC(p, i+1)
				if !complete || !valid {
					return 0, 0, false, false
				}
				i = end
				break
			}
			if i >= len(p) || p[i] != '[' {
				return 0, 0, false, false
			}
			i++
			start := i
			for i < len(p) && (p[i] < 0x40 || p[i] > 0x7e) {
				i++
			}
			if i >= len(p) {
				return 0, 0, false, false
			}
			params, action := string(p[start:i]), p[i]
			i++
			switch action {
			case 'm':
				for _, c := range params {
					if (c < '0' || c > '9') && c != ';' && c != ':' {
						return 0, 0, false, false
					}
				}
			case 'G':
				column := 1
				if params != "" {
					var err error
					column, err = strconv.Atoi(params)
					if err != nil || column < 1 || column > r.width {
						return 0, 0, false, false
					}
				}
				col, wrap = column-1, false
			case 'K':
				if params != "" && params != "0" && params != "1" && params != "2" {
					return 0, 0, false, false
				}
			default:
				return 0, 0, false, false
			}
		case c == '\r':
			col, wrap = 0, false
		case c == '\n':
			if col != 0 {
				return 0, 0, false, false
			}
			distance++
			wrap = false
		case c == '\b':
			if col > 0 {
				col--
			}
			wrap = false
		case (c >= 32 && c <= 126) || (unicode.Is(unicode.Latin, c) && unicode.IsLetter(c)) || (c >= 0x2800 && c <= 0x28ff) || c == '–' || c == '—':
			if wrap {
				distance++
				col, wrap = 0, false
			}
			if col == r.width-1 {
				wrap = true
			} else {
				col++
			}
		default:
			return 0, 0, false, false
		}
		if distance >= r.height {
			return 0, 0, false, false
		}
	}
	return distance, col, wrap, true
}

func (r *Renderer) writeLog(p []byte, format func([]byte) []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return 0, r.err
	}
	anchored := r.valid()
	if anchored {
		if !utf8.Valid(p) {
			r.active = false
			r.known = false
			anchored = false
		}
		for _, c := range string(p) {
			if c != '\n' && !(c >= 32 && c <= 126) && !(unicode.Is(unicode.Latin, c) && unicode.IsLetter(c)) && c != '–' && c != '—' {
				r.active = false
				r.known = false
				anchored = false
				break
			}
		}
	}
	if anchored {
		for _, c := range string(p) {
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
		if r.handoffHeading && !r.partial && utf8.Valid(p) {
			lineStart := true
			for _, c := range string(p) {
				if c != '\n' && (c < 32 || c == 127) {
					lineStart = false
					break
				}
			}
			if lineStart {
				r.known = true
			}
		}
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
	if r.terminalHeadings && r.valid() && !r.wrap {
		seq := fmt.Sprintf("\r\x1b[%dA\r\x1b[2K%s\r\x1b[%dB\r", r.distance, r.formatAnchor(r.message, "37", r.width), r.distance)
		if r.col > 0 {
			seq += fmt.Sprintf("\x1b[%dC", r.col)
		}
		if _, err := r.write([]byte(seq)); err != nil {
			return err
		}
		r.handoffHeading = true
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
