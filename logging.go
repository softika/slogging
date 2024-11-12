package slogging

import (
	"context"
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
func Slogger(h ...slog.Handler) *slog.Logger {
	once.Do(func() {
		if len(h) > 0 {
			logger = slog.New(h[0])
			return
		}

		logLevel := slog.LevelInfo

		env := os.Getenv("ENVIRONMENT")
		switch env {
		case "local", "development":
			logLevel = slog.LevelDebug
		case "production":
			logLevel = slog.LevelError
		}

		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
		logger = slog.New(newContextJsonHandler(jsonHandler))

	})
	return logger
}

type contextJsonHandler struct {
	handler *slog.JSONHandler
}

func newContextJsonHandler(handler *slog.JSONHandler) slog.Handler {
	return &contextJsonHandler{handler: handler}
}

func (h *contextJsonHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestId, ok := ctx.Value(RequestIdKey).(string); ok {
		r.AddAttrs(slog.String(string(RequestIdKey), requestId))
	}
	if correlationId, ok := ctx.Value(CorrelationIdKey).(string); ok {
		r.AddAttrs(slog.String(string(CorrelationIdKey), correlationId))
	}
	if userId, ok := ctx.Value(UserIdKey).(string); ok {
		r.AddAttrs(slog.String(string(UserIdKey), userId))
	}
	if accountId, ok := ctx.Value(AccountIdKey).(string); ok {
		r.AddAttrs(slog.String(string(AccountIdKey), accountId))
	}
	if orgId, ok := ctx.Value(OrgIdKey).(string); ok {
		r.AddAttrs(slog.String(string(OrgIdKey), orgId))
	}
	return h.handler.Handle(ctx, r)
}

func (h *contextJsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.handler.WithAttrs(attrs)
}

func (h *contextJsonHandler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}

func (h *contextJsonHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}
