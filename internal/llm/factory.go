// Package llm wires LLM adapters from environment configuration.
package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/llm/openai"
)

// NewClientFromEnv returns an LLM client based on LLM_PROVIDER (default: openai-compatible).
// Supports OpenAI, OpenRouter, LM Studio, and any OpenAI-compatible HTTP API.
func NewClientFromEnv() (ports.LLMClient, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	switch provider {
	case "", "openai-compatible", "openai":
		cfg, err := openai.LoadConfig()
		if err != nil {
			return nil, err
		}
		return openai.NewClient(cfg, nil), nil
	default:
		return nil, fmt.Errorf("unsupported LLM_PROVIDER %q (use openai-compatible)", provider)
	}
}
