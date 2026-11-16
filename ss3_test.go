package readline

import (
	"bufio"
	"strings"
	"testing"
)

// TestEscapeKeySS3: terminals in application-cursor-key mode (what nano/vim
// leave behind on some clients) send arrows as ESC O A/B/C/D. Without mapping
// them the parser returns the literal 'O' and unreads the letter, so the
// prompt shows "OA"/"OB" instead of moving the cursor.
func TestEscapeKeySS3(t *testing.T) {
	cases := map[rune]rune{
		'A': CharPrev,
		'B': CharNext,
		'C': CharForward,
		'D': CharBackward,
		'H': CharLineStart,
		'F': CharLineEnd,
	}
	for letter, want := range cases {
		buf := bufio.NewReader(strings.NewReader(string(letter)))
		if got := escapeKey('O', buf); got != want {
			t.Errorf("ESC O %c => %q, want %q", letter, got, want)
		}
	}
}

// TestEscapeKeySS3Unknown: an unknown SS3 letter keeps the old behaviour — the
// letter is unread and stays literal after the 'O'.
func TestEscapeKeySS3Unknown(t *testing.T) {
	buf := bufio.NewReader(strings.NewReader("X"))
	if got := escapeKey('O', buf); got != 'O' {
		t.Fatalf("ESC O X => %q, want literal 'O'", got)
	}
	next, _, _ := buf.ReadRune()
	if next != 'X' {
		t.Fatalf("next rune = %q, want the unread 'X'", next)
	}
}
