// Package domain holds core entity types persisted in storage.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Document status values stored in documents.status.
type DocumentStatus string

const (
	DocumentStatusDraft    DocumentStatus = "draft"
	DocumentStatusIngested DocumentStatus = "ingested"
	DocumentStatusPaid     DocumentStatus = "paid"
)

// User is a Telegram user with check balance.
type User struct {
	ID           uuid.UUID
	TelegramID   int64
	Locale       string
	CheckBalance int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Document is one analyzable agreement session.
type Document struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Status        DocumentStatus
	CheckConsumed bool
	OriginalText  *string
	CleanText     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AnalysisResult is a cached or fresh analysis output for a document.
type AnalysisResult struct {
	ID           uuid.UUID
	DocumentID   uuid.UUID
	AnalysisType string
	Locale       string
	Payload      json.RawMessage
	Meta         json.RawMessage
	Cached       bool
	CreatedAt    time.Time
}

// SourceKind is the type of document input fragment.
type SourceKind string

const (
	SourceKindText  SourceKind = "text"
	SourceKindURL   SourceKind = "url"
	SourceKindImage SourceKind = "image"
)

// DocumentSource is one piece of input attached to a document.
type DocumentSource struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Kind       SourceKind
	Content    *string
	SourceURL  *string
	Sequence   int
	CreatedAt  time.Time
}

// LedgerEntry records a check balance change.
type LedgerEntry struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DocumentID *uuid.UUID
	PurchaseID *uuid.UUID
	Delta      int
	Reason     string
	CreatedAt  time.Time
}
