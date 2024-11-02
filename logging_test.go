package slogging_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/softika/slogging"
)

func TestLoggerEnvironment(t *testing.T) {
	t.Parallel()

	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	logger := slogging.Slogger()

	logger.Info("application info", slog.String("module", "logging"))

	logger.Debug("application debug", slog.String("module", "logging"))

	logger.Warn("application warning", slog.String("module", "logging"))

	ctx := context.WithValue(context.Background(), "correlation_id", "unique_id_value")
	logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))

	ctx = context.WithValue(ctx, "user_id", "unique_id_value")
	logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))
}

func TestLogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
	}{
		{
			name:  "log info",
			level: slog.LevelInfo,
		},
		{
			name:  "log debug",
			level: slog.LevelDebug,
		},
		{
			name:  "log warn",
			level: slog.LevelWarn,
		},
		{
			name:  "log error",
			level: slog.LevelError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := slogging.Slogger()
			if got == nil {
				t.Fatalf("Logger() = nil; want not nil")
			}

			got.WithGroup("logger test")
			switch tc.level {
			case slog.LevelInfo:
				got.Info("test info")
			case slog.LevelDebug:
				got.Debug("test debug")
			case slog.LevelWarn:
				got.Warn("test warn")
			case slog.LevelError:
				err := errors.New("test error message")
				got.Error("general message", "error", err)
			}
		})
	}
}

func TestLoggerWithHandler(t *testing.T) {
	t.Parallel()

	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slogging.Slogger(handler)

	logger.Error("error message", "error", errors.New("error details"))
}
