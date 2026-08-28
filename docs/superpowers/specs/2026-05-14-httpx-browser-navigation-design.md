# httpx Browser Navigation State Design

## Goal

Enhance `httpx` browser navigation behavior so browser-style headers and browser navigation state model real page loads separately from XHR/fetch/subresource requests.

## Current Context

`httpx.WithBrowserNavigation()` installs request hooks in `httpx/browser_navigation.go`. It currently derives browser headers from the previous URL and request kind, but every successful request updates the client navigation state. This makes XHR/API requests become the next `Referer` base, which differs from browser behavior.

## Requirements

1. Navigation requests update the current page URL after the request succeeds.
2. XHR/fetch/subresource requests use the current page URL for `Referer`, but do not update the current page URL.
3. `AsXHR()` must produce XHR headers and suppress navigation state updates.
4. `AsNavigation()` must produce document/navigation headers and allow navigation state updates.
5. `BrowserRequestAuto` must infer the effective kind from existing rules, then use that effective kind for both headers and state-update policy.
6. Callers can explicitly anchor the current page context at client level with `Client.SetBrowserCurrentURL(...)`.
7. Callers can explicitly anchor a single request context with `Request.WithBrowserPageURL(...)`; this affects the request `Referer` base without changing client state by itself.
8. Callers can suppress client navigation state updates for one request with `Request.WithBrowserStateUpdate(false)` or `Request.WithoutBrowserNavigationStateUpdate()`.
9. Existing browser header behavior remains intact: navigation `Accept`, XHR `Accept`, POST `Origin`, `Sec-Fetch-Site`, no XHR `Sec-Fetch-User`/`Upgrade-Insecure-Requests`, navigation `Sec-Fetch-Mode: navigate`, and navigation `Sec-Fetch-Dest: document`.

## API Design

- `func (c *Client) SetBrowserCurrentURL(rawURL string) *Client`
  - Parses `rawURL` and stores it as the client's current browser page URL.
  - Invalid or empty URLs clear the stored current URL, matching the package's existing nil-safe style.
- `func (r *Request) WithBrowserPageURL(rawURL string) *Request`
  - Parses `rawURL` and uses it as the `Referer`/`Sec-Fetch-Site` base for that request.
  - Does not mutate the client current URL by itself.
- `func (r *Request) WithBrowserStateUpdate(enabled bool) *Request`
  - Overrides whether this request may write the client navigation state.
- `func (r *Request) WithoutBrowserNavigationStateUpdate() *Request`
  - Convenience wrapper for `WithBrowserStateUpdate(false)`.

## Architecture

Keep browser request semantics in `httpx/browser_navigation.go`. The installed hook owns a shared `browserNavigationState` per req client, and the `httpx.Client` wrapper keeps an unexported pointer to that state so `SetBrowserCurrentURL` can seed it without any global registry. Cloned clients keep the same state pointer to preserve the existing shared navigation-state behavior. Request APIs communicate page URL and update overrides through context values/internal headers so public API remains small.

Request processing resolves three independent pieces of state:

1. Previous page URL: request-level page URL override if set, otherwise client current URL.
2. Effective request kind: explicit kind if set, otherwise auto detection from headers/content type.
3. State update permission: explicit request override if set, otherwise `true` only for effective navigation requests.

After the round trip, the client state is updated only when the effective request kind permits it and request-level state update has not disabled it. Redirects continue to remember the final URL for navigation requests. Retry handling restores the prior client URL only when that request had changed it.

## Testing

Add tests in `httpx/browser_navigation_test.go` for:

- navigation GET updates the page URL;
- XHR POST uses current page as `Referer`;
- XHR POST does not update current page URL;
- multiple XHRs do not pollute the final navigation POST `Referer`;
- explicit client and request page contexts drive `Referer`;
- `AsXHR()`, `AsNavigation()`, and `Auto` keep headers and state update policy consistent;
- per-request state update suppression prevents navigation requests from changing client state.

## Impact and Risk

GitNexus impact analysis found `installBrowserNavigation` LOW risk and `browserNavigationState.apply`/`remember` HIGH risk because they sit on the browser navigation path used by client factory creation flows. The implementation must therefore keep changes focused, preserve existing header semantics, and rely on targeted regression tests.
