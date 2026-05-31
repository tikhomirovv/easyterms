package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type ledgerRepo struct {
	pool *pgxpool.Pool
}

func (r *ledgerRepo) Insert(ctx context.Context, entry *domain.LedgerEntry) error {
	const q = `
		INSERT INTO check_ledger (user_id, document_id, purchase_id, delta, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, q,
		entry.UserID, entry.DocumentID, entry.PurchaseID, entry.Delta, entry.Reason,
	).Scan(&entry.ID, &entry.CreatedAt)
}

func (r *ledgerRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.LedgerEntry, error) {
	const q = `
		SELECT id, user_id, document_id, purchase_id, delta, reason, created_at
		FROM check_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.LedgerEntry
	for rows.Next() {
		var e domain.LedgerEntry
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.DocumentID, &e.PurchaseID,
			&e.Delta, &e.Reason, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
