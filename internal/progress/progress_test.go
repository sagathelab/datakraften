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
			case 'K':
				s.rows[s.y] = []rune(strings.Repeat(" ", s.w))
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
