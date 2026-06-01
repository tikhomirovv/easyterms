// Package document implements document lifecycle: create, add sources, ingest, history.
package document

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/ingest/urlfetch"
)

// maxIngestChars limits LLM input size (local models often have smaller context windows).
const maxIngestChars = 100_000

// Billing gates and records check consumption.
type Billing interface {
	HasChecks(ctx context.Context, userID uuid.UUID) (bool, error)
	ConsumeCheck(ctx context.Context, userID, documentID uuid.UUID) error
}

// URLFetcher downloads a page and returns plain text for ingest.
type URLFetcher interface {
	FetchText(ctx context.Context, rawURL string) (string, error)
}

type defaultURLFetcher struct{}

func (defaultURLFetcher) FetchText(ctx context.Context, rawURL string) (string, error) {
	return urlfetch.FetchText(ctx, rawURL, nil)
}

// Service handles document use cases.
type Service struct {
	users   ports.UserRepository
	docs    ports.DocumentRepository
	sources ports.DocumentSourceRepository
	billing Billing
	urls    URLFetcher
	llm     ports.LLMClient
}

// NewService wires document operations.
func NewService(
	users ports.UserRepository,
	docs ports.DocumentRepository,
	sources ports.DocumentSourceRepository,
	billing Billing,
	llm ports.LLMClient,
) *Service {
	return NewServiceWithURLFetcher(users, docs, sources, billing, defaultURLFetcher{}, llm)
}

// NewServiceWithURLFetcher allows injecting a URL fetcher (tests).
func NewServiceWithURLFetcher(
	users ports.UserRepository,
	docs ports.DocumentRepository,
	sources ports.DocumentSourceRepository,
	billing Billing,
	urls URLFetcher,
	llm ports.LLMClient,
) *Service {
	return &Service{
		users:   users,
		docs:    docs,
		sources: sources,
		billing: billing,
		urls:    urls,
		llm:     llm,
	}
}

// CreateDocument starts a new draft document for the user.
func (s *Service) CreateDocument(ctx context.Context, userID uuid.UUID) (*domain.Document, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	doc := &domain.Document{
		UserID:        userID,
		Status:        domain.DocumentStatusDraft,
		CheckConsumed: false,
	}
	if err := s.docs.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// AddTextSource appends pasted text to a draft document.
func (s *Service) AddTextSource(ctx context.Context, userID, documentID uuid.UUID, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("text is empty")
	}
	doc, err := s.loadOwnedDraft(ctx, userID, documentID)
	if err != nil {
		return err
	}
	seq, err := s.nextSequence(ctx, doc.ID)
	if err != nil {
		return err
	}
	return s.sources.Create(ctx, &domain.DocumentSource{
		DocumentID: doc.ID,
		Kind:       domain.SourceKindText,
		Content:    &text,
		Sequence:   seq,
	})
}

// AddURLSource appends a URL source to a draft document (fetch happens in ingest pipeline later).
func (s *Service) AddURLSource(ctx context.Context, userID, documentID uuid.UUID, rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return fmt.Errorf("url is empty")
	}
	doc, err := s.loadOwnedDraft(ctx, userID, documentID)
	if err != nil {
		return err
	}
	seq, err := s.nextSequence(ctx, doc.ID)
	if err != nil {
		return err
	}
	return s.sources.Create(ctx, &domain.DocumentSource{
		DocumentID: doc.ID,
		Kind:       domain.SourceKindURL,
		SourceURL:  &rawURL,
		Sequence:   seq,
	})
}

// Ingest runs LLM extraction, persists clean text, and consumes one check on first success.
func (s *Service) Ingest(ctx context.Context, userID, documentID uuid.UUID) (*domain.Document, error) {
	doc, err := s.docs.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, core.ErrForbidden
	}

	// Already ingested with consumed check — return cached document without calling LLM.
	if doc.CheckConsumed && doc.CleanText != nil && *doc.CleanText != "" {
		return doc, nil
	}

	srcs, err := s.sources.ListByDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if len(srcs) == 0 {
		return nil, core.ErrNoSources
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !doc.CheckConsumed {
		ok, err := s.billing.HasChecks(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, core.ErrInsufficientBalance
		}
	}

	extractReq, err := s.buildExtractRequest(ctx, srcs, user.Locale, doc.ID.String())
	if err != nil {
		slog.Warn("ingest: build extract request failed",
			slog.String("document_id", doc.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	extractReq.RawText = truncateForLLM(extractReq.RawText)
	if strings.TrimSpace(extractReq.RawText) == "" {
		slog.Warn("ingest: empty content after fetch",
			slog.String("document_id", doc.ID.String()),
			slog.Int("sources", len(srcs)),
		)
		return nil, fmt.Errorf("ingest: no text content from sources")
	}

	slog.Info("ingest: calling llm",
		slog.String("document_id", doc.ID.String()),
		slog.Int("input_chars", len(extractReq.RawText)),
	)
	resp, err := s.llm.ExtractCleanText(ctx, extractReq)
	if err != nil {
		slog.Warn("ingest: llm extract failed",
			slog.String("document_id", doc.ID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("ingest llm: %w", err)
	}

	original := assembleOriginalText(srcs)
	doc.OriginalText = &original
	clean := strings.TrimSpace(resp.CleanText)
	doc.CleanText = &clean
	doc.Status = domain.DocumentStatusIngested

	if !doc.CheckConsumed {
		if err := s.billing.ConsumeCheck(ctx, userID, doc.ID); err != nil {
			return nil, err
		}
		doc.CheckConsumed = true
	}

	if err := s.docs.Update(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// ListHistory returns the user's documents newest first.
func (s *Service) ListHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Document, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.docs.ListByUser(ctx, userID, limit, offset)
}

// GetDocument returns a document if it belongs to the user.
func (s *Service) GetDocument(ctx context.Context, userID, documentID uuid.UUID) (*domain.Document, error) {
	doc, err := s.docs.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, core.ErrForbidden
	}
	return doc, nil
}

func (s *Service) loadOwnedDraft(ctx context.Context, userID, documentID uuid.UUID) (*domain.Document, error) {
	doc, err := s.docs.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, core.ErrForbidden
	}
	if doc.Status != domain.DocumentStatusDraft {
		return nil, core.ErrInvalidState
	}
	return doc, nil
}

func (s *Service) nextSequence(ctx context.Context, documentID uuid.UUID) (int, error) {
	srcs, err := s.sources.ListByDocument(ctx, documentID)
	if err != nil {
		return 0, err
	}
	return len(srcs), nil
}

func (s *Service) buildExtractRequest(ctx context.Context, srcs []domain.DocumentSource, locale, documentID string) (ports.ExtractRequest, error) {
	var parts []string
	for _, src := range srcs {
		switch src.Kind {
		case domain.SourceKindText:
			if src.Content != nil {
				parts = append(parts, *src.Content)
			}
		case domain.SourceKindURL:
			if src.SourceURL == nil {
				continue
			}
			pageText, err := s.urls.FetchText(ctx, *src.SourceURL)
			if err != nil {
				return ports.ExtractRequest{}, fmt.Errorf("fetch url %s: %w", *src.SourceURL, err)
			}
			parts = append(parts, pageText)
		}
	}
	raw := strings.Join(parts, "\n\n")
	slog.Info("ingest: sources prepared",
		slog.String("document_id", documentID),
		slog.Int("parts", len(parts)),
		slog.Int("raw_chars", len(raw)),
	)
	return ports.ExtractRequest{
		Locale:     locale,
		DocumentID: documentID,
		RawText:    raw,
	}, nil
}

func truncateForLLM(s string) string {
	if len(s) <= maxIngestChars {
		return s
	}
	return s[:maxIngestChars] + "\n\n[truncated for model context limit]"
}

func assembleOriginalText(srcs []domain.DocumentSource) string {
	var parts []string
	for _, s := range srcs {
		switch s.Kind {
		case domain.SourceKindText:
			if s.Content != nil {
				parts = append(parts, *s.Content)
			}
		case domain.SourceKindURL:
			if s.SourceURL != nil {
				parts = append(parts, "URL: "+*s.SourceURL)
			}
		}
	}
	return strings.Join(parts, "\n\n")
}
