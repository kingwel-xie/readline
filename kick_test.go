package readline

import (
	"io"
	"testing"
	"time"
)

type kickNopWriteCloser struct{ io.Writer }

func (kickNopWriteCloser) Close() error { return nil }

// waitReading polls IsReading until it matches want or the deadline passes.
func waitReading(t *testing.T, term *Terminal, want bool) bool {
	t.Helper()
	for i := 0; i < 100; i++ {
		if term.IsReading() == want {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TestNoStaleKickAfterLine: a kick that arrives while the loop is consuming a
// line (continuous-read mode) must not start a fresh pending read once the
// line completes. A standing pending read shares the tty with any child
// process, which is what starved the full-screen editor (todo: sys cj edit).
func TestNoStaleKickAfterLine(t *testing.T) {
	pr, pw := io.Pipe()
	rl, err := NewEx(&Config{
		Stdin:          pr,
		Stdout:         kickNopWriteCloser{io.Discard},
		FuncIsTerminal: func() bool { return true },
		FuncMakeRaw:    func() error { return nil },
		FuncExitRaw:    func() error { return nil },
		FuncGetWidth:   func() int { return 80 },
	})
	if err != nil {
		t.Fatalf("readline: %v", err)
	}
	t.Cleanup(func() {
		_ = pw.Close()
		_ = rl.Close()
	})

	// completed lines are pushed to the operation's outchan; nobody is in
	// Readline() here, so drain them to keep the operation loop moving
	go func() {
		for range rl.Operation.OutChan() {
		}
	}()

	term := rl.Terminal

	// start a line and let the loop enter continuous-read mode
	term.KickRead()
	if _, err := pw.Write([]byte("a")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitReading(t, term, true) {
		t.Fatal("the loop did not start reading")
	}

	// kick again while the line is being read, then finish it with Enter
	term.KickRead()
	if _, err := pw.Write([]byte("\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !waitReading(t, term, false) {
		t.Fatal("a stale kick left a pending read after the line completed")
	}

	// the next line must still be read normally
	term.KickRead()
	if _, err := pw.Write([]byte("b")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitReading(t, term, true) {
		t.Fatal("the loop stopped reading after the stale-kick drain")
	}
}
