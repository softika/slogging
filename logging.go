package slogging

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

const (
	requestIdKey     = "request_id"
	correlationIdKey = "correlation_id"
	userIdKey        = "user_id"
	accountIdKey     = "account_id"
	orgIdKey         = "org_id"
)

var (
	logger *slog.Logger

	once sync.Once
)

func Slogger() *slog.Logger {
	once.Do(func() {
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
	if requestId, ok := ctx.Value(requestIdKey).(string); ok {
		r.AddAttrs(slog.String(requestIdKey, requestId))
	}
	if correlationId, ok := ctx.Value(correlationIdKey).(string); ok {
		r.AddAttrs(slog.String(correlationIdKey, correlationId))
	}
	if userId, ok := ctx.Value(userIdKey).(string); ok {
		r.AddAttrs(slog.String(userIdKey, userId))
	}
	if accountId, ok := ctx.Value(accountIdKey).(string); ok {
		r.AddAttrs(slog.String(accountIdKey, accountId))
	}
	if orgId, ok := ctx.Value(orgIdKey).(string); ok {
		r.AddAttrs(slog.String(orgIdKey, orgId))
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
