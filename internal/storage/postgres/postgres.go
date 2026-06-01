// Package postgres implements storage ports against PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

// Store bundles PostgreSQL repository implementations.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore connects to PostgreSQL and returns repositories backed by the pool.
func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close releases the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Pool exposes the underlying pool for tests and transactions.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) Users() ports.UserRepository {
	return &userRepo{pool: s.pool}
}

func (s *Store) Documents() ports.DocumentRepository {
	return &documentRepo{pool: s.pool}
}

func (s *Store) DocumentSources() ports.DocumentSourceRepository {
	return &documentSourceRepo{pool: s.pool}
}

func (s *Store) AnalysisResults() ports.AnalysisResultRepository {
	return &analysisResultRepo{pool: s.pool}
}

func (s *Store) Ledger() ports.LedgerRepository {
	return &ledgerRepo{pool: s.pool}
}

func mapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	}
	return err
}
