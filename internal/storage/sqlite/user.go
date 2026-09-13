package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
)

type userRepo struct {
	db *sql.DB
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	const q = `
		INSERT INTO users (id, telegram_id, locale, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, user.ID.String(), user.TelegramID, user.Locale, formatTime(now), formatTime(now))
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, telegram_id, locale, created_at, updated_at
		FROM users WHERE id = ?`
	var u domain.User
	var idStr, createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, q, id.String()).Scan(
		&idStr, &u.TelegramID, &u.Locale, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	u.ID = uuid.MustParse(idStr)
	u.CreatedAt = parseTime(createdAt)
	u.UpdatedAt = parseTime(updatedAt)
	return &u, nil
}

func (r *userRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	const q = `
		SELECT id, telegram_id, locale, created_at, updated_at
		FROM users WHERE telegram_id = ?`
	var u domain.User
	var idStr, createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, q, telegramID).Scan(
		&idStr, &u.TelegramID, &u.Locale, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err)
	}
	u.ID = uuid.MustParse(idStr)
	u.CreatedAt = parseTime(createdAt)
	u.UpdatedAt = parseTime(updatedAt)
	return &u, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC()
	user.UpdatedAt = now
	const q = `
		UPDATE users
		SET locale = ?, updated_at = ?
		WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, user.Locale, formatTime(now), user.ID.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return mapNotFound(sql.ErrNoRows)
	}
	return nil
}
