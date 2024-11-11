package slogging_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/softika/slogging"
)

func TestSloggerSingleton(t *testing.T) {
	t.Parallel()

	logger1 := slogging.Slogger()
	logger2 := slogging.Slogger()

	if logger1 != logger2 {
		t.Errorf("expected logger1 to be equal to logger2")
	}

	logger3 := slogging.Slogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	if logger1 != logger3 {
		t.Errorf("expected logger1 to be equal to logger3")
	}
}

func TestSloggerEnvironment(t *testing.T) {
	t.Parallel()

	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	logger := slogging.Slogger()

	logger.Info("application info", slog.String("module", "logging"))

	logger.Debug("application debug", slog.String("module", "logging"))

	logger.Warn("application warning", slog.String("module", "logging"))

	ctx := context.WithValue(context.Background(), slogging.CorrelationIdKey, "e1156cc4-57bf-4b09-926f-1112b8f89a03")
	logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))

	ctx = context.WithValue(ctx, slogging.UserIdKey, "425b0d3a-b3b6-4e21-882d-1ca94c798df0")
	logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))
}

func TestSlogger(t *testing.T) {
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
