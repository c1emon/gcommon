# httpx

`httpx` 是基于 [`github.com/imroc/req/v3`](https://github.com/imroc/req) 的 HTTP client factory 封装，核心是 `ClientFactory` + 命名 profile。

适用场景：
- 需要为多个下游 HTTP 客户端维护不同配置（BaseURL、UA、拦截器、重试、限流等）
- 需要用统一的全局默认配置创建 caller-owned client
- 需要为登录、票据兑换、浏览器式会话创建相互隔离的 cookie jar
- 需要按 profile 或单个 client 覆盖重试、业务错误映射、跳转策略

> 说明：`Result/Pagination` 等 VO 已迁移到根目录 `vo` 包（`github.com/c1emon/gcommon/v2/vo`）。

## 安装

```bash
go get github.com/c1emon/gcommon/httpx/v2
```

## 快速开始

```go
package main

import (
	"fmt"
	"time"

	"github.com/c1emon/gcommon/httpx/v2"
)

func main() {
	factory := httpx.NewClientFactory(
		httpx.WithGlobalHeader("X-App", "demo"),
		httpx.WithGlobalRetry(httpx.RetryPolicy{
			Enabled:    true,
			MaxRetries: 2,
			MinBackoff: 100 * time.Millisecond,
			MaxBackoff: 1 * time.Second,
		}),
	)

	factory.RegisterProfile("github",
		httpx.WithBaseURL("https://api.github.com"),
		httpx.WithTimeout(5*time.Second),
		httpx.WithHeader("Accept", "application/vnd.github+json"),
	)

	client := factory.MustNewClient("github")
	resp, err := client.Req().Get("/users/imroc")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.StatusCode)
}
```

## 包级默认 ClientFactory

若整个进程只需要一个共享 `ClientFactory`（例如 main 里初始化一次，其它包通过 getter 取），可先调用 `InitDefaultClientFactory`，再用 `GetDefaultClientFactory`：

```go
httpx.InitDefaultClientFactory(
	httpx.WithGlobalHeader("X-App", "demo"),
)

factory := httpx.GetDefaultClientFactory()
if factory == nil {
	panic("call InitDefaultClientFactory before GetDefaultClientFactory")
}
factory.RegisterProfile("api", httpx.WithBaseURL("https://api.example.com"))
c := factory.MustNewClient("api")
_ = c
```

说明：

- `InitDefaultClientFactory` 可多次调用，每次都会**替换**为新的 `ClientFactory` 实例。
- 未调用 `InitDefaultClientFactory` 时，`GetDefaultClientFactory()` 返回 **`nil`**，调用方需自行判空或保证初始化顺序。

## 核心概念

- `ClientFactory`
  - 维护命名 profile，而不是维护 client 实例
  - 每次 `NewClient(profileName)` 都创建新的 `*Client`
  - 创建出的 client 生命周期由调用者管理
- `Client`
  - 对 `*req.Client` 的轻量包装
  - 通过 `Req()` 发起请求
  - 可用 `Clone()` 创建同一会话内的临时变体
- `ClientOption` / `FactoryOption`
  - 沿用 `util.Option` 风格
  - 创建 client 时按“全局默认 -> profile 配置 -> 实例覆盖”合并

## 常用 API

### ClientFactory

- `NewClientFactory(opts ...FactoryOption) *ClientFactory`
- `InitDefaultClientFactory(opts ...FactoryOption)`：用 `NewClientFactory(opts...)` 构建并**替换**包级默认 `ClientFactory`
- `GetDefaultClientFactory() *ClientFactory`：返回当前包级默认 `ClientFactory`；若从未调用过 `InitDefaultClientFactory`，返回 **`nil`**
- `(*ClientFactory).RegisterProfile(name string, opts ...ClientOption)`：创建或替换命名 profile，只保存配置，不返回 client
- `(*ClientFactory).NewClient(name string, opts ...ClientOption) (*Client, bool)`：为已注册 profile 创建新的 caller-owned client
- `(*ClientFactory).MustNewClient(name string, opts ...ClientOption) *Client`：profile 不存在时 panic
- `(*ClientFactory).ProfileNames() []string`

### FactoryOption

- `WithGlobalHeader(key, val string)`
- `WithGlobalLogger(l *slog.Logger)`
- `WithGlobalLimiter(l *rate.Limiter)`
- `WithGlobalRetry(p RetryPolicy)`
- `WithGlobalBrowser(p BrowserProfile)`
- `WithGlobalReqInterceptor(i ReqInterceptor)`
- `WithGlobalRespInterceptor(i RespInterceptor)`
- `WithGlobalStrictJSONContentType()`
- `WithGlobalBusinessError()`
- `DisableGlobalBusinessError()`

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
- `WithBusinessError()` / `DisableBusinessError()`
- `WithCookieJar(jar http.CookieJar)` / `WithCookieJarFactory(factory CookieJarFactory)`
- `WithRedirectPolicy(policies ...RedirectPolicy)`
- `WithLogger(l *slog.Logger)`（传 `nil` 可禁用该 client 日志）

### Client 方法

- `(*Client).Name() string`
- `(*Client).Req() *Request`
- `(*Client).Clone() *Client`
- `(*Client).SetCookieJar(jar http.CookieJar) *Client`
- `(*Client).SetCookieJarFactory(factory CookieJarFactory) *Client`
- `(*Client).GetCookies(rawURL string) ([]*http.Cookie, error)`
- `(*Client).SetRedirectPolicy(policies ...RedirectPolicy) *Client`

### Cookie 和 Redirect

- `type CookieJarFactory func() *cookiejar.Jar`
- `type RedirectPolicy = req.RedirectPolicy`
- `NoRedirectPolicy() RedirectPolicy`
- `DefaultRedirectPolicy() RedirectPolicy`
- `MaxRedirectPolicy(noOfRedirect int) RedirectPolicy`
- `SameDomainRedirectPolicy() RedirectPolicy`
- `SameHostRedirectPolicy() RedirectPolicy`
- `AllowedHostRedirectPolicy(hosts ...string) RedirectPolicy`
- `AllowedDomainRedirectPolicy(hosts ...string) RedirectPolicy`
- `AlwaysCopyHeaderRedirectPolicy(headers ...string) RedirectPolicy`

## 拦截器

拦截器类型：

```go
type ReqInterceptor func(client *httpx.Client, req *httpx.Request) error
type RespInterceptor func(client *httpx.Client, resp *httpx.Response) error
```

示例（注册请求拦截器）：

```go
factory := httpx.NewClientFactory()
factory.RegisterProfile("biz",
	httpx.WithReqInterceptor(func(client *httpx.Client, req *httpx.Request) error {
		req.SetHeader("X-Trace-Id", "trace-id")
		return nil
	}),
)
c := factory.MustNewClient("biz")
_ = c
```

## 限流

使用 `golang.org/x/time/rate`：

- 支持全局 limiter：`WithGlobalLimiter(...)`
- 支持单 client limiter：`WithLimiter(...)`
- 默认禁用（`nil`）

```go
import "golang.org/x/time/rate"

factory := httpx.NewClientFactory(
	httpx.WithGlobalLimiter(rate.NewLimiter(rate.Every(50*time.Millisecond), 10)),
)

factory.RegisterProfile("slow-api",
	httpx.WithLimiter(rate.NewLimiter(rate.Every(100*time.Millisecond), 5)),
)
c := factory.MustNewClient("slow-api")
_ = c
```

> 限流 hook 会在请求链最前注册，先于用户请求拦截器执行。

## 重试

`RetryPolicy` 通过 `WithGlobalRetry` 或 `WithRetry` 配置：

```go
factory := httpx.NewClientFactory(
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

登录、支付、表单提交、票据兑换、信任登录等非幂等请求不建议自动重试。为这些流程注册独立 profile，并显式使用 `DisableRetry()`，让业务层自己判断是否可以重放请求。

## 浏览器伪装

支持 profile：

- `BrowserNone`
- `BrowserChrome`
- `BrowserFirefox`
- `BrowserSafari`

```go
factory := httpx.NewClientFactory()
factory.RegisterProfile("anti-bot",
	httpx.WithBrowser(httpx.BrowserChrome),
)
c := factory.MustNewClient("anti-bot")
_ = c
```

当同时设置 `WithBrowser(...)` 与 `WithUserAgent(...)` 时，优先使用 `BrowserProfile`。

## 浏览器式会话

`ClientFactory` 管 profile，不管实例。每次用户登录、票据兑换或上游会话都应调用 `MustNewClient` 创建独立 client；会话结束后由调用方丢弃该实例。

```go
package main

import (
	"net/http/cookiejar"
	"time"

	"github.com/c1emon/gcommon/httpx/v2"
	"golang.org/x/net/publicsuffix"
)

func main() {
	factory := httpx.NewClientFactory()
	factory.RegisterProfile("browser-session",
		httpx.WithTimeout(10*time.Second),
		httpx.DisableRetry(),
		httpx.DisableBusinessError(),
		httpx.WithCookieJarFactory(func() *cookiejar.Jar {
			jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
			if err != nil {
				panic(err)
			}
			return jar
		}),
	)

	session := factory.MustNewClient("browser-session")
	noRedirect := session.Clone().SetRedirectPolicy(httpx.NoRedirectPolicy())
	_, _ = session, noRedirect
}
```

## 业务错误映射

`httpx` 可以按需启用 gcommon 业务 JSON 信封（`code/msg/data`）错误映射。默认不启用，避免把第三方上游 JSON 误判为 gcommon 业务错误。

```go
factory := httpx.NewClientFactory(httpx.WithGlobalBusinessError())

single := httpx.NewClientFactory()
single.RegisterProfile("biz", httpx.WithBusinessError())
```

严格 Content-Type 模式只在业务错误映射启用时生效：

- 默认：非严格，业务错误映射启用后，即使 Content-Type 不是 JSON 也会尝试解析
- 严格：业务错误映射启用后，只在 JSON Content-Type（`application/json` 或 `+json`）时解析

全局开启严格模式：

```go
factory := httpx.NewClientFactory(
	httpx.WithGlobalBusinessError(),
	httpx.WithGlobalStrictJSONContentType(),
)
```

单 profile 开启/关闭严格模式：

```go
factory := httpx.NewClientFactory(httpx.WithGlobalBusinessError())

factory.RegisterProfile("a",
	httpx.WithStrictJSONContentType(),
)

factory.RegisterProfile("b",
	httpx.WithoutStrictJSONContentType(),
)
```

## 与 vo 包配合

如果你的接口返回统一结构，推荐使用 `vo`：

```go
import "github.com/c1emon/gcommon/v2/vo"

ok := vo.NewResultOK(map[string]string{"status": "ok"})
msg := vo.NewMsgResult(4001, "bad request")
_, _ = ok, msg
```
