# logx

`logx` is a pure `slog` helper package for consistent logger setup.

## Features

- Pure stdlib logger handlers (`slog.NewJSONHandler` / `slog.NewTextHandler`)
- Unified logger initialization with `Config`
- Package-level default logger management (`Init` / `Default`)
- Context attrs propagation (`WithAttrs` / `FromContext`)
- Fixed attrs injection via handler (`Config.FixedAttrs`)
- Common attrs keys and constructors (`TraceID`, `UserID`, `Err`, etc.)

## Quick Start

```go
package main

import (
	"context"
	"log/slog"

	"github.com/c1emon/gcommon/logx/v2"
)

func main() {
	logx.Init(logx.Config{
		Format:    logx.FormatJSON,
		Level:     slog.LevelInfo,
		AddSource: false,
		FixedAttrs: []slog.Attr{
			slog.String("service", "order-service"),
			slog.String("version", "v1.0.0"),
		},
	})

	ctx := logx.WithAttrs(context.Background(),
		logx.TraceID("trace-123"),
		logx.UserID("user-42"),
	)

	logx.FromContext(ctx).InfoContext(ctx, "request completed",
		logx.Method("GET"),
		logx.Path("/healthz"),
		logx.Status(200),
		logx.DurationMs(12),
	)
}
```

## Level Parsing

```go
lv, err := logx.ParseLogLevel("info")
if err != nil {
	// handle invalid level input
}
_ = lv
```

## Usage Convention

- Prefer `Default().InfoContext(ctx, ...)` in hot paths where you already have `ctx`.
- Use `FromContext(ctx)` when you want call sites to read as "logger from request context".
- Always pass the same `ctx` to `InfoContext/ErrorContext` so handler-level context attrs are injected.

```go
func Handle(ctx context.Context) {
	ctx = logx.WithAttrs(ctx, logx.TraceID("trace-abc"))

	// Preferred direct style
	logx.Default().InfoContext(ctx, "handle started")

	// Equivalent explicit style
	logx.FromContext(ctx).InfoContext(ctx, "handle finished")
}
```
