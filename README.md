![go workflow](https://github.com/softika/slogging/actions/workflows/test.yml/badge.svg)
![lint workflow](https://github.com/softika/slogging/actions/workflows/lint.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/softika/slogging)](https://goreportcard.com/report/github.com/softika/slogging)

# Logging Library

This package provides a singleton logger using the `slog` package, configured with different logging levels based on the application environment. 
By default, it outputs logs in JSON format. Optionally, you can inject your own `slog.Handler` to the `slogger` to change the format of log entries.


## Features

- JSON-formatted logs for structured logging.
- Configurable log level based on the `ENVIRONMENT` variable.
    - `local`, `development`: Debug level.
    - `production`: Error level.
    - Default: Info level.
- Singleton logger instance to ensure only one logger is created.
- Logging with context.
- Custom `slog.Handler` injection.

## Installation

```bash
go get github.com/softika/logging
```

## Usage

### 1. Import the Package

Import the `slogging` package in your Go code:

```go
import "github.com/softika/slogging"
```

### 2. Configure the Environment Variable

Set the `ENVIRONMENT` environment variable to control the log level:

```bash
export ENVIRONMENT=development  # Options: local, development, production
```

### 3. Use the Logger

Retrieve the singleton logger instance by calling `Slogger()`, and use it for logging messages.

```go
package main

import (
    "log/slog"
    "context"
    "errors"
	
    "github.com/softika/slogging"
)

func main() {
    // Get the logger instance
    logger := slogging.Slogger()

    // Log an Info message
    logger.Info("application info", slog.String("module", "logging"))

    // Log a Debug message (only in local or development environments)
    logger.Debug("application debug", slog.String("module", "logging"))

    // Log a Warning message
    logger.Warn("application warning", slog.String("module", "logging"))
    
    // Log error with context
    ctx := context.WithValue(context.Background(), "correlation_id", "unique_id_value")
    logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))

    ctx = context.WithValue(ctx, "user_id", "unique_id_value")
    logger.ErrorContext(ctx, "error message", "error", errors.New("error details"))
}
```

## Example Output

With `ENVIRONMENT=development`, the output for the above logs would look like:

```json
{"time":"2024-11-02T22:39:45.732646+01:00","level":"INFO","msg":"application info","module":"logging"}
{"time":"2024-11-02T22:39:45.732812+01:00","level":"DEBUG","msg":"application debug","module":"logging"}
{"time":"2024-11-02T22:39:45.732815+01:00","level":"WARN","msg":"application warning","module":"logging"}
{"time":"2024-11-02T22:39:45.732818+01:00","level":"ERROR","msg":"error message","error":"error details","correlation_id":"unique_id_value"}
{"time":"2024-11-02T22:39:45.732823+01:00","level":"ERROR","msg":"error message","error":"error details","correlation_id":"unique_id_value","user_id":"unique_id_value"}
```

In production, only `ERROR` level messages will appear in the output.

## Custom Handler

You can inject your own `slog.Handler` to the logger instance. For example, you can use the `slog.TextHandler` to output logs in text format:

```go
package main

import (
    "errors"
    "log/slog"
    "os"
    
    "github.com/softika/slogging"
)

func main() {
    // create new slog.Handler	
    handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelError,
    })
	
    // init logger with custom handler
    logger := slogging.Slogger(handler)

    // log error
    logger.Error("error message", "error", errors.New("error details"))
}
```

And the output would be:

```
time=2024-11-02T22:59:19.256+01:00 level=ERROR msg="error message" error="error details"
```