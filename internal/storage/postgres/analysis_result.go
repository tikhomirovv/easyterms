package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type analysisResultRepo struct {
	pool *pgxpool.Pool
}

func (r *analysisResultRepo) Upsert(ctx context.Context, result *domain.AnalysisResult) error {
	const q = `
		INSERT INTO analysis_results (document_id, analysis_type, locale, payload, meta, cached)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (document_id, analysis_type) DO UPDATE SET
			locale = EXCLUDED.locale,
			payload = EXCLUDED.payload,
			meta = EXCLUDED.meta,
			cached = EXCLUDED.cached
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, q,
		result.DocumentID, result.AnalysisType, result.Locale,
		result.Payload, result.Meta, result.Cached,
	).Scan(&result.ID, &result.CreatedAt)
}

func (r *analysisResultRepo) GetByDocumentAndType(ctx context.Context, documentID uuid.UUID, analysisType string) (*domain.AnalysisResult, error) {
	const q = `
		SELECT id, document_id, analysis_type, locale, payload, meta, cached, created_at
		FROM analysis_results
		WHERE document_id = $1 AND analysis_type = $2`
	var a domain.AnalysisResult
	err := r.pool.QueryRow(ctx, q, documentID, analysisType).Scan(
		&a.ID, &a.DocumentID, &a.AnalysisType, &a.Locale,
		&a.Payload, &a.Meta, &a.Cached, &a.CreatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &a, nil
}
