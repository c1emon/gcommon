# server

`server` 在 [`service`](https://github.com/c1emon/gcommon/tree/main/service) 之上提供**多后台服务编排**：用 `errgroup` 并发执行各服务的 `Run`，通过根 `context` 统一触发停止，并支持关机前后钩子、超时与系统信号。

## 功能概览

| 能力 | 说明 |
|------|------|
| [`New`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/server#New) | 绑定 `ServiceRepo`、`*slog.Logger` 与可选 `serverOption`（如 `PreRunFunc`、`WithShutdownTimeout` 等），内部会 `Init` |
| [`Run`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/server#Server.Run) | `preRun` → 为每个已注册服务启动 goroutine 跑 `ServiceRunner.Run` → `postRun`，最后 `Wait`（进程常驻入口） |
| [`Shutdown`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/server#Server.Shutdown) | `preStop` → `cancel` 根上下文 → 等待 `Run` 侧收尾 → `postStop`；可重复调用但仅第一次生效（`sync.Once`） |
| [`ListenToSystemSignals`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/server#Server.ListenToSystemSignals) | 监听 `SIGINT` / `SIGTERM`（及 `SIGHUP` 占位），收到终止类信号后带超时调用 `Shutdown` |

关机时传给每个服务 `Stop` 的超时会略小于 `shutdownTimeout`，以便外层 `Shutdown` 的 `ctx` 先结束（见源码 `serviceStopTimeout`）。

## 构造与选项

```go
srv, err := server.New(repo, logger,
	server.WithShutdownTimeout(30*time.Second),
	server.PreRunFunc(func(ctx context.Context) error {
		// 在所有服务 Run 之前
		return nil
	}),
	server.PostRunFunc(func(ctx context.Context) error {
		// 在所有服务 goroutine 已启动之后（仍阻塞在 Wait 之前）
		return nil
	}),
	server.PreStopFunc(func(ctx context.Context) error {
		return nil
	}),
	server.PostStopFunc(func(ctx context.Context) error {
		return nil
	}),
)
```

- **`logger`**：用于生命周期与错误日志，应非 `nil`（与项目其余包一致，建议用 `logx` 或 `slog` 自建）。
- **`WithShutdownTimeout`**：小于 2 秒时会被抬到 **2 秒**。

## 推荐运行方式

推荐在**单独 goroutine**里调用 [`ListenToSystemSignals`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/server#Server.ListenToSystemSignals)：该方法会**阻塞**直到收到 `SIGINT` / `SIGTERM`，再按 `WithShutdownTimeout` 构造带超时的上下文并调用 `Shutdown`。主 goroutine 仍阻塞在 **`Run`** 上。

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/c1emon/gcommon/v2/server"
	"github.com/c1emon/gcommon/v2/service"
)

func main() {
	repo := service.NewServiceRepo()
	// repo.Register(...)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	srv, err := server.New(repo, logger, server.WithShutdownTimeout(15*time.Second))
	if err != nil {
		panic(err)
	}

	go func() {
		srv.ListenToSystemSignals(context.Background())
	}()

	if err := srv.Run(); err != nil {
		logger.Error("server run", "error", err)
		os.Exit(1)
	}
}
```

传给 `ListenToSystemSignals` 的 `ctx` 会作为 `Shutdown` 外层等待的父上下文（见 `handleShutdownSignal`）；一般用 `context.Background()` 即可。

### 备选：`signal.NotifyContext` + 手动 `Shutdown`

若需要与进程内其它逻辑共享同一 `ctx`、或自行组合多种信号，可自行 `signal.NotifyContext` 后在 `<-ctx.Done()` 里调用 `Shutdown`，行为与内置监听等价，但需自己保持**关机超时**与 `WithShutdownTimeout` 一致，避免重复或竞态。

## 与 `service` 包的关系

- 你只实现并注册 [`service.Service`](https://pkg.go.dev/github.com/c1emon/gcommon/v2/service#Service)。
- `server` 通过 `WrapDefault` 得到 `Run(ctx, timeout)`，在关机路径上驱动 `Stop`。

更细的接口说明见 [`service/README.md`](../service/README.md)。

## 仓库内参考示例

阻塞式集成 smoke：`test/http_svc_test.go`（`INTEGRATION=1 go test -v ./test -run Test_http_svc`）。
