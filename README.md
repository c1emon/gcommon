# gcommon

Shared Go utilities and small framework pieces for services (HTTP, lifecycle, logging, discovery, etc.).

This repository is a **multi-module workspace**: the root module holds lightweight packages; heavier integrations live in nested modules with their own `go.mod` files so consumers only download what they import.

## Modules

| Module path | Directory | Role |
|-------------|-----------|------|
| `github.com/c1emon/gcommon/v2` | [`.`](.) | Core: `util`, `optional`, `errorx`, `data`, `stack`, `reader`, `service`, `server`, `pinger` |
| `github.com/c1emon/gcommon/health/v2` | [`health`](health/) | Health JSON ([hellofresh/health-go](https://github.com/hellofresh/health-go)); nested packages [`health/http`](health/http), [`health/gin`](health/gin) — see [`health/README.md`](health/README.md) |
| `github.com/c1emon/gcommon/httpx/v2` | [`httpx`](httpx/) | HTTP client (`imroc/req`) |
| `github.com/c1emon/gcommon/ginx/v2` | [`ginx`](ginx/) | Gin helpers |
| `github.com/c1emon/gcommon/cloud/v2` | [`cloud`](cloud/) | Service discovery / Consul |
| `github.com/c1emon/gcommon/gormx/v2` | [`gormx`](gormx/) | GORM helpers |
| `github.com/c1emon/gcommon/cachex/v2` | [`cachex`](cachex/) | Cache interfaces + memory backend |
| `github.com/c1emon/gcommon/logx/v2` | [`logx`](logx/) | Pure [`log/slog`](https://pkg.go.dev/log/slog) helpers (handlers, `Init` / `Default`, context attrs) |
| `github.com/c1emon/gcommon/test/v2` | [`test`](test/) | Integration tests (optional `INTEGRATION=1`) |

## Using as a dependency

Import only the paths you need. For example, only the root module:

```go
import "github.com/c1emon/gcommon/v2/util"
```

```bash
go get github.com/c1emon/gcommon/v2@v2.0.0
```

HTTP client stack (root + httpx):

```bash
go get github.com/c1emon/gcommon/v2@v2.0.0 github.com/c1emon/gcommon/httpx/v2@v2.3.0
```

Health endpoints (nested module; import `health/http` or `health/gin`):

```bash
go get github.com/c1emon/gcommon/health/v2/http@v2.2.0
# or: go get github.com/c1emon/gcommon/health/v2/gin@v2.2.0
```

Nested modules declare real released versions for internal dependencies and keep local `replace` directives for in-repo development. Published consumers resolve directory-prefixed module tags such as `httpx/v2.3.0` and `ginx/v2.2.0`.

## Contributing (this repo)

Use the committed [`go.work`](go.work) so all modules resolve together.

Run tests for **every** module (root `go test ./...` only covers the main module):

```bash
go test ./... \
  ./httpx/... ./ginx/... ./health/... ./cloud/... ./gormx/... ./cachex/... ./logx/... ./test/... \
  -count=1 -timeout 120s
```

Avoid `go work sync` here: submodules use `replace` for the monorepo layout, while published consumers should resolve the released module versions.

## Requirements

Go 1.25+ (see root [`go.mod`](go.mod); nested modules use the same `go` line).
