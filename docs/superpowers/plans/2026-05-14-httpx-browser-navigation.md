# httpx Browser Navigation State Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `httpx` browser navigation state distinguish page navigations from XHR/fetch requests while exposing explicit page context and state-update controls.

**Architecture:** Keep browser semantics in `httpx/browser_navigation.go`, and store the navigation state pointer on the `httpx.Client` wrapper created in `httpx/factory.go`. Resolve effective request kind once per request, use it for both header generation and state update policy, and expose client/request helpers that write to the existing browser navigation state through controlled internal mechanisms.

**Tech Stack:** Go, `github.com/imroc/req/v3`, existing `httpx` package tests with `net/http/httptest`.

---

## File Structure

- Modify: `httpx/browser_navigation.go` — add page context/state-update APIs and make state writes conditional on effective request kind.
- Modify: `httpx/client.go` — carry the unexported browser navigation state pointer through `Client` and `Clone`.
- Modify: `httpx/factory.go` — attach the installed browser navigation state to newly created `Client` wrappers.
- Modify: `httpx/browser_navigation_test.go` — add behavior-first tests for the requested navigation/XHR/current-page semantics.
- Create: `docs/superpowers/specs/2026-05-14-httpx-browser-navigation-design.md` — approved design spec.
- Create: `docs/superpowers/plans/2026-05-14-httpx-browser-navigation.md` — this implementation plan.

## Task 1: Add failing tests for navigation state semantics

**Files:**
- Modify: `httpx/browser_navigation_test.go`

- [ ] **Step 1: Add tests**

Add tests that perform `GET /login` as navigation, multiple `AsXHR()` POSTs, and final `AsNavigation()` POST, asserting XHR requests use `/login` as `Referer` and do not become the next page URL.

- [ ] **Step 2: Run focused test and verify RED**

Run: `go test ./httpx -run 'TestBrowserNavigation_(XHRDoesNotUpdateCurrentPage|MultipleXHRsKeepFinalNavigationReferer)' -count=1`

Expected: FAIL because current implementation remembers XHR URLs.

## Task 2: Add failing tests for explicit page context APIs

**Files:**
- Modify: `httpx/browser_navigation_test.go`

- [ ] **Step 1: Add tests**

Add tests for `Client.SetBrowserCurrentURL(rawURL)` and `Request.WithBrowserPageURL(rawURL)` asserting `Referer` is based on the explicit page context.

- [ ] **Step 2: Run focused test and verify RED**

Run: `go test ./httpx -run 'TestBrowserNavigation_(SetBrowserCurrentURL|WithBrowserPageURL)' -count=1`

Expected: FAIL to compile because APIs do not exist yet.

## Task 3: Add failing tests for request state-update override

**Files:**
- Modify: `httpx/browser_navigation_test.go`

- [ ] **Step 1: Add tests**

Add a test where a navigation request calls `WithoutBrowserNavigationStateUpdate()` and the next navigation still uses the original page as `Referer`.

- [ ] **Step 2: Run focused test and verify RED**

Run: `go test ./httpx -run TestBrowserNavigation_WithoutBrowserNavigationStateUpdate -count=1`

Expected: FAIL to compile because API does not exist yet.

## Task 4: Implement minimal API and state-update policy

**Files:**
- Modify: `httpx/browser_navigation.go`

- [ ] **Step 1: Add context keys and helpers**

Add internal context values for request page URL and request state update override.

- [ ] **Step 2: Add public methods**

Implement `Client.SetBrowserCurrentURL`, `Request.WithBrowserPageURL`, `Request.WithBrowserStateUpdate`, and `Request.WithoutBrowserNavigationStateUpdate`.

- [ ] **Step 3: Resolve request metadata once**

Make `apply` return effective request metadata or otherwise store it in context so post-round-trip update logic can update only navigation requests unless explicitly overridden.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `go test ./httpx -run TestBrowserNavigation -count=1`

Expected: PASS.

## Task 5: Run full verification

**Files:**
- All modified files

- [ ] **Step 1: Run package tests**

Run: `go test ./httpx/... -count=1`

Expected: PASS.

- [ ] **Step 2: Run GitNexus change detection**

Run: `gitnexus_detect_changes(scope="all", repo="gcommon")`

Expected: changed symbols are limited to httpx browser navigation behavior and docs/tests.

## Self-Review

- Spec coverage: all 6 items from `httpx.md` map to tasks 1-4.
- Placeholder scan: no placeholders remain.
- Type consistency: public methods consistently use `Client` and `Request` wrappers already defined in `httpx/client.go`.
