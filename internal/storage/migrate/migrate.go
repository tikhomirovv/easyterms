// Package migrate applies SQL migrations from internal/storage/migrations.
package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/tikhomirovv/easyterms/internal/storage/migrations"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER NOT NULL PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

// Up applies all pending migrations. databasePath is a path to the SQLite file.
func Up(databasePath string) error {
	return apply(databasePath, true)
}

// Down rolls back all migrations.
func Down(databasePath string) error {
	return apply(databasePath, false)
}

func apply(databasePath string, up bool) error {
	if err := ensureParentDir(databasePath); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", sqliteDSN(databasePath))
	if err != nil {
		return fmt.Errorf("sqlite open: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(schemaMigrationsTable); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	files, err := listMigrationFiles(up)
	if err != nil {
		return err
	}
	if !up {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	for _, name := range files {
		version, err := parseVersion(name)
		if err != nil {
			return err
		}
		applied, err := isApplied(db, version)
		if err != nil {
			return err
		}
		if up {
			if applied {
				continue
			}
			if err := execMigrationFile(db, name); err != nil {
				return err
			}
			if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
				return fmt.Errorf("record migration %d: %w", version, err)
			}
		} else {
			if !applied {
				continue
			}
			if err := execMigrationFile(db, name); err != nil {
				return err
			}
			if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = ?`, version); err != nil {
				return fmt.Errorf("remove migration %d: %w", version, err)
			}
		}
	}
	return nil
}

func ensureParentDir(databasePath string) error {
	dir := filepath.Dir(databasePath)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func sqliteDSN(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return "file:" + filepath.ToSlash(abs) + "?_pragma=foreign_keys(1)"
}

func listMigrationFiles(up bool) ([]string, error) {
	suffix := ".up.sql"
	if !up {
		suffix = ".down.sql"
	}
	var names []string
	err := fs.WalkDir(migrations.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, suffix) {
			return nil
		}
		names = append(names, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	if len(names) == 0 {
		return nil, errors.New("no migration files found")
	}
	// Ensure deterministic order (000001 before 000002).
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names, nil
}

func parseVersion(name string) (int, error) {
	base := filepath.Base(name)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid migration name %q", name)
	}
	var v int
	if _, err := fmt.Sscanf(parts[0], "%d", &v); err != nil {
		return 0, fmt.Errorf("parse migration version %q: %w", name, err)
	}
	return v, nil
}

func isApplied(db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count)
	return count > 0, err
}

func execMigrationFile(db *sql.DB, name string) error {
	body, err := migrations.Files.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}
	if _, err := db.Exec(string(body)); err != nil {
		return fmt.Errorf("exec migration %s: %w", name, err)
	}
	return nil
}
