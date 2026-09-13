// Package sqlite implements storage ports against SQLite.
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

// Store bundles SQLite repository implementations.
type Store struct {
	db *sql.DB
}

// NewStore opens SQLite at databasePath and returns repositories backed by the DB.
func NewStore(databasePath string) (*Store, error) {
	if err := ensureParentDir(databasePath); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(databasePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite path: %w", err)
	}
	dsn := "file:" + filepath.ToSlash(abs) + "?_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB exposes the underlying database for tests.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) Users() ports.UserRepository {
	return &userRepo{db: s.db}
}

func (s *Store) Documents() ports.DocumentRepository {
	return &documentRepo{db: s.db}
}

func (s *Store) DocumentSources() ports.DocumentSourceRepository {
	return &documentSourceRepo{db: s.db}
}

func (s *Store) AnalysisResults() ports.AnalysisResultRepository {
	return &analysisResultRepo{db: s.db}
}

func ensureParentDir(databasePath string) error {
	dir := filepath.Dir(databasePath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func mapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ports.ErrNotFound
	}
	return err
}
