// Package analysis runs document analysis modes with DB caching.
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/llmlimits"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

// Supported analysis mode identifiers.
const (
	TypePlain      = "plain"
	TypeHighlights = "highlights"
)

// Service runs analysis modes on ingested documents.
type Service struct {
	users   ports.UserRepository
	docs    ports.DocumentRepository
	results ports.AnalysisResultRepository
	llm     ports.LLMClient
}

// NewService wires analysis operations.
func NewService(
	users ports.UserRepository,
	docs ports.DocumentRepository,
	results ports.AnalysisResultRepository,
	llm ports.LLMClient,
) *Service {
	return &Service{users: users, docs: docs, results: results, llm: llm}
}

// Run executes an analysis mode for a document owned by the user.
// Results are cached per document and analysis type; repeat calls skip the LLM.
func (s *Service) Run(ctx context.Context, userID, documentID uuid.UUID, analysisType string) (*domain.AnalysisResult, error) {
	slog.Debug("analysis: start",
		slog.String("document_id", documentID.String()),
		slog.String("user_id", userID.String()),
		slog.String("type", analysisType),
	)
	if !isSupportedType(analysisType) {
		return nil, fmt.Errorf("unsupported analysis type %q", analysisType)
	}

	doc, err := s.docs.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, core.ErrForbidden
	}
	if err := requireIngested(doc); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cached, err := s.results.GetByDocumentAndType(ctx, documentID, analysisType); err == nil {
		slog.Debug("analysis: cache hit", slog.String("document_id", documentID.String()), slog.String("type", analysisType))
		out := *cached
		out.Cached = true
		return &out, nil
	} else if err != ports.ErrNotFound {
		return nil, err
	}

	clean := ""
	if doc.CleanText != nil {
		clean = *doc.CleanText
	}
	before := len(clean)
	clean, truncated := llmlimits.TruncateForLLM(clean)
	if truncated {
		slog.Debug("analysis: input truncated",
			slog.Int("raw_chars", before),
			slog.Int("sent_chars", len(clean)),
		)
	}
	if strings.TrimSpace(clean) == "" {
		return nil, fmt.Errorf("analysis: empty clean text")
	}
	slog.Debug("analysis: calling llm",
		slog.String("document_id", documentID.String()),
		slog.Int("input_chars", len(clean)),
		slog.Bool("json_mode", true),
	)
	resp, err := s.llm.Analyze(ctx, ports.AnalyzeRequest{
		CleanText:    clean,
		AnalysisType: analysisType,
		Locale:       user.Locale,
		DocumentID:   doc.ID.String(),
	})
	if err != nil {
		slog.Warn("analysis: llm failed",
			slog.String("document_id", documentID.String()),
			slog.String("type", analysisType),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("analysis llm: %w", err)
	}
	slog.Debug("analysis: llm ok",
		slog.String("document_id", documentID.String()),
		slog.Int("payload_bytes", len(resp.Payload)),
	)

	meta, _ := json.Marshal(map[string]string{"prompt_version": "v1"})
	result := &domain.AnalysisResult{
		DocumentID:   documentID,
		AnalysisType: analysisType,
		Locale:       user.Locale,
		Payload:      resp.Payload,
		Meta:         meta,
		Cached:       false,
	}
	if err := s.results.Upsert(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func isSupportedType(t string) bool {
	switch t {
	case TypePlain, TypeHighlights:
		return true
	default:
		return false
	}
}

func requireIngested(doc *domain.Document) error {
	if doc.Status != domain.DocumentStatusIngested && doc.Status != domain.DocumentStatusPaid {
		return core.ErrInvalidState
	}
	if !doc.CheckConsumed {
		return core.ErrInvalidState
	}
	if doc.CleanText == nil || *doc.CleanText == "" {
		return core.ErrInvalidState
	}
	return nil
}
