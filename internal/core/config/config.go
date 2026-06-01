// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds runtime settings shared by all entrypoints.
type Config struct {
	// LogLevel is slog level name: debug, info, warn, error.
	LogLevel string

	// TelegramBotToken is required when running the Telegram bot (set in later issues).
	TelegramBotToken string

	// DatabaseURL is the PostgreSQL connection string (used from issue #3+).
	DatabaseURL string
}

// Load reads configuration from environment variables.
// If a .env file exists in the process working directory, it is loaded first
// (existing environment variables are not overwritten).
// Required vars for a given entrypoint are validated at startup in cmd, not here.
func Load() (Config, error) {
	loadDotEnv()
	cfg := Config{
		LogLevel:         envOrDefault("LOG_LEVEL", "info"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
	}
	return cfg, nil
}

// NewLogger builds a slog.Logger from cfg.LogLevel.
func (c Config) NewLogger() (*slog.Logger, error) {
	level, err := parseLogLevel(c.LogLevel)
	if err != nil {
		return nil, err
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler), nil
}

// loadDotEnv loads .env when present; missing file is ignored.
func loadDotEnv() {
	_ = godotenv.Load(".env")
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(name string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q", name)
	}
}
