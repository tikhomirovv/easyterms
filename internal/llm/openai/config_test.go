package openai_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/llm/openai"
)

func TestLoadConfig_defaults(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("LLM_API_KEY", "sk-test")
	t.Setenv("LLM_MODEL", "")

	cfg, err := openai.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "gpt-5.6-luna" {
		t.Fatalf("model = %q", cfg.Model)
	}
}

func TestHostLabel(t *testing.T) {
	if openai.HostLabel("https://api.openai.com/v1") != "api.openai.com" {
		t.Fatal("unexpected host label")
	}
}

func TestLoadConfig_localPlaceholder(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:1234/v1")
	t.Setenv("LLM_API_KEY", "")

	cfg, err := openai.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "local" {
		t.Fatalf("api key = %q", cfg.APIKey)
	}
}

func TestLoadConfig_missingKeyForOpenAI(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "https://api.openai.com/v1")
	t.Setenv("LLM_API_KEY", "")

	_, err := openai.LoadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}
