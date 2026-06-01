package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type documentSourceRepo struct {
	pool *pgxpool.Pool
}

func (r *documentSourceRepo) Create(ctx context.Context, source *domain.DocumentSource) error {
	const q = `
		INSERT INTO document_sources (document_id, kind, content, source_url, sequence)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, q,
		source.DocumentID, string(source.Kind), source.Content, source.SourceURL, source.Sequence,
	).Scan(&source.ID, &source.CreatedAt)
}

func (r *documentSourceRepo) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]domain.DocumentSource, error) {
	const q = `
		SELECT id, document_id, kind, content, source_url, sequence, created_at
		FROM document_sources
		WHERE document_id = $1
		ORDER BY sequence ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DocumentSource
	for rows.Next() {
		var s domain.DocumentSource
		var kind string
		if err := rows.Scan(
			&s.ID, &s.DocumentID, &kind, &s.Content, &s.SourceURL, &s.Sequence, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.Kind = domain.SourceKind(kind)
		out = append(out, s)
	}
	return out, rows.Err()
}
