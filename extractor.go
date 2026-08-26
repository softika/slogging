package slogging

import (
	"context"
	"log/slog"
)

// Extractor pulls attributes off a context so they can be stamped onto every
// log record.
//
// It is the extension point that keeps this package dependency free. Anything
// context-scoped -- a tenant, a tracing id, a feature flag -- is added by
// registering an Extractor with WithExtractor, rather than by teaching this
// package about it and releasing a new version.
//
// Returning nil is fine and costs nothing; an Extractor should never panic on
// a context that does not carry its value.
type Extractor func(context.Context) []slog.Attr

// contextIdKeys are the identity keys this package defines, in the order they
// are stamped onto a record.
var contextIdKeys = []key{
	RequestIdKey,
	CorrelationIdKey,
	UserIdKey,
	AccountIdKey,
	OrgIdKey,
}

// ContextIds extracts the identity values this package defines.
//
// It is the default Extractor, so a handler built with no WithExtractor option
// behaves the way this package always has.
func ContextIds(ctx context.Context) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(contextIdKeys))

	for _, k := range contextIdKeys {
		if v, ok := ctx.Value(k).(string); ok {
			attrs = append(attrs, slog.String(string(k), v))
		}
	}

	return attrs
}
