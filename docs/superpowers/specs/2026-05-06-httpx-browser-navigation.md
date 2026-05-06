# httpx Browser Navigation Spec

## 背景

`httpx` 现有 `WithBrowser(BrowserChrome)` 已经把 TLS/HTTP2 指纹、浏览器默认 header 和 header order 交给 `github.com/imroc/req/v3` 的浏览器模拟能力处理。`httpx.md` 描述的新增需求不是继续增加业务方手写 header，而是在单个 `httpx.Client` 内维护浏览器式导航状态，让调用方可以用同一个 caller-owned client 表达一段顺序浏览器会话。

该能力的直接下游是 `sso-broker/internal/upstream.Session`，但 `httpx` 只提供通用浏览器导航语义，不包含 ISC SSO、PMS30、ticket、cookie、账号密码、业务错误解析或日志脱敏规则。

## 目标

- 提供 `WithBrowserNavigation(...) ClientOption`，为某个新建 `httpx.Client` 开启浏览器导航 header 维护。
- 在最终发送请求前，根据当前请求和上一跳最终 URL 补全 `Referer`、`Sec-Fetch-Site`、POST `Origin`、HTML navigation header、JSON/XHR header。
- 每次 `ClientFactory.NewClient` / `MustNewClient` 创建独立导航状态；`ClientFactory` 只保存 profile，不共享导航状态。
- `Client.Clone()` 保留导航能力，并共享同一导航状态，支持 no-redirect 临时 clone 继续推进同一会话。
- 单次请求显式设置的 header 优先级最高；导航层可以覆盖 `WithBrowser(...)` 或 `WithHeader(...)` 注入的 profile 级默认动态 header。
- redirect 后记录最终响应 URL，下一跳以最终 URL 计算 `Referer` 和 `Sec-Fetch-Site`。

## 非目标

- 不下沉 ISC SSO、PMS30 或其他系统的 URL、步骤顺序、ticket/cookie 语义。
- 不处理账号、密码、ticket、cookie 等业务敏感字段日志或脱敏策略。
- 不解释上游响应 body 的业务状态码。
- 不实现项目专用 WAF 绕过规则。
- 不把浏览器导航状态放进 `ClientFactory`，也不让 profile 之间共享上一跳 URL。

## Public API

新增 API：

```go
type BrowserNavigationOption util.Option[browserNavigationConfig]

func WithBrowserNavigation(opts ...BrowserNavigationOption) ClientOption

type ReferrerPolicy string

const (
	ReferrerPolicyNoReferrer                   ReferrerPolicy = "no-referrer"
	ReferrerPolicyOrigin                       ReferrerPolicy = "origin"
	ReferrerPolicyStrictOriginWhenCrossOrigin  ReferrerPolicy = "strict-origin-when-cross-origin"
)

func WithReferrerPolicy(policy ReferrerPolicy) BrowserNavigationOption
func WithXHRForJSON(enabled bool) BrowserNavigationOption
func WithDefaultXRequestedWith(enabled bool) BrowserNavigationOption
func WithDefaultSecFetchUser(enabled bool) BrowserNavigationOption

type BrowserRequestKind int

const (
	BrowserRequestAuto BrowserRequestKind = iota
	BrowserRequestNavigation
	BrowserRequestXHR
)

func (r *Request) WithBrowserRequestKind(kind BrowserRequestKind) *Request
func (r *Request) AsXHR() *Request
func (r *Request) AsNavigation() *Request
```

默认配置：

```go
browserNavigationConfig{
	referrerPolicy:        ReferrerPolicyStrictOriginWhenCrossOrigin,
	xhrForJSON:            true,
	defaultXRequestedWith: false,
	defaultSecFetchUser:   true,
}
```

示例：

```go
factory.RegisterProfile("flow-upstream",
	httpx.WithBrowser(httpx.BrowserChrome),
	httpx.WithBrowserNavigation(
		httpx.WithReferrerPolicy(httpx.ReferrerPolicyStrictOriginWhenCrossOrigin),
		httpx.WithXHRForJSON(true),
	),
)

client := factory.MustNewClient("flow-upstream")
_, _ = client.Req().Get("https://example.com/login")
_, _ = client.Req().AsXHR().SetBodyJsonMarshal(map[string]string{"ok": "1"}).Post("https://example.com/api")
```

## Header 语义

### 显式 header 识别

导航层必须区分“单次请求显式设置”和“client/profile/common header 注入”：

- 在 `req.Client.OnBeforeRequest` 阶段捕获 `req.Request.Headers` 中已有的 header key。该阶段发生在 req 内部 `parseRequestHeader` 把 client common headers 合并进 request 之前。
- 捕获 hook 应注册在用户 `WithReqInterceptor` 之后，使用户在请求拦截器中设置的 header 也视为显式 header。
- 动态导航 header 在 `req.Client.WrapRoundTripFunc` 阶段应用，此时 URL、method、body 和 content type 已经解析完成。
- 显式 header key 一律不覆盖；非显式 header 可以被导航层覆盖，即使该值来自 `WithBrowser(...)` 或 `WithHeader(...)`。

### Referer

- 首次请求没有上一跳 URL 时，不设置 `Referer`。
- `ReferrerPolicyNoReferrer` 永不自动设置 `Referer`。
- `ReferrerPolicyOrigin` 使用上一跳 origin，例如 `https://example.com/`。
- `ReferrerPolicyStrictOriginWhenCrossOrigin`：
  - 同源请求使用上一跳完整 URL。
  - 跨源请求使用上一跳 origin。
- 单次请求显式 `Referer` 不覆盖。

### Sec-Fetch-Site

- 首次请求没有上一跳 URL 时设置 `none`。
- 当前 URL 与上一跳 URL 同 scheme 且同 host 时设置 `same-origin`。
- 其他情况设置 `cross-site`。
- 单次请求显式 `Sec-Fetch-Site` 不覆盖。

### Origin

- POST 请求没有显式 `Origin` 时，按当前请求 URL 生成 origin。
- 非 POST 请求默认不补 `Origin`。

### HTML Navigation

当请求类型为 navigation 时，默认补：

```text
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
Sec-Fetch-Mode: navigate
Sec-Fetch-Dest: document
Sec-Fetch-User: ?1
Upgrade-Insecure-Requests: 1
```

`Sec-Fetch-User` 受 `WithDefaultSecFetchUser(false)` 控制。单次请求显式 header 不覆盖。

### JSON/XHR

请求类型判定：

- `Request.AsXHR()` 或 `WithBrowserRequestKind(BrowserRequestXHR)` 显式标记为 XHR。
- `WithXHRForJSON(true)` 且 `Content-Type` 以 `application/json` 开头时，自动判定为 XHR。
- `Request.AsNavigation()` 显式标记为 navigation，即使 content type 是 JSON 也按 navigation 处理。

XHR 默认补：

```text
Accept: application/json, text/plain, */*
Sec-Fetch-Mode: cors
Sec-Fetch-Dest: empty
```

`X-Requested-With: XMLHttpRequest` 仅在 `WithDefaultXRequestedWith(true)` 时默认补。单次请求显式 header 不覆盖。

## 状态模型

- `browserNavigationState` 至少包含 `lastURL *url.URL` 和 `sync.Mutex`。
- 状态属于单个 `ClientFactory.NewClient` / `MustNewClient` 创建出的 `httpx.Client`。
- `Client.Clone()` 通过 req clone 复制 wrapper 闭包，闭包捕获同一个 state 指针，因此 clone 共享上一跳 URL。
- 状态更新以响应最终 URL 为准：优先使用 `resp.Request.URL`，没有响应时不推进状态。
- 该 client 语义上对应一个顺序浏览器会话。实现需要基本并发保护，但文档应声明不建议跨用户、跨登录会话并发复用同一 browser navigation client。

## 实现边界

- `WithBrowserNavigation` 写入 `clientRegisterOpts.browserNavigation *browserNavigationConfig`。
- `BrowserNavigationOption` 沿用现有 `ClientOption` / `FactoryOption` 的 `util.Option` 风格。
- `clientRegisterOpts.clone()` 深拷贝配置值，但不拷贝运行时状态。
- `ClientFactory.buildReqClient` 在每次创建 client 时按配置创建新的 `browserNavigationState` 并安装导航 wrapper。
- `Request.WithBrowserRequestKind` 使用内部 sentinel header 在 req 请求对象上传递单次请求类型；导航 wrapper 发送前删除该 sentinel header。
- `WithBrowser(...)` 继续负责浏览器指纹、header order 和静态默认 header；导航层只负责动态导航字段。

## GitNexus 影响面基线

本 spec 制定前已按 AGENTS 要求对计划会涉及的核心符号运行上游影响分析：

| Symbol | Risk | Direct callers | Affected processes |
| --- | --- | --- | --- |
| `Client` (`httpx/client.go`) | LOW | `Client.Clone` | 无 |
| `Client.Clone` (`httpx/client.go`) | LOW | 无 | 无 |
| `ClientFactory.buildReqClient` (`httpx/factory.go`) | LOW | `ClientFactory.NewClient` | `MustNewClient`, low-confidence `cloud/module/consul/client.go:New` |
| `clientRegisterOpts` (`httpx/client_options.go`) | LOW | `newClientRegisterOpts` | `RegisterProfile` |
| `applyBrowserProfile` (`httpx/browser.go`) | LOW | `ClientFactory.buildReqClient` | `MustNewClient`, low-confidence `cloud/module/consul/client.go:New` |

没有 HIGH 或 CRITICAL 风险。后续真正编辑这些符号前仍需重新运行对应 impact，提交前必须运行 `gitnexus_detect_changes()` / MCP `detect_changes`。

## 测试要求

- 首次请求：`Sec-Fetch-Site=none`，无 `Referer`。
- 同源第二次请求：`Sec-Fetch-Site=same-origin`，`Referer` 为上一跳完整 URL。
- 跨源请求：`Sec-Fetch-Site=cross-site`，`Referer` 按策略收敛为 origin。
- POST 请求自动补当前请求 origin。
- JSON 请求自动补 XHR 风格 header。
- HTML/form 请求自动补 navigation 风格 header。
- `Request.AsXHR()` 可强制 XHR；`Request.AsNavigation()` 可强制 navigation。
- 单次请求显式 header 不被覆盖。
- 与 `WithBrowser(BrowserChrome)` 同时启用时，动态字段能覆盖 browser profile 的静态 `Sec-Fetch-Site`。
- `NewClient` 之间状态隔离。
- `Clone` 后 no-redirect 请求共享导航状态。
- redirect 后记录最终 URL。
