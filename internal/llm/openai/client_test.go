package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/llm/openai"
)

func TestExtractCleanText(t *testing.T) {
	var gotAuth string
	var gotBody chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Content: "  Clean terms text.  "}}},
		})
	}))
	defer srv.Close()

	client := openai.NewClient(openai.Config{
		BaseURL:  srv.URL + "/v1",
		APIKey:   "test-key",
		Model:    "test-model",
		JSONMode: true,
	}, srv.Client())

	resp, err := client.ExtractCleanText(context.Background(), ports.ExtractRequest{
		RawText: "messy terms",
		Locale:  "ru",
	})
	if err != nil {
		t.Fatalf("ExtractCleanText: %v", err)
	}
	if resp.CleanText != "Clean terms text." {
		t.Fatalf("clean text = %q", resp.CleanText)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotBody.Model != "test-model" {
		t.Fatalf("model = %q", gotBody.Model)
	}
}

func TestAnalyze_JSONPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body chatRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ResponseFormat == nil || body.ResponseFormat.Type != "json_object" {
			t.Fatal("expected json_object response format")
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{
				Message: chatMessage{Content: `{"summary":"Simple explanation."}`},
			}},
		})
	}))
	defer srv.Close()

	client := openai.NewClient(openai.Config{
		BaseURL:  srv.URL + "/v1",
		APIKey:   "key",
		Model:    "m",
		JSONMode: true,
	}, srv.Client())

	resp, err := client.Analyze(context.Background(), ports.AnalyzeRequest{
		CleanText:     "Long legal text.",
		AnalysisType:  "plain",
		Locale:        "en",
		PromptVersion: "v1",
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(resp.Payload, &parsed); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if parsed["summary"] != "Simple explanation." {
		t.Fatalf("summary = %q", parsed["summary"])
	}
}

func TestChat_APIErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
	}))
	defer srv.Close()

	client := openai.NewClient(openai.Config{
		BaseURL: srv.URL + "/v1",
		APIKey:  "bad",
		Model:   "m",
	}, srv.Client())

	_, err := client.ExtractCleanText(context.Background(), ports.ExtractRequest{RawText: "x", Locale: "en"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAnalyze_withoutJSONMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body chatRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ResponseFormat != nil {
			t.Fatal("expected no response_format when JSONMode disabled")
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{
				Message: chatMessage{Content: "```json\n{\"summary\":\"ok\"}\n```"},
			}},
		})
	}))
	defer srv.Close()

	client := openai.NewClient(openai.Config{
		BaseURL:  srv.URL + "/v1",
		APIKey:   "local",
		Model:    "m",
		JSONMode: false,
	}, srv.Client())

	resp, err := client.Analyze(context.Background(), ports.AnalyzeRequest{
		CleanText:    "text",
		AnalysisType: "plain",
		Locale:       "en",
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(resp.Payload, &parsed); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if parsed["summary"] != "ok" {
		t.Fatalf("summary = %q", parsed["summary"])
	}
}

// Mirror chat types for decoding request bodies in tests (avoid exporting internals).
type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}
