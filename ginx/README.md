# ginx

`ginx` provides a small integration layer around Gin with a fixed default middleware pipeline and `service.Service` adapter.

## Logging Convention

- Pass a non-nil `*slog.Logger` into `Logger(logger)`, `Recovery(logger)`, and `DefaultEngineConfig.Logger` (used by `NewDefaultEngine`). These APIs do not call `logx.Default()` implicitly.
- You may still use `logx` to build that logger, for example:

```go
logx.Init(logx.Config{
	Format: logx.FormatJSON,
	Level:  slog.LevelInfo,
})
logger := logx.Default()
```

Or construct a `*slog.Logger` with `log/slog` only; either way the logger is supplied explicitly at call sites.

### Debug Logging Notes

At `debug` level, `ginx.Logger(logger)` logs extra request/response details:

- full request headers
- full response headers
- request and response body (truncated to `4096` bytes)

Use debug logging only in trusted environments, because headers/body can include sensitive data.

## Engine APIs

- `NewBareEngine()` creates a bare `*gin.Engine`.
- `NewDefaultEngine(DefaultEngineConfig)` creates an engine with:
  - `Logger(cfg.Logger)`
  - `ErrorResponder()`
  - `Recovery(cfg.Logger)`
  - plus `DefaultEngineConfig.Middlewares` in-order (`cfg.Logger` must be non-nil).
  - It also calls [`SetGinSlogWriters`](https://pkg.go.dev/github.com/c1emon/gcommon/ginx#SetGinSlogWriters) so [`gin.DefaultWriter`](https://pkg.go.dev/github.com/gin-gonic/gin#DefaultWriter) and [`gin.DefaultErrorWriter`](https://pkg.go.dev/github.com/gin-gonic/gin#DefaultErrorWriter) emit **one slog record per line** (info vs error) instead of raw stdout/stderr.
- `SetGinSlogWriters(logger *slog.Logger)` assigns those Gin package globals; use it yourself if you build engines with [`NewBareEngine`](https://pkg.go.dev/github.com/c1emon/gcommon/ginx#NewBareEngine) but still want Gin’s internal/debug output routed through the same logger.

### Example: Build a Default Engine

```go
import (
	"log/slog"

	"github.com/c1emon/gcommon/logx"
	"github.com/c1emon/gcommon/ginx"
)

logx.Init(logx.Config{
	Format: logx.FormatText,
	Level:  slog.LevelInfo,
})

logger := logx.Default()

r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{Logger: logger})
r.GET("/healthz", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
})
```

## HTTP Service API

Use `NewHTTPService(HTTPServiceConfig)` to wrap an `*http.Server` as `service.Service`. You must set `Logger` (non-nil); the constructor does not fall back to `logx.Default()`.

```go
import "log/slog"

logger := slog.Default()

r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{Logger: logger})
r.GET("/healthz", func(c *gin.Context) {
	c.String(http.StatusOK, "ok")
})

svc := ginx.NewHTTPService(ginx.HTTPServiceConfig{
	Name:   "api-http",
	Logger: logger,
	Server: &http.Server{
		Addr:    ":8080",
		Handler: r,
	},
})
```

`svc.ServeErrors()` can be used to observe async `Serve` errors.

### Example: Create Service Directly From Gin Engine

Use `NewGinService(GinServiceConfig)` when you only want to provide `addr + engine` (and a non-nil `Logger`):

```go
import "log/slog"

logger := slog.Default()

r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{Logger: logger})
r.GET("/healthz", func(c *gin.Context) {
	c.String(http.StatusOK, "ok")
})

svc := ginx.NewGinService(ginx.GinServiceConfig{
	Name:   "api-http",
	Addr:   ":8080",
	Engine: r,
	Logger: logger,
})
```

### Example: Observe Async Serve Errors

```go
go func() {
	if err := <-svc.ServeErrors(); err != nil {
		logx.Default().Error("http serve exited", logx.Err(err))
	}
}()
```
