package openai_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/llm/openai"
)

func TestLoadConfig_openAIRequiresKey(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("LLM_API_KEY", "")
	_, err := openai.LoadConfig()
	if err == nil {
		t.Fatal("expected error without API key for OpenAI cloud")
	}
}

func TestLoadConfig_localPlaceholderKey(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://100.64.0.7:1234/v1")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_MODEL", "qwen/qwen3.6-27b")
	t.Setenv("LLM_JSON_MODE", "false")

	cfg, err := openai.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "local" {
		t.Fatalf("APIKey = %q, want local placeholder", cfg.APIKey)
	}
	if cfg.JSONMode {
		t.Fatal("expected JSONMode false")
	}
	if cfg.Model != "qwen/qwen3.6-27b" {
		t.Fatalf("model = %q", cfg.Model)
	}
}

func TestLoadConfig_customKeyPreserved(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:1234/v1")
	t.Setenv("LLM_API_KEY", "lm-studio")
	cfg, err := openai.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "lm-studio" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
}
