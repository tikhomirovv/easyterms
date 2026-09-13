package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type documentRepo struct {
	db *sql.DB
}

func (r *documentRepo) Create(ctx context.Context, doc *domain.Document) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	const q = `
		INSERT INTO documents (id, user_id, status, original_text, clean_text, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		doc.ID.String(), doc.UserID.String(), string(doc.Status),
		doc.OriginalText, doc.CleanText, formatTime(now), formatTime(now),
	)
	return err
}

func (r *documentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	const q = `
		SELECT id, user_id, status, original_text, clean_text, created_at, updated_at
		FROM documents WHERE id = ?`
	return scanDocument(r.db.QueryRowContext(ctx, q, id.String()))
}

func (r *documentRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Document, error) {
	const q = `
		SELECT id, user_id, status, original_text, clean_text, created_at, updated_at
		FROM documents
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, q, userID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Document
	for rows.Next() {
		d, err := scanDocumentRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *documentRepo) Update(ctx context.Context, doc *domain.Document) error {
	now := time.Now().UTC()
	doc.UpdatedAt = now
	const q = `
		UPDATE documents
		SET status = ?, original_text = ?, clean_text = ?, updated_at = ?
		WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q,
		string(doc.Status), doc.OriginalText, doc.CleanText, formatTime(now), doc.ID.String(),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return mapNotFound(sql.ErrNoRows)
	}
	return nil
}

func scanDocument(row *sql.Row) (*domain.Document, error) {
	var d domain.Document
	var idStr, userIDStr, status, createdAt, updatedAt string
	err := row.Scan(
		&idStr, &userIDStr, &status, &d.OriginalText, &d.CleanText, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	d.ID = uuid.MustParse(idStr)
	d.UserID = uuid.MustParse(userIDStr)
	d.Status = domain.DocumentStatus(status)
	d.CreatedAt = parseTime(createdAt)
	d.UpdatedAt = parseTime(updatedAt)
	return &d, nil
}

func scanDocumentRow(rows *sql.Rows) (*domain.Document, error) {
	var d domain.Document
	var idStr, userIDStr, status, createdAt, updatedAt string
	err := rows.Scan(
		&idStr, &userIDStr, &status, &d.OriginalText, &d.CleanText, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.ID = uuid.MustParse(idStr)
	d.UserID = uuid.MustParse(userIDStr)
	d.Status = domain.DocumentStatus(status)
	d.CreatedAt = parseTime(createdAt)
	d.UpdatedAt = parseTime(updatedAt)
	return &d, nil
}
