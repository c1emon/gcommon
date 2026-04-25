# httpx

`httpx` 是基于 [`github.com/imroc/req/v3`](https://github.com/imroc/req) 的多客户端管理封装，核心是 `Manager` + 命名 `Client`。

适用场景：
- 需要同时维护多个下游 HTTP 客户端（不同 BaseURL、UA、拦截器等）
- 需要统一注入基础能力（日志、错误拦截、重试、限流、浏览器伪装）
- 需要按客户端做差异化覆盖

> 说明：`Result/Pagination` 等 VO 已迁移到根目录 `vo` 包（`github.com/c1emon/gcommon/vo`）。

## 安装

```bash
go get github.com/c1emon/gcommon/httpx
```

## 快速开始

```go
package main

import (
	"fmt"
	"time"

	"github.com/c1emon/gcommon/httpx"
)

func main() {
	m := httpx.NewManager(
		httpx.WithGlobalHeader("X-App", "demo"),
		httpx.WithGlobalRetry(httpx.RetryPolicy{
			Enabled:    true,
			MaxRetries: 2,
			MinBackoff: 100 * time.Millisecond,
			MaxBackoff: 1 * time.Second,
		}),
	)

	client := m.Register("github",
		httpx.WithBaseURL("https://api.github.com"),
		httpx.WithTimeout(5*time.Second),
		httpx.WithHeader("Accept", "application/vnd.github+json"),
	)

	resp, err := client.Req().Get("/users/imroc")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode)
}
```

## 包级默认 Manager

若整个进程只需要一个共享 `Manager`（例如 main 里初始化一次，其它包通过 getter 取），可先调用 `InitDefaultManager`，再用 `GetDefaultManager`：

```go
httpx.InitDefaultManager(
	httpx.WithGlobalHeader("X-App", "demo"),
)

m := httpx.GetDefaultManager()
if m == nil {
	panic("call InitDefaultManager before GetDefaultManager")
}
c := m.Register("api", httpx.WithBaseURL("https://api.example.com"))
_ = c
```

说明：

- `InitDefaultManager` 可多次调用，每次都会**替换**为新的 `Manager` 实例。
- 未调用 `InitDefaultManager` 时，`GetDefaultManager()` 返回 **`nil`**，调用方需自行判空或保证初始化顺序。

## 核心概念

- `Manager`
  - 维护 `map[string]*Client`
  - 管理全局默认配置（重试、限流、浏览器、拦截器、header、logger）
  - 可选：通过 `InitDefaultManager` / `GetDefaultManager` 使用包级默认实例（适合进程内只需一个共享 `Manager` 的场景）
- `Client`
  - 对 `*req.Client` 的轻量包装
  - 通过 `Req()` 发起请求
- `ClientOption` / `ManagerOption`
  - 沿用 `util.Option` 风格
  - 注册时按“全局默认 -> 客户端覆盖”合并

## 常用 API

### Manager

- `NewManager(opts ...ManagerOption) *Manager`
- `InitDefaultManager(opts ...ManagerOption)`：用 `NewManager(opts...)` 构建并**替换**包级默认 `Manager`
- `GetDefaultManager() *Manager`：返回当前包级默认 `Manager`；若从未调用过 `InitDefaultManager`，返回 **`nil`**
- `(*Manager).Register(name string, opts ...ClientOption) *Client`
- `(*Manager).Client(name string) (*Client, bool)`
- `(*Manager).MustClient(name string) *Client`
- `(*Manager).Names() []string`

### ManagerOption

- `WithGlobalHeader(key, val string)`
- `WithGlobalLogger(l *slog.Logger)`
- `WithGlobalLimiter(l *rate.Limiter)`
- `WithGlobalRetry(p RetryPolicy)`
- `WithGlobalBrowser(p BrowserProfile)`
- `WithGlobalReqInterceptor(i ReqInterceptor)`
- `WithGlobalRespInterceptor(i RespInterceptor)`
- `WithGlobalStrictJSONContentType()`

### ClientOption

- `WithBaseURL(url string)`
- `WithTimeout(t time.Duration)`
- `WithHeader(key, val string)` / `WithHeaders(map[string]string)`
- `WithUserAgent(ua string)`
- `WithBrowser(p BrowserProfile)`
- `WithLimiter(l *rate.Limiter)` / `DisableLimiter()`
- `WithRetry(p RetryPolicy)` / `DisableRetry()`
- `WithReqInterceptor(i ReqInterceptor)` / `WithRespInterceptor(i RespInterceptor)`
- `WithStrictJSONContentType()` / `WithoutStrictJSONContentType()`
- `WithLogger(l *slog.Logger)`（传 `nil` 可禁用该 client 日志）

## 拦截器

拦截器类型：

```go
type ReqInterceptor func(client *httpx.Client, req *httpx.Request) error
type RespInterceptor func(client *httpx.Client, resp *httpx.Response) error
```

示例（注册请求拦截器）：

```go
m := httpx.NewManager()
c := m.Register("biz",
	httpx.WithReqInterceptor(func(client *httpx.Client, req *httpx.Request) error {
		req.SetHeader("X-Trace-Id", "trace-id")
		return nil
	}),
)
_ = c
```

## 限流

使用 `golang.org/x/time/rate`：

- 支持全局 limiter：`WithGlobalLimiter(...)`
- 支持单 client limiter：`WithLimiter(...)`
- 默认禁用（`nil`）

```go
import "golang.org/x/time/rate"

m := httpx.NewManager(
	httpx.WithGlobalLimiter(rate.NewLimiter(rate.Every(50*time.Millisecond), 10)),
)

// 该 client 再叠加一层本地 limiter
c := m.Register("slow-api",
	httpx.WithLimiter(rate.NewLimiter(rate.Every(100*time.Millisecond), 5)),
)
_ = c
```

> 限流 hook 会在请求链最前注册，先于用户请求拦截器执行。

## 重试

`RetryPolicy` 通过 `WithGlobalRetry` 或 `WithRetry` 配置：

```go
m := httpx.NewManager(
	httpx.WithGlobalRetry(httpx.RetryPolicy{
		Enabled:    true,
		MaxRetries: 3,
		MinBackoff: 100 * time.Millisecond,
		MaxBackoff: 2 * time.Second,
	}),
)
```

默认重试判定：
- 传输错误（`err != nil`）
- HTTP `429`
- HTTP `5xx`

也可以通过 `RetryIf` 追加自定义条件。

## 浏览器伪装

支持 profile：

- `BrowserNone`
- `BrowserChrome`
- `BrowserFirefox`
- `BrowserSafari`

```go
c := httpx.NewManager().Register("anti-bot",
	httpx.WithBrowser(httpx.BrowserChrome),
)
_ = c
```

当同时设置 `WithBrowser(...)` 与 `WithUserAgent(...)` 时，优先使用 `BrowserProfile`。

## 严格 JSON Content-Type 模式

`httpx` 内置错误拦截器会解析业务 JSON 信封（`code/msg/data`）。

- 默认：非严格，Content-Type 不是 JSON 时也会尝试解析
- 严格：只在 JSON Content-Type（`application/json` 或 `+json`）时解析

全局开启：

```go
m := httpx.NewManager(httpx.WithGlobalStrictJSONContentType())
```

单 client 开启/关闭：

```go
c := httpx.NewManager().Register("a",
	httpx.WithStrictJSONContentType(),
)

d := httpx.NewManager(httpx.WithGlobalStrictJSONContentType()).Register("b",
	httpx.WithoutStrictJSONContentType(),
)

_, _ = c, d
```

## 与 vo 包配合

如果你的接口返回统一结构，推荐使用 `vo`：

```go
import "github.com/c1emon/gcommon/vo"

ok := vo.NewResultOK(map[string]string{"status": "ok"})
msg := vo.NewMsgResult(4001, "bad request")
_, _ = ok, msg
```
