// Command telegram is the EasyTerms Telegram bot entrypoint.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tikhomirovv/easyterms/internal/core/config"
	"github.com/tikhomirovv/easyterms/internal/core/services/analysis"
	"github.com/tikhomirovv/easyterms/internal/core/services/billing"
	"github.com/tikhomirovv/easyterms/internal/core/services/document"
	"github.com/tikhomirovv/easyterms/internal/llm"
	"github.com/tikhomirovv/easyterms/internal/payment/manual"
	"github.com/tikhomirovv/easyterms/internal/storage/postgres"
	"github.com/tikhomirovv/easyterms/internal/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	log, err := cfg.NewLogger()
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	llmClient, err := llm.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}

	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer store.Close()

	bill := billing.NewService(store.Users(), store.Ledger(), store.Purchases(), manual.NewProvider())
	docs := document.NewService(store.Users(), store.Documents(), store.DocumentSources(), bill, llmClient)
	analyze := analysis.NewService(store.Users(), store.Documents(), store.AnalysisResults(), llmClient)

	app := telegram.NewApp(store.Users(), docs, bill, analyze, log)
	log.Info("easyterms telegram starting", slog.String("log_level", cfg.LogLevel))
	return telegram.Run(ctx, cfg.TelegramBotToken, app)
}
