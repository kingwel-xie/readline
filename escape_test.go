package readline

import (
	"bufio"
	"strings"
	"testing"
)

// TestReadEscKeyPrivateMarkers: CSI replies carry private markers before the
// payload ("?84;0;0c" for a DA query, ">84;0;0c" for a secondary one). Without
// skipping the marker the marker itself becomes the key type and the payload
// ("84;0;0c") leaks into the input as plain keystrokes — visible garbage at
// the next prompt.
func TestReadEscKeyPrivateMarkers(t *testing.T) {
	buf := bufio.NewReader(strings.NewReader("84;0;0c"))
	key := readEscKey('?', buf)
	if key.typ != 'c' {
		t.Fatalf("typ = %q, want 'c'", key.typ)
	}
	if key.attr != "84;0;0" {
		t.Fatalf("attr = %q, want 84;0;0", key.attr)
	}
	if got := escapeExKey(key); got != 0 {
		t.Fatalf("escapeExKey = %v, want 0 (dropped)", got)
	}

	buf = bufio.NewReader(strings.NewReader("84;0;0c"))
	key = readEscKey('>', buf)
	if key.typ != 'c' || key.attr != "84;0;0" {
		t.Fatalf("secondary DA: typ=%q attr=%q", key.typ, key.attr)
	}
}

// TestReadEscKeyStillParsesKnownKeys: the marker skip must not break the
// regular escape keys.
func TestReadEscKeyStillParsesKnownKeys(t *testing.T) {
	// delete: "\x1b[3~" — buf holds "3~"
	key := readEscKey('3', bufio.NewReader(strings.NewReader("~")))
	if key.typ != '~' || key.attr != "3" {
		t.Fatalf("delete: typ=%q attr=%q", key.typ, key.attr)
	}
	// modified arrow "\x1b[1;5A" — buf holds "1;5A"
	key = readEscKey('1', bufio.NewReader(strings.NewReader(";5A")))
	if key.typ != 'A' || key.attr != "1;5" {
		t.Fatalf("modified arrow: typ=%q attr=%q", key.typ, key.attr)
	}
}
