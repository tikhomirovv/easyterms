package openai

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Config holds OpenAI-compatible API settings (OpenAI, OpenRouter, LM Studio, Ollama shim, etc.).
type Config struct {
	BaseURL       string
	APIKey        string
	Model         string
	ProviderLabel string
	// JSONMode requests response_format json_object when supported (disable for some local models).
	JSONMode bool
}

// LoadConfig reads LLM settings from environment variables.
func LoadConfig() (Config, error) {
	baseURL := strings.TrimRight(envOrDefault("LLM_BASE_URL", defaultBaseURL), "/")
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))

	cfg := Config{
		BaseURL:       baseURL,
		APIKey:        resolveAPIKey(baseURL, apiKey),
		Model:         envOrDefault("LLM_MODEL", "gpt-4o-mini"),
		ProviderLabel: envOrDefault("LLM_PROVIDER_LABEL", "openai-compatible"),
		JSONMode:      envBoolDefault("LLM_JSON_MODE", true),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("LLM_API_KEY is required for %s", baseURL)
	}
	return cfg, nil
}

// resolveAPIKey returns the API key to send. Cloud OpenAI requires a real key;
// local OpenAI-compatible servers (LM Studio, etc.) accept any placeholder.
func resolveAPIKey(baseURL, apiKey string) string {
	if apiKey != "" {
		return apiKey
	}
	if isOpenAICloudHost(baseURL) {
		return ""
	}
	// LM Studio and similar: Bearer token is often ignored; "local" is a common default.
	return "local"
}

func isOpenAICloudHost(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return strings.Contains(baseURL, "api.openai.com")
	}
	host := strings.ToLower(u.Hostname())
	return host == "api.openai.com" || strings.HasSuffix(host, ".openai.azure.com")
}

func envBoolDefault(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch v {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
