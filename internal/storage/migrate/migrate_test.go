package migrate

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/storage/migrations"
)

func TestMigrationsEmbedded(t *testing.T) {
	var ups, downs int
	err := fs.WalkDir(migrations.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch {
		case strings.HasSuffix(path, ".up.sql"):
			ups++
		case strings.HasSuffix(path, ".down.sql"):
			downs++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk migrations: %v", err)
	}
	if ups == 0 {
		t.Fatal("expected at least one .up.sql migration")
	}
	if ups != downs {
		t.Fatalf("up/down migration count mismatch: %d up, %d down", ups, downs)
	}
}
