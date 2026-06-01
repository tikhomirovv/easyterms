package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// UserRepository persists Telegram users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// DocumentSourceRepository persists document input fragments.
type DocumentSourceRepository interface {
	Create(ctx context.Context, source *domain.DocumentSource) error
	ListByDocument(ctx context.Context, documentID uuid.UUID) ([]domain.DocumentSource, error)
}

// DocumentRepository persists documents.
type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Document, error)
	Update(ctx context.Context, doc *domain.Document) error
}

// AnalysisResultRepository persists analysis outputs (upsert by document + type).
type AnalysisResultRepository interface {
	Upsert(ctx context.Context, result *domain.AnalysisResult) error
	GetByDocumentAndType(ctx context.Context, documentID uuid.UUID, analysisType string) (*domain.AnalysisResult, error)
}

// LedgerRepository appends and lists check balance movements.
type LedgerRepository interface {
	Insert(ctx context.Context, entry *domain.LedgerEntry) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.LedgerEntry, error)
	ExistsByReason(ctx context.Context, userID uuid.UUID, reason string) (bool, error)
}

// PurchaseRepository persists payment attempts.
type PurchaseRepository interface {
	Create(ctx context.Context, p *domain.Purchase) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Purchase, error)
	Update(ctx context.Context, p *domain.Purchase) error
}
