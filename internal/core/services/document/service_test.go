package document_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/core/services/document"
)

type mockLLM struct {
	extract func(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error)
}

func (m *mockLLM) ExtractCleanText(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
	return m.extract(ctx, req)
}

func (m *mockLLM) Analyze(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error) {
	return ports.AnalyzeResponse{}, nil
}

type memStore struct {
	users    map[uuid.UUID]*domain.User
	docs     map[uuid.UUID]*domain.Document
	sources  []domain.DocumentSource
	ledger   []domain.LedgerEntry
}

func newMemStore() *memStore {
	return &memStore{
		users: make(map[uuid.UUID]*domain.User),
		docs:  make(map[uuid.UUID]*domain.Document),
	}
}

func (m *memStore) userRepo() ports.UserRepository       { return &memUsers{m} }
func (m *memStore) docRepo() ports.DocumentRepository    { return &memDocs{m} }
func (m *memStore) sourceRepo() ports.DocumentSourceRepository { return &memSources{m} }
func (m *memStore) ledgerRepo() ports.LedgerRepository   { return &memLedger{m} }

type memUsers struct{ m *memStore }

func (r *memUsers) Create(ctx context.Context, u *domain.User) error {
	u.ID = uuid.New()
	r.m.users[u.ID] = u
	return nil
}
func (r *memUsers) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.m.users[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return u, nil
}
func (r *memUsers) GetByTelegramID(ctx context.Context, _ int64) (*domain.User, error) {
	return nil, ports.ErrNotFound
}
func (r *memUsers) Update(ctx context.Context, u *domain.User) error {
	r.m.users[u.ID] = u
	return nil
}

type memDocs struct{ m *memStore }

func (r *memDocs) Create(ctx context.Context, d *domain.Document) error {
	d.ID = uuid.New()
	cp := *d
	r.m.docs[d.ID] = &cp
	return nil
}
func (r *memDocs) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	d, ok := r.m.docs[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	cp := *d
	return &cp, nil
}
func (r *memDocs) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Document, error) {
	var out []domain.Document
	for _, d := range r.m.docs {
		if d.UserID == userID {
			out = append(out, *d)
		}
	}
	return out, nil
}
func (r *memDocs) Update(ctx context.Context, d *domain.Document) error {
	cp := *d
	r.m.docs[d.ID] = &cp
	return nil
}

type memSources struct{ m *memStore }

func (r *memSources) Create(ctx context.Context, s *domain.DocumentSource) error {
	s.ID = uuid.New()
	r.m.sources = append(r.m.sources, *s)
	return nil
}
func (r *memSources) ListByDocument(ctx context.Context, documentID uuid.UUID) ([]domain.DocumentSource, error) {
	var out []domain.DocumentSource
	for _, s := range r.m.sources {
		if s.DocumentID == documentID {
			out = append(out, s)
		}
	}
	return out, nil
}

type memLedger struct{ m *memStore }

func (r *memLedger) Insert(ctx context.Context, e *domain.LedgerEntry) error {
	e.ID = uuid.New()
	r.m.ledger = append(r.m.ledger, *e)
	return nil
}
func (r *memLedger) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.LedgerEntry, error) {
	return nil, nil
}

func TestDocumentService_ingestConsumesCheckOnce(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 1, Locale: "en", CheckBalance: 1}
	_ = mem.userRepo().Create(ctx, user)

	llmCalls := 0
	llm := &mockLLM{
		extract: func(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
			llmCalls++
			return ports.ExtractResponse{CleanText: "Clean agreement text."}, nil
		},
	}
	svc := document.NewService(mem.userRepo(), mem.docRepo(), mem.sourceRepo(), mem.ledgerRepo(), llm)

	doc, err := svc.CreateDocument(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.AddTextSource(ctx, user.ID, doc.ID, "raw terms"); err != nil {
		t.Fatal(err)
	}

	ingested, err := svc.Ingest(ctx, user.ID, doc.ID)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if ingested.Status != domain.DocumentStatusIngested || !ingested.CheckConsumed {
		t.Fatalf("doc = %+v", ingested)
	}
	if llmCalls != 1 {
		t.Fatalf("llm calls = %d", llmCalls)
	}

	u, _ := mem.userRepo().GetByID(ctx, user.ID)
	if u.CheckBalance != 0 {
		t.Fatalf("balance = %d", u.CheckBalance)
	}
	if len(mem.ledger) != 1 || mem.ledger[0].Delta != -1 {
		t.Fatalf("ledger = %+v", mem.ledger)
	}

	// Second ingest must not call LLM or charge again.
	_, err = svc.Ingest(ctx, user.ID, doc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if llmCalls != 1 {
		t.Fatalf("expected no second llm call, got %d", llmCalls)
	}
}

func TestDocumentService_insufficientBalance(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 2, Locale: "en", CheckBalance: 0}
	_ = mem.userRepo().Create(ctx, user)

	svc := document.NewService(mem.userRepo(), mem.docRepo(), mem.sourceRepo(), mem.ledgerRepo(), &mockLLM{
		extract: func(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
			return ports.ExtractResponse{CleanText: "x"}, nil
		},
	})
	doc, _ := svc.CreateDocument(ctx, user.ID)
	_ = svc.AddTextSource(ctx, user.ID, doc.ID, "text")

	_, err := svc.Ingest(ctx, user.ID, doc.ID)
	if err != core.ErrInsufficientBalance {
		t.Fatalf("err = %v", err)
	}
}

func TestDocumentService_listHistory(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 3, Locale: "ru", CheckBalance: 5}
	_ = mem.userRepo().Create(ctx, user)
	svc := document.NewService(mem.userRepo(), mem.docRepo(), mem.sourceRepo(), mem.ledgerRepo(), &mockLLM{})

	_, _ = svc.CreateDocument(ctx, user.ID)
	_, _ = svc.CreateDocument(ctx, user.ID)
	list, err := svc.ListHistory(ctx, user.ID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d", len(list))
	}
}
