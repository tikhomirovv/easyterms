package telegram

import "testing"

func TestNormalizeInputURL(t *testing.T) {
	if got := normalizeInputURL("example.com/terms"); got != "https://example.com/terms" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeInputURL("https://example.com/a"); got != "https://example.com/a" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeInputURL("hello world"); got != "hello world" {
		t.Fatalf("got %q", got)
	}
}
