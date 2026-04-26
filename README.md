# gcommon

Shared Go utilities and small framework pieces for services (HTTP, lifecycle, logging, discovery, etc.).

This repository is a **multi-module workspace**: the root module holds lightweight packages; heavier integrations live in nested modules with their own `go.mod` files so consumers only download what they import.

## Modules

| Module path | Directory | Role |
|-------------|-----------|------|
| `github.com/c1emon/gcommon` | [`.`](.) | Core: `util`, `optional`, `errorx`, `data`, `stack`, `reader`, `service`, `server`, `pinger` |
| `github.com/c1emon/gcommon/health` | [`health`](health/) | Health JSON ([hellofresh/health-go](https://github.com/hellofresh/health-go)); nested packages [`health/http`](health/http), [`health/gin`](health/gin) — see [`health/README.md`](health/README.md) |
| `github.com/c1emon/gcommon/httpx` | [`httpx`](httpx/) | HTTP client (`imroc/req`) |
| `github.com/c1emon/gcommon/ginx` | [`ginx`](ginx/) | Gin helpers |
| `github.com/c1emon/gcommon/cloud` | [`cloud`](cloud/) | Service discovery / Consul |
| `github.com/c1emon/gcommon/gormx` | [`gormx`](gormx/) | GORM helpers |
| `github.com/c1emon/gcommon/cachex` | [`cachex`](cachex/) | Cache interfaces + memory backend |
| `github.com/c1emon/gcommon/logx` | [`logx`](logx/) | Pure [`log/slog`](https://pkg.go.dev/log/slog) helpers (handlers, `Init` / `Default`, context attrs) |
| `github.com/c1emon/gcommon/test` | [`test`](test/) | Integration tests (optional `INTEGRATION=1`) |

## Using as a dependency

Import only the paths you need. For example, only the root module:

```go
import "github.com/c1emon/gcommon/util"
```

```bash
go get github.com/c1emon/gcommon@latest
```

HTTP client stack (root + httpx):

```bash
go get github.com/c1emon/gcommon@latest github.com/c1emon/gcommon/httpx@latest
```

Health endpoints (nested module; import `health/http` or `health/gin`):

```bash
go get github.com/c1emon/gcommon/health/http@latest
# or: go get github.com/c1emon/gcommon/health/gin@latest
```

Nested modules declare `require github.com/c1emon/gcommon v0.0.0` and a **local `replace => ../`** in-repo for development. For published versions, tag releases and align `require` versions across modules (or drop `replace` once tagged pseudo-versions exist on the proxy).

## Contributing (this repo)

Use the committed [`go.work`](go.work) so all modules resolve together.

Run tests for **every** module (root `go test ./...` only covers the main module):

```bash
go test ./... \
  ./httpx/... ./ginx/... ./health/... ./cloud/... ./gormx/... ./cachex/... ./logx/... ./test/... \
  -count=1 -timeout 120s
```

Avoid `go work sync` here: it tries to resolve `v0.0.0` from the network while submodules use `replace` for the monorepo layout.

## Requirements

Go 1.25+ (see root [`go.mod`](go.mod); nested modules use the same `go` line).
