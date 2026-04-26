# service

`service` 提供进程内后台服务的统一生命周期抽象：**注册 → 启动 → 在上下文取消后带超时的停止**。  
根模块里的 [`server`](https://github.com/c1emon/gcommon/tree/main/server) 会消费 `ServiceRepo`，按 `errgroup` 并发跑多个服务，并在关机时取消上下文、触发 `Stop`。

## 功能概览

| 能力 | 说明 |
|------|------|
| [`Service`](https://pkg.go.dev/github.com/c1emon/gcommon/service#Service) | 三个方法：`Name()`、`Start()`（同步启动）、`Stop(ctx)`（优雅停止） |
| [`ServiceRepo`](https://pkg.go.dev/github.com/c1emon/gcommon/service#ServiceRepo) | 线程安全的注册表，`Register` 追加服务 |
| [`ServiceRunner`](https://pkg.go.dev/github.com/c1emon/gcommon/service#ServiceRunner) | `Run(ctx, timeout)`：在 `ctx` 取消后用给定超时调用 `Stop`；由 `WrapDefault` 包装 `Service` 得到 |

典型用法：实现 `Service`（例如包装 `http.Server`、队列消费者等），`Register` 进 repo，再交给 `server.New` 驱动（见下方「与 server 配合」）。

## 核心类型

### `Service`

```go
type Service interface {
	Name() string
	Start() error
	Stop(timeOutCtx context.Context) error
}
```

- **`Start`**：应快速返回；长时间运行的逻辑放在自己的 goroutine 里（例如 `http.Server.Serve`）。
- **`Stop`**：在 `timeOutCtx` 截止前完成释放（例如 `http.Server.Shutdown`）。

### `ServiceRepo`

```go
repo := service.NewServiceRepo()
repo.Register(mySvc)
```

`Register` 顺序即后续 `Services()` 的迭代顺序（实现细节以当前代码为准）。

### `WrapDefault` 与 `Run`

`server` 通过 `Services()` 拿到的是已包装好的 [`ServiceRunner`](https://pkg.go.dev/github.com/c1emon/gcommon/service#ServiceRunner)。`Run` 的行为要点：

1. 同步调用 `Start()`；失败则返回错误。
2. 在后台等待 `ctx.Done()`（由上层 `server` 在关机时 `cancel`）。
3. 使用独立的超时上下文调用 `Stop(timeout)`，避免与外层关机上下文混淆。

因此 **`Run` 会阻塞到 `ctx` 被取消并完成 `Stop`**，适合被 `errgroup` 一类调度器管理。

## 示例：注册多个服务

```go
package main

import (
	"context"
	"log/slog"

	"github.com/c1emon/gcommon/server"
	"github.com/c1emon/gcommon/service"
)

type noopService struct{ name string }

func (n *noopService) Name() string { return n.name }
func (n *noopService) Start() error { return nil }
func (n *noopService) Stop(_ context.Context) error { return nil }

func main() {
	repo := service.NewServiceRepo()
	repo.Register(&noopService{name: "a"})
	repo.Register(&noopService{name: "b"})

	logger := slog.Default()
	srv, err := server.New(repo, logger)
	if err != nil {
		panic(err)
	}
	// 随后用 goroutine 跑 ListenToSystemSignals，主流程 srv.Run()，见 ../server/README.md
	_ = srv
}
```

## 与 ginx 的 `HTTPService` 配合

[`ginx`](https://github.com/c1emon/gcommon/tree/main/ginx) 的 `HTTPService` 实现了 `Service`（`Start` 起监听、`Stop` 调 `Shutdown`）。完整 smoke 可参考仓库内：

`test/http_svc_test.go`（`INTEGRATION=1` 时手动跑）。
