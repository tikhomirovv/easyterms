package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type analysisResultRepo struct {
	db *sql.DB
}

func (r *analysisResultRepo) Upsert(ctx context.Context, result *domain.AnalysisResult) error {
	now := time.Now().UTC()
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}
	result.CreatedAt = now
	const q = `
		INSERT INTO analysis_results (id, document_id, analysis_type, locale, payload, meta, cached, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (document_id, analysis_type) DO UPDATE SET
			locale = excluded.locale,
			payload = excluded.payload,
			meta = excluded.meta,
			cached = excluded.cached`
	cached := 0
	if result.Cached {
		cached = 1
	}
	_, err := r.db.ExecContext(ctx, q,
		result.ID.String(), result.DocumentID.String(), result.AnalysisType, result.Locale,
		string(result.Payload), string(result.Meta), cached, formatTime(now),
	)
	return err
}

func (r *analysisResultRepo) GetByDocumentAndType(ctx context.Context, documentID uuid.UUID, analysisType string) (*domain.AnalysisResult, error) {
	const q = `
		SELECT id, document_id, analysis_type, locale, payload, meta, cached, created_at
		FROM analysis_results
		WHERE document_id = ? AND analysis_type = ?`
	var a domain.AnalysisResult
	var idStr, docIDStr, payload, meta string
	var cached int
	var createdAt string
	err := r.db.QueryRowContext(ctx, q, documentID.String(), analysisType).Scan(
		&idStr, &docIDStr, &a.AnalysisType, &a.Locale,
		&payload, &meta, &cached, &createdAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	a.ID = uuid.MustParse(idStr)
	a.DocumentID = uuid.MustParse(docIDStr)
	a.Payload = []byte(payload)
	a.Meta = []byte(meta)
	a.Cached = cached != 0
	a.CreatedAt = parseTime(createdAt)
	return &a, nil
}
