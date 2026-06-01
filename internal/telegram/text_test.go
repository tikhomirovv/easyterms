package telegram

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateRunes_preservesValidUTF8(t *testing.T) {
	// Cyrillic is 2 bytes per rune; byte slice truncation often breaks UTF-8.
	s := strings.Repeat("абв", 1200)
	out := truncateRunes(s, 100)
	if !utf8.ValidString(out) {
		t.Fatal("truncated string is not valid UTF-8")
	}
	if utf8.RuneCountInString(out) != 101 { // 100 runes + ellipsis
		t.Fatalf("rune count = %d, want 101", utf8.RuneCountInString(out))
	}
}

func TestTruncateRunes_avoidsByteSliceBug(t *testing.T) {
	s := strings.Repeat("а", 2000) // 2 bytes per rune; len(s)=4000
	broken := s[:3501]               // odd byte count → invalid UTF-8
	if utf8.ValidString(broken) {
		t.Fatal("test setup: expected invalid UTF-8 from byte truncation")
	}
	out := truncateRunes(s, 3500)
	if !utf8.ValidString(out) {
		t.Fatal("truncateRunes must keep UTF-8 valid")
	}
}

func TestSanitizeUTF8(t *testing.T) {
	invalid := "ok\xff\xfe"
	out := sanitizeUTF8(invalid)
	if !utf8.ValidString(out) {
		t.Fatal("expected valid UTF-8")
	}
	if !strings.HasPrefix(out, "ok") {
		t.Fatalf("got %q", out)
	}
}
