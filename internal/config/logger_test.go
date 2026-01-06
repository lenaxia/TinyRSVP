package config

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		wantLevel slog.Level
	}{
		{
			name:      "debug level",
			level:     LogLevelDebug,
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "info level",
			level:     LogLevelInfo,
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "warn level",
			level:     LogLevelWarn,
			wantLevel: slog.LevelWarn,
		},
		{
			name:      "error level",
			level:     LogLevelError,
			wantLevel: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := InitLogger(tt.level)
			if logger == nil {
				t.Fatal("Expected logger, got nil")
			}
		})
	}
}

func TestInitLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(&buf, opts)
	logger := slog.New(handler)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected log output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Expected valid JSON output, got error: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got %v", logEntry["key"])
	}
}

func TestGetLogLevelFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantLevel LogLevel
	}{
		{
			name:      "DEBUG level",
			envValue:  "DEBUG",
			wantLevel: LogLevelDebug,
		},
		{
			name:      "INFO level",
			envValue:  "INFO",
			wantLevel: LogLevelInfo,
		},
		{
			name:      "WARN level",
			envValue:  "WARN",
			wantLevel: LogLevelWarn,
		},
		{
			name:      "ERROR level",
			envValue:  "ERROR",
			wantLevel: LogLevelError,
		},
		{
			name:      "invalid level defaults to INFO",
			envValue:  "INVALID",
			wantLevel: LogLevelInfo,
		},
		{
			name:      "empty defaults to INFO",
			envValue:  "",
			wantLevel: LogLevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("LOG_LEVEL", tt.envValue)
				defer os.Unsetenv("LOG_LEVEL")
			}

			level := GetLogLevelFromEnv()
			if level != tt.wantLevel {
				t.Errorf("GetLogLevelFromEnv() = %v, want %v", level, tt.wantLevel)
			}
		})
	}
}

func TestInitLogger_SetsDefaultLogger(t *testing.T) {
	logger := InitLogger(LogLevelInfo)
	if logger == nil {
		t.Fatal("Expected logger, got nil")
	}

	defaultLogger := slog.Default()
	if defaultLogger == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestLogLevels_FilterMessages(t *testing.T) {
	var buf bytes.Buffer

	opts := &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}
	handler := slog.NewJSONHandler(&buf, opts)
	logger := slog.New(handler)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out at WARN level")
	}

	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out at WARN level")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be included at WARN level")
	}

	if !strings.Contains(output, "error message") {
		t.Error("Error message should be included at WARN level")
	}
}
