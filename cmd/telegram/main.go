// Command telegram is the EasyTerms Telegram bot entrypoint.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tikhomirovv/easyterms/internal/core/config"
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

	log, err := cfg.NewLogger()
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	log.Info("easyterms telegram starting",
		slog.String("log_level", cfg.LogLevel),
		slog.Bool("has_bot_token", cfg.TelegramBotToken != ""),
		slog.Bool("has_database_url", cfg.DatabaseURL != ""),
	)

	// Bot wiring (handlers, core services) is added in issue #11.
	<-ctx.Done()
	log.Info("shutdown signal received")
	return nil
}
