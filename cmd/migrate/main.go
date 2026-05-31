// Command migrate applies database migrations (up or down).
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/tikhomirovv/easyterms/internal/core/config"
	"github.com/tikhomirovv/easyterms/internal/storage/migrate"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	log, err := cfg.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	slog.SetDefault(log)

	switch *direction {
	case "up":
		err = migrate.Up(cfg.DatabaseURL)
	case "down":
		err = migrate.Down(cfg.DatabaseURL)
	default:
		fmt.Fprintf(os.Stderr, "unknown direction %q\n", *direction)
		os.Exit(1)
	}
	if err != nil {
		slog.Error("migration failed", slog.String("direction", *direction), slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("migration complete", slog.String("direction", *direction))
}
