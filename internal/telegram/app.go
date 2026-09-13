package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/core/services/document"
	"github.com/tikhomirovv/easyterms/internal/telegram/i18n"
)

// App wires Telegram handlers to core services.
type App struct {
	users     ports.UserRepository
	docs      *document.Service
	analysis  analysisRunner
	allowlist Allowlist
	log       *slog.Logger
}

type analysisRunner interface {
	Run(ctx context.Context, userID, documentID uuid.UUID, analysisType string) (*domain.AnalysisResult, error)
}

// NewApp constructs the handler app.
func NewApp(
	users ports.UserRepository,
	docs *document.Service,
	analyze analysisRunner,
	allowlist Allowlist,
	log *slog.Logger,
) *App {
	return &App{users: users, docs: docs, analysis: analyze, allowlist: allowlist, log: log}
}

func (a *App) ensureUser(ctx context.Context, telegramID int64, locale string) (*domain.User, error) {
	u, err := a.users.GetByTelegramID(ctx, telegramID)
	if err == nil {
		if locale != "" && u.Locale != locale {
			u.Locale = locale
			_ = a.users.Update(ctx, u)
		}
		return u, nil
	}
	if !errors.Is(err, ports.ErrNotFound) {
		return nil, err
	}
	u = &domain.User{
		TelegramID: telegramID,
		Locale:     locale,
	}
	if err := a.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *App) activeDraft(ctx context.Context, userID uuid.UUID) (*domain.Document, error) {
	docs, err := a.docs.ListHistory(ctx, userID, 20, 0)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if d.Status == domain.DocumentStatusDraft {
			return &d, nil
		}
	}
	return nil, ports.ErrNotFound
}

func (a *App) latestIngested(ctx context.Context, userID uuid.UUID) (*domain.Document, error) {
	docs, err := a.docs.ListHistory(ctx, userID, 20, 0)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if d.Status == domain.DocumentStatusIngested {
			return &d, nil
		}
	}
	return nil, ports.ErrNotFound
}

func isURL(text string) bool {
	text = strings.TrimSpace(text)
	u, err := url.Parse(text)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func normalizeInputURL(text string) string {
	text = strings.TrimSpace(text)
	if isURL(text) {
		return text
	}
	if strings.Contains(text, " ") {
		return text
	}
	candidate := "https://" + strings.TrimPrefix(text, "//")
	u, err := url.Parse(candidate)
	if err != nil || u.Host == "" || !strings.Contains(u.Host, ".") {
		return text
	}
	return candidate
}

func formatPlainPayload(payload []byte) string {
	var v struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(payload, &v); err == nil && v.Summary != "" {
		return sanitizeUTF8(v.Summary)
	}
	return sanitizeUTF8(string(payload))
}

func formatHighlightsPayload(payload []byte) string {
	var v struct {
		Highlights []struct {
			Title       string `json:"title"`
			Explanation string `json:"explanation"`
			Severity    string `json:"severity"`
		} `json:"highlights"`
	}
	if err := json.Unmarshal(payload, &v); err != nil || len(v.Highlights) == 0 {
		return sanitizeUTF8(string(payload))
	}
	var b strings.Builder
	for i, h := range v.Highlights {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "• %s (%s)\n%s",
			sanitizeUTF8(h.Title),
			sanitizeUTF8(h.Severity),
			sanitizeUTF8(h.Explanation),
		)
	}
	return b.String()
}

func userFacingErr(locale string, err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, core.ErrNoSources):
		return i18n.T(locale, "err_ingest_failed")
	case errors.Is(err, core.ErrForbidden), errors.Is(err, core.ErrInvalidState):
		return i18n.T(locale, "error_generic")
	case strings.Contains(msg, "fetch url"):
		return i18n.T(locale, "err_url_fetch")
	case strings.Contains(msg, "ingest llm"):
		return i18n.T(locale, "err_ingest_failed")
	case strings.Contains(msg, "ingest: no text"):
		return i18n.T(locale, "err_ingest_failed")
	default:
		return i18n.T(locale, "error_generic")
	}
}

func withDisclaimer(locale, text string) string {
	return text + "\n\n" + i18n.T(locale, "disclaimer")
}
