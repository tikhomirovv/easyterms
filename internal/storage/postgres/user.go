package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (telegram_id, locale, check_balance)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, q, user.TelegramID, user.Locale, user.CheckBalance).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, telegram_id, locale, check_balance, created_at, updated_at
		FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.TelegramID, &u.Locale, &u.CheckBalance, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &u, nil
}

func (r *userRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	const q = `
		SELECT id, telegram_id, locale, check_balance, created_at, updated_at
		FROM users WHERE telegram_id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, telegramID).Scan(
		&u.ID, &u.TelegramID, &u.Locale, &u.CheckBalance, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &u, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	const q = `
		UPDATE users
		SET locale = $2, check_balance = $3, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q, user.ID, user.Locale, user.CheckBalance).Scan(&user.UpdatedAt)
	return mapNotFound(err)
}
