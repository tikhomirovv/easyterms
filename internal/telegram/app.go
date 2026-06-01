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
	"github.com/tikhomirovv/easyterms/internal/core/services/analysis"
	"github.com/tikhomirovv/easyterms/internal/core/services/billing"
	"github.com/tikhomirovv/easyterms/internal/core/services/document"
	"github.com/tikhomirovv/easyterms/internal/payment/manual"
	"github.com/tikhomirovv/easyterms/internal/telegram/i18n"
)

// App wires Telegram handlers to core services.
type App struct {
	users    ports.UserRepository
	docs     *document.Service
	billing  *billing.Service
	analysis *analysis.Service
	log      *slog.Logger
}

// NewApp constructs the handler app.
func NewApp(
	users ports.UserRepository,
	docs *document.Service,
	bill *billing.Service,
	analyze *analysis.Service,
	log *slog.Logger,
) *App {
	return &App{users: users, docs: docs, billing: bill, analysis: analyze, log: log}
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
		TelegramID:   telegramID,
		Locale:       locale,
		CheckBalance: 0,
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
		if d.Status == domain.DocumentStatusIngested || d.Status == domain.DocumentStatusPaid {
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

func formatPlainPayload(payload []byte) string {
	var v struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(payload, &v); err == nil && v.Summary != "" {
		return v.Summary
	}
	return string(payload)
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
		return string(payload)
	}
	var b strings.Builder
	for i, h := range v.Highlights {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "• %s (%s)\n%s", h.Title, h.Severity, h.Explanation)
	}
	return b.String()
}

func userFacingErr(locale string, err error) string {
	switch {
	case errors.Is(err, core.ErrInsufficientBalance):
		return i18n.T(locale, "insufficient_balance")
	case errors.Is(err, core.ErrForbidden), errors.Is(err, core.ErrInvalidState), errors.Is(err, core.ErrNoSources):
		return err.Error()
	default:
		return i18n.T(locale, "error_generic")
	}
}

func (a *App) startPurchase(ctx context.Context, userID uuid.UUID, packageID, locale string) (string, error) {
	sess, err := a.billing.StartPurchase(ctx, userID, packageID, manual.ProviderID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(i18n.T(locale, "buy_manual"), packageID, sess.PaymentID), nil
}
