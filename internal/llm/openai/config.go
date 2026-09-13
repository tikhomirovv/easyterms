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
	BaseURL string
	APIKey  string
	Model   string
}

// LoadConfig reads LLM settings from environment variables.
func LoadConfig() (Config, error) {
	baseURL := strings.TrimRight(envOrDefault("LLM_BASE_URL", defaultBaseURL), "/")
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))

	cfg := Config{
		BaseURL: baseURL,
		APIKey:  resolveAPIKey(baseURL, apiKey),
		Model:   envOrDefault("LLM_MODEL", "gpt-5.6-luna"),
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("LLM_API_KEY is required for %s", baseURL)
	}
	return cfg, nil
}

// HostLabel returns a short provider label derived from the API base URL hostname.
func HostLabel(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil || u.Hostname() == "" {
		return "unknown"
	}
	return u.Hostname()
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

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
