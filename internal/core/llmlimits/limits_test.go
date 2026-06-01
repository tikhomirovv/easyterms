package llmlimits_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/core/llmlimits"
)

func TestMaxInputChars_default(t *testing.T) {
	t.Setenv("LLM_MAX_INPUT_CHARS", "")
	if got := llmlimits.MaxInputChars(); got != llmlimits.DefaultMaxInputChars {
		t.Fatalf("got %d", got)
	}
}

func TestMaxInputChars_override(t *testing.T) {
	t.Setenv("LLM_MAX_INPUT_CHARS", "12000")
	if got := llmlimits.MaxInputChars(); got != 12000 {
		t.Fatalf("got %d", got)
	}
}
