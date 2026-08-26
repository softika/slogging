package slogging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/softika/slogging"
)

// decode returns the single JSON record written to buf.
func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid json: %v\nraw: %s", err, buf.String())
	}

	return got
}

// ctxWithIds returns a context carrying the request and correlation ids.
func ctxWithIds() context.Context {
	ctx := context.WithValue(context.Background(), slogging.RequestIdKey, "req-1")
	return context.WithValue(ctx, slogging.CorrelationIdKey, "corr-1")
}

func TestNewHandlerIncludesContextIds(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(slogging.WithWriter(buf)))

	logger.InfoContext(ctxWithIds(), "hello")

	got := decode(t, buf)
	if got["X-Request-Id"] != "req-1" {
		t.Errorf("X-Request-Id = %v; want req-1", got["X-Request-Id"])
	}
	if got["X-Correlation-Id"] != "corr-1" {
		t.Errorf("X-Correlation-Id = %v; want corr-1", got["X-Correlation-Id"])
	}
}

// TestWithAttrsPreservesContextIds pins the regression that made a derived
// logger silently drop context enrichment.
func TestWithAttrsPreservesContextIds(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(slogging.WithWriter(buf))).With("component", "db")

	logger.InfoContext(ctxWithIds(), "hello")

	got := decode(t, buf)
	if got["component"] != "db" {
		t.Errorf("component = %v; want db", got["component"])
	}
	if got["X-Request-Id"] != "req-1" {
		t.Errorf("X-Request-Id = %v; want req-1 (derived logger dropped context)", got["X-Request-Id"])
	}
}

// TestWithGroupPreservesContextIds covers the second half of the same bug.
//
// Attributes added while a group is open are qualified by that group, which is
// what slog.Handler's contract requires, so the ids are asserted inside it.
func TestWithGroupPreservesContextIds(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(slogging.WithWriter(buf))).WithGroup("db")

	logger.InfoContext(ctxWithIds(), "hello")

	got := decode(t, buf)
	group, ok := got["db"].(map[string]any)
	if !ok {
		t.Fatalf("db group missing; got %v", got)
	}
	if group["X-Request-Id"] != "req-1" {
		t.Errorf("db.X-Request-Id = %v; want req-1 (grouped logger dropped context)", group["X-Request-Id"])
	}
}

func TestWithExtractorReplacesDefault(t *testing.T) {
	t.Parallel()

	tenant := func(ctx context.Context) []slog.Attr {
		v, ok := ctx.Value(tenantKey{}).(string)
		if !ok {
			return nil
		}
		return []slog.Attr{slog.String("tenant", v)}
	}

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(
		slogging.WithWriter(buf),
		slogging.WithExtractor(tenant),
	))

	ctx := context.WithValue(ctxWithIds(), tenantKey{}, "acme")
	logger.InfoContext(ctx, "hello")

	got := decode(t, buf)
	if got["tenant"] != "acme" {
		t.Errorf("tenant = %v; want acme", got["tenant"])
	}
	if _, found := got["X-Request-Id"]; found {
		t.Error("X-Request-Id present; an explicit extractor must replace the default")
	}
}

type tenantKey struct{}

func TestExtractorsCompose(t *testing.T) {
	t.Parallel()

	extra := func(context.Context) []slog.Attr {
		return []slog.Attr{slog.String("extra", "yes")}
	}

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(
		slogging.WithWriter(buf),
		slogging.WithExtractor(slogging.ContextIds, extra),
	))

	logger.InfoContext(ctxWithIds(), "hello")

	got := decode(t, buf)
	if got["X-Request-Id"] != "req-1" {
		t.Errorf("X-Request-Id = %v; want req-1", got["X-Request-Id"])
	}
	if got["extra"] != "yes" {
		t.Errorf("extra = %v; want yes", got["extra"])
	}
}

func TestWithLevelFilters(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(
		slogging.WithWriter(buf),
		slogging.WithLevel(slog.LevelError),
	))

	logger.Info("dropped")

	if buf.Len() != 0 {
		t.Errorf("info record emitted at error level: %s", buf.String())
	}
}

// TestWithLevelIsDynamic proves WithLevel accepts a slog.Leveler, so the level
// can be changed at runtime without rebuilding the logger.
func TestWithLevelIsDynamic(t *testing.T) {
	t.Parallel()

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelError)

	buf := new(bytes.Buffer)
	logger := slog.New(slogging.NewHandler(
		slogging.WithWriter(buf),
		slogging.WithLevel(lvl),
	))

	logger.Info("dropped")
	if buf.Len() != 0 {
		t.Fatalf("info record emitted at error level: %s", buf.String())
	}

	lvl.Set(slog.LevelInfo)

	logger.Info("kept")
	if buf.Len() == 0 {
		t.Error("info record dropped after level was lowered")
	}
}

// TestNewHandlerIsIndependent guards the property that makes the handler
// testable: unlike Slogger, it is not a process-wide singleton.
func TestNewHandlerIsIndependent(t *testing.T) {
	t.Parallel()

	first, second := new(bytes.Buffer), new(bytes.Buffer)

	slog.New(slogging.NewHandler(slogging.WithWriter(first))).Info("one")
	slog.New(slogging.NewHandler(slogging.WithWriter(second))).Info("two")

	if first.Len() == 0 || second.Len() == 0 {
		t.Fatal("both handlers should have written independently")
	}
	if bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Error("handlers wrote identical output; they are not independent")
	}
}
