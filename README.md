# gcommon

Shared Go utilities and small framework pieces for services (HTTP, lifecycle, logging, discovery, etc.).

## Package layout (dependency tiers)

| Tier | Packages | Notes |
|------|------------|--------|
| **Light / std-first** | [`optional`](optional/), [`stack`](stack/), parts of [`data`](data/) | Prefer for leaf libraries. |
| **Options & helpers** | [`util`](util/) (`FuncOption` pattern), [`optional`](optional/) | [`httpx`](httpx/), [`ginx`](ginx/) build on `util` options. |
| **HTTP** | [`httpx`](httpx/), [`ginx`](ginx/) | Pulls `req`, `gin` (see `go.mod`). |
| **Process lifecycle** | [`service`](service/), [`server`](server/) | `errgroup`, `slog`. |
| **Integrations** | [`cloud`](cloud/) (Consul, Etcd), [`gormx`](gormx/), [`cachex`](cachex/), [`logx`](logx/) | Heavy transitive dependencies; import only what you need. |

This is a **single module**: importing any path still resolves the whole `go.mod`. For minimal graphs, prefer copying small helpers or splitting into submodules in the future.

## Testing

- Default: `go test ./...`
- Blocking HTTP smoke test under [`test`](test/): set `INTEGRATION=1` (see `test/doc.go`).

## Requirements

Go 1.23+ (see [`go.mod`](go.mod)).
