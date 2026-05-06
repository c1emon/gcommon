# health

基于 [`github.com/hellofresh/health-go/v5`](https://github.com/hellofresh/health-go) 的健康检查封装：在 JSON 中返回 **组件信息**（服务名、版本）以及 **Go 运行时基础指标**（通过库的 `WithSystemInfo`：Go 版本、协程数、堆分配等）。

## 独立模块

本目录包含 **`go.mod`**，模块路径为 **`github.com/c1emon/gcommon/health/v2`**，与仓库根模块并列。

- 在本仓库内开发：使用仓库根目录的 [`go.work`](https://go.dev/ref/mod#workspaces)，其中已 `use` 本模块（与其它子模块一并列出）。
- 仅依赖本模块时：`go get github.com/c1emon/gcommon/health/v2/http@<version>`（或 `.../gin`）；若从源码路径引用，可在主模块中加  
  `replace github.com/c1emon/gcommon/health/v2 => ../health`（路径按实际调整）。

本模块在 **`health/` 下拆为多个 Go 子包**，按需引入：

| 子包 | 用途 |
|------|------|
| [`health`](https://pkg.go.dev/github.com/c1emon/gcommon/health/v2) | 共享 [`Config`](https://pkg.go.dev/github.com/c1emon/gcommon/health/v2#Config)（仅类型，无 Handler） |
| [`health/bridge`](https://pkg.go.dev/github.com/c1emon/gcommon/health/v2/bridge) | 对接 hellofresh 的构建逻辑；通常由 `http` / `gin` 间接使用 |
| [`health/http`](https://pkg.go.dev/github.com/c1emon/gcommon/health/v2/http) | 标准库 [`net/http`](https://pkg.go.dev/net/http#Handler) |
| [`health/gin`](https://pkg.go.dev/github.com/c1emon/gcommon/health/v2/gin) | [Gin](https://github.com/gin-gonic/gin) [`HandlerFunc`](https://pkg.go.dev/github.com/gin-gonic/gin#HandlerFunc) |

详细用法与示例见各子目录 **README**。
