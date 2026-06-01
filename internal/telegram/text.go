package telegram

import (
	"strings"
	"unicode/utf8"
)

// sanitizeUTF8 ensures text is valid UTF-8 for the Telegram Bot API.
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "\uFFFD")
}

// truncateRunes limits string length by rune count so UTF-8 stays valid after truncation.
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "…"
}
