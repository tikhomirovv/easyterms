package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type purchaseRepo struct {
	pool *pgxpool.Pool
}

func (r *purchaseRepo) Create(ctx context.Context, p *domain.Purchase) error {
	const q = `
		INSERT INTO purchases (user_id, provider_id, package_id, amount_cents, currency, status, external_payment_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q,
		p.UserID, p.ProviderID, p.PackageID, p.AmountCents, p.Currency,
		string(p.Status), p.ExternalPaymentID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *purchaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Purchase, error) {
	const q = `
		SELECT id, user_id, provider_id, package_id, amount_cents, currency, status, external_payment_id, created_at, updated_at
		FROM purchases WHERE id = $1`
	var p domain.Purchase
	var status string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.UserID, &p.ProviderID, &p.PackageID, &p.AmountCents, &p.Currency,
		&status, &p.ExternalPaymentID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	p.Status = domain.PurchaseStatus(status)
	return &p, nil
}

func (r *purchaseRepo) Update(ctx context.Context, p *domain.Purchase) error {
	const q = `
		UPDATE purchases SET status = $2, external_payment_id = $3, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q, p.ID, string(p.Status), p.ExternalPaymentID).Scan(&p.UpdatedAt)
	return mapNotFound(err)
}
