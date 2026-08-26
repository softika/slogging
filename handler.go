package slogging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Option configures a handler built by NewHandler.
type Option func(*options)

type options struct {
	writer     io.Writer
	level      slog.Leveler
	extractors []Extractor
}

// WithWriter sets the destination. Defaults to os.Stdout.
//
// Passing a buffer is what makes log output assertable in a test.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}

// WithLevel sets the minimum level. Defaults to slog.LevelInfo.
//
// It takes a slog.Leveler rather than a slog.Level, so passing a *slog.LevelVar
// allows the level to be changed at runtime without rebuilding the logger.
func WithLevel(l slog.Leveler) Option {
	return func(o *options) {
		o.level = l
	}
}

// WithExtractor registers Extractors, in the order given.
//
// Supplying any Extractor replaces the ContextIds default rather than adding to
// it, so a caller that wants both must name both:
//
//	slogging.WithExtractor(slogging.ContextIds, myExtractor)
//
// Repeated calls accumulate.
func WithExtractor(e ...Extractor) Option {
	return func(o *options) {
		o.extractors = append(o.extractors, e...)
	}
}

// NewHandler returns a JSON handler that stamps context values onto every
// record.
//
// Unlike Slogger, it is not a singleton: each call returns an independent
// handler, so tests can capture output and a process can run more than one.
func NewHandler(opts ...Option) slog.Handler {
	o := options{
		writer: os.Stdout,
		level:  slog.LevelInfo,
	}

	for _, opt := range opts {
		opt(&o)
	}

	if len(o.extractors) == 0 {
		o.extractors = []Extractor{ContextIds}
	}

	base := slog.NewJSONHandler(o.writer, &slog.HandlerOptions{Level: o.level})

	return &contextHandler{handler: base, extractors: o.extractors}
}

// contextHandler enriches records with context values before delegating.
type contextHandler struct {
	handler    slog.Handler
	extractors []Extractor
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, extract := range h.extractors {
		if attrs := extract(ctx); len(attrs) > 0 {
			r.AddAttrs(attrs...)
		}
	}

	return h.handler.Handle(ctx, r)
}

// WithAttrs wraps the derived handler rather than returning it bare.
//
// Returning h.handler.WithAttrs(attrs) directly would hand back an unwrapped
// handler, and every logger derived with slog.With would silently stop
// reporting request and correlation ids.
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{
		handler:    h.handler.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

// WithGroup wraps the derived handler, for the same reason as WithAttrs.
//
// Context attributes are added while the group is open, so slog qualifies them
// with the group name. That follows slog.Handler's contract: attributes added
// after WithGroup belong to that group.
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{
		handler:    h.handler.WithGroup(name),
		extractors: h.extractors,
	}
}
