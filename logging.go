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

func Slogger(handlers ...slog.Handler) *slog.Logger {
	once.Do(func() {
		logLevel := slog.LevelInfo

		env := os.Getenv("ENVIRONMENT")
		switch env {
		case "local", "development":
			logLevel = slog.LevelDebug
		case "production":
			logLevel = slog.LevelError
		}

		if len(handlers) > 0 {
			logger = slog.New(handlers[0])
			return
		} else {
			jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: logLevel,
			})
			logger = slog.New(newContextJsonHandler(jsonHandler))
		}

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
