//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/storage/migrate"
	"github.com/tikhomirovv/easyterms/internal/storage/postgres"
)

var testStore *postgres.Store

func TestMain(m *testing.M) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		os.Exit(0)
	}
	ctx := context.Background()
	if err := migrate.Up(url); err != nil {
		panic(err)
	}
	store, err := postgres.NewStore(ctx, url)
	if err != nil {
		panic(err)
	}
	testStore = store
	code := m.Run()
	store.Close()
	os.Exit(code)
}

func cleanDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testStore.Pool().Exec(ctx, `
		TRUNCATE check_ledger, analysis_results, document_sources, documents, purchases, users CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func TestUserRepository_CRUD(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}
	cleanDB(t)
	ctx := context.Background()
	repo := testStore.Users()

	u := &domain.User{TelegramID: 12345, Locale: "ru", CheckBalance: 3}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == uuid.Nil {
		t.Fatal("expected generated id")
	}

	got, err := repo.GetByTelegramID(ctx, 12345)
	if err != nil {
		t.Fatalf("get by telegram: %v", err)
	}
	if got.CheckBalance != 3 {
		t.Fatalf("balance = %d, want 3", got.CheckBalance)
	}

	got.CheckBalance = 5
	got.Locale = "en"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	again, err := repo.GetByID(ctx, got.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if again.CheckBalance != 5 || again.Locale != "en" {
		t.Fatalf("after update: %+v", again)
	}
}

func TestDocumentRepository_CRUD(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}
	cleanDB(t)
	ctx := context.Background()

	u := &domain.User{TelegramID: 99, Locale: "en", CheckBalance: 1}
	if err := testStore.Users().Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	text := "Terms and conditions apply."
	doc := &domain.Document{
		UserID:        u.ID,
		Status:        domain.DocumentStatusDraft,
		CheckConsumed: false,
		OriginalText:  &text,
	}
	if err := testStore.Documents().Create(ctx, doc); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	got, err := testStore.Documents().GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.OriginalText == nil || *got.OriginalText != text {
		t.Fatalf("original text mismatch")
	}

	clean := "Simple terms."
	got.Status = domain.DocumentStatusIngested
	got.CheckConsumed = true
	got.CleanText = &clean
	if err := testStore.Documents().Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	list, err := testStore.Documents().ListByUser(ctx, u.ID, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Status != domain.DocumentStatusIngested {
		t.Fatalf("list = %+v", list)
	}
}

func TestAnalysisResultRepository_Upsert(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}
	cleanDB(t)
	ctx := context.Background()

	u := &domain.User{TelegramID: 1, Locale: "en", CheckBalance: 0}
	_ = testStore.Users().Create(ctx, u)
	doc := &domain.Document{UserID: u.ID, Status: domain.DocumentStatusIngested}
	_ = testStore.Documents().Create(ctx, doc)

	repo := testStore.AnalysisResults()
	payload, _ := json.Marshal(map[string]string{"summary": "ok"})
	res := &domain.AnalysisResult{
		DocumentID:   doc.ID,
		AnalysisType: "plain",
		Locale:       "en",
		Payload:      payload,
		Meta:         json.RawMessage(`{}`),
		Cached:       false,
	}
	if err := repo.Upsert(ctx, res); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.GetByDocumentAndType(ctx, doc.ID, "plain")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !jsonEqual(got.Payload, payload) {
		t.Fatalf("payload = %s, want %s", got.Payload, payload)
	}

	res.Cached = true
	res.Payload, _ = json.Marshal(map[string]string{"summary": "cached"})
	if err := repo.Upsert(ctx, res); err != nil {
		t.Fatalf("upsert again: %v", err)
	}
	got2, _ := repo.GetByDocumentAndType(ctx, doc.ID, "plain")
	if !got2.Cached {
		t.Fatal("expected cached flag true")
	}
}

func TestLedgerRepository_InsertList(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}
	cleanDB(t)
	ctx := context.Background()

	u := &domain.User{TelegramID: 42, Locale: "ru", CheckBalance: 0}
	_ = testStore.Users().Create(ctx, u)

	repo := testStore.Ledger()
	e1 := &domain.LedgerEntry{UserID: u.ID, Delta: 3, Reason: "purchase"}
	if err := repo.Insert(ctx, e1); err != nil {
		t.Fatalf("insert: %v", err)
	}
	e2 := &domain.LedgerEntry{UserID: u.ID, Delta: -1, Reason: "consume"}
	if err := repo.Insert(ctx, e2); err != nil {
		t.Fatalf("insert: %v", err)
	}

	entries, err := repo.ListByUser(ctx, u.ID, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len = %d", len(entries))
	}
	if entries[0].Delta != -1 {
		t.Fatalf("expected newest first, got %+v", entries)
	}
}

func jsonEqual(a, b []byte) bool {
	var va, vb any
	if err := json.Unmarshal(a, &va); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return false
	}
	ab, _ := json.Marshal(va)
	bb, _ := json.Marshal(vb)
	return string(ab) == string(bb)
}

func TestUserRepository_NotFound(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}
	cleanDB(t)
	ctx := context.Background()
	_, err := testStore.Users().GetByID(ctx, uuid.New())
	if err != ports.ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
