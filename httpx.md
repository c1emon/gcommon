# httpx 需要补全的能力

## 1. 区分“页面导航”和“子资源 / XHR 请求”的导航状态更新

需要 httpx 的 browser navigation 能力支持：

- `Navigation` 请求可以更新当前页面 URL / 后续 Referer 基准。
- `XHR` / `fetch` / 子资源请求只使用当前页面 URL 生成 Referer，不应把自身 URL 写回为当前页面 URL。

期望链路：

```text
GET  /login                 (navigation)  -> 当前页面 = /login
POST /preLogin/check        (xhr)         -> Referer = /login，当前页面仍是 /login
POST /accountEntrust/check  (xhr)         -> Referer = /login，当前页面仍是 /login
POST /login                 (navigation)  -> Referer = /login
```

## 2. 显式可控的请求类型

现有 `AsXHR()` / `AsNavigation()` 方向是对的，但需要确保请求类型不仅影响请求头，还影响 browser navigation state 的写入策略。

需要保证：

- `AsXHR()`：生成 XHR 头；不更新当前页面 URL。
- `AsNavigation()`：生成 document/navigation 头；成功发出后可更新当前页面 URL。
- `Auto`：根据 Content-Type / Accept 等规则推断类型，并使用推断后的类型决定是否更新当前页面 URL。

## 3. 支持固定当前页面上下文 / Referer 基准

需要提供一种 API，让调用方显式声明“这些请求都从某个页面上下文发起”。例如：

```go
client.SetBrowserCurrentURL(loginURL)
// or
req.WithBrowserPageURL(loginURL)
```

用途：登录流程中多个 AJAX 和最终表单提交都来自同一个页面，但代码层面是多次独立请求。调用方应能把当前页面锚定到初始登录页。

## 4. 提供不污染全局导航状态的临时请求

需要请求级能力：某次请求使用 browser headers，但不改变 client 的 navigation state。

示例语义：

```go
req.AsXHR().WithoutBrowserNavigationStateUpdate()
```

或：

```go
req.WithBrowserStateUpdate(false)
```

这可覆盖 XHR、探测请求、接口预校验等不会改变页面位置的请求。

## 5. 保留真实浏览器风格的 header 生成

在补全上述能力时，应继续保持当前已有行为：

- navigation 请求生成 `Accept: text/html,...`
- XHR 请求生成 JSON / axios 风格 `Accept`
- POST 请求设置 `Origin`
- 同源请求设置 `Sec-Fetch-Site: same-origin`
- XHR 请求不带 `Sec-Fetch-User` / `Upgrade-Insecure-Requests`
- navigation 请求带 `Sec-Fetch-Mode: navigate`、`Sec-Fetch-Dest: document`

## 6. 测试覆盖要求

需要在 httpx 上游补测试锁定以下行为：

1. navigation GET 后，当前页面 URL 更新。
2. XHR POST 使用当前页面作为 Referer。
3. XHR POST 完成后不更新当前页面 URL。
4. 连续多个 XHR 后，最终 navigation POST 的 Referer 仍然是原始页面。
5. 显式设置当前页面上下文时，Referer 以该上下文为准。
6. `AsXHR()` / `AsNavigation()` / `Auto` 三种模式下，header 与 state update 策略一致。
