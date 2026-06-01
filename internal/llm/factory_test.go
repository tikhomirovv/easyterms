package llm_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/llm"
)

func TestNewClientFromEnv_openaiCompatible(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "openai-compatible")
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:1234/v1")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_MODEL", "qwen/qwen3.6-27b")
	t.Setenv("LLM_JSON_MODE", "false")

	c, err := llm.NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected client")
	}
}

func TestNewClientFromEnv_unknownProvider(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "anthropic")
	_, err := llm.NewClientFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}
