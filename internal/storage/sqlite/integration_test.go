package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/storage/migrate"
	"github.com/tikhomirovv/easyterms/internal/storage/sqlite"
)

func TestStore_userDocumentFlow(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "easyterms.db")
	if err := migrate.Up(dbPath); err != nil {
		t.Fatal(err)
	}

	store, err := sqlite.NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	user := &domain.User{TelegramID: 4242, Locale: "en"}
	if err := store.Users().Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	doc := &domain.Document{UserID: user.ID, Status: domain.DocumentStatusDraft}
	if err := store.Documents().Create(ctx, doc); err != nil {
		t.Fatal(err)
	}

	text := "sample terms"
	if err := store.DocumentSources().Create(ctx, &domain.DocumentSource{
		DocumentID: doc.ID,
		Kind:       domain.SourceKindText,
		Content:    &text,
		Sequence:   0,
	}); err != nil {
		t.Fatal(err)
	}

	got, err := store.Users().GetByTelegramID(ctx, 4242)
	if err != nil || got.Locale != "en" {
		t.Fatalf("user: %v", err)
	}
}
