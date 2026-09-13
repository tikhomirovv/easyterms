package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_PATH", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.DatabasePath != DefaultDatabasePath {
		t.Errorf("DatabasePath = %q", cfg.DatabasePath)
	}
}

func TestNewLogger_levels(t *testing.T) {
	tests := []struct {
		level string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cfg := Config{LogLevel: tt.level}
			_, err := cfg.NewLogger()
			if err != nil {
				t.Fatalf("NewLogger: %v", err)
			}
		})
	}
}

func TestLoad_fromDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	const wantDB = "data/test.db"
	content := "LOG_LEVEL=debug\nDATABASE_PATH=" + wantDB + "\nALLOWED_TELEGRAM_IDS=42,99\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("ALLOWED_TELEGRAM_IDS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.DatabasePath != wantDB {
		t.Errorf("DatabasePath = %q", cfg.DatabasePath)
	}
	if len(cfg.AllowedTelegramIDs) != 2 || cfg.AllowedTelegramIDs[0] != 42 {
		t.Fatalf("ids = %v", cfg.AllowedTelegramIDs)
	}
}

func TestNewLogger_invalid(t *testing.T) {
	cfg := Config{LogLevel: "verbose"}
	if _, err := cfg.NewLogger(); err == nil {
		t.Fatal("expected error for unknown log level")
	}
}
