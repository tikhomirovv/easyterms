package analysis_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/core/services/analysis"
)

type mockLLM struct {
	analyze func(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error)
	calls   int
}

func (m *mockLLM) ExtractCleanText(context.Context, ports.ExtractRequest) (ports.ExtractResponse, error) {
	return ports.ExtractResponse{}, nil
}

func (m *mockLLM) Analyze(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error) {
	m.calls++
	return m.analyze(ctx, req)
}

type mem struct {
	users   map[uuid.UUID]*domain.User
	docs    map[uuid.UUID]*domain.Document
	results map[string]*domain.AnalysisResult // key: docID + type
}

func newMem() *mem {
	return &mem{
		users:   make(map[uuid.UUID]*domain.User),
		docs:    make(map[uuid.UUID]*domain.Document),
		results: make(map[string]*domain.AnalysisResult),
	}
}

func resultKey(docID uuid.UUID, t string) string {
	return docID.String() + ":" + t
}

type memUsers struct{ m *mem }

func (r *memUsers) Create(_ context.Context, u *domain.User) error {
	u.ID = uuid.New()
	r.m.users[u.ID] = u
	return nil
}
func (r *memUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.m.users[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	cp := *u
	return &cp, nil
}
func (r *memUsers) GetByTelegramID(context.Context, int64) (*domain.User, error) {
	return nil, ports.ErrNotFound
}
func (r *memUsers) Update(_ context.Context, u *domain.User) error {
	r.m.users[u.ID] = u
	return nil
}

type memDocs struct{ m *mem }

func (r *memDocs) Create(_ context.Context, d *domain.Document) error {
	d.ID = uuid.New()
	cp := *d
	r.m.docs[d.ID] = &cp
	return nil
}
func (r *memDocs) GetByID(_ context.Context, id uuid.UUID) (*domain.Document, error) {
	d, ok := r.m.docs[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	cp := *d
	return &cp, nil
}
func (r *memDocs) ListByUser(context.Context, uuid.UUID, int, int) ([]domain.Document, error) {
	return nil, nil
}
func (r *memDocs) Update(_ context.Context, d *domain.Document) error {
	cp := *d
	r.m.docs[d.ID] = &cp
	return nil
}

type memResults struct{ m *mem }

func (r *memResults) Upsert(_ context.Context, res *domain.AnalysisResult) error {
	res.ID = uuid.New()
	cp := *res
	r.m.results[resultKey(res.DocumentID, res.AnalysisType)] = &cp
	return nil
}
func (r *memResults) GetByDocumentAndType(_ context.Context, docID uuid.UUID, t string) (*domain.AnalysisResult, error) {
	res, ok := r.m.results[resultKey(docID, t)]
	if !ok {
		return nil, ports.ErrNotFound
	}
	cp := *res
	return &cp, nil
}

func ingestedDoc(userID uuid.UUID) *domain.Document {
	clean := "Terms and conditions apply."
	return &domain.Document{
		UserID:        userID,
		Status:        domain.DocumentStatusIngested,
		CheckConsumed: true,
		CleanText:     &clean,
	}
}

func TestRun_plain_callsLLMOnce(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{Locale: "ru"}
	_ = (&memUsers{m}).Create(ctx, u)
	doc := ingestedDoc(u.ID)
	_ = (&memDocs{m}).Create(ctx, doc)

	llm := &mockLLM{
		analyze: func(_ context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error) {
			if req.AnalysisType != analysis.TypePlain {
				t.Fatalf("type = %q", req.AnalysisType)
			}
			if req.Locale != "ru" {
				t.Fatalf("locale = %q", req.Locale)
			}
			payload, _ := json.Marshal(map[string]string{"summary": "ok"})
			return ports.AnalyzeResponse{Payload: payload}, nil
		},
	}
	svc := analysis.NewService(&memUsers{m}, &memDocs{m}, &memResults{m}, llm)

	got, err := svc.Run(ctx, u.ID, doc.ID, analysis.TypePlain)
	if err != nil {
		t.Fatal(err)
	}
	if got.Cached {
		t.Fatal("expected fresh result")
	}

	got2, err := svc.Run(ctx, u.ID, doc.ID, analysis.TypePlain)
	if err != nil {
		t.Fatal(err)
	}
	if !got2.Cached {
		t.Fatal("expected cached result on second call")
	}
	if llm.calls != 1 {
		t.Fatalf("llm calls = %d", llm.calls)
	}
}

func TestRun_highlights_separateCache(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{Locale: "en"}
	_ = (&memUsers{m}).Create(ctx, u)
	doc := ingestedDoc(u.ID)
	_ = (&memDocs{m}).Create(ctx, doc)

	llm := &mockLLM{
		analyze: func(_ context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResponse, error) {
			payload, _ := json.Marshal(map[string]string{"mode": req.AnalysisType})
			return ports.AnalyzeResponse{Payload: payload}, nil
		},
	}
	svc := analysis.NewService(&memUsers{m}, &memDocs{m}, &memResults{m}, llm)

	_, _ = svc.Run(ctx, u.ID, doc.ID, analysis.TypePlain)
	_, _ = svc.Run(ctx, u.ID, doc.ID, analysis.TypeHighlights)
	if llm.calls != 2 {
		t.Fatalf("llm calls = %d", llm.calls)
	}
}

func TestRun_notIngested(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{}
	_ = (&memUsers{m}).Create(ctx, u)
	doc := &domain.Document{UserID: u.ID, Status: domain.DocumentStatusDraft}
	_ = (&memDocs{m}).Create(ctx, doc)

	svc := analysis.NewService(&memUsers{m}, &memDocs{m}, &memResults{m}, &mockLLM{})
	_, err := svc.Run(ctx, u.ID, doc.ID, analysis.TypePlain)
	if err != core.ErrInvalidState {
		t.Fatalf("err = %v", err)
	}
}

func TestRun_forbidden(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	owner := &domain.User{}
	other := &domain.User{}
	_ = (&memUsers{m}).Create(ctx, owner)
	_ = (&memUsers{m}).Create(ctx, other)
	doc := ingestedDoc(owner.ID)
	_ = (&memDocs{m}).Create(ctx, doc)

	svc := analysis.NewService(&memUsers{m}, &memDocs{m}, &memResults{m}, &mockLLM{})
	_, err := svc.Run(ctx, other.ID, doc.ID, analysis.TypePlain)
	if err != core.ErrForbidden {
		t.Fatalf("err = %v", err)
	}
}
