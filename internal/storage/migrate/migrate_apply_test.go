package migrate_test

import (
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/storage/migrate"
)

func TestUpDown_roundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "easyterms.db")
	if err := migrate.Up(dbPath); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := migrate.Down(dbPath); err != nil {
		t.Fatalf("down: %v", err)
	}
	if err := migrate.Up(dbPath); err != nil {
		t.Fatalf("up again: %v", err)
	}
}
