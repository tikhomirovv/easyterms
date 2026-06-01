// Package billing manages check balance, ledger, and payment providers.
package billing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core"
	"github.com/tikhomirovv/easyterms/internal/core/domain"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
	"github.com/tikhomirovv/easyterms/internal/payment/manual"
)

// Service implements billing use cases.
type Service struct {
	users     ports.UserRepository
	ledger    ports.LedgerRepository
	purchases ports.PurchaseRepository
	providers map[string]ports.PaymentProvider
}

// NewService registers payment providers by ID.
func NewService(
	users ports.UserRepository,
	ledger ports.LedgerRepository,
	purchases ports.PurchaseRepository,
	providers ...ports.PaymentProvider,
) *Service {
	m := make(map[string]ports.PaymentProvider, len(providers))
	for _, p := range providers {
		m[p.ID()] = p
	}
	return &Service{users: users, ledger: ledger, purchases: purchases, providers: m}
}

// Balance returns the user's current check balance.
func (s *Service) Balance(ctx context.Context, userID uuid.UUID) (int, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return u.CheckBalance, nil
}

// HasChecks reports whether the user can run a billable LLM operation.
func (s *Service) HasChecks(ctx context.Context, userID uuid.UUID) (bool, error) {
	bal, err := s.Balance(ctx, userID)
	return bal > 0, err
}

// CreditChecks adds checks to the user balance (idempotent by key).
func (s *Service) CreditChecks(ctx context.Context, userID uuid.UUID, amount int, idempotencyKey string) error {
	if amount <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}
	reason := "credit:" + idempotencyKey
	exists, err := s.ledger.ExistsByReason(ctx, userID, reason)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	u.CheckBalance += amount
	if err := s.users.Update(ctx, u); err != nil {
		return err
	}
	return s.ledger.Insert(ctx, &domain.LedgerEntry{
		UserID: userID,
		Delta:  amount,
		Reason: reason,
	})
}

// ConsumeCheck debits one check for a document ingest (idempotent per document).
func (s *Service) ConsumeCheck(ctx context.Context, userID, documentID uuid.UUID) error {
	reason := "document_ingest:" + documentID.String()
	exists, err := s.ledger.ExistsByReason(ctx, userID, reason)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.CheckBalance < 1 {
		return core.ErrInsufficientBalance
	}
	u.CheckBalance--
	if err := s.users.Update(ctx, u); err != nil {
		return err
	}
	docID := documentID
	return s.ledger.Insert(ctx, &domain.LedgerEntry{
		UserID:     userID,
		DocumentID: &docID,
		Delta:      -1,
		Reason:     reason,
	})
}

// ListPackages returns offers from all registered providers (deduped by package id).
func (s *Service) ListPackages() []ports.PackageOffer {
	seen := make(map[string]struct{})
	var out []ports.PackageOffer
	for _, p := range s.providers {
		for _, pkg := range p.SupportedPackages() {
			if _, ok := seen[pkg.ID]; ok {
				continue
			}
			seen[pkg.ID] = struct{}{}
			out = append(out, pkg)
		}
	}
	return out
}

// StartPurchase creates a pending purchase and returns a payment session.
func (s *Service) StartPurchase(ctx context.Context, userID uuid.UUID, packageID, providerID string) (ports.PaymentSession, error) {
	prov, ok := s.providers[providerID]
	if !ok {
		return ports.PaymentSession{}, fmt.Errorf("unknown provider %q", providerID)
	}
	checks, err := manual.PackageChecks(packageID)
	if err != nil {
		return ports.PaymentSession{}, err
	}
	_ = checks

	purchase := &domain.Purchase{
		UserID:     userID,
		ProviderID: providerID,
		PackageID:  packageID,
		Status:     domain.PurchaseStatusPending,
		Currency:   "RUB",
	}
	if err := s.purchases.Create(ctx, purchase); err != nil {
		return ports.PaymentSession{}, err
	}
	sess, err := prov.CreatePayment(ctx, ports.CreatePaymentRequest{
		UserID:     userID.String(),
		PackageID:  packageID,
		ProviderID: providerID,
	})
	if err != nil {
		return ports.PaymentSession{}, err
	}
	sess.PaymentID = purchase.ID.String()
	return sess, nil
}

// ConfirmPayment completes a purchase and credits checks (idempotent).
func (s *Service) ConfirmPayment(ctx context.Context, purchaseID uuid.UUID) error {
	p, err := s.purchases.GetByID(ctx, purchaseID)
	if err != nil {
		return err
	}
	if p.Status == domain.PurchaseStatusCompleted {
		return nil
	}
	checks, err := manual.PackageChecks(p.PackageID)
	if err != nil {
		return err
	}
	key := "purchase:" + purchaseID.String()
	if err := s.CreditChecks(ctx, p.UserID, checks, key); err != nil {
		return err
	}
	p.Status = domain.PurchaseStatusCompleted
	return s.purchases.Update(ctx, p)
}
