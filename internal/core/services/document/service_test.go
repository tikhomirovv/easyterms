package document_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	users   map[uuid.UUID]*domain.User
	docs    map[uuid.UUID]*domain.Document
	sources []domain.DocumentSource
}

func newMemStore() *memStore {
	return &memStore{
		users: make(map[uuid.UUID]*domain.User),
		docs:  make(map[uuid.UUID]*domain.Document),
	}
}

func (m *memStore) userRepo() ports.UserRepository            { return &memUsers{m} }
func (m *memStore) docRepo() ports.DocumentRepository         { return &memDocs{m} }
func (m *memStore) sourceRepo() ports.DocumentSourceRepository { return &memSources{m} }

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

func TestDocumentService_ingestOnce(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 1, Locale: "en"}
	_ = mem.userRepo().Create(ctx, user)

	llmCalls := 0
	llm := &mockLLM{
		extract: func(ctx context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
			llmCalls++
			return ports.ExtractResponse{CleanText: "Clean agreement text."}, nil
		},
	}
	svc := document.NewService(mem.userRepo(), mem.docRepo(), mem.sourceRepo(), llm)

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
	if ingested.Status != domain.DocumentStatusIngested {
		t.Fatalf("doc = %+v", ingested)
	}
	if llmCalls != 1 {
		t.Fatalf("llm calls = %d", llmCalls)
	}

	_, err = svc.Ingest(ctx, user.ID, doc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if llmCalls != 1 {
		t.Fatalf("expected no second llm call, got %d", llmCalls)
	}
}

func TestDocumentService_listHistory(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 3, Locale: "ru"}
	_ = mem.userRepo().Create(ctx, user)
	svc := document.NewService(mem.userRepo(), mem.docRepo(), mem.sourceRepo(), &mockLLM{})

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

type stubURLFetcher struct {
	text string
	err  error
}

func (f stubURLFetcher) FetchText(context.Context, string) (string, error) {
	return f.text, f.err
}

func TestDocumentService_ingestURLSource(t *testing.T) {
	ctx := context.Background()
	mem := newMemStore()
	user := &domain.User{TelegramID: 4, Locale: "en"}
	_ = mem.userRepo().Create(ctx, user)

	var gotRaw string
	llm := &mockLLM{
		extract: func(_ context.Context, req ports.ExtractRequest) (ports.ExtractResponse, error) {
			gotRaw = req.RawText
			return ports.ExtractResponse{CleanText: "clean"}, nil
		},
	}
	svc := document.NewServiceWithURLFetcher(
		mem.userRepo(), mem.docRepo(), mem.sourceRepo(),
		stubURLFetcher{text: "fetched page text"},
		llm,
	)

	doc, _ := svc.CreateDocument(ctx, user.ID)
	_ = svc.AddURLSource(ctx, user.ID, doc.ID, "https://example.com/terms")
	_, err := svc.Ingest(ctx, user.ID, doc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotRaw != "fetched page text" {
		t.Fatalf("raw = %q", gotRaw)
	}
}
