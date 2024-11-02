package slogging

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type key string

const (
	requestIdKey     key = "request_id"
	correlationIdKey key = "correlation_id"
	userIdKey        key = "user_id"
	accountIdKey     key = "account_id"
	orgIdKey         key = "org_id"
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

type options struct {
	handler slog.Handler
}

type Option func(*options)

func WithHandler(handler slog.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}

type contextJsonHandler struct {
	handler *slog.JSONHandler
}

func newContextJsonHandler(handler *slog.JSONHandler) slog.Handler {
	return &contextJsonHandler{handler: handler}
}

func (h *contextJsonHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestId, ok := ctx.Value(requestIdKey).(string); ok {
		r.AddAttrs(slog.String(string(requestIdKey), requestId))
	}
	if correlationId, ok := ctx.Value(correlationIdKey).(string); ok {
		r.AddAttrs(slog.String(string(correlationIdKey), correlationId))
	}
	if userId, ok := ctx.Value(userIdKey).(string); ok {
		r.AddAttrs(slog.String(string(userIdKey), userId))
	}
	if accountId, ok := ctx.Value(accountIdKey).(string); ok {
		r.AddAttrs(slog.String(string(accountIdKey), accountId))
	}
	if orgId, ok := ctx.Value(orgIdKey).(string); ok {
		r.AddAttrs(slog.String(string(orgIdKey), orgId))
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
