package slogging

import (
	"log/slog"
	"os"
	"sync"
)

type key string

const (
	RequestIdKey     key = "X-Request-Id"
	CorrelationIdKey key = "X-Correlation-Id"
	UserIdKey        key = "X-User-Id"
	AccountIdKey     key = "X-Account-Id"
	OrgIdKey         key = "X-Org-Id"
)

var (
	logger *slog.Logger

	once sync.Once
)

// Slogger initializes or retrieves a singleton instance of slog.Logger with a structured JSONHandler.
// By default, it configures the log level based on the ENVIRONMENT variable; if ENVIRONMENT is unset,
// it defaults to INFO level.
// This function supports injecting a custom handler on the first call, allowing for flexible logging configurations.
//
// Example usage with a custom handler:
//
//	logger := slogging.Slogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
//	      Level: slog.LevelError,
//	}))
//
// Slogger is a process-wide singleton: the first call wins and later calls
// return that same logger, whatever arguments they pass. Prefer NewHandler when
// you need an independent logger, want the level to come from configuration
// rather than the environment, or need to assert on output in a test.
func Slogger(h ...slog.Handler) *slog.Logger {
	once.Do(func() {
		if len(h) > 0 {
			logger = slog.New(h[0])
			return
		}

		logger = slog.New(NewHandler(WithLevel(levelFromEnv())))
	})
	return logger
}

// levelFromEnv reports the log level implied by the ENVIRONMENT variable.
//
// Note that production maps to Error, which discards Info and Warn records.
// This is long-standing behaviour and is kept for compatibility; build a
// handler with NewHandler and WithLevel to choose the level yourself.
func levelFromEnv() slog.Level {
	switch os.Getenv("ENVIRONMENT") {
	case "local", "development":
		return slog.LevelDebug
	case "production":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
