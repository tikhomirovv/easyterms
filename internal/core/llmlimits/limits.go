// Package llmlimits reads LLM-related size limits from environment variables.
package llmlimits

import (
	"os"
	"strconv"
)

const (
	// DefaultMaxInputChars is a safe default for ~4k context local models (LM Studio).
	DefaultMaxInputChars = 8_000
)

// MaxInputChars returns the maximum characters sent to the LLM per request (ingest + analysis).
func MaxInputChars() int {
	n := envInt("LLM_MAX_INPUT_CHARS", DefaultMaxInputChars)
	if n < 500 {
		return 500
	}
	return n
}

// TruncateForLLM caps text to MaxInputChars for model context limits.
func TruncateForLLM(s string) (out string, truncated bool) {
	max := MaxInputChars()
	if len(s) <= max {
		return s, false
	}
	return s[:max] + "\n\n[truncated for model context limit]", true
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
