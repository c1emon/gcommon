# V2 Independent Module Versioning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish the root module and nested modules with independent v2 versions while keeping the repository buildable as a Go workspace.

**Architecture:** Use Go's multi-module release convention: the root module is tagged with root tags and each nested module is tagged with a directory-prefixed tag. Because the requested versions are v2, all module paths and imports must use the `/v2` suffix to satisfy Go Modules semantic import versioning.

**Tech Stack:** Go 1.25, Go workspaces, Git tags, GitNexus, `go test`, `go list`.

---

## Release Target

Requested release versions:

| Module | Directory | Requested version | Required tag |
| --- | --- | --- | --- |
| `github.com/c1emon/gcommon/v2` | `.` | `v2.0.0` | `v2.0.0` |
| `github.com/c1emon/gcommon/httpx/v2` | `httpx` | `v2.3.0` | `httpx/v2.3.0` |
| `github.com/c1emon/gcommon/ginx/v2` | `ginx` | `v2.2.0` | `ginx/v2.2.0` |
| `github.com/c1emon/gcommon/cloud/v2` | `cloud` | `v2.2.0` | `cloud/v2.2.0` |
| `github.com/c1emon/gcommon/gormx/v2` | `gormx` | `v2.2.0` | `gormx/v2.2.0` |
| `github.com/c1emon/gcommon/cachex/v2` | `cachex` | `v2.2.0` | `cachex/v2.2.0` |
| `github.com/c1emon/gcommon/health/v2` | `health` | `v2.2.0` | `health/v2.2.0` |
| `github.com/c1emon/gcommon/logx/v2` | `logx` | `v2.2.0` | `logx/v2.2.0` |
| `github.com/c1emon/gcommon/test/v2` | `test` | `v2.2.0` | `test/v2.2.0` |

## Hard Constraint: Existing Root Tag

The local repository already has root tags `v2.0.0`, `v2.1.0`, `v2.2.0`, and `v2.3.0`.

Before implementation, decide one of these two paths:

1. Preferred: change the root release target from `v2.0.0` to the next unused root version, such as `v2.4.0`.
2. Risky: move/recreate `v2.0.0` at current `main`. Only do this after confirming the tag has not been pushed, consumed, or cached by the Go proxy.

Do not retag `v2.0.0` silently.

Implementation decision: the user explicitly approved retagging the root `v2.0.0` at current `main`.

## Files To Modify

- Modify: `go.mod`
  - Change root module path to `github.com/c1emon/gcommon/v2`.
- Modify: `httpx/go.mod`
  - Change module path to `github.com/c1emon/gcommon/httpx/v2`.
  - Change root dependency to `github.com/c1emon/gcommon/v2 v2.0.0` or the approved replacement root version.
- Modify: `ginx/go.mod`
  - Change module path to `github.com/c1emon/gcommon/ginx/v2`.
  - Change root dependency to `github.com/c1emon/gcommon/v2 v2.0.0` or the approved replacement root version.
  - Change logx dependency to `github.com/c1emon/gcommon/logx/v2 v2.2.0`.
- Modify: `cloud/go.mod`
  - Change module path to `github.com/c1emon/gcommon/cloud/v2`.
  - Change root dependency to `github.com/c1emon/gcommon/v2 v2.0.0` or the approved replacement root version.
- Modify: `gormx/go.mod`
  - Change module path to `github.com/c1emon/gcommon/gormx/v2`.
- Modify: `cachex/go.mod`
  - Change module path to `github.com/c1emon/gcommon/cachex/v2`.
- Modify: `health/go.mod`
  - Change module path to `github.com/c1emon/gcommon/health/v2`.
- Modify: `logx/go.mod`
  - Change module path to `github.com/c1emon/gcommon/logx/v2`.
- Modify: `test/go.mod`
  - Change module path to `github.com/c1emon/gcommon/test/v2`.
  - Change root dependency to `github.com/c1emon/gcommon/v2 v2.0.0` or the approved replacement root version.
  - Change ginx dependency to `github.com/c1emon/gcommon/ginx/v2 v2.2.0`.
  - Change logx dependency to `github.com/c1emon/gcommon/logx/v2 v2.2.0`.
- Modify: Go source and test files that import `github.com/c1emon/gcommon...`.
  - Update imports to the matching `/v2` module path.
- Modify: README files and plan/docs examples that mention old import paths or `v0.0.0`.
  - Update install commands, import examples, pkg.go.dev links, and release notes.
- Modify: `go.work` only if `go` tooling rewrites formatting.
- Modify: `go.work.sum` only if verification commands update it legitimately.

## GitNexus Controls

- Before editing source files, run targeted impact analysis for representative public symbols because import path migration touches many files:

```bash
gitnexus impact --repo gcommon --target Client --direction upstream
gitnexus impact --repo gcommon --target Service --direction upstream
gitnexus impact --repo gcommon --target Result --direction upstream
gitnexus impact --repo gcommon --target Handler --direction upstream
```

- If any impact result is HIGH or CRITICAL, stop and report the result before editing.
- Before any commit, run GitNexus change detection:

```bash
gitnexus detect-changes --repo gcommon --scope all
```

If the CLI names differ in this local installation, use the equivalent MCP tools: `impact` and `detect_changes`.

## Task 1: Confirm Release Version Decision

**Files:**
- Read: `go.mod`
- Read: all nested `go.mod` files
- Read: local Git tags

- [ ] **Step 1: Confirm current HEAD and existing v2 tags**

Run:

```bash
git rev-parse HEAD
git tag --list 'v2*' 'httpx/*' 'ginx/*' 'cloud/*' 'gormx/*' 'cachex/*' 'health/*' 'logx/*' 'test/*'
git rev-list -n 1 v2.0.0
```

Expected:

```text
Current HEAD is the commit intended for release.
Root v2.0.0 already exists.
No nested module v2 tags exist unless someone created them after this plan.
```

- [ ] **Step 2: Choose the root tag path**

Record one of:

```text
Root version decision: use v2.4.0 because v2.0.0 already exists.
```

or:

```text
Root version decision: retag v2.0.0 after confirming it is unpublished and uncached.
```

Do not continue until this line is recorded in the implementation notes or final response.

## Task 2: Run Pre-Edit GitNexus Impact Checks

**Files:**
- No file edits.

- [ ] **Step 1: Run impact checks with MCP or CLI**

Preferred MCP targets:

```text
impact(repo: "gcommon", target: "Client", direction: "upstream")
impact(repo: "gcommon", target: "Service", direction: "upstream")
impact(repo: "gcommon", target: "Result", direction: "upstream")
impact(repo: "gcommon", target: "Handler", direction: "upstream")
```

Expected:

```text
Risk level recorded for each target.
Direct callers and affected processes summarized before edits.
```

- [ ] **Step 2: Stop on HIGH or CRITICAL**

If any result is HIGH or CRITICAL, report:

```text
Impact check returned HIGH/CRITICAL for <target>. Direct callers: <list>. Affected processes: <list>. Waiting before editing.
```

Continue only after explicit approval.

## Task 3: Update Module Paths And Internal Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `httpx/go.mod`
- Modify: `ginx/go.mod`
- Modify: `cloud/go.mod`
- Modify: `gormx/go.mod`
- Modify: `cachex/go.mod`
- Modify: `health/go.mod`
- Modify: `logx/go.mod`
- Modify: `test/go.mod`

- [ ] **Step 1: Update root `go.mod`**

Change:

```go
module github.com/c1emon/gcommon
```

to:

```go
module github.com/c1emon/gcommon/v2
```

- [ ] **Step 2: Update `httpx/go.mod`**

Use the approved root version in place of `v2.0.0` if Task 1 selected a replacement.

```go
module github.com/c1emon/gcommon/httpx/v2

go 1.25.0

require (
	github.com/c1emon/gcommon/v2 v2.0.0
	github.com/imroc/req/v3 v3.57.0
)
```

Keep the existing indirect dependencies and keep the `quic-go` replacement. Change the local root replacement to:

```go
replace (
	github.com/c1emon/gcommon/v2 => ../
	github.com/quic-go/quic-go => github.com/quic-go/quic-go v0.57.1
)
```

- [ ] **Step 3: Update `ginx/go.mod`**

Use:

```go
module github.com/c1emon/gcommon/ginx/v2

go 1.25.0

require (
	github.com/c1emon/gcommon/logx/v2 v2.2.0
	github.com/c1emon/gcommon/v2 v2.0.0
	github.com/gin-gonic/gin v1.12.0
)
```

Keep the existing indirect dependencies. Change local replacements to:

```go
replace (
	github.com/c1emon/gcommon/logx/v2 => ../logx
	github.com/c1emon/gcommon/v2 => ../
)
```

- [ ] **Step 4: Update `cloud/go.mod`**

Use:

```go
module github.com/c1emon/gcommon/cloud/v2

go 1.25.0

require (
	github.com/c1emon/gcommon/v2 v2.0.0
	github.com/hashicorp/consul/api v1.30.0
)
```

Change the local replacement to:

```go
replace github.com/c1emon/gcommon/v2 => ../
```

- [ ] **Step 5: Update standalone nested module declarations**

Change module declarations exactly:

```go
module github.com/c1emon/gcommon/gormx/v2
```

```go
module github.com/c1emon/gcommon/cachex/v2
```

```go
module github.com/c1emon/gcommon/health/v2
```

```go
module github.com/c1emon/gcommon/logx/v2
```

For `logx/go.mod`, remove the stale root replacement if no package in `logx` imports root packages:

```go
module github.com/c1emon/gcommon/logx/v2

go 1.25.0
```

- [ ] **Step 6: Update `test/go.mod`**

Use:

```go
module github.com/c1emon/gcommon/test/v2

go 1.25.0

require (
	github.com/c1emon/gcommon/ginx/v2 v2.2.0
	github.com/c1emon/gcommon/logx/v2 v2.2.0
	github.com/c1emon/gcommon/v2 v2.0.0
	github.com/gin-gonic/gin v1.12.0
)
```

Keep the existing indirect dependencies. Change local replacements to:

```go
replace (
	github.com/c1emon/gcommon/ginx/v2 => ../ginx
	github.com/c1emon/gcommon/logx/v2 => ../logx
	github.com/c1emon/gcommon/v2 => ../
)
```

## Task 4: Update Source Imports

**Files:**
- Modify: every `.go` file that imports `github.com/c1emon/gcommon`.

- [ ] **Step 1: List old imports**

Run:

```bash
rg -n 'github.com/c1emon/gcommon' -g '*.go'
```

Expected:

```text
All old import paths are listed before the rewrite.
```

- [ ] **Step 2: Rewrite root-module imports**

Apply these mappings:

```text
github.com/c1emon/gcommon/v2/data              -> github.com/c1emon/gcommon/v2/data
github.com/c1emon/gcommon/v2/errorx            -> github.com/c1emon/gcommon/v2/errorx
github.com/c1emon/gcommon/v2/optional          -> github.com/c1emon/gcommon/v2/optional
github.com/c1emon/gcommon/v2/pinger            -> github.com/c1emon/gcommon/v2/pinger
github.com/c1emon/gcommon/v2/reader/...        -> github.com/c1emon/gcommon/v2/reader/...
github.com/c1emon/gcommon/v2/server            -> github.com/c1emon/gcommon/v2/server
github.com/c1emon/gcommon/v2/service           -> github.com/c1emon/gcommon/v2/service
github.com/c1emon/gcommon/v2/stack             -> github.com/c1emon/gcommon/v2/stack
github.com/c1emon/gcommon/v2/util              -> github.com/c1emon/gcommon/v2/util
github.com/c1emon/gcommon/v2/vo                -> github.com/c1emon/gcommon/v2/vo
```

- [ ] **Step 3: Rewrite nested-module imports**

Apply these mappings:

```text
github.com/c1emon/gcommon/httpx/v2             -> github.com/c1emon/gcommon/httpx/v2
github.com/c1emon/gcommon/httpx/v2/interceptors -> github.com/c1emon/gcommon/httpx/v2/interceptors
github.com/c1emon/gcommon/ginx/v2              -> github.com/c1emon/gcommon/ginx/v2
github.com/c1emon/gcommon/logx/v2              -> github.com/c1emon/gcommon/logx/v2
github.com/c1emon/gcommon/cloud/v2             -> github.com/c1emon/gcommon/cloud/v2
github.com/c1emon/gcommon/cloud/v2/registry    -> github.com/c1emon/gcommon/cloud/v2/registry
github.com/c1emon/gcommon/health/v2            -> github.com/c1emon/gcommon/health/v2
github.com/c1emon/gcommon/health/v2/bridge     -> github.com/c1emon/gcommon/health/v2/bridge
github.com/c1emon/gcommon/health/v2/http       -> github.com/c1emon/gcommon/health/v2/http
github.com/c1emon/gcommon/health/v2/gin        -> github.com/c1emon/gcommon/health/v2/gin
github.com/c1emon/gcommon/cachex/v2            -> github.com/c1emon/gcommon/cachex/v2
github.com/c1emon/gcommon/gormx/v2             -> github.com/c1emon/gcommon/gormx/v2
```

- [ ] **Step 4: Verify no old Go imports remain**

Run:

```bash
rg -n 'github.com/c1emon/gcommon(?!/v2|/.+/v2)' -g '*.go'
```

Expected:

```text
No matches.
```

If `rg` does not support look-around in this environment, run:

```bash
rg -n 'github.com/c1emon/gcommon' -g '*.go'
```

and manually confirm every match has the correct `/v2` segment.

## Task 5: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `httpx/README.md`
- Modify: `ginx/README.md`
- Modify: `health/README.md`
- Modify: `health/http/README.md`
- Modify: `health/gin/README.md`
- Modify: `server/README.md`
- Modify: `service/README.md`
- Modify: `logx/README.md`
- Modify: `docs/superpowers/plans/2026-05-05-httpx-session-client.md` only if keeping historical plan examples current is desired.

- [ ] **Step 1: Update root README module table**

Use `/v2` module paths in the module table:

```markdown
| `github.com/c1emon/gcommon/v2` | [`.`](.) | Core: `util`, `optional`, `errorx`, `data`, `stack`, `reader`, `service`, `server`, `pinger` |
| `github.com/c1emon/gcommon/health/v2` | [`health`](health/) | Health JSON ([hellofresh/health-go](https://github.com/hellofresh/health-go)); nested packages [`health/http`](health/http), [`health/gin`](health/gin) — see [`health/README.md`](health/README.md) |
| `github.com/c1emon/gcommon/httpx/v2` | [`httpx`](httpx/) | HTTP client (`imroc/req`) |
| `github.com/c1emon/gcommon/ginx/v2` | [`ginx`](ginx/) | Gin helpers |
| `github.com/c1emon/gcommon/cloud/v2` | [`cloud`](cloud/) | Service discovery / Consul |
| `github.com/c1emon/gcommon/gormx/v2` | [`gormx`](gormx/) | GORM helpers |
| `github.com/c1emon/gcommon/cachex/v2` | [`cachex`](cachex/) | Cache interfaces + memory backend |
| `github.com/c1emon/gcommon/logx/v2` | [`logx`](logx/) | Pure [`log/slog`](https://pkg.go.dev/log/slog) helpers (handlers, `Init` / `Default`, context attrs) |
| `github.com/c1emon/gcommon/test/v2` | [`test`](test/) | Integration tests (optional `INTEGRATION=1`) |
```

- [ ] **Step 2: Update root README install examples**

Use:

```bash
go get github.com/c1emon/gcommon/v2@v2.0.0
go get github.com/c1emon/gcommon/v2@v2.0.0 github.com/c1emon/gcommon/httpx/v2@v2.3.0
go get github.com/c1emon/gcommon/health/v2/http@v2.2.0
go get github.com/c1emon/gcommon/health/v2/gin@v2.2.0
```

If Task 1 selected a different root version, replace `v2.0.0` consistently.

- [ ] **Step 3: Replace `v0.0.0` development-version wording**

Replace the old sentence about nested modules requiring `v0.0.0` with:

```markdown
Nested modules declare real released versions for internal dependencies and keep local `replace` directives for in-repo development. Published consumers resolve directory-prefixed module tags such as `httpx/v2.3.0` and `ginx/v2.2.0`.
```

- [ ] **Step 4: Update each nested README import example**

Examples:

```go
import "github.com/c1emon/gcommon/httpx/v2"
```

```go
import "github.com/c1emon/gcommon/logx/v2"
```

```go
import health "github.com/c1emon/gcommon/health/v2"
import healthgin "github.com/c1emon/gcommon/health/v2/gin"
```

```go
import httphealth "github.com/c1emon/gcommon/health/v2/http"
```

- [ ] **Step 5: Verify old docs paths**

Run:

```bash
rg -n 'github.com/c1emon/gcommon' -g '*.md'
```

Expected:

```text
Every live README example uses the correct /v2 path.
Historical plan files may either be updated or explicitly left as historical snapshots.
```

## Task 6: Normalize Go Modules

**Files:**
- Modify: affected `go.mod` and `go.sum` files only if Go tooling changes them.
- Modify: `go.work.sum` only if Go tooling changes it.

- [ ] **Step 1: Run module tidy per module**

Run from each module directory:

```bash
go mod tidy
```

Directories:

```text
.
httpx
ginx
cloud
gormx
cachex
health
logx
test
```

Expected:

```text
Each command exits 0.
Only expected go.mod/go.sum/go.work.sum changes appear.
```

- [ ] **Step 2: Do not run `go work sync`**

Do not run:

```bash
go work sync
```

The repository README already warns that this can try to resolve local development versions from the network.

## Task 7: Verify Build And Tests

**Files:**
- No intentional edits.

- [ ] **Step 1: Verify package listing**

Run:

```bash
go list ./...
go list ./... ./httpx/... ./ginx/... ./health/... ./cloud/... ./gormx/... ./cachex/... ./logx/... ./test/...
```

Expected:

```text
All packages list successfully with /v2 import paths.
No package references github.com/c1emon/gcommon without the correct /v2 segment.
```

- [ ] **Step 2: Run all module tests**

Run:

```bash
go test ./... \
  ./httpx/... ./ginx/... ./health/... ./cloud/... ./gormx/... ./cachex/... ./logx/... ./test/... \
  -count=1 -timeout 120s
```

Expected:

```text
All packages pass.
```

- [ ] **Step 3: Verify no old runtime import paths remain**

Run:

```bash
rg -n 'github.com/c1emon/gcommon' -g '*.go' -g 'go.mod'
```

Expected:

```text
All matches are either:
- github.com/c1emon/gcommon/v2
- github.com/c1emon/gcommon/<nested-module>/v2
```

## Task 8: GitNexus Final Scope Check

**Files:**
- No intentional edits.

- [ ] **Step 1: Detect affected symbols and flows**

Run:

```bash
gitnexus detect-changes --repo gcommon --scope all
```

or use MCP:

```text
detect_changes(repo: "gcommon", scope: "all")
```

Expected:

```text
Affected changes are module-path/import/documentation changes.
No unexpected behavioral flow changes are reported.
```

- [ ] **Step 2: Review git diff**

Run:

```bash
git status --short
git diff -- go.mod '*/go.mod' '*.go' '*.md' go.work go.work.sum
```

Expected:

```text
Diff contains only v2 module path migration, internal dependency version updates, documentation updates, and legitimate module metadata changes.
```

## Task 9: Commit

**Files:**
- Commit only files touched for the v2 module migration.

- [ ] **Step 1: Stage scoped files**

Run:

```bash
git add go.mod go.sum go.work go.work.sum \
  httpx ginx cloud gormx cachex health logx test \
  README.md server/README.md service/README.md docs/superpowers/plans/2026-05-06-v2-independent-module-versioning.md
```

If `docs/superpowers/plans/2026-05-05-httpx-session-client.md` was updated, include it deliberately.

- [ ] **Step 2: Commit**

Run:

```bash
git commit -m "chore: prepare independent v2 module releases"
```

Expected:

```text
Commit succeeds and excludes unrelated files.
```

## Task 10: Tag Releases

**Files:**
- No file edits.

- [ ] **Step 1: Create root tag only after Task 1 decision**

If using the requested root version and it is safe to retag:

```bash
git tag -f v2.0.0
```

If using the preferred replacement version:

```bash
git tag v2.4.0
```

- [ ] **Step 2: Create nested module tags**

Run:

```bash
git tag httpx/v2.3.0
git tag ginx/v2.2.0
git tag cloud/v2.2.0
git tag gormx/v2.2.0
git tag cachex/v2.2.0
git tag health/v2.2.0
git tag logx/v2.2.0
git tag test/v2.2.0
```

- [ ] **Step 3: Verify tags point at the intended commit**

Run:

```bash
git rev-parse HEAD
git rev-list -n 1 httpx/v2.3.0
git rev-list -n 1 ginx/v2.2.0
git rev-list -n 1 cloud/v2.2.0
git rev-list -n 1 gormx/v2.2.0
git rev-list -n 1 cachex/v2.2.0
git rev-list -n 1 health/v2.2.0
git rev-list -n 1 logx/v2.2.0
git rev-list -n 1 test/v2.2.0
```

Expected:

```text
Every nested tag resolves to the release commit.
```

## Task 11: Optional Remote/Proxy Validation

**Files:**
- No file edits.

- [ ] **Step 1: Push tags after human approval**

Run only after confirming the tag policy:

```bash
git push origin <root-tag>
git push origin httpx/v2.3.0 ginx/v2.2.0 cloud/v2.2.0 gormx/v2.2.0 cachex/v2.2.0 health/v2.2.0 logx/v2.2.0 test/v2.2.0
```

- [ ] **Step 2: Validate from a temporary module**

Run in a temporary directory outside the repository:

```bash
go mod init tmp-gcommon-v2-check
go get github.com/c1emon/gcommon/v2@<root-version>
go get github.com/c1emon/gcommon/httpx/v2@v2.3.0
go get github.com/c1emon/gcommon/ginx/v2@v2.2.0
go get github.com/c1emon/gcommon/health/v2@v2.2.0
go list -m all
```

Expected:

```text
All requested modules resolve from the remote without local replace directives.
```

## Self-Review

- Spec coverage: the plan covers independent nested module tags, the requested `httpx` version, the requested `v2.2.0` versions for the other nested modules, internal dependency versions, source imports, docs, tests, and GitNexus checks.
- Placeholder scan: no task uses TBD/TODO/fill-in placeholders.
- Type consistency: all planned module paths use the Go v2 semantic import suffix consistently.
- Known unresolved decision: the requested root `v2.0.0` conflicts with an existing local root tag. This must be resolved before implementation or tagging.
