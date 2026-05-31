package config

import (
	"log/slog"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
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

func TestNewLogger_invalid(t *testing.T) {
	cfg := Config{LogLevel: "verbose"}
	if _, err := cfg.NewLogger(); err == nil {
		t.Fatal("expected error for unknown log level")
	}
}
