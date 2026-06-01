package billing_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/core/services/billing"
	"github.com/tikhomirovv/easyterms/internal/payment/manual"
)

type mem struct {
	users     map[uuid.UUID]*domain.User
	ledger    []domain.LedgerEntry
	purchases map[uuid.UUID]*domain.Purchase
}

func newMem() *mem {
	return &mem{
		users:     make(map[uuid.UUID]*domain.User),
		purchases: make(map[uuid.UUID]*domain.Purchase),
	}
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

type memLedger struct{ m *mem }

func (r *memLedger) Insert(_ context.Context, e *domain.LedgerEntry) error {
	e.ID = uuid.New()
	r.m.ledger = append(r.m.ledger, *e)
	return nil
}
func (r *memLedger) ListByUser(context.Context, uuid.UUID, int) ([]domain.LedgerEntry, error) {
	return nil, nil
}
func (r *memLedger) ExistsByReason(_ context.Context, userID uuid.UUID, reason string) (bool, error) {
	for _, e := range r.m.ledger {
		if e.UserID == userID && e.Reason == reason {
			return true, nil
		}
	}
	return false, nil
}

type memPurchases struct{ m *mem }

func (r *memPurchases) Create(_ context.Context, p *domain.Purchase) error {
	p.ID = uuid.New()
	cp := *p
	r.m.purchases[p.ID] = &cp
	return nil
}
func (r *memPurchases) GetByID(_ context.Context, id uuid.UUID) (*domain.Purchase, error) {
	p, ok := r.m.purchases[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	cp := *p
	return &cp, nil
}
func (r *memPurchases) Update(_ context.Context, p *domain.Purchase) error {
	r.m.purchases[p.ID] = p
	return nil
}

func TestCreditChecks_idempotent(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{CheckBalance: 0}
	_ = (&memUsers{m}).Create(ctx, u)

	svc := billing.NewService(&memUsers{m}, &memLedger{m}, &memPurchases{m}, manual.NewProvider())

	if err := svc.CreditChecks(ctx, u.ID, 3, "admin-1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreditChecks(ctx, u.ID, 3, "admin-1"); err != nil {
		t.Fatal(err)
	}
	bal, _ := svc.Balance(ctx, u.ID)
	if bal != 3 {
		t.Fatalf("balance = %d", bal)
	}
	if len(m.ledger) != 1 {
		t.Fatalf("ledger entries = %d", len(m.ledger))
	}
}

func TestConsumeCheck_idempotent(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{CheckBalance: 2}
	_ = (&memUsers{m}).Create(ctx, u)
	svc := billing.NewService(&memUsers{m}, &memLedger{m}, &memPurchases{m}, manual.NewProvider())

	docID := uuid.New()
	if err := svc.ConsumeCheck(ctx, u.ID, docID); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConsumeCheck(ctx, u.ID, docID); err != nil {
		t.Fatal(err)
	}
	bal, _ := svc.Balance(ctx, u.ID)
	if bal != 1 {
		t.Fatalf("balance = %d, want 1", bal)
	}
	if len(m.ledger) != 1 {
		t.Fatalf("ledger entries = %d", len(m.ledger))
	}
}

func TestConfirmPayment_idempotent(t *testing.T) {
	ctx := context.Background()
	m := newMem()
	u := &domain.User{CheckBalance: 0}
	_ = (&memUsers{m}).Create(ctx, u)
	svc := billing.NewService(&memUsers{m}, &memLedger{m}, &memPurchases{m}, manual.NewProvider())

	sess, err := svc.StartPurchase(ctx, u.ID, "checks_3", manual.ProviderID)
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := uuid.Parse(sess.PaymentID)
	_ = svc.ConfirmPayment(ctx, pid)
	_ = svc.ConfirmPayment(ctx, pid)

	bal, _ := svc.Balance(ctx, u.ID)
	if bal != 3 {
		t.Fatalf("balance = %d", bal)
	}
}
