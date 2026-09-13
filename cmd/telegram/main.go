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
	"github.com/tikhomirovv/easyterms/internal/core/services/document"
	"github.com/tikhomirovv/easyterms/internal/llm"
	"github.com/tikhomirovv/easyterms/internal/llm/openai"
	"github.com/tikhomirovv/easyterms/internal/storage/migrate"
	"github.com/tikhomirovv/easyterms/internal/storage/sqlite"
	"github.com/tikhomirovv/easyterms/internal/telegram"
	"github.com/tikhomirovv/easyterms/internal/version"
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

	log, err := cfg.NewLogger()
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	if err := migrate.Up(cfg.DatabasePath); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	llmClient, err := llm.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}
	llmCfg, _ := openai.LoadConfig()
	log.Info("llm configured",
		slog.String("provider", openai.HostLabel(llmCfg.BaseURL)),
		slog.String("base_url", llmCfg.BaseURL),
		slog.String("model", llmCfg.Model),
	)

	if cfg.AllowlistOpen() {
		log.Warn("SECURITY: ALLOWED_TELEGRAM_IDS is empty — bot is PUBLIC; anyone can use your LLM API key. Set ALLOWED_TELEGRAM_IDS for self-hosted use.")
	} else {
		log.Info("telegram allowlist enabled", slog.Int("allowed_users", len(cfg.AllowedTelegramIDs)))
	}

	store, err := sqlite.NewStore(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}
	defer store.Close()

	docs := document.NewService(store.Users(), store.Documents(), store.DocumentSources(), llmClient)
	analyze := analysis.NewService(store.Users(), store.Documents(), store.AnalysisResults(), llmClient)

	app := telegram.NewApp(store.Users(), docs, analyze, telegram.NewAllowlist(cfg.AllowedTelegramIDs), log)
	log.Info("easyterms telegram starting",
		slog.String("version", version.Version),
		slog.String("log_level", cfg.LogLevel),
		slog.String("database_path", cfg.DatabasePath),
	)
	return telegram.Run(ctx, cfg.TelegramBotToken, cfg.TelegramProxy, app)
}
