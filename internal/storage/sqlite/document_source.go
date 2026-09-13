package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type documentSourceRepo struct {
	db *sql.DB
}

func (r *documentSourceRepo) Create(ctx context.Context, source *domain.DocumentSource) error {
	if source.ID == uuid.Nil {
		source.ID = uuid.New()
	}
	now := time.Now().UTC()
	source.CreatedAt = now
	const q = `
		INSERT INTO document_sources (id, document_id, kind, content, source_url, sequence, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		source.ID.String(), source.DocumentID.String(), string(source.Kind),
		source.Content, source.SourceURL, source.Sequence, formatTime(now),
	)
	return err
}

func (r *documentSourceRepo) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]domain.DocumentSource, error) {
	const q = `
		SELECT id, document_id, kind, content, source_url, sequence, created_at
		FROM document_sources
		WHERE document_id = ?
		ORDER BY sequence ASC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, documentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DocumentSource
	for rows.Next() {
		var s domain.DocumentSource
		var idStr, docIDStr, kind, createdAt string
		if err := rows.Scan(
			&idStr, &docIDStr, &kind, &s.Content, &s.SourceURL, &s.Sequence, &createdAt,
		); err != nil {
			return nil, err
		}
		s.ID = uuid.MustParse(idStr)
		s.DocumentID = uuid.MustParse(docIDStr)
		s.Kind = domain.SourceKind(kind)
		s.CreatedAt = parseTime(createdAt)
		out = append(out, s)
	}
	return out, rows.Err()
}
