package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type documentRepo struct {
	pool *pgxpool.Pool
}

func (r *documentRepo) Create(ctx context.Context, doc *domain.Document) error {
	const q = `
		INSERT INTO documents (user_id, status, check_consumed, original_text, clean_text)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q,
		doc.UserID, string(doc.Status), doc.CheckConsumed, doc.OriginalText, doc.CleanText,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
}

func (r *documentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	const q = `
		SELECT id, user_id, status, check_consumed, original_text, clean_text, created_at, updated_at
		FROM documents WHERE id = $1`
	var d domain.Document
	var status string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&d.ID, &d.UserID, &status, &d.CheckConsumed,
		&d.OriginalText, &d.CleanText, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	d.Status = domain.DocumentStatus(status)
	return &d, nil
}

func (r *documentRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Document, error) {
	const q = `
		SELECT id, user_id, status, check_consumed, original_text, clean_text, created_at, updated_at
		FROM documents
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Document
	for rows.Next() {
		var d domain.Document
		var status string
		if err := rows.Scan(
			&d.ID, &d.UserID, &status, &d.CheckConsumed,
			&d.OriginalText, &d.CleanText, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		d.Status = domain.DocumentStatus(status)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *documentRepo) Update(ctx context.Context, doc *domain.Document) error {
	const q = `
		UPDATE documents
		SET status = $2, check_consumed = $3, original_text = $4, clean_text = $5, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q,
		doc.ID, string(doc.Status), doc.CheckConsumed, doc.OriginalText, doc.CleanText,
	).Scan(&doc.UpdatedAt)
	return mapNotFound(err)
}
