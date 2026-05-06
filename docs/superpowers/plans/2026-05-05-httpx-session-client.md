# httpx Client Factory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework `httpx` into a general-purpose client factory that registers named profiles and creates independent `*httpx.Client` instances for browser-style sessions.

**Architecture:** Replace the current `Manager` instance-registry model with `ClientFactory`: the factory stores profile configuration, and every `NewClient(profileName)` builds a fresh `req.Client` wrapper that is not tracked by the factory. `Client.Clone()` remains available only for short-lived variants inside one caller-owned session, such as temporarily disabling redirects. Business-envelope error handling becomes opt-in so third-party upstream JSON is not interpreted as gcommon business errors by default.

**Tech Stack:** Go 1.25, `github.com/imroc/req/v3`, `net/http`, `net/http/cookiejar`, `net/http/httptest`, `golang.org/x/net/publicsuffix`, GitNexus.

---

## Scope Decisions

- Breaking change: rename the public lifecycle concept from `Manager` to `ClientFactory`.
- Breaking change: `RegisterProfile` stores configuration only and does not return a `*Client`.
- Breaking change: `NewClientFactory()` does not enable gcommon business-error interception by default. Clients that need `{code,msg,data}` mapping must opt in with `WithBusinessError()` or `WithGlobalBusinessError()`.
- `ClientFactory` never stores, lists, closes, reuses, or pools created client instances. Instance lifecycle belongs to the caller, such as `sso-broker/internal/upstream.Session`.
- Session isolation should use `factory.MustNewClient("flow-upstream")` as the primary path. `Client.Clone()` is only for temporary variants of an existing session client.
- `httpx` exposes generic cookie, redirect, retry, interceptor, and business-error policies only. It does not implement `Referer`, `Sec-Fetch-*`, ISC SSO, PMS30, or downstream session semantics.
- `httpx.md` remains the input requirement document. Do not edit it during implementation unless the user explicitly asks to update requirements.

## GitNexus Impact Baseline

- `Manager` rename impact: LOW. Direct upstream is `NewManager`; indirect upstream is `InitDefaultManager`.
- `NewManager` rename impact: LOW. Direct upstream is `InitDefaultManager`.
- `InitDefaultManager` and `GetDefaultManager` impact: LOW. GitNexus found no upstream callers in this repository.
- `Manager.buildReqClient` impact: LOW. Direct upstream is `Manager.Register`; affected flow group is the `Register` construction path.
- `Manager.Register` impact: LOW. GitNexus found no non-test upstream callers in this repository; existing `httpx` tests call it directly.
- `Client` impact: LOW. No upstream symbols detected by GitNexus; adding explicit methods mainly shadows promoted `req.Client` methods with `httpx` return types.
- `clientRegisterOpts` impact: LOW. Direct upstream is `newClientRegisterOpts`, then the registration/build path.
- `interceptors.Error` impact: LOW. Direct upstream is the client build path. Error conversion flows include `Error -> HttpError`, `Error -> IOError`, `Error -> WithCause`, and `Error -> CommonError`.

No HIGH or CRITICAL impact was reported for the gcommon-side symbols. The downstream `sso-broker` `upstream.Session` migration remains CRITICAL and should be handled in that repository after this package is validated.

## Target Public API

```go
type ClientFactory struct {}
type FactoryOption util.Option[ClientFactory]

func NewClientFactory(opts ...FactoryOption) *ClientFactory
func InitDefaultClientFactory(opts ...FactoryOption)
func GetDefaultClientFactory() *ClientFactory

func (f *ClientFactory) RegisterProfile(name string, opts ...ClientOption)
func (f *ClientFactory) NewClient(name string, opts ...ClientOption) (*Client, bool)
func (f *ClientFactory) MustNewClient(name string, opts ...ClientOption) *Client
func (f *ClientFactory) ProfileNames() []string

func (c *Client) Clone() *Client
func (c *Client) SetCookieJar(jar http.CookieJar) *Client
func (c *Client) SetCookieJarFactory(factory CookieJarFactory) *Client
func (c *Client) GetCookies(rawURL string) ([]*http.Cookie, error)
func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client
```

Primary session usage:

```go
factory := httpx.NewClientFactory()
factory.RegisterProfile("flow-upstream",
	httpx.DisableRetry(),
	httpx.DisableBusinessError(),
	httpx.WithCookieJarFactory(newCookieJar),
)

sessionClient := factory.MustNewClient("flow-upstream")
noRedirect := sessionClient.Clone().SetRedirectPolicy(httpx.NoRedirectPolicy())
_, _ = sessionClient, noRedirect
```

## Requirement Coverage

| Requirement from `httpx.md` | Planned coverage |
| --- | --- |
| Business error interceptor can be disabled | Covered by opt-in business-error policy plus `DisableBusinessError()` and `DisableGlobalBusinessError()` explicit overrides. |
| Independent cookie jar per SSO login | Covered by `ClientFactory.NewClient` creating a fresh client per profile request and by `WithCookieJarFactory` running per created client. |
| Safe clone | Covered by `Client.Clone()` returning an unmanaged wrapper for temporary per-session variants. |
| Redirect policy per client or temporary clone | Covered by `WithRedirectPolicy` at profile/instance creation and `Client.SetRedirectPolicy` on session clones. |
| Non-idempotent requests should disable retry | Covered in `httpx/README.md` with a browser/session profile example using `DisableRetry()`. |
| Keep `sso-broker` session boundary | Covered as an integration handoff: `upstream.Session` owns the created `*httpx.Client` instance lifecycle. |
| Existing tests and new tests | Covered by focused `httptest` tests for factory lifecycle, business error, cookie isolation, clone isolation, and redirect behavior. |

## File Map

- Move `httpx/manager.go` to `httpx/factory.go`: `ClientFactory`, profile registry, default factory, and build logic.
- Move `httpx/manager_options.go` to `httpx/factory_options.go`: `FactoryOption` and global options.
- Modify `httpx/client_options.go`: add business-error, cookie jar, redirect policy state, and option-copy helpers.
- Modify `httpx/client.go`: add explicit clone, cookie, and redirect wrapper methods returning `*httpx.Client`.
- Create `httpx/cookie.go`: cookie jar option helpers and client method wrappers.
- Create `httpx/redirect.go`: public redirect type aliases and redirect policy helpers.
- Move `httpx/manager_test.go` to `httpx/factory_test.go`: factory profile/new-client tests.
- Modify `httpx/interceptor_test.go`: update business-error tests for opt-in behavior and disabled behavior.
- Modify `httpx/client_test.go`: add created-client and clone isolation tests.
- Create `httpx/redirect_test.go`: redirect policy tests.
- Modify `httpx/README.md`: document client factory, caller-owned sessions, business-error policy, cookie jars, redirects, and non-idempotent retry guidance.

### Task 1: Replace Manager With ClientFactory Profiles

**Files:**
- Move: `httpx/manager.go` -> `httpx/factory.go`
- Move: `httpx/manager_options.go` -> `httpx/factory_options.go`
- Modify: `httpx/client_options.go`
- Move: `httpx/manager_test.go` -> `httpx/factory_test.go`

- [ ] **Step 1: Write failing factory lifecycle tests**

Move `httpx/manager_test.go` to `httpx/factory_test.go` and replace the manager-specific tests with:

```go
package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFactory_profileHeaderAndInstanceOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Trace")))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalHeader("X-Trace", "global"))
	f.RegisterProfile("api", WithBaseURL(srv.URL), WithHeader("X-Trace", "profile"))

	c, ok := f.NewClient("api", WithHeader("X-Trace", "instance"))
	if !ok {
		t.Fatal("expected profile to exist")
	}
	resp, err := c.R().Get("/")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "instance" {
		t.Fatalf("want instance header to override profile/global, got %q", b)
	}
}

func TestClientFactory_newClientReturnsIndependentInstances(t *testing.T) {
	a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("a"))
	}))
	t.Cleanup(a.Close)

	f := NewClientFactory()
	f.RegisterProfile("api", WithBaseURL(a.URL))

	first := f.MustNewClient("api")
	second := f.MustNewClient("api")
	if first == second {
		t.Fatal("MustNewClient returned the same wrapper instance twice")
	}
	if first.Client == second.Client {
		t.Fatal("MustNewClient returned wrappers around the same req client")
	}
}

func TestClientFactory_unknownProfile(t *testing.T) {
	f := NewClientFactory()
	if c, ok := f.NewClient("missing"); ok || c != nil {
		t.Fatalf("NewClient missing profile: got client=%v ok=%v", c, ok)
	}
}

func TestDefaultClientFactory_initAndGet(t *testing.T) {
	defaultClientFactoryMu.Lock()
	prev := defaultClientFactory
	defaultClientFactoryMu.Unlock()
	t.Cleanup(func() {
		defaultClientFactoryMu.Lock()
		defaultClientFactory = prev
		defaultClientFactoryMu.Unlock()
	})

	InitDefaultClientFactory(WithGlobalHeader("X-Default-Test", "1"))
	got := GetDefaultClientFactory()
	if got == nil {
		t.Fatal("GetDefaultClientFactory returned nil after InitDefaultClientFactory")
	}
	got.RegisterProfile("smoke", WithBaseURL("https://example.com"))
	if names := got.ProfileNames(); len(names) != 1 || names[0] != "smoke" {
		t.Fatalf("expected one registered profile, got %+v", names)
	}

	InitDefaultClientFactory()
	if got2 := GetDefaultClientFactory(); got2 == got {
		t.Fatal("expected InitDefaultClientFactory to replace default factory instance")
	}
}
```

- [ ] **Step 2: Run tests and confirm old API does not satisfy the new lifecycle**

Run:

```bash
go test ./httpx -run 'TestClientFactory|TestDefaultClientFactory' -count=1
```

Expected result before implementation: build fails because `NewClientFactory`, `RegisterProfile`, `NewClient`, `MustNewClient`, and default factory symbols do not exist.

- [ ] **Step 3: Implement profile option copying**

In `httpx/client_options.go`, add a copy helper so stored profiles cannot be mutated when per-instance options are applied:

```go
func (o clientRegisterOpts) clone() clientRegisterOpts {
	out := o
	if o.headers != nil {
		out.headers = make(map[string]string, len(o.headers))
		for k, v := range o.headers {
			out.headers[k] = v
		}
	}
	out.clientReqInterceptors = append([]ReqInterceptor(nil), o.clientReqInterceptors...)
	out.clientRespInterceptors = append([]RespInterceptor(nil), o.clientRespInterceptors...)
	out.redirectPolicies = append([]RedirectPolicy(nil), o.redirectPolicies...)
	return out
}

func newClientRegisterOptsFrom(opts ...ClientOption) clientRegisterOpts {
	o := newClientRegisterOpts()
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return o
}
```

- [ ] **Step 4: Replace `Manager` with `ClientFactory`**

Rewrite the type and default singleton in `httpx/factory.go`:

```go
type ClientFactory struct {
	mu sync.RWMutex

	profiles map[string]clientRegisterOpts

	globalLimiter *rate.Limiter
	logger        *slog.Logger

	globalHeaders map[string]string

	globalReqInterceptors  []ReqInterceptor
	globalRespInterceptors []RespInterceptor

	globalRetry             RetryPolicy
	globalBrowserProfile    BrowserProfile
	hasGlobalBrowserProfile bool
	strictJSONType          bool
	globalBusinessError     bool
}

func NewClientFactory(opts ...FactoryOption) *ClientFactory {
	f := &ClientFactory{
		profiles: make(map[string]clientRegisterOpts),
	}
	for _, opt := range opts {
		opt.Apply(f)
	}
	return f
}

var (
	defaultClientFactoryMu sync.RWMutex
	defaultClientFactory   *ClientFactory
)

func InitDefaultClientFactory(opts ...FactoryOption) {
	f := NewClientFactory(opts...)
	defaultClientFactoryMu.Lock()
	defaultClientFactory = f
	defaultClientFactoryMu.Unlock()
}

func GetDefaultClientFactory() *ClientFactory {
	defaultClientFactoryMu.RLock()
	defer defaultClientFactoryMu.RUnlock()
	return defaultClientFactory
}
```

Do not keep `NewManager`, `Manager`, `InitDefaultManager`, or `GetDefaultManager` unless the user later asks for a compatibility layer.

- [ ] **Step 5: Add profile registration and instance creation**

Add these methods to `httpx/factory.go`:

```go
func (f *ClientFactory) RegisterProfile(name string, opts ...ClientOption) {
	o := newClientRegisterOptsFrom(opts...)

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.profiles == nil {
		f.profiles = make(map[string]clientRegisterOpts)
	}
	f.profiles[name] = o
}

func (f *ClientFactory) NewClient(name string, opts ...ClientOption) (*Client, bool) {
	f.mu.RLock()
	base, ok := f.profiles[name]
	f.mu.RUnlock()
	if !ok {
		return nil, false
	}

	o := base.clone()
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return &Client{Client: f.buildReqClient(name, &o), name: name}, true
}

func (f *ClientFactory) MustNewClient(name string, opts ...ClientOption) *Client {
	c, ok := f.NewClient(name, opts...)
	if !ok {
		panic(fmt.Sprintf("httpx: unknown client profile %q", name))
	}
	return c
}

func (f *ClientFactory) ProfileNames() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]string, 0, len(f.profiles))
	for n := range f.profiles {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
```

Import `sort` in `httpx/factory.go` so `ProfileNames()` is deterministic.

- [ ] **Step 6: Rename manager options to factory options**

In `httpx/factory_options.go`, replace:

```go
type ManagerOption util.Option[Manager]
```

with:

```go
type FactoryOption util.Option[ClientFactory]
```

Update every global option closure receiver from `func(m *Manager)` to `func(f *ClientFactory)`, and write fields on `f`.

- [ ] **Step 7: Keep build logic private to the factory**

Rename `func (m *Manager) buildReqClient(...)` to:

```go
func (f *ClientFactory) buildReqClient(name string, o *clientRegisterOpts) *req.Client
```

Inside the method, replace references to manager fields with factory fields. Keep the existing order: logger, base URL, timeout, headers, browser profile, retry, limiter, request hooks, logging hooks, response hooks, and business-error hook after Task 2.

- [ ] **Step 8: Verify factory lifecycle**

Run:

```bash
go test ./httpx -run 'TestClientFactory|TestDefaultClientFactory' -count=1
```

Expected result: all factory lifecycle tests pass.

- [ ] **Step 9: Commit**

```bash
git add httpx/factory.go httpx/factory_options.go httpx/client_options.go httpx/factory_test.go
git add -u httpx/manager.go httpx/manager_options.go httpx/manager_test.go
git commit -m "feat(httpx): introduce client factory profiles"
```

### Task 2: Make Business-Envelope Errors Explicit

**Files:**
- Modify: `httpx/client_options.go`
- Modify: `httpx/factory_options.go`
- Modify: `httpx/factory.go`
- Modify: `httpx/interceptor_test.go`

- [ ] **Step 1: Write failing tests for opt-in and disabled business errors**

Append these tests to `httpx/interceptor_test.go`:

```go
func TestBusinessError_defaultDoesNotMapEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"msg":"third-party","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("third-party")
	resp, err := f.MustNewClient("third-party").R().Get(srv.URL)
	if err != nil {
		t.Fatalf("default client should not map third-party envelope: %v", err)
	}
	if string(resp.Bytes()) != `{"code":400,"msg":"third-party","ts":1}` {
		t.Fatalf("response body changed: %q", string(resp.Bytes()))
	}
}

func TestBusinessError_clientOptInMapsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"msg":"bad","ts":1}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("biz", WithBusinessError())
	_, err := f.MustNewClient("biz").R().Get(srv.URL)
	if err == nil {
		t.Fatal("want business error when WithBusinessError is enabled")
	}
}

func TestBusinessError_globalEnableClientDisable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":123,"msg":"x"}`))
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory(WithGlobalBusinessError())
	f.RegisterProfile("raw", DisableBusinessError())
	resp, err := f.MustNewClient("raw").R().Get(srv.URL)
	if err != nil {
		t.Fatalf("DisableBusinessError should leave response raw: %v", err)
	}
	if string(resp.Bytes()) != `{"code":123,"msg":"x"}` {
		t.Fatalf("body changed: %q", string(resp.Bytes()))
	}
}
```

- [ ] **Step 2: Run tests and confirm the current behavior fails the new default**

Run:

```bash
go test ./httpx -run 'TestBusinessError_' -count=1
```

Expected result before implementation: `TestBusinessError_defaultDoesNotMapEnvelope` fails until the unconditional `interceptors.Error(...)` hook is removed.

- [ ] **Step 3: Add option state and options**

In `clientRegisterOpts`, add:

```go
	businessErrorSet bool
	businessError    bool
```

In `httpx/client_options.go`, add:

```go
func WithBusinessError() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.businessErrorSet = true
		o.businessError = true
	})
}

func DisableBusinessError() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.businessErrorSet = true
		o.businessError = false
	})
}
```

In `httpx/factory_options.go`, add:

```go
func WithGlobalBusinessError() FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalBusinessError = true
	})
}

func DisableGlobalBusinessError() FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalBusinessError = false
	})
}
```

- [ ] **Step 4: Apply business-error policy in `buildReqClient`**

At the end of `buildReqClient`, replace the unconditional error hook with:

```go
	businessError := f.globalBusinessError
	if o.businessErrorSet {
		businessError = o.businessError
	}
	if businessError {
		strictJSONType := f.strictJSONType
		if o.strictJSONTypeSet {
			strictJSONType = o.strictJSONType
		}
		c.OnAfterResponse(interceptors.Error(strictJSONType))
	}
```

Strict JSON options should have an effect only when business-error mapping is enabled.

- [ ] **Step 5: Update existing error-interceptor tests**

Change existing tests that expect business-envelope mapping to use `WithBusinessError()` or `WithGlobalBusinessError()`. For strict content-type tests, initialize the factory with:

```go
f := NewClientFactory(WithGlobalBusinessError(), WithGlobalStrictJSONContentType())
f.RegisterProfile("t")
c := f.MustNewClient("t")
```

- [ ] **Step 6: Verify business-error behavior**

Run:

```bash
go test ./httpx -run 'TestErrorInterceptor|TestBusinessError' -count=1
```

Expected result: all interceptor tests pass.

- [ ] **Step 7: Commit**

```bash
git add httpx/client_options.go httpx/factory_options.go httpx/factory.go httpx/interceptor_test.go
git commit -m "feat(httpx): make business errors explicit"
```

### Task 3: Add Cookie Jar Factory Semantics and Client Clone

**Files:**
- Modify: `httpx/client.go`
- Modify: `httpx/client_options.go`
- Modify: `httpx/factory.go`
- Create: `httpx/cookie.go`
- Modify: `httpx/client_test.go`

- [ ] **Step 1: Write failing tests for independent created clients**

Append these tests to `httpx/client_test.go`:

```go
func TestClientFactory_cookieJarFactoryCreatesIndependentClients(t *testing.T) {
	var calls atomic.Int32
	factory := func() *cookiejar.Jar {
		calls.Add(1)
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err)
		}
		return jar
	}

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithCookieJarFactory(factory))

	first := f.MustNewClient("browser")
	second := f.MustNewClient("browser")

	u, err := url.Parse("https://example.test/")
	if err != nil {
		t.Fatal(err)
	}
	first.Client.GetClient().Jar.SetCookies(u, []*http.Cookie{{Name: "sid", Value: "first"}})
	second.Client.GetClient().Jar.SetCookies(u, []*http.Cookie{{Name: "sid", Value: "second"}})

	firstCookies, err := first.GetCookies("https://example.test/")
	if err != nil {
		t.Fatal(err)
	}
	secondCookies, err := second.GetCookies("https://example.test/")
	if err != nil {
		t.Fatal(err)
	}
	if firstCookies[0].Value != "first" || secondCookies[0].Value != "second" {
		t.Fatalf("created clients should have independent jars: first=%+v second=%+v", firstCookies, secondCookies)
	}
	if calls.Load() != 2 {
		t.Fatalf("factory should run once per NewClient call, got %d", calls.Load())
	}
}

func TestClientClone_preservesNameAndIsUnmanagedVariant(t *testing.T) {
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser")
	c := f.MustNewClient("browser")

	clone := c.Clone()
	if clone == c {
		t.Fatal("Clone returned original wrapper")
	}
	if clone.Name() != "browser" {
		t.Fatalf("clone name: got %q", clone.Name())
	}
	if clone.Req() == nil {
		t.Fatal("clone Req returned nil")
	}
	if names := f.ProfileNames(); len(names) != 1 || names[0] != "browser" {
		t.Fatalf("clone should not register a new profile, got %+v", names)
	}
}
```

Imports required in `httpx/client_test.go`:

```go
import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync/atomic"
)
```

- [ ] **Step 2: Run tests and confirm wrappers are missing**

Run:

```bash
go test ./httpx -run 'TestClientFactory_cookieJarFactory|TestClientClone_' -count=1
```

Expected result before implementation: build fails because `WithCookieJarFactory`, `GetCookies`, and explicit `*httpx.Client.Clone()` are missing.

- [ ] **Step 3: Add cookie option state**

In `httpx/client_options.go`, import `net/http` and `net/http/cookiejar`, then add:

```go
type cookieJarMode int

const (
	cookieJarUnset cookieJarMode = iota
	cookieJarStatic
	cookieJarFactory
)
```

Add fields to `clientRegisterOpts`:

```go
	cookieJarMode    cookieJarMode
	cookieJar        http.CookieJar
	cookieJarFactory CookieJarFactory
```

- [ ] **Step 4: Create cookie wrappers**

Create `httpx/cookie.go`:

```go
package httpx

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/c1emon/gcommon/v2/util"
)

type CookieJarFactory func() *cookiejar.Jar

func WithCookieJar(jar http.CookieJar) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.cookieJarMode = cookieJarStatic
		o.cookieJar = jar
		o.cookieJarFactory = nil
	})
}

func WithCookieJarFactory(factory CookieJarFactory) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.cookieJarMode = cookieJarFactory
		o.cookieJar = nil
		o.cookieJarFactory = factory
	})
}

func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.Client.SetCookieJar(jar)
	return c
}

func (c *Client) SetCookieJarFactory(factory CookieJarFactory) *Client {
	c.Client.SetCookieJarFactory(factory)
	return c
}

func (c *Client) GetCookies(rawURL string) ([]*http.Cookie, error) {
	return c.Client.GetCookies(rawURL)
}
```

- [ ] **Step 5: Apply cookie options during each `NewClient` build**

In `buildReqClient`, after timeout and before hooks, add:

```go
	switch o.cookieJarMode {
	case cookieJarStatic:
		c.SetCookieJar(o.cookieJar)
	case cookieJarFactory:
		c.SetCookieJarFactory(o.cookieJarFactory)
	}
```

Because `NewClient` calls `buildReqClient` every time, `WithCookieJarFactory` creates a new jar per created client.

- [ ] **Step 6: Add explicit `httpx.Client.Clone`**

In `httpx/client.go`, add:

```go
func (c *Client) Clone() *Client {
	if c == nil || c.Client == nil {
		return nil
	}
	return &Client{
		Client: c.Client.Clone(),
		name:   c.name,
	}
}
```

Clone names intentionally match their source profile for logging and traceability. The clone is not registered in any factory.

- [ ] **Step 7: Verify cookie and clone behavior**

Run:

```bash
go test ./httpx -run 'TestClientFactory_cookieJarFactory|TestClientClone_' -count=1
```

Expected result: each `MustNewClient` call receives an independent cookie jar, and clone does not alter factory profiles.

- [ ] **Step 8: Commit**

```bash
git add httpx/client.go httpx/client_options.go httpx/factory.go httpx/cookie.go httpx/client_test.go
git commit -m "feat(httpx): create isolated session clients"
```

### Task 4: Add Redirect Policy Wrappers

**Files:**
- Modify: `httpx/client_options.go`
- Modify: `httpx/factory.go`
- Create: `httpx/redirect.go`
- Create: `httpx/redirect_test.go`

- [ ] **Step 1: Write failing redirect tests**

Create `httpx/redirect_test.go`:

```go
package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectPolicy_profileNoRedirectReadsLocation(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("no-redirect", WithRedirectPolicy(NoRedirectPolicy()))
	resp, err := f.MustNewClient("no-redirect").R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("want 302 without following redirect, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != target.URL+"/next" {
		t.Fatalf("Location: got %q", got)
	}
}

func TestRedirectPolicy_defaultFollowsRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("default")
	resp, err := f.MustNewClient("default").R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want followed redirect 200, got %d", resp.StatusCode)
	}
	if string(resp.Bytes()) != "target" {
		t.Fatalf("body: got %q", string(resp.Bytes()))
	}
}

func TestRedirectPolicy_cloneDoesNotMutateSessionClient(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("target"))
	}))
	t.Cleanup(target.Close)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", target.URL+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	f := NewClientFactory()
	f.RegisterProfile("browser")
	session := f.MustNewClient("browser")
	noRedirect := session.Clone().SetRedirectPolicy(NoRedirectPolicy())

	cloneResp, err := noRedirect.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if cloneResp.StatusCode != http.StatusFound {
		t.Fatalf("clone should not follow redirect, got %d", cloneResp.StatusCode)
	}

	sessionResp, err := session.R().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("session should still follow redirect, got %d", sessionResp.StatusCode)
	}
}
```

- [ ] **Step 2: Run tests and confirm redirect wrappers are missing**

Run:

```bash
go test ./httpx -run 'TestRedirectPolicy_' -count=1
```

Expected result before implementation: build fails because `WithRedirectPolicy` and `NoRedirectPolicy` do not exist in `httpx`.

- [ ] **Step 3: Add redirect option state**

In `clientRegisterOpts`, add:

```go
	redirectPolicySet bool
	redirectPolicies  []RedirectPolicy
```

- [ ] **Step 4: Create redirect wrappers**

Create `httpx/redirect.go`:

```go
package httpx

import (
	"github.com/c1emon/gcommon/v2/util"
	"github.com/imroc/req/v3"
)

type RedirectPolicy = req.RedirectPolicy

func WithRedirectPolicy(policies ...RedirectPolicy) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.redirectPolicySet = true
		o.redirectPolicies = append([]RedirectPolicy(nil), policies...)
	})
}

func NoRedirectPolicy() RedirectPolicy { return req.NoRedirectPolicy() }
func DefaultRedirectPolicy() RedirectPolicy { return req.DefaultRedirectPolicy() }
func MaxRedirectPolicy(noOfRedirect int) RedirectPolicy { return req.MaxRedirectPolicy(noOfRedirect) }
func SameDomainRedirectPolicy() RedirectPolicy { return req.SameDomainRedirectPolicy() }
func SameHostRedirectPolicy() RedirectPolicy { return req.SameHostRedirectPolicy() }
func AllowedHostRedirectPolicy(hosts ...string) RedirectPolicy { return req.AllowedHostRedirectPolicy(hosts...) }
func AllowedDomainRedirectPolicy(hosts ...string) RedirectPolicy { return req.AllowedDomainRedirectPolicy(hosts...) }
func AlwaysCopyHeaderRedirectPolicy(headers ...string) RedirectPolicy {
	return req.AlwaysCopyHeaderRedirectPolicy(headers...)
}

func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	c.Client.SetRedirectPolicy(policies...)
	return c
}
```

- [ ] **Step 5: Apply redirect options during each client build**

In `buildReqClient`, after timeout and before hooks, add:

```go
	if o.redirectPolicySet {
		c.SetRedirectPolicy(o.redirectPolicies...)
	}
```

- [ ] **Step 6: Verify redirect behavior**

Run:

```bash
go test ./httpx -run 'TestRedirectPolicy_' -count=1
```

Expected result: no-redirect profile exposes `Location`, default profile follows redirects, and clone redirect changes do not affect the original session client.

- [ ] **Step 7: Commit**

```bash
git add httpx/client_options.go httpx/factory.go httpx/redirect.go httpx/redirect_test.go
git commit -m "feat(httpx): wrap redirect policies"
```

### Task 5: Update README for Client Factory Sessions

**Files:**
- Modify: `httpx/README.md`

- [ ] **Step 1: Update core concepts and API lists**

Replace the old `Manager` concept with:

```markdown
- `ClientFactory`
  - 维护命名 profile，而不是维护 client 实例
  - 每次 `NewClient(profileName)` 都创建新的 `*Client`
  - 创建出的 client 生命周期由调用者管理
- `Client`
  - 对 `*req.Client` 的轻量包装
  - 通过 `Req()` 发起请求
  - 可用 `Clone()` 创建同一会话内的临时变体
```

Document APIs:

```markdown
- `NewClientFactory(opts ...FactoryOption) *ClientFactory`
- `InitDefaultClientFactory(opts ...FactoryOption)`
- `GetDefaultClientFactory() *ClientFactory`
- `(*ClientFactory).RegisterProfile(name string, opts ...ClientOption)`
- `(*ClientFactory).NewClient(name string, opts ...ClientOption) (*Client, bool)`
- `(*ClientFactory).MustNewClient(name string, opts ...ClientOption) *Client`
- `(*ClientFactory).ProfileNames() []string`
```

- [ ] **Step 2: Document business-error opt-in**

Add:

````markdown
`httpx` 可以按需启用 gcommon 业务 JSON 信封（`code/msg/data`）错误映射。默认不启用，避免把第三方上游 JSON 误判为 gcommon 业务错误。

```go
factory := httpx.NewClientFactory(httpx.WithGlobalBusinessError())

single := httpx.NewClientFactory()
single.RegisterProfile("biz", httpx.WithBusinessError())
```

严格 Content-Type 模式只在业务错误映射启用时生效。
````

- [ ] **Step 3: Document non-idempotent retry guidance**

In the retry section, add:

```markdown
登录、支付、表单提交、票据兑换、信任登录等非幂等请求不建议自动重试。为这些流程注册独立 profile，并显式使用 `DisableRetry()`，让业务层自己判断是否可以重放请求。
```

- [ ] **Step 4: Add browser/session profile example**

Add:

````markdown
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
````

- [ ] **Step 5: Verify README update with package tests**

Run:

```bash
go test ./httpx -count=1
```

Expected result: package tests pass after docs changes.

- [ ] **Step 6: Commit**

```bash
git add httpx/README.md
git commit -m "docs(httpx): document client factory sessions"
```

### Task 6: Downstream `sso-broker` Handoff Plan

**Files in this repository:**
- No changes.

**Expected downstream files in `sso-broker`:**
- Modify: `internal/http.go` or the actual file containing `internal.NewHTTPClientFactory`
- Modify: `internal/upstream/session.go`
- Modify: constructors under `internal/flow/iscsso`
- Modify: constructors under `internal/flow/pms30`
- Test: `internal/upstream`
- Test: `internal/flow/iscsso`
- Test: `internal/flow/pms30`
- Test: `internal/bootstrap`

- [ ] **Step 1: Register a flow profile downstream**

Use the new factory API:

```go
const FlowHTTPProfileName = "flow-upstream"

factory.RegisterProfile(FlowHTTPProfileName,
	httpx.WithTimeout(timeout),
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
```

- [ ] **Step 2: Keep `upstream.Session` as the flow boundary**

Change the session-owned field from `*http.Client` to `*httpx.Client`:

```go
type Session struct {
	client *httpx.Client
}
```

Construct each session with a newly created client:

```go
func NewSession(client *httpx.Client) *Session {
	return &Session{client: client}
}
```

The bootstrap or factory layer should call:

```go
session := upstream.NewSession(factory.MustNewClient(FlowHTTPProfileName))
```

- [ ] **Step 3: Preserve no-redirect method behavior inside the session**

Use a temporary clone for no-redirect calls:

```go
func (s *Session) noRedirectClient() *httpx.Client {
	return s.client.Clone().SetRedirectPolicy(httpx.NoRedirectPolicy())
}
```

`GetNoRedirect` and `PostFormNoRedirect` should call `s.noRedirectClient()` so the normal session client remains default-redirect capable.

- [ ] **Step 4: Avoid flow call-site scatter**

Update flow constructors to accept either `*httpx.Client` or a small session factory:

```go
type SessionFactory func() *upstream.Session
```

Keep `internal/flow/iscsso` and `internal/flow/pms30` request methods calling `Session.Get`, `Session.PostForm`, `Session.PostJSON`, `Session.GetNoRedirect`, and `Session.PostFormNoRedirect`.

- [ ] **Step 5: Validate downstream with local replace only during testing**

Use a temporary local dependency replacement during validation:

```bash
go mod edit -replace github.com/c1emon/gcommon/httpx/v2=/Users/clemon/Workplace/gcommon/httpx
go test ./internal/upstream ./internal/flow/iscsso ./internal/flow/pms30 ./internal/bootstrap
go test ./...
go mod edit -dropreplace github.com/c1emon/gcommon/httpx/v2
```

Expected result: downstream tests pass and no local replace remains in the final diff unless the release strategy explicitly requires it.

### Task 7: Final Verification

**Files:**
- All modified files from Tasks 1 through 5.

- [ ] **Step 1: Run package-level tests**

Run:

```bash
go test ./httpx -count=1
```

Expected result: all `httpx` package tests pass.

- [ ] **Step 2: Run workspace tests**

Run:

```bash
go test ./...
```

Expected result: all repository tests pass.

- [ ] **Step 3: Run GitNexus change detection before final commit**

Run:

```bash
npx gitnexus analyze
```

Then use GitNexus `detect_changes` with `scope: "all"`.

Expected result: changed symbols are limited to `httpx` factory/client/options/README areas, and affected flows match the planned factory construction and error-interceptor construction paths.

- [ ] **Step 4: Commit final verification fixes if needed**

If test or GitNexus verification required adjustments, commit only those scoped files:

```bash
git add httpx
git commit -m "test(httpx): verify client factory sessions"
```

## Self-Review

- Requirement coverage: every requirement in `httpx.md` maps to a task or downstream handoff section.
- Factory semantics: the factory stores profiles only; created client instances are caller-owned and unmanaged by the factory.
- Clone semantics: clone is retained only for temporary variants within an existing session, not as the primary way to create independent sessions.
- Generic wrapper priority: APIs are cookie, redirect, retry, clone, and business-error policy wrappers, with no SSO-specific request headers or flow methods in `httpx`.
- Breaking-change stance: documented explicitly; `Manager` naming is replaced with `ClientFactory`, and default business-error behavior changes to safer third-party handling.
- Risk controls: per-`NewClient` cookie jar isolation and clone redirect isolation have dedicated tests.
- GitNexus controls: impact baseline is recorded, and final implementation requires `detect_changes` before any commit.
