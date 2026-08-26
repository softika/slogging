![go workflow](https://github.com/softika/slogging/actions/workflows/test.yml/badge.svg)
![lint workflow](https://github.com/softika/slogging/actions/workflows/lint.yml/badge.svg)
![security workflow](https://github.com/softika/slogging/actions/workflows/security.yml/badge.svg)

# Logging Library

A **zero dependency** structured logger built on `log/slog`. It emits JSON and
stamps context-scoped values -- request ids, correlation ids, user ids -- onto
every record, so logs from one request can be tied together across layers and
across systems.

## Features

- JSON-formatted logs for structured logging.
- Context values stamped onto every record, including on loggers derived with
  `With` and `WithGroup`.
- Pluggable extractors, so anything on the context can be logged without
  changing this package.
- Level from configuration (`NewHandler`) or from the environment (`Slogger`).
- Optional singleton, for applications that want one global logger.

## Installation

```bash
go get github.com/softika/slogging
```

Requires **Go 1.27** or later.

## Usage

There are two entry points. Use `NewHandler` unless you specifically want a
process-wide singleton.

### `NewHandler` -- configurable, independent

```go
package main

import (
	"context"
	"log/slog"

	"github.com/softika/slogging"
)

func main() {
	slog.SetDefault(slog.New(slogging.NewHandler(
		slogging.WithLevel(slog.LevelInfo),
	)))

	ctx := context.WithValue(context.Background(), slogging.CorrelationIdKey, "unique_id_value")

	slog.InfoContext(ctx, "application info", slog.String("module", "logging"))
}
```

```json
{"time":"2024-11-02T22:39:45.732646+01:00","level":"INFO","msg":"application info","module":"logging","X-Correlation-Id":"unique_id_value"}
```

Options:

| Option | Default | Purpose |
| --- | --- | --- |
| `WithWriter(io.Writer)` | `os.Stdout` | Destination. Pass a buffer to assert on output in tests |
| `WithLevel(slog.Leveler)` | `slog.LevelInfo` | Minimum level. Accepts `*slog.LevelVar`, so the level can change at runtime |
| `WithExtractor(...Extractor)` | `ContextIds` | Which context values to stamp onto records |

Each call returns an independent handler, so a process can build more than one
and a test can capture its output.

### `Slogger` -- the singleton

```go
logger := slogging.Slogger()
slog.SetDefault(logger)
```

The first call wins; later calls return that same logger whatever arguments they
pass. Its level comes from the `ENVIRONMENT` variable:

| `ENVIRONMENT` | Level |
| --- | --- |
| `local`, `development` | Debug |
| `production` | Error |
| unset or anything else | Info |

> **Note:** `production` maps to Error, which discards Info and Warn records --
> including startup, shutdown and most diagnostics. This is long-standing
> behaviour, kept for compatibility. To choose the level yourself, build a
> handler with `NewHandler` and `WithLevel`.

## Context values

Set values using the exported key constants:

```go
ctx = context.WithValue(ctx, slogging.RequestIdKey, "unique_id_value")
```

Use the constants, **not** a plain string. `context.Value` matches on the key's
dynamic type as well as its value, and these keys have an unexported type -- so
`context.WithValue(ctx, "X-Request-Id", v)` silently never matches and the value
never reaches your logs.

The keys stamped by default:

| Constant | Log field |
| --- | --- |
| `RequestIdKey` | `X-Request-Id` |
| `CorrelationIdKey` | `X-Correlation-Id` |
| `UserIdKey` | `X-User-Id` |
| `AccountIdKey` | `X-Account-Id` |
| `OrgIdKey` | `X-Org-Id` |

## Custom extractors

An `Extractor` reads a context and returns attributes:

```go
type Extractor func(context.Context) []slog.Attr
```

This is how anything context-scoped gets logged without this package having to
know about it -- a tenant, a feature flag, or a tracing id:

```go
func traceAttrs(ctx context.Context) []slog.Attr {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []slog.Attr{
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	}
}

handler := slogging.NewHandler(
	slogging.WithExtractor(slogging.ContextIds, traceAttrs),
)
```

Supplying any extractor **replaces** the `ContextIds` default rather than adding
to it, so name both when you want both. Keeping the OpenTelemetry import on the
caller's side is what lets this package stay dependency free.

## Custom handler

`Slogger` also accepts a `slog.Handler` on its first call, replacing the JSON
handler entirely:

```go
handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelError,
})

logger := slogging.Slogger(handler)
logger.Error("error message", "error", errors.New("error details"))
```

```
time=2024-11-02T22:59:19.256+01:00 level=ERROR msg="error message" error="error details"
```

Note that an injected handler is used as given: context extraction is not
applied to it, since that is the wrapping this package would otherwise provide.

## A note on groups

Attributes added while a group is open are qualified by that group, which is
what `slog.Handler`'s contract requires. Context values are stamped as the
record is handled, so under `WithGroup("db")` they appear as
`{"db":{"X-Request-Id":"..."}}`. Use `With` rather than `WithGroup` if you need
these ids to stay at the top level.
