package ports

import "context"

// PaymentProvider abstracts a payment backend (manual stub, YooKassa, etc.).
type PaymentProvider interface {
	ID() string
	DisplayName() string
	SupportedPackages() []PackageOffer
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (PaymentSession, error)
}

// PackageOffer describes a purchasable check bundle.
type PackageOffer struct {
	ID    string
	Checks int
	// PriceCents is display price in minor units (e.g. kopecks); 0 for stub/MVP.
	PriceCents int
	Currency   string
}

// CreatePaymentRequest starts a purchase flow.
type CreatePaymentRequest struct {
	UserID    string
	PackageID string
	ProviderID string
}

// PaymentSessionKind describes how the user completes payment.
type PaymentSessionKind string

const (
	PaymentSessionKindLink            PaymentSessionKind = "link"
	PaymentSessionKindTelegramInvoice PaymentSessionKind = "telegram_invoice"
	PaymentSessionKindManualPending   PaymentSessionKind = "manual_pending"
)

// PaymentSession is returned when a purchase is started.
type PaymentSession struct {
	PaymentID string
	Kind      PaymentSessionKind
	URL       string
}
