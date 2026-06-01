package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tikhomirovv/easyterms/internal/core/prompts"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

// Client implements ports.LLMClient via an OpenAI-compatible HTTP API.
type Client struct {
	cfg    Config
	http   *http.Client
}

// NewClient creates an LLM client. httpClient may be nil to use a default timeout.
func NewClient(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	return &Client{cfg: cfg, http: httpClient}
}

// ExtractCleanText calls the chat API to normalize input into clean document text.
func (c *Client) ExtractCleanText(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
	system, user := prompts.ExtractMessages(req, "")
	content, err := c.chat(ctx, system, user, false)
	if err != nil {
		return ports.ExtractResponse{}, err
	}
	return ports.ExtractResponse{CleanText: strings.TrimSpace(content)}, nil
}

// Analyze calls the chat API for a structured analysis result (JSON payload).
func (c *Client) Analyze(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error) {
	system, user, jsonMode := prompts.AnalyzeMessages(req, "")
	content, err := c.chat(ctx, system, user, jsonMode)
	if err != nil {
		return ports.AnalyzeResponse{}, err
	}
	content = strings.TrimSpace(content)
	payload := extractJSONPayload(content)
	if !json.Valid(payload) {
		return ports.AnalyzeResponse{}, fmt.Errorf("llm: analyze response is not valid JSON")
	}
	return ports.AnalyzeResponse{Payload: payload}, nil
}

// extractJSONPayload returns JSON from raw model text (handles markdown fences from local models).
func extractJSONPayload(content string) []byte {
	if json.Valid([]byte(content)) {
		return []byte(content)
	}
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		candidate := content[start : end+1]
		if json.Valid([]byte(candidate)) {
			return []byte(candidate)
		}
	}
	return []byte(content)
}

func (c *Client) chat(ctx context.Context, system, user string, jsonMode bool) (string, error) {
	body := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	if jsonMode && c.cfg.JSONMode {
		body.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("llm: marshal request: %w", err)
	}

	url := c.cfg.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm: api status %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("llm: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("llm: api error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm: empty choices in response")
	}
	return parsed.Choices[0].Message.Content, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ProviderLabel returns the configured provider name for logging and metadata.
func (c *Client) ProviderLabel() string {
	return c.cfg.ProviderLabel
}
