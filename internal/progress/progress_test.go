package progress

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"golang.org/x/term"
)

type terminalOutput struct {
	bytes.Buffer
	fd uintptr
}

func (w *terminalOutput) Fd() uintptr { return w.fd }

func TestNoColorDisablesAnchoring(t *testing.T) {
	tty, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("PTY unavailable: %v", err)
	}
	defer tty.Close()
	if !term.IsTerminal(int(tty.Fd())) {
		t.Skip("PTY is not a terminal")
	}
	t.Setenv("TERM", "xterm")
	for _, value := range []string{"", "1"} {
		t.Run(fmt.Sprintf("NO_COLOR=%q", value), func(t *testing.T) {
			t.Setenv("NO_COLOR", value)
			out := &terminalOutput{fd: tty.Fd()}
			r := NewAtLineStart(out)
			if r.size != nil || r.known || r.color {
				t.Fatal("NO_COLOR must disable dynamic terminal output")
			}
			if err := r.Running("work"); err != nil {
				t.Fatal(err)
			}
			if err := r.Done("done"); err != nil {
				t.Fatal(err)
			}
			if got, want := out.String(), "[...] work\n[OK] done\n"; got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestAppendOnlyEventsAndLogs(t *testing.T) {
	var out bytes.Buffer
	r := &Renderer{out: &out, color: true}
	_ = r.Running("first")
	_, _ = r.Write([]byte("log\npartial"))
	_ = r.Running("second")
	_ = r.Fail("bad\nthing")
	_, _ = r.Write([]byte("later\n"))
	_ = r.Close()
	want := "[...] first\nlog\npartial\n[...] second\n[FAIL] bad thing\nlater\n"
	if got := out.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNoScreenPaddingOrCursorMovement(t *testing.T) {
	var out bytes.Buffer
	r := New(&out)
	for i := 0; i < 3; i++ {
		_ = r.Running("work")
		_ = r.Done("done")
	}
	want := strings.Repeat("[...] work\n[OK] done\n", 3)
	if got := out.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHandoff(t *testing.T) {
	var out bytes.Buffer
	r := New(&out)
	_, _ = r.Write([]byte("partial"))
	wantErr := errors.New("prompt failed")
	if err := r.Handoff(func() error { _, _ = out.WriteString("prompt\n"); return wantErr }); !errors.Is(err, wantErr) {
		t.Fatal(err)
	}
	_ = r.Fail("failed")
	if got, want := out.String(), "partial\nprompt\n[FAIL] failed\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWritePreservesBytes(t *testing.T) {
	var out bytes.Buffer
	r := New(&out)
	input := []byte("a\tæ\x1b[2J\r\n")
	n, err := r.Write(input)
	if err != nil || n != len(input) || !bytes.Equal(out.Bytes(), input) {
		t.Fatalf("n=%d err=%v output=%q", n, err, out.Bytes())
	}
}

type brokenWriter struct {
	calls int
	err   error
	short bool
}

func (w *brokenWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.short {
		return 0, nil
	}
	return 0, w.err
}

func TestStickyWriteErrors(t *testing.T) {
	for _, short := range []bool{false, true} {
		want := errors.New("failed")
		w := &brokenWriter{err: want, short: short}
		if short {
			want = io.ErrShortWrite
		}
		r := New(w)
		if err := r.Running("x"); !errors.Is(err, want) {
			t.Fatal(err)
		}
		if _, err := r.Write([]byte("log")); !errors.Is(err, want) {
			t.Fatal(err)
		}
		if err := r.Fail("x"); !errors.Is(err, want) {
			t.Fatal(err)
		}
		if err := r.Handoff(func() error { t.Error("unexpected callback"); return nil }); !errors.Is(err, want) {
			t.Fatal(err)
		}
		if err := r.Close(); !errors.Is(err, want) {
			t.Fatal(err)
		}
		if w.calls != 1 {
			t.Fatalf("writes: %d", w.calls)
		}
	}
}

func TestConcurrentPlain(t *testing.T) {
	var out bytes.Buffer
	r := New(&out)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); _ = r.Running(fmt.Sprint(i)) }(i)
	}
	wg.Wait()
	if n := strings.Count(out.String(), "\n"); n != 100 {
		t.Fatal(n)
	}
}

type terminalModel struct {
	w, h, x, y int
	rows       [][]rune
	wrap       bool
	raw        bytes.Buffer
}

func model(w, h, y int) *terminalModel {
	s := &terminalModel{w: w, h: h, y: y}
	for i := 0; i < h; i++ {
		s.rows = append(s.rows, []rune(strings.Repeat(" ", w)))
	}
	return s
}
func (s *terminalModel) down() {
	if s.y == s.h-1 {
		s.rows = append(s.rows[1:], []rune(strings.Repeat(" ", s.w)))
	} else {
		s.y++
	}
}
func (s *terminalModel) Write(p []byte) (int, error) {
	s.raw.Write(p)
	for i := 0; i < len(p); {
		c := p[i]
		i++
		switch c {
		case '\x1b':
			if i >= len(p) || p[i] != '[' {
				panic("bad escape")
			}
			i++
			start := i
			for i < len(p) && (p[i] < '@' || p[i] > '~') {
				i++
			}
			if i >= len(p) {
				panic("short escape")
			}
			v, _ := strconv.Atoi(string(p[start:i]))
			if v == 0 {
				v = 1
			}
			cmd := p[i]
			i++
			switch cmd {
			case 'A':
				s.y -= v
			case 'B':
				s.y += v
			case 'C':
				s.x += v
			case 'G':
				s.x = v - 1
			case 'K':
				if string(p[start:i-1]) == "2" {
					s.rows[s.y] = []rune(strings.Repeat(" ", s.w))
				} else {
					for x := s.x; x < s.w; x++ {
						s.rows[s.y][x] = ' '
					}
				}
			}
			if cmd != 'm' {
				s.wrap = false
			}
		case '\r':
			s.x = 0
			s.wrap = false
		case '\n':
			s.down()
			s.wrap = false
		default:
			ch := rune(c)
			if c >= 128 {
				var n int
				ch, n = utf8.DecodeRune(p[i-1:])
				i += n - 1
			}
			if s.wrap {
				s.down()
				s.x = 0
				s.wrap = false
			}
			s.rows[s.y][s.x] = ch
			if s.x == s.w-1 {
				s.wrap = true
			} else {
				s.x++
			}
		}
	}
	return len(p), nil
}
func (s *terminalModel) row(y int) string { return strings.TrimRight(string(s.rows[y]), " ") }
func anchored(s *terminalModel) *Renderer {
	return &Renderer{out: s, color: true, known: true, size: func() (int, int, error) { return s.w, s.h, nil }}
}
func TestAnchoredTerminalModel(t *testing.T) {
	for _, initial := range []int{0, 3, 4} {
		t.Run(fmt.Sprint(initial), func(t *testing.T) {
			s := model(24, 5, initial)
			r := anchored(s)
			_ = r.Running("working")
			if bytes.Count(s.raw.Bytes(), []byte("\n")) != 1 {
				t.Fatal("unexpected padding")
			}
			_, _ = r.Write([]byte("log\npart"))
			_ = r.Fail("broken")
			anchor := initial
			if anchor > 2 {
				anchor = 2
			}
			if got := s.row(anchor); got != "● broken" {
				t.Fatalf("anchor %d: %q", anchor, got)
			}
			if got := s.row(anchor + 1); got != "log" {
				t.Fatalf("log: %q", got)
			}
			_, _ = r.Write([]byte("ial"))
			if got := s.row(anchor + 2); got != "partial" {
				t.Fatalf("column: %q", got)
			}
			data := s.raw.String()
			if !strings.Contains(data, "\x1b[90m●\x1b[0m") || !strings.Contains(data, "\x1b[31m●\x1b[0m") {
				t.Fatalf("colors: %q", data)
			}
		})
	}
}
func TestAnchorDoneAndFallback(t *testing.T) {
	t.Run("done", func(t *testing.T) {
		s := model(20, 5, 0)
		r := anchored(s)
		_ = r.Running("x")
		_ = r.Done("done")
		if got := s.row(0); got != "● done" {
			t.Fatal(got)
		}
		if !strings.Contains(s.raw.String(), "\x1b[32m●\x1b[0m") {
			t.Fatal("missing green dot")
		}
	})
	t.Run("scroll", func(t *testing.T) {
		s := model(20, 3, 2)
		r := anchored(s)
		_ = r.Running("x")
		_, _ = r.Write([]byte("one\ntwo\n"))
		if r.active {
			t.Fatal("anchor survived scrolling")
		}
		_ = r.Fail("failed")
		if !strings.Contains(s.raw.String(), "[FAIL] failed\n") {
			t.Fatal("missing fallback")
		}
	})
	t.Run("resize", func(t *testing.T) {
		s := model(20, 5, 0)
		r := anchored(s)
		_ = r.Running("x")
		s.w = 21
		_ = r.Done("done")
		if !strings.Contains(s.raw.String(), "[OK] done\n") {
			t.Fatal("missing fallback")
		}
	})
	t.Run("prompt", func(t *testing.T) {
		s := model(20, 5, 0)
		r := anchored(s)
		_ = r.Running("x")
		_ = r.Handoff(func() error { return nil })
		_ = r.Fail("failed")
		if !strings.Contains(s.raw.String(), "[FAIL] failed\n") {
			t.Fatal("missing fallback")
		}
	})
	t.Run("wrap", func(t *testing.T) {
		s := model(20, 5, 0)
		r := anchored(s)
		_ = r.Running("x")
		_, _ = r.Write([]byte(strings.Repeat("a", 20)))
		_ = r.Done("done")
		if !strings.Contains(s.raw.String(), "[OK] done\n") {
			t.Fatal("missing fallback")
		}
	})
	t.Run("control", func(t *testing.T) {
		s := model(20, 5, 0)
		r := anchored(s)
		_ = r.Running("x")
		_, _ = r.Write([]byte("\t"))
		_ = r.Done("done")
		if !strings.Contains(s.raw.String(), "[OK] done\n") {
			t.Fatal("missing fallback")
		}
	})
}

func TestHandoffKeepsRunningHeadingVisible(t *testing.T) {
	for _, failed := range []bool{false, true} {
		s := model(60, 12, 0)
		r := anchored(s)
		r.terminalHeadings = true
		_ = r.Running("Updating npm")
		_, _ = r.Write([]byte("    Updating packages...\n"))
		if err := r.Handoff(func() error {
			if got := s.row(0); got != "● Updating npm" {
				t.Fatalf("heading changed during handoff: %q", got)
			}
			_, err := s.Write([]byte("  sudo? yes\r\n"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
		if failed {
			_ = r.Fail("Updating npm")
		} else {
			_ = r.Done("Updating npm")
		}
		count := 0
		for i := range s.rows {
			if s.row(i) == "● Updating npm" {
				count++
			}
		}
		if count != 1 || !strings.Contains(s.raw.String(), "\x1b[37m●\x1b[0m Updating npm") || (failed && !strings.Contains(s.raw.String(), "[FAIL] Updating npm")) {
			t.Fatalf("failed=%v headings=%d output=%q", failed, count, s.raw.String())
		}
	}
}

func TestTerminalWriterTracksInteractiveOutput(t *testing.T) {
	s := model(50, 12, 0)
	r := anchored(s)
	r.terminalHeadings = true
	_ = r.Running("Updating npm")
	w := r.TerminalWriter()
	for _, part := range []string{"  Continue? ", "y\r\n", "[sudo] password: ", "\r\n", "\x1b[32mchanged\x1b[0m 2 packages\r\n"} {
		if _, err := w.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if !r.active {
		t.Fatal("lost anchor while tracking prompt")
	}
	_ = r.Done("Updating npm")
	if s.row(0) != "● Updating npm" || !strings.Contains(s.raw.String(), "\x1b[32m●\x1b[0m Updating npm") {
		t.Fatalf("missing green heading: %q", s.raw.String())
	}
	if s.row(1) != "  Continue? y" || s.row(2) != "[sudo] password:" || s.row(3) != "changed 2 packages" {
		t.Fatalf("damaged terminal output: %q, %q, %q", s.row(1), s.row(2), s.row(3))
	}
}

func TestTerminalWriterFallsBackBeforeUnsafeOutput(t *testing.T) {
	s := model(50, 12, 0)
	r := anchored(s)
	r.terminalHeadings = true
	_ = r.Running("Updating npm")
	_, _ = r.TerminalWriter().Write([]byte("unsupported: 😀\r\n"))
	if r.active || !r.handoffHeading {
		t.Fatal("unsafe text retained anchor")
	}
	_ = r.Done("Updating npm")
	count := 0
	for i := range s.rows {
		if s.row(i) == "● Updating npm" {
			count++
		}
	}
	if count != 1 || !strings.Contains(s.raw.String(), "\x1b[37m●\x1b[0m Updating npm") {
		t.Fatalf("unsafe fallback: count=%d output=%q", count, s.raw.String())
	}
}

func TestTerminalWriterTracksWrapAndScroll(t *testing.T) {
	s := model(12, 4, 0)
	r := anchored(s)
	r.terminalHeadings = true
	_ = r.Running("work")
	_, _ = r.TerminalWriter().Write([]byte(strings.Repeat("x", 13) + "\r\n"))
	if !r.active || r.distance != 3 {
		t.Fatalf("wrap tracking: active=%v distance=%d", r.active, r.distance)
	}
	_, _ = r.TerminalWriter().Write([]byte("more\r\n"))
	if r.active || !r.handoffHeading {
		t.Fatal("expected conservative fallback before scroll")
	}
}

func TestTerminalWriterTracksBackspace(t *testing.T) {
	s := model(20, 10, 0)
	r := anchored(s)
	_ = r.Running("work")
	_, _ = r.TerminalWriter().Write([]byte("typo\b\bOK\r\n"))
	if !r.active || r.distance != 2 || r.col != 0 {
		t.Fatalf("backspace tracking: active=%v distance=%d col=%d", r.active, r.distance, r.col)
	}
}

func TestTerminalWriterPlainOutput(t *testing.T) {
	var out bytes.Buffer
	r := New(&out)
	input := []byte("password: \r\n\x1b[31merror\x1b[0m\r\n")
	if n, err := r.TerminalWriter().Write(input); err != nil || n != len(input) || !bytes.Equal(out.Bytes(), input) {
		t.Fatalf("terminal output n=%d err=%v got=%q", n, err, out.Bytes())
	}
}

func TestTerminalWriterTracksFragmentedSequences(t *testing.T) {
	for _, tc := range []struct{ first, second string }{
		{"\x1b[", "32mtext\x1b[0m\r\n"},
		{"\xc3", "\xa6\r\n"},
	} {
		var out bytes.Buffer
		r := &Renderer{out: &out, color: true, terminalHeadings: true, known: true, size: func() (int, int, error) { return 80, 20, nil }}
		_ = r.Running("work")
		w := r.TerminalWriter()
		_, _ = w.Write([]byte(tc.first))
		if !r.active {
			t.Fatalf("fragment lost anchor: %q", tc.first)
		}
		_, _ = w.Write([]byte(tc.second))
		if !r.active {
			t.Fatalf("completed fragment lost anchor: %q", tc.first)
		}
		_ = r.Done("work")
		if !strings.Contains(out.String(), tc.first+tc.second) {
			t.Fatalf("fragmented bytes changed: %q", out.String())
		}
	}
}

func TestTerminalWriterTracksNpmSpinner(t *testing.T) {
	s := model(80, 20, 0)
	r := anchored(s)
	_ = r.Running("Updating npm")
	w := r.TerminalWriter()
	for _, part := range []string{"\x1b[", "1G\x1b[0K\xe2", "\xa0\x99", "\x1b[1G\x1b[0K", "changed 2 packages\r\n"} {
		if _, err := w.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if !r.active {
		t.Fatal("npm spinner lost the progress anchor")
	}
	_ = r.Done("Updating npm")
	if s.row(0) != "● Updating npm" || s.row(1) != "changed 2 packages" || !strings.Contains(s.raw.String(), "\x1b[32m●\x1b[0m Updating npm") {
		t.Fatalf("unexpected screen: %q %q; %q", s.row(0), s.row(1), s.raw.String())
	}
}

func TestTerminalWriterTracksFragmentedOSC(t *testing.T) {
	for _, sequence := range []string{"\x1b]0;npm update\a", "\x1b]9;4;1;50\x1b\\", "\x1b]8;;https://example.test\x1b\\"} {
		var out bytes.Buffer
		r := &Renderer{out: &out, color: true, terminalHeadings: true, known: true, size: func() (int, int, error) { return 212, 47, nil }}
		_ = r.Running("Updating npm")
		w := r.TerminalWriter()
		for _, b := range []byte(sequence) {
			if _, err := w.Write([]byte{b}); err != nil {
				t.Fatal(err)
			}
		}
		_, _ = w.Write([]byte("changed 2 packages\r\n"))
		if !r.active {
			t.Fatalf("OSC sequence lost progress: %q", sequence)
		}
		_ = r.Done("Updating npm")
		if !strings.Contains(out.String(), sequence) || !strings.Contains(out.String(), "\x1b[32m●\x1b[0m Updating npm") {
			t.Fatalf("missing OSC sequence or completed heading: %q", out.String())
		}
	}
}

func TestTerminalWriterFallsBackForUnsupportedOSC(t *testing.T) {
	var out bytes.Buffer
	r := &Renderer{out: &out, color: true, terminalHeadings: true, known: true, size: func() (int, int, error) { return 80, 20, nil }}
	_ = r.Running("work")
	sequence := "\x1b]1337;File=example\a"
	_, _ = r.TerminalWriter().Write([]byte(sequence))
	if r.active || !r.handoffHeading || !strings.Contains(out.String(), sequence) {
		t.Fatalf("unsupported OSC was not safely forwarded: %q", out.String())
	}
}

func TestTerminalWriterFlushesIncompleteSequences(t *testing.T) {
	var out bytes.Buffer
	r := &Renderer{out: &out, color: true, terminalHeadings: true, known: true, size: func() (int, int, error) { return 80, 20, nil }}
	_ = r.Running("work")
	w := r.TerminalWriter().(*terminalWriter)
	_, _ = w.Write([]byte("\x1b["))
	if err := w.Flush(); err != nil || r.active || !strings.HasSuffix(out.String(), "\x1b[") {
		t.Fatalf("incomplete output: active=%v err=%v output=%q", r.active, err, out.String())
	}
}

func TestTerminalWriterCanStartNextStep(t *testing.T) {
	for _, unsupported := range []bool{false, true} {
		s := model(60, 20, 0)
		r := anchored(s)
		r.terminalHeadings = true
		_ = r.Running("first")
		output := "done\r\n"
		if unsupported {
			output = "done 😀\r\n"
		}
		_, _ = r.TerminalWriter().Write([]byte(output))
		_, _ = r.Write([]byte("    [OK] Command completed\n"))
		_ = r.Done("first")
		_ = r.Running("second")
		if !r.active {
			t.Fatalf("could not anchor next step after fallback=%v", unsupported)
		}
		_ = r.Done("second")
		if !strings.Contains(s.raw.String(), "\x1b[32m●\x1b[0m second") {
			t.Fatalf("next step not completed after fallback=%v: %q", unsupported, s.raw.String())
		}
	}
}

func TestTerminalHeadingsSurviveUntrackableOutput(t *testing.T) {
	s := model(80, 20, 0)
	r := anchored(s)
	r.terminalHeadings = true
	_ = r.Running("AI apps")
	_, _ = r.Write([]byte("    \t skipped\n"))
	_ = r.Done("AI apps")
	_ = r.Running("Docker")
	_ = r.Done("Docker")
	got := s.raw.String()
	if !strings.Contains(got, "\x1b[32m●\x1b[0m AI apps") || !strings.Contains(got, "\x1b[32m●\x1b[0m Docker") || strings.Contains(got, "[...] Docker") {
		t.Fatalf("inconsistent headings: %q", got)
	}
}

func TestAnchorTracksCommonUnicodeText(t *testing.T) {
	s := model(80, 20, 0)
	r := anchored(s)
	_ = r.Running("AI apps")
	_, _ = r.Write([]byte("    – Codex: på macOS\n"))
	if !r.active {
		t.Fatal("lost anchor on common text")
	}
	_ = r.Done("AI apps")
	if !strings.Contains(s.raw.String(), "\x1b[32m●\x1b[0m AI apps") {
		t.Fatal("missing completed heading")
	}
}

func TestFormattedLogPreservesAnchor(t *testing.T) {
	s := model(80, 20, 0)
	r := anchored(s)
	_ = r.Running("work")
	input := []byte("    [OK] done\n")
	n, err := r.WriteFormatted(input, func([]byte) []byte { return []byte("    \x1b[32m[OK]\x1b[0m done\n") })
	if err != nil || n != len(input) || !r.active {
		t.Fatalf("formatted write: n=%d err=%v active=%v", n, err, r.active)
	}
	_ = r.Done("work")
	if !strings.Contains(s.raw.String(), "\x1b[32m[OK]\x1b[0m") {
		t.Fatal("missing formatted status")
	}
}

type failAfter struct {
	out          bytes.Buffer
	calls, limit int
	err          error
}

func (w *failAfter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls > w.limit {
		return 0, w.err
	}
	return w.out.Write(p)
}
func TestAnchorErrorSticky(t *testing.T) {
	want := errors.New("update failed")
	w := &failAfter{limit: 1, err: want}
	r := &Renderer{out: w, known: true, size: func() (int, int, error) { return 20, 5, nil }}
	if err := r.Running("work"); err != nil {
		t.Fatal(err)
	}
	if err := r.Fail("broken"); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err := r.Done("later"); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err := r.Close(); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if w.calls != 2 {
		t.Fatal(w.calls)
	}
}
func TestWrappedLogsRemainAnchored(t *testing.T) {
	s := model(10, 6, 0)
	r := anchored(s)
	_ = r.Running("x")
	_, _ = r.Write([]byte("1234567890123"))
	_ = r.Done("done")
	if got := s.row(0); got != "● done" {
		t.Fatal(got)
	}
	_, _ = r.Write([]byte("45"))
	if got := s.row(2); got != "12345" {
		t.Fatalf("cursor: %q", got)
	}
}
