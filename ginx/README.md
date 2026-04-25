# ginx

`ginx` provides a small integration layer around Gin with a fixed default middleware pipeline and `service.Service` adapter.

## Logging Convention

- `ginx` logs always use `logx.Default()`.
- Initialize global logging once at startup:

```go
logx.Init(logx.Config{
	Format: logx.FormatJSON,
	Level:  slog.LevelInfo,
})
```

If `logx.Init(...)` is not called, `logx.Default()` falls back to `slog.Default()`.

### Debug Logging Notes

At `debug` level, `ginx.Logger()` logs extra request/response details:

- full request headers
- full response headers
- request and response body (truncated to `4096` bytes)

Use debug logging only in trusted environments, because headers/body can include sensitive data.

## Engine APIs

- `NewBareEngine()` creates a bare `*gin.Engine`.
- `NewDefaultEngine(DefaultEngineConfig)` creates an engine with:
  - `Logger()`
  - `ErrorResponder()`
  - `Recovery()`
  - plus `DefaultEngineConfig.Middlewares` in-order.

### Example: Build a Default Engine

```go
logx.Init(logx.Config{
	Format: logx.FormatText,
	Level:  slog.LevelInfo,
})

r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{})
r.GET("/healthz", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
})
```

## HTTP Service API

Use `NewHTTPService(HTTPServiceConfig)` to wrap an `*http.Server` as `service.Service`.

```go
r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{})
r.GET("/healthz", func(c *gin.Context) {
	c.String(http.StatusOK, "ok")
})

svc := ginx.NewHTTPService(ginx.HTTPServiceConfig{
	Name: "api-http",
	Server: &http.Server{
		Addr:    ":8080",
		Handler: r,
	},
})
```

`svc.ServeErrors()` can be used to observe async `Serve` errors.

### Example: Create Service Directly From Gin Engine

Use `NewGinService(GinServiceConfig)` when you only want to provide `addr + engine`:

```go
r := ginx.NewDefaultEngine(ginx.DefaultEngineConfig{})
r.GET("/healthz", func(c *gin.Context) {
	c.String(http.StatusOK, "ok")
})

svc := ginx.NewGinService(ginx.GinServiceConfig{
	Name:   "api-http",
	Addr:   ":8080",
	Engine: r,
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
