package application

import (
	"regexp"
	"testing"
)

func TestTextHashFormatIsSixteenHex(t *testing.T) {
	h := TextHash("hello world")
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(h) {
		t.Fatalf("got %q, want 16-char lowercase hex", h)
	}
}

func TestTextHashStableUnderWhitespaceAndCase(t *testing.T) {
	base := TextHash("Hello World")
	cases := []string{
		"hello world",
		"HELLO  WORLD",
		"  hello   world  ",
		"\thello\n\tworld\n",
	}
	for _, c := range cases {
		if got := TextHash(c); got != base {
			t.Errorf("TextHash(%q) = %q, want %q", c, got, base)
		}
	}
}

func TestTextHashChangesOnContentEdit(t *testing.T) {
	a := TextHash("Senior Go backend engineer")
	b := TextHash("Senior Rust backend engineer")
	if a == b {
		t.Fatalf("hash collided across distinct content: %q", a)
	}
}

func TestTextHashEmptyInputStable(t *testing.T) {
	a := TextHash("")
	b := TextHash("   \t\n")
	if a != b {
		t.Fatalf("empty/whitespace variants produced different hashes: %q vs %q", a, b)
	}
}
