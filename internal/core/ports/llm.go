// Package ports defines domain interfaces for external dependencies.
// Core depends only on these interfaces; adapters live outside internal/core.
package ports

import "context"

// LLMClient abstracts text extraction and analysis calls to an LLM provider.
type LLMClient interface {
	ExtractCleanText(ctx context.Context, req ExtractRequest) (ExtractResponse, error)
	Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error)
}

// ExtractRequest is input for ingest: raw content → clean document text.
type ExtractRequest struct {
	// RawText is pasted plain text (optional if URL is set).
	RawText string
	// URL is a page to fetch and extract (optional if RawText is set).
	URL string
	// Locale is the user's language (e.g. Telegram language code).
	Locale string
	// DocumentID links the request to a document for logging and billing.
	DocumentID string
}

// ExtractResponse is the cleaned text produced during ingest.
type ExtractResponse struct {
	CleanText string
}

// AnalyzeRequest runs an analysis mode on already-ingested clean text.
type AnalyzeRequest struct {
	CleanText    string
	AnalysisType string // plain, highlights, ...
	Locale       string
	PromptVersion string
	DocumentID   string
}

// AnalyzeResponse is a structured analysis result from the LLM.
type AnalyzeResponse struct {
	// Payload is mode-specific structured output (stored as JSONB later).
	Payload []byte
}
