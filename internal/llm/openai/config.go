package openai

import (
	"fmt"
	"os"
	"strings"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Config holds OpenAI-compatible API settings (OpenAI, OpenRouter, etc.).
type Config struct {
	BaseURL       string
	APIKey        string
	Model         string
	ProviderLabel string
}

// LoadConfig reads LLM settings from environment variables.
func LoadConfig() (Config, error) {
	cfg := Config{
		BaseURL:       strings.TrimRight(envOrDefault("LLM_BASE_URL", defaultBaseURL), "/"),
		APIKey:        strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		Model:         envOrDefault("LLM_MODEL", "gpt-4o-mini"),
		ProviderLabel: envOrDefault("LLM_PROVIDER_LABEL", "openai-compatible"),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("LLM_API_KEY is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
