// Package manual implements a stub PaymentProvider for MVP (admin credits checks).
package manual

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

// ProviderID is the manual payment provider identifier.
const ProviderID = "manual"

// Provider implements manual / admin payment flow.
type Provider struct{}

// NewProvider returns a manual payment provider.
func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) ID() string { return ProviderID }

func (p *Provider) DisplayName() string { return "Manual (admin)" }

func (p *Provider) SupportedPackages() []ports.PackageOffer {
	return []ports.PackageOffer{
		{ID: "checks_1", Checks: 1, PriceCents: 0, Currency: "RUB"},
		{ID: "checks_3", Checks: 3, PriceCents: 0, Currency: "RUB"},
		{ID: "checks_10", Checks: 10, PriceCents: 0, Currency: "RUB"},
	}
}

func (p *Provider) CreatePayment(ctx context.Context, req ports.CreatePaymentRequest) (ports.PaymentSession, error) {
	_ = ctx
	return ports.PaymentSession{
		PaymentID: uuid.New().String(),
		Kind:      ports.PaymentSessionKindManualPending,
		URL:       "",
	}, nil
}

// PackageChecks returns the number of checks for a package id.
func PackageChecks(packageID string) (int, error) {
	for _, o := range (&Provider{}).SupportedPackages() {
		if o.ID == packageID {
			return o.Checks, nil
		}
	}
	return 0, fmt.Errorf("unknown package %q", packageID)
}
