// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// DefaultDatabasePath is used when DATABASE_PATH is unset.
const DefaultDatabasePath = "data/easyterms.db"

// Config holds runtime settings shared by all entrypoints.
type Config struct {
	LogLevel           string
	TelegramBotToken   string
	TelegramProxy      string
	DatabasePath       string
	AllowedTelegramIDs []int64
}

// Load reads configuration from environment variables.
// If a .env file exists in the process working directory, it is loaded first
// (existing environment variables are not overwritten).
func Load() (Config, error) {
	loadDotEnv()
	cfg := Config{
		LogLevel:         envOrDefault("LOG_LEVEL", "info"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramProxy:    strings.TrimSpace(os.Getenv("TELEGRAM_PROXY")),
		DatabasePath:     envOrDefault("DATABASE_PATH", DefaultDatabasePath),
	}
	ids, err := parseTelegramIDs(os.Getenv("ALLOWED_TELEGRAM_IDS"))
	if err != nil {
		return Config{}, err
	}
	cfg.AllowedTelegramIDs = ids
	return cfg, nil
}

// AllowlistOpen reports whether the bot accepts any Telegram user (no allowlist configured).
func (c Config) AllowlistOpen() bool {
	return len(c.AllowedTelegramIDs) == 0
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

func loadDotEnv() {
	_ = godotenv.Load(".env")
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parseTelegramIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var ids []int64
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ALLOWED_TELEGRAM_IDS: invalid id %q", part)
		}
		ids = append(ids, id)
	}
	return ids, nil
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
