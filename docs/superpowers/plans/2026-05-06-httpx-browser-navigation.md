# httpx Browser Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a generic browser navigation layer to `httpx.Client` so caller-owned clients can maintain `Referer`, `Origin`, `Sec-Fetch-*`, and HTML/XHR headers across a sequential browser-like session.

**Architecture:** Keep browser fingerprinting in `WithBrowser(...)` and add `WithBrowserNavigation(...)` as a per-client profile option. Each `NewClient` gets a fresh navigation state; `Client.Clone()` shares that state through req wrapper closures. Capture request-explicit headers before req merges common headers, then apply dynamic headers in the client round-trip wrapper after URL/body parsing and before transport send.

**Tech Stack:** Go 1.25, `github.com/imroc/req/v3`, `net/http`, `net/url`, `sync`, `net/http/httptest`, GitNexus.

---

## Spec

Use `docs/superpowers/specs/2026-05-06-httpx-browser-navigation.md` as the source of truth. `httpx.md` remains the requirements input and should not be edited unless the user explicitly asks to revise requirements.

## File Map

- Create: `httpx/browser_navigation.go`
  - Public navigation options, `ReferrerPolicy`, request kind API, stateful wrapper installation, header computation helpers.
- Create: `httpx/browser_navigation_test.go`
  - End-to-end header behavior tests with `httptest`.
- Modify: `httpx/client_options.go`
  - Add `browserNavigation *browserNavigationConfig` to `clientRegisterOpts`.
  - Deep-copy the config in `clone()`.
- Modify: `httpx/factory.go`
  - Install navigation on each created req client after request interceptors and before logging/response hooks return.
- Modify: `httpx/client.go`
  - Keep `Clone()` behavior; add no public state fields. The clone sharing behavior comes from req copying wrapper closures.
- Modify: `httpx/doc.go`
  - Mention browser navigation and the request hook/round-trip ordering.
- Modify: `httpx/README.md`
  - Document public API, header precedence, clone/session semantics, and browser session example.

## GitNexus Controls

Pre-plan impact baseline already returned LOW risk for:

- `Client`
- `Client.Clone`
- `ClientFactory.buildReqClient`
- `clientRegisterOpts`
- `applyBrowserProfile`

Before implementation edits, rerun impact for any symbol being modified:

```text
impact(repo: "gcommon", target: "Client", file_path: "httpx/client.go", kind: "Struct", direction: "upstream")
impact(repo: "gcommon", target: "Clone", file_path: "httpx/client.go", kind: "Method", direction: "upstream")
impact(repo: "gcommon", target: "buildReqClient", file_path: "httpx/factory.go", kind: "Method", direction: "upstream")
impact(repo: "gcommon", target: "clientRegisterOpts", file_path: "httpx/client_options.go", kind: "Struct", direction: "upstream")
impact(repo: "gcommon", target: "applyBrowserProfile", file_path: "httpx/browser.go", kind: "Function", direction: "upstream")
```

If any result is HIGH or CRITICAL, stop and report direct callers, affected processes, and risk before editing. Before any commit, run:

```text
detect_changes(repo: "gcommon", scope: "all")
```

## Task 1: Add Public API Tests

**Files:**
- Create: `httpx/browser_navigation_test.go`

- [ ] **Step 1: Add tests for first request, same-origin, cross-site, POST origin, and HTML defaults**

Create `httpx/browser_navigation_test.go` with:

```go
package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c1emon/gcommon/httpx/v2"
)

type seenNavigationHeaders struct {
	Accept                   string `json:"accept"`
	Referer                  string `json:"referer"`
	Origin                   string `json:"origin"`
	SecFetchSite             string `json:"secFetchSite"`
	SecFetchMode             string `json:"secFetchMode"`
	SecFetchDest             string `json:"secFetchDest"`
	SecFetchUser             string `json:"secFetchUser"`
	UpgradeInsecureRequests string `json:"upgradeInsecureRequests"`
	XRequestedWith           string `json:"xRequestedWith"`
}

func navigationRecorder(t *testing.T) (*httptest.Server, *[]seenNavigationHeaders) {
	t.Helper()
	var seen []seenNavigationHeaders
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := seenNavigationHeaders{
			Accept:                   r.Header.Get("Accept"),
			Referer:                  r.Header.Get("Referer"),
			Origin:                   r.Header.Get("Origin"),
			SecFetchSite:             r.Header.Get("Sec-Fetch-Site"),
			SecFetchMode:             r.Header.Get("Sec-Fetch-Mode"),
			SecFetchDest:             r.Header.Get("Sec-Fetch-Dest"),
			SecFetchUser:             r.Header.Get("Sec-Fetch-User"),
			UpgradeInsecureRequests: r.Header.Get("Upgrade-Insecure-Requests"),
			XRequestedWith:           r.Header.Get("X-Requested-With"),
		}
		seen = append(seen, h)
		_ = json.NewEncoder(w).Encode(h)
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

func TestBrowserNavigation_firstAndSameOriginNavigationHeaders(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(srv.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/login?next=%2Fhome"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/home"); err != nil {
		t.Fatal(err)
	}

	first := (*seen)[0]
	if first.SecFetchSite != "none" {
		t.Fatalf("first Sec-Fetch-Site: want none, got %q", first.SecFetchSite)
	}
	if first.Referer != "" {
		t.Fatalf("first Referer: want empty, got %q", first.Referer)
	}
	if first.SecFetchMode != "navigate" || first.SecFetchDest != "document" || first.SecFetchUser != "?1" {
		t.Fatalf("first navigation headers: %+v", first)
	}
	if first.UpgradeInsecureRequests != "1" {
		t.Fatalf("first Upgrade-Insecure-Requests: want 1, got %q", first.UpgradeInsecureRequests)
	}

	second := (*seen)[1]
	if second.SecFetchSite != "same-origin" {
		t.Fatalf("second Sec-Fetch-Site: want same-origin, got %q", second.SecFetchSite)
	}
	wantReferer := srv.URL + "/login?next=%2Fhome"
	if second.Referer != wantReferer {
		t.Fatalf("second Referer: want %q, got %q", wantReferer, second.Referer)
	}
}

func TestBrowserNavigation_crossOriginRefererUsesOrigin(t *testing.T) {
	first, _ := navigationRecorder(t)
	second, seenSecond := navigationRecorder(t)

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get(first.URL + "/start?ticket=abc"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get(second.URL + "/next"); err != nil {
		t.Fatal(err)
	}

	got := (*seenSecond)[0]
	if got.SecFetchSite != "cross-site" {
		t.Fatalf("Sec-Fetch-Site: want cross-site, got %q", got.SecFetchSite)
	}
	if got.Referer != first.URL+"/" {
		t.Fatalf("Referer: want %q, got %q", first.URL+"/", got.Referer)
	}
}

func TestBrowserNavigation_postAddsCurrentOrigin(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().Post("/submit"); err != nil {
		t.Fatal(err)
	}
	if (*seen)[0].Origin != srv.URL {
		t.Fatalf("Origin: want %q, got %q", srv.URL, (*seen)[0].Origin)
	}
}
```

- [ ] **Step 2: Run the focused tests and confirm they fail**

Run:

```bash
go test ./httpx -run BrowserNavigation -count=1
```

Expected: build fails because `WithBrowserNavigation` does not exist.

## Task 2: Add API Types And Option Storage

**Files:**
- Create: `httpx/browser_navigation.go`
- Modify: `httpx/client_options.go`

- [ ] **Step 1: Add public API and config types**

Create `httpx/browser_navigation.go` with this initial content:

```go
package httpx

import (
	"net/url"
	"sync"

	"github.com/c1emon/gcommon/v2/util"
)

type BrowserNavigationOption util.Option[browserNavigationConfig]

type ReferrerPolicy string

const (
	ReferrerPolicyNoReferrer                  ReferrerPolicy = "no-referrer"
	ReferrerPolicyOrigin                      ReferrerPolicy = "origin"
	ReferrerPolicyStrictOriginWhenCrossOrigin ReferrerPolicy = "strict-origin-when-cross-origin"
)

type BrowserRequestKind int

const (
	BrowserRequestAuto BrowserRequestKind = iota
	BrowserRequestNavigation
	BrowserRequestXHR
)

const browserRequestKindHeader = "X-Httpx-Browser-Request-Kind"

type browserNavigationConfig struct {
	referrerPolicy        ReferrerPolicy
	xhrForJSON            bool
	defaultXRequestedWith bool
	defaultSecFetchUser   bool
}

type browserNavigationState struct {
	mu      sync.Mutex
	lastURL *url.URL
}

func defaultBrowserNavigationConfig() *browserNavigationConfig {
	return &browserNavigationConfig{
		referrerPolicy:        ReferrerPolicyStrictOriginWhenCrossOrigin,
		xhrForJSON:            true,
		defaultXRequestedWith: false,
		defaultSecFetchUser:   true,
	}
}

func (c *browserNavigationConfig) clone() *browserNavigationConfig {
	if c == nil {
		return nil
	}
	out := *c
	return &out
}

func WithBrowserNavigation(opts ...BrowserNavigationOption) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		cfg := defaultBrowserNavigationConfig()
		for _, opt := range opts {
			opt.Apply(cfg)
		}
		o.browserNavigation = cfg
	})
}

func WithReferrerPolicy(policy ReferrerPolicy) BrowserNavigationOption {
	return util.WrapFuncOption(func(c *browserNavigationConfig) {
		if policy != "" {
			c.referrerPolicy = policy
		}
	})
}

func WithXHRForJSON(enabled bool) BrowserNavigationOption {
	return util.WrapFuncOption(func(c *browserNavigationConfig) {
		c.xhrForJSON = enabled
	})
}

func WithDefaultXRequestedWith(enabled bool) BrowserNavigationOption {
	return util.WrapFuncOption(func(c *browserNavigationConfig) {
		c.defaultXRequestedWith = enabled
	})
}

func WithDefaultSecFetchUser(enabled bool) BrowserNavigationOption {
	return util.WrapFuncOption(func(c *browserNavigationConfig) {
		c.defaultSecFetchUser = enabled
	})
}

func (r *Request) WithBrowserRequestKind(kind BrowserRequestKind) *Request {
	if r == nil || r.Request == nil {
		return r
	}
	switch kind {
	case BrowserRequestNavigation:
		r.SetHeader(browserRequestKindHeader, "navigation")
	case BrowserRequestXHR:
		r.SetHeader(browserRequestKindHeader, "xhr")
	default:
		r.SetHeader(browserRequestKindHeader, "auto")
	}
	return r
}

func (r *Request) AsXHR() *Request {
	return r.WithBrowserRequestKind(BrowserRequestXHR)
}

func (r *Request) AsNavigation() *Request {
	return r.WithBrowserRequestKind(BrowserRequestNavigation)
}
```

- [ ] **Step 2: Store config on client registration options**

In `httpx/client_options.go`, add this field to `clientRegisterOpts`:

```go
browserNavigation *browserNavigationConfig
```

In `func (o clientRegisterOpts) clone() clientRegisterOpts`, add:

```go
out.browserNavigation = o.browserNavigation.clone()
```

- [ ] **Step 3: Run tests and confirm API compiles but behavior is missing**

Run:

```bash
go test ./httpx -run BrowserNavigation -count=1
```

Expected: tests fail at runtime because no navigation wrapper is installed yet.

## Task 3: Implement Header Computation Helpers

**Files:**
- Modify: `httpx/browser_navigation.go`

- [ ] **Step 1: Add helper unit tests for request kind and referrer policy**

Append to `httpx/browser_navigation_test.go`:

```go
func TestBrowserNavigation_jsonDefaultsToXHR(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().SetBodyJsonMarshal(map[string]string{"hello": "world"}).Post("/api"); err != nil {
		t.Fatal(err)
	}
	got := (*seen)[0]
	if got.Accept != "application/json, text/plain, */*" {
		t.Fatalf("Accept: want JSON accept, got %q", got.Accept)
	}
	if got.SecFetchMode != "cors" || got.SecFetchDest != "empty" {
		t.Fatalf("XHR fetch headers: %+v", got)
	}
	if got.SecFetchUser != "" || got.UpgradeInsecureRequests != "" {
		t.Fatalf("XHR should not include navigation-only headers: %+v", got)
	}
}

func TestBrowserNavigation_explicitRequestKindOverridesContentType(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().AsNavigation().SetBodyJsonMarshal(map[string]string{"hello": "world"}).Post("/form"); err != nil {
		t.Fatal(err)
	}
	got := (*seen)[0]
	if got.SecFetchMode != "navigate" || got.SecFetchDest != "document" {
		t.Fatalf("explicit navigation should win over JSON content type: %+v", got)
	}
}
```

- [ ] **Step 2: Add helper functions**

Add `net/http` and `strings` to the import list in `httpx/browser_navigation.go`.

Append these helpers to `httpx/browser_navigation.go`:

```go
func sameOrigin(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

func originString(u *url.URL) string {
	if u == nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func originReferer(u *url.URL) string {
	if o := originString(u); o != "" {
		return o + "/"
	}
	return ""
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	out := *u
	return &out
}

func explicitHeaderKeys(h http.Header) map[string]struct{} {
	if len(h) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(h))
	for k := range h {
		out[http.CanonicalHeaderKey(k)] = struct{}{}
	}
	delete(out, http.CanonicalHeaderKey(browserRequestKindHeader))
	return out
}

func hasExplicitHeader(explicit map[string]struct{}, key string) bool {
	_, ok := explicit[http.CanonicalHeaderKey(key)]
	return ok
}

func setDynamicHeader(h http.Header, explicit map[string]struct{}, key, value string) {
	if value == "" || hasExplicitHeader(explicit, key) {
		return
	}
	h.Set(key, value)
}

func browserRequestKindFromHeader(h http.Header) BrowserRequestKind {
	switch strings.ToLower(h.Get(browserRequestKindHeader)) {
	case "navigation":
		return BrowserRequestNavigation
	case "xhr":
		return BrowserRequestXHR
	default:
		return BrowserRequestAuto
	}
}

func isJSONContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	return ct == "application/json" || strings.HasPrefix(ct, "application/json;")
}

func (c *browserNavigationConfig) effectiveRequestKind(h http.Header) BrowserRequestKind {
	switch browserRequestKindFromHeader(h) {
	case BrowserRequestNavigation:
		return BrowserRequestNavigation
	case BrowserRequestXHR:
		return BrowserRequestXHR
	}
	if c.xhrForJSON && isJSONContentType(h.Get("Content-Type")) {
		return BrowserRequestXHR
	}
	return BrowserRequestNavigation
}

func (c *browserNavigationConfig) refererFor(prev, current *url.URL) string {
	if prev == nil {
		return ""
	}
	switch c.referrerPolicy {
	case ReferrerPolicyNoReferrer:
		return ""
	case ReferrerPolicyOrigin:
		return originReferer(prev)
	case ReferrerPolicyStrictOriginWhenCrossOrigin, "":
		if sameOrigin(prev, current) {
			return prev.String()
		}
		return originReferer(prev)
	default:
		if sameOrigin(prev, current) {
			return prev.String()
		}
		return originReferer(prev)
	}
}

func secFetchSite(prev, current *url.URL) string {
	if prev == nil {
		return "none"
	}
	if sameOrigin(prev, current) {
		return "same-origin"
	}
	return "cross-site"
}
```

- [ ] **Step 3: Run focused tests**

Run:

```bash
go test ./httpx -run BrowserNavigation -count=1
```

Expected: tests still fail until the wrapper applies helpers.

## Task 4: Install Navigation Wrapper

**Files:**
- Modify: `httpx/browser_navigation.go`
- Modify: `httpx/factory.go`

- [ ] **Step 1: Add explicit-header capture and round-trip application**

Append to `httpx/browser_navigation.go`:

```go
const (
	htmlAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
	jsonAccept = "application/json, text/plain, */*"
)

type explicitHeadersKey struct{}

func installBrowserNavigation(c *req.Client, cfg *browserNavigationConfig) {
	if c == nil || cfg == nil {
		return
	}
	state := &browserNavigationState{}

	c.OnBeforeRequest(func(_ *req.Client, r *req.Request) error {
		if r == nil {
			return nil
		}
		r.SetContext(contextWithExplicitHeaders(r.Context(), explicitHeaderKeys(r.Headers)))
		return nil
	})

	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(r *req.Request) (*req.Response, error) {
			if r == nil {
				return rt.RoundTrip(r)
			}
			explicit := explicitHeadersFromContext(r.Context())
			state.apply(cfg, r, explicit)
			resp, err := rt.RoundTrip(r)
			if resp != nil && resp.Response != nil && resp.Response.Request != nil {
				state.remember(resp.Response.Request.URL)
			}
			return resp, err
		}
	})
}

func contextWithExplicitHeaders(ctx context.Context, explicit map[string]struct{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, explicitHeadersKey{}, explicit)
}

func explicitHeadersFromContext(ctx context.Context) map[string]struct{} {
	if ctx == nil {
		return nil
	}
	explicit, _ := ctx.Value(explicitHeadersKey{}).(map[string]struct{})
	return explicit
}

func (s *browserNavigationState) remember(u *url.URL) {
	if s == nil || u == nil {
		return
	}
	s.mu.Lock()
	s.lastURL = cloneURL(u)
	s.mu.Unlock()
}

func (s *browserNavigationState) previous() *url.URL {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneURL(s.lastURL)
}

func (s *browserNavigationState) apply(cfg *browserNavigationConfig, r *req.Request, explicit map[string]struct{}) {
	if s == nil || cfg == nil || r == nil || r.URL == nil {
		return
	}
	h := r.Headers
	if h == nil {
		h = make(http.Header)
		r.Headers = h
	}
	current := r.URL
	prev := s.previous()

	setDynamicHeader(h, explicit, "Referer", cfg.refererFor(prev, current))
	setDynamicHeader(h, explicit, "Sec-Fetch-Site", secFetchSite(prev, current))
	if strings.EqualFold(r.Method, http.MethodPost) {
		setDynamicHeader(h, explicit, "Origin", originString(current))
	}

	switch cfg.effectiveRequestKind(h) {
	case BrowserRequestXHR:
		setDynamicHeader(h, explicit, "Accept", jsonAccept)
		setDynamicHeader(h, explicit, "Sec-Fetch-Mode", "cors")
		setDynamicHeader(h, explicit, "Sec-Fetch-Dest", "empty")
		if cfg.defaultXRequestedWith {
			setDynamicHeader(h, explicit, "X-Requested-With", "XMLHttpRequest")
		}
	default:
		setDynamicHeader(h, explicit, "Accept", htmlAccept)
		setDynamicHeader(h, explicit, "Sec-Fetch-Mode", "navigate")
		setDynamicHeader(h, explicit, "Sec-Fetch-Dest", "document")
		if cfg.defaultSecFetchUser {
			setDynamicHeader(h, explicit, "Sec-Fetch-User", "?1")
		}
		setDynamicHeader(h, explicit, "Upgrade-Insecure-Requests", "1")
	}
	h.Del(browserRequestKindHeader)
}
```

Add `context` and `github.com/imroc/req/v3` to the import list in `httpx/browser_navigation.go`.

- [ ] **Step 2: Install wrapper in factory build**

In `httpx/factory.go`, after registering global/client request interceptors and before request logging, add:

```go
if o.browserNavigation != nil {
	installBrowserNavigation(c, o.browserNavigation)
}
```

This ordering captures headers after user request interceptors have had a chance to mark explicit headers, and applies dynamic headers later in the round-trip wrapper.

- [ ] **Step 3: Run focused tests**

Run:

```bash
go test ./httpx -run BrowserNavigation -count=1
```

Expected: Task 1 and Task 3 tests pass.

## Task 5: Cover Precedence, Browser Profile, Isolation, Clone, And Redirect

**Files:**
- Modify: `httpx/browser_navigation_test.go`

- [ ] **Step 1: Add header precedence test**

Append:

```go
func TestBrowserNavigation_doesNotOverrideSingleRequestExplicitHeaders(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(srv.URL),
		httpx.WithBrowserNavigation(),
		httpx.WithHeader("Sec-Fetch-Site", "profile-default"),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().
		SetHeader("Sec-Fetch-Site", "manual").
		SetHeader("Referer", "https://manual.example/path").
		SetHeader("Accept", "text/custom").
		Get("/manual"); err != nil {
		t.Fatal(err)
	}

	got := (*seen)[0]
	if got.SecFetchSite != "manual" || got.Referer != "https://manual.example/path" || got.Accept != "text/custom" {
		t.Fatalf("explicit headers should win, got %+v", got)
	}
}
```

- [ ] **Step 2: Add browser profile dynamic override test**

Append:

```go
func TestBrowserNavigation_overridesBrowserProfileDynamicDefaults(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(srv.URL),
		httpx.WithBrowser(httpx.BrowserChrome),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	if (*seen)[1].SecFetchSite != "same-origin" {
		t.Fatalf("navigation should override browser profile static Sec-Fetch-Site, got %+v", (*seen)[1])
	}
}
```

- [ ] **Step 3: Add independent-client and clone-state tests**

Append:

```go
func TestBrowserNavigation_newClientStateIsolation(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())

	first := f.MustNewClient("browser")
	second := f.MustNewClient("browser")
	if _, err := first.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	if (*seen)[1].Referer != "" || (*seen)[1].SecFetchSite != "none" {
		t.Fatalf("second client should start with empty navigation state, got %+v", (*seen)[1])
	}
}

func TestBrowserNavigation_cloneSharesState(t *testing.T) {
	srv, seen := navigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())
	original := f.MustNewClient("browser")

	if _, err := original.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	clone := original.Clone()
	if _, err := clone.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	wantReferer := srv.URL + "/first"
	if (*seen)[1].Referer != wantReferer {
		t.Fatalf("clone should share navigation state, want Referer %q, got %+v", wantReferer, (*seen)[1])
	}
}
```

- [ ] **Step 4: Add redirect final URL test**

Append:

```go
func TestBrowserNavigation_remembersFinalURLAfterRedirect(t *testing.T) {
	var seen []seenNavigationHeaders
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/landing", http.StatusFound)
		default:
			seen = append(seen, seenNavigationHeaders{
				Referer:      r.Header.Get("Referer"),
				SecFetchSite: r.Header.Get("Sec-Fetch-Site"),
			})
			_, _ = w.Write([]byte("ok"))
		}
	}))
	t.Cleanup(srv.Close)

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBaseURL(srv.URL), httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/start"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/next"); err != nil {
		t.Fatal(err)
	}

	wantReferer := srv.URL + "/landing"
	if seen[1].Referer != wantReferer {
		t.Fatalf("Referer should use final redirect URL, want %q, got %+v", wantReferer, seen[1])
	}
}
```

- [ ] **Step 5: Run focused and package tests**

Run:

```bash
go test ./httpx -run BrowserNavigation -count=1
go test ./httpx -count=1
```

Expected: all `httpx` tests pass.

## Task 6: Update Documentation

**Files:**
- Modify: `httpx/doc.go`
- Modify: `httpx/README.md`

- [ ] **Step 1: Update package doc**

In `httpx/doc.go`, replace the feature sentence:

```go
// limiting (golang.org/x/time/rate), browser impersonation profiles, cookie jars,
// redirect policies, and opt-in JSON envelope error handling via interceptors.Error.
```

with:

```go
// limiting (golang.org/x/time/rate), browser impersonation profiles, browser
// navigation headers, cookie jars, redirect policies, and opt-in JSON envelope
// error handling via interceptors.Error.
```

Append to the ordering paragraph:

```go
// Browser navigation captures request-explicit headers before req merges common
// headers, then applies dynamic navigation headers in the client round-trip wrapper
// after URL and body parsing.
```

- [ ] **Step 2: Update README API list**

In `httpx/README.md`, add this item under `ClientOption`:

```markdown
- `WithBrowserNavigation(opts ...BrowserNavigationOption)`
```

Add a new `BrowserNavigationOption` list near the browser navigation section:

```markdown
### BrowserNavigationOption

- `WithReferrerPolicy(policy ReferrerPolicy)`
- `WithXHRForJSON(enabled bool)`
- `WithDefaultXRequestedWith(enabled bool)`
- `WithDefaultSecFetchUser(enabled bool)`
```

Under `Client 方法`, add:

```markdown
- `(*Request).WithBrowserRequestKind(kind BrowserRequestKind) *Request`
- `(*Request).AsXHR() *Request`
- `(*Request).AsNavigation() *Request`
```

- [ ] **Step 3: Add README browser navigation section**

Add a section after the browser profile section:

```markdown
## 浏览器导航 header

`WithBrowser(...)` 负责 TLS/HTTP2 指纹、浏览器默认 header 和 header order；`WithBrowserNavigation(...)` 负责单个 caller-owned client 内的动态导航 header：

- 首次请求自动设置 `Sec-Fetch-Site: none`，不设置 `Referer`
- 同源后续请求用上一跳完整 URL 作为 `Referer`
- 跨源后续请求按 `strict-origin-when-cross-origin` 收敛为上一跳 origin
- POST 请求自动补当前请求 `Origin`
- JSON 请求默认使用 XHR 风格 `Accept` / `Sec-Fetch-Mode` / `Sec-Fetch-Dest`
- HTML/form 请求默认使用 navigation 风格 header
- 单次请求显式设置的 header 不会被覆盖

```go
factory.RegisterProfile("flow-upstream",
	httpx.WithBrowser(httpx.BrowserChrome),
	httpx.WithBrowserNavigation(
		httpx.WithReferrerPolicy(httpx.ReferrerPolicyStrictOriginWhenCrossOrigin),
		httpx.WithXHRForJSON(true),
	),
)

session := factory.MustNewClient("flow-upstream")
_, _ = session.Req().Get("https://login.example.com/")
_, _ = session.Req().AsXHR().SetBodyJsonMarshal(map[string]string{"step": "check"}).Post("https://login.example.com/api/check")

noRedirect := session.Clone().SetRedirectPolicy(httpx.NoRedirectPolicy())
_, _ = noRedirect.Req().Get("https://login.example.com/redirect-once")
```

`ClientFactory` 不保存导航状态。每次 `NewClient` / `MustNewClient` 都是新的浏览器会话；`Clone()` 用于同一会话内的临时变体，并共享上一跳 URL。
```

- [ ] **Step 4: Run docs-adjacent tests**

Run:

```bash
go test ./httpx -count=1
```

Expected: all `httpx` tests pass.

## Task 7: Final Verification

**Files:**
- No planned edits beyond previous tasks.

- [ ] **Step 1: Run module tests**

Run:

```bash
go test ./httpx -count=1
```

Expected: PASS.

- [ ] **Step 2: Run workspace-level package listing**

Run:

```bash
go list ./... ./httpx/...
```

Expected: all packages list successfully.

- [ ] **Step 3: Run GitNexus change detection**

Run MCP:

```text
detect_changes(repo: "gcommon", scope: "all")
```

Expected: changes limited to `httpx` browser navigation symbols and docs. Any unexpected affected process must be inspected before commit.

## Spec Coverage Review

- Dynamic `Referer`, `Sec-Fetch-Site`, POST `Origin`: Task 1, Task 3, Task 4.
- HTML navigation and JSON/XHR defaults: Task 1, Task 3, Task 4.
- Single-request explicit header priority: Task 4 and Task 5.
- `WithBrowser(...)` coexistence and dynamic override: Task 5.
- New-client state isolation: Task 5.
- Clone shared state: Task 5.
- Redirect final URL: Task 5.
- Docs and session semantics: Task 6.
- GitNexus controls: GitNexus Controls section and Task 7.
