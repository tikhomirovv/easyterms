// Command credit is an admin CLI to credit checks to a user (MVP manual billing).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/tikhomirovv/easyterms/internal/core/config"
	"github.com/tikhomirovv/easyterms/internal/core/services/billing"
	"github.com/tikhomirovv/easyterms/internal/payment/manual"
	"github.com/tikhomirovv/easyterms/internal/storage/postgres"
)

func main() {
	telegramID := flag.Int64("telegram-id", 0, "Telegram user id (mutually exclusive with -user-id)")
	userIDStr := flag.String("user-id", "", "internal user UUID (mutually exclusive with -telegram-id)")
	amount := flag.Int("amount", 0, "number of checks to credit (required)")
	key := flag.String("key", "", "idempotency key for this credit operation (required)")
	flag.Parse()

	if *amount <= 0 {
		fmt.Fprintln(os.Stderr, "-amount must be positive")
		os.Exit(1)
	}
	if *key == "" {
		fmt.Fprintln(os.Stderr, "-key is required (idempotency)")
		os.Exit(1)
	}
	hasTG := *telegramID != 0
	hasUID := *userIDStr != ""
	if hasTG == hasUID {
		fmt.Fprintln(os.Stderr, "specify exactly one of -telegram-id or -user-id")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	var userID uuid.UUID
	if hasUID {
		userID, err = uuid.Parse(*userIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid user-id: %v\n", err)
			os.Exit(1)
		}
		if _, err := store.Users().GetByID(ctx, userID); err != nil {
			fmt.Fprintf(os.Stderr, "user not found: %v\n", err)
			os.Exit(1)
		}
	} else {
		u, err := store.Users().GetByTelegramID(ctx, *telegramID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "user not found for telegram-id %d: %v\n", *telegramID, err)
			os.Exit(1)
		}
		userID = u.ID
	}

	svc := billing.NewService(store.Users(), store.Ledger(), store.Purchases(), manual.NewProvider())
	if err := svc.CreditChecks(ctx, userID, *amount, *key); err != nil {
		fmt.Fprintf(os.Stderr, "credit failed: %v\n", err)
		os.Exit(1)
	}

	bal, err := svc.Balance(ctx, userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "balance read failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("credited %d checks to user %s (balance=%d)\n", *amount, userID, bal)
}
