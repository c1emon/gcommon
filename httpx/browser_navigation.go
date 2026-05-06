package httpx

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/c1emon/gcommon/v2/util"
	"github.com/imroc/req/v3"
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

const (
	browserRequestKindHeader = "X-Httpx-Browser-Request-Kind"
	htmlAccept               = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
	jsonAccept               = "application/json, text/plain, */*"
)

type explicitHeadersKey struct{}

type browserRequestKindKey struct{}

type browserPreviousURLKey struct{}

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

func installBrowserNavigation(c *req.Client, cfg *browserNavigationConfig) {
	if c == nil || cfg == nil {
		return
	}

	state := &browserNavigationState{}
	c.OnBeforeRequest(func(_ *req.Client, r *req.Request) error {
		ctx := r.Context()
		if explicitHeadersFromContext(ctx) != nil {
			return nil
		}
		ctx = contextWithExplicitHeaders(ctx, explicitHeaderKeys(r.Headers))
		ctx = contextWithBrowserRequestKind(ctx, browserRequestKindFromHeader(r.Headers))
		r.SetContext(ctx)
		return nil
	})
	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(r *req.Request) (*req.Response, error) {
			r.SetContext(contextWithPreviousBrowserURL(r.Context(), state.previous()))
			state.apply(cfg, r, explicitHeadersFromContext(r.Context()), browserRequestKindFromContext(r.Context()))
			resp, err := rt.RoundTrip(r)
			if resp != nil && resp.Response != nil && resp.Response.Request != nil {
				state.remember(resp.Response.Request.URL)
			}
			return resp, err
		}
	})
	c.AddCommonRetryHook(func(resp *req.Response, _ error) {
		if resp == nil || resp.Request == nil {
			return
		}
		prev, ok := previousBrowserURLFromContext(resp.Request.Context())
		if !ok {
			return
		}
		state.remember(prev)
	})
}

func contextWithExplicitHeaders(ctx context.Context, explicit map[string]struct{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, explicitHeadersKey{}, explicit)
}

func contextWithBrowserRequestKind(ctx context.Context, kind BrowserRequestKind) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, browserRequestKindKey{}, kind)
}

func contextWithPreviousBrowserURL(ctx context.Context, u *url.URL) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, browserPreviousURLKey{}, cloneURL(u))
}

func explicitHeadersFromContext(ctx context.Context) map[string]struct{} {
	if ctx == nil {
		return nil
	}
	explicit, _ := ctx.Value(explicitHeadersKey{}).(map[string]struct{})
	return explicit
}

func browserRequestKindFromContext(ctx context.Context) BrowserRequestKind {
	if ctx == nil {
		return BrowserRequestAuto
	}
	kind, _ := ctx.Value(browserRequestKindKey{}).(BrowserRequestKind)
	return kind
}

func previousBrowserURLFromContext(ctx context.Context) (*url.URL, bool) {
	if ctx == nil {
		return nil, false
	}
	u, ok := ctx.Value(browserPreviousURLKey{}).(*url.URL)
	return cloneURL(u), ok
}

func (s *browserNavigationState) remember(u *url.URL) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastURL = cloneURL(u)
}

func (s *browserNavigationState) previous() *url.URL {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneURL(s.lastURL)
}

func (s *browserNavigationState) apply(cfg *browserNavigationConfig, r *req.Request, explicit map[string]struct{}, kind BrowserRequestKind) {
	if s == nil || cfg == nil || r == nil || r.URL == nil {
		return
	}
	defer deleteInternalBrowserRequestKindHeader(r.Headers)
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}

	current := r.URL
	prev := s.previous()
	setOrDeleteDynamicHeader(r.Headers, explicit, "Referer", cfg.refererFor(prev, current))
	setDynamicHeader(r.Headers, explicit, "Sec-Fetch-Site", secFetchSite(prev, current))
	if strings.EqualFold(r.Method, http.MethodPost) {
		setDynamicHeader(r.Headers, explicit, "Origin", originString(current))
	} else {
		deleteDynamicHeader(r.Headers, explicit, "Origin")
	}

	if kind == BrowserRequestAuto {
		kind = cfg.effectiveRequestKind(r.Headers)
	}
	switch kind {
	case BrowserRequestXHR:
		setDynamicHeader(r.Headers, explicit, "Accept", jsonAccept)
		setDynamicHeader(r.Headers, explicit, "Sec-Fetch-Mode", "cors")
		setDynamicHeader(r.Headers, explicit, "Sec-Fetch-Dest", "empty")
		if cfg.defaultXRequestedWith {
			setDynamicHeader(r.Headers, explicit, "X-Requested-With", "XMLHttpRequest")
		}
		deleteDynamicHeader(r.Headers, explicit, "Sec-Fetch-User")
		deleteDynamicHeader(r.Headers, explicit, "Upgrade-Insecure-Requests")
	case BrowserRequestNavigation:
		setDynamicHeader(r.Headers, explicit, "Accept", htmlAccept)
		setDynamicHeader(r.Headers, explicit, "Sec-Fetch-Mode", "navigate")
		setDynamicHeader(r.Headers, explicit, "Sec-Fetch-Dest", "document")
		if cfg.defaultSecFetchUser {
			setDynamicHeader(r.Headers, explicit, "Sec-Fetch-User", "?1")
		}
		setDynamicHeader(r.Headers, explicit, "Upgrade-Insecure-Requests", "1")
		deleteDynamicHeader(r.Headers, explicit, "X-Requested-With")
	}
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
		c.referrerPolicy = policy
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
		delete(r.Request.Headers, browserRequestKindHeader)
	}
	return r
}

func (r *Request) AsXHR() *Request {
	return r.WithBrowserRequestKind(BrowserRequestXHR)
}

func (r *Request) AsNavigation() *Request {
	return r.WithBrowserRequestKind(BrowserRequestNavigation)
}

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
	origin := originString(u)
	if origin == "" {
		return ""
	}
	return origin + "/"
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	out := *u
	return &out
}

func explicitHeaderKeys(h http.Header) map[string]struct{} {
	explicit := make(map[string]struct{}, len(h))
	for key := range h {
		canonicalKey := http.CanonicalHeaderKey(key)
		if canonicalKey == browserRequestKindHeader {
			continue
		}
		explicit[canonicalKey] = struct{}{}
	}
	return explicit
}

func hasExplicitHeader(explicit map[string]struct{}, key string) bool {
	if explicit == nil {
		return false
	}
	_, ok := explicit[http.CanonicalHeaderKey(key)]
	return ok
}

func setDynamicHeader(h http.Header, explicit map[string]struct{}, key, value string) {
	if value == "" || hasExplicitHeader(explicit, key) {
		return
	}
	h.Set(key, value)
}

func setOrDeleteDynamicHeader(h http.Header, explicit map[string]struct{}, key, value string) {
	if value == "" {
		deleteDynamicHeader(h, explicit, key)
		return
	}
	setDynamicHeader(h, explicit, key, value)
}

func deleteDynamicHeader(h http.Header, explicit map[string]struct{}, key string) {
	if h == nil || hasExplicitHeader(explicit, key) {
		return
	}
	h.Del(key)
}

func deleteInternalBrowserRequestKindHeader(h http.Header) {
	for key := range h {
		if strings.EqualFold(key, browserRequestKindHeader) {
			delete(h, key)
		}
	}
}

func browserRequestKindFromHeader(h http.Header) BrowserRequestKind {
	if kind := parseBrowserRequestKindValues(h[browserRequestKindHeader]); kind != BrowserRequestAuto {
		return kind
	}
	for key, values := range h {
		if key == browserRequestKindHeader || !strings.EqualFold(key, browserRequestKindHeader) {
			continue
		}
		if kind := parseBrowserRequestKindValues(values); kind != BrowserRequestAuto {
			return kind
		}
	}
	return BrowserRequestAuto
}

func parseBrowserRequestKindValues(values []string) BrowserRequestKind {
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "navigation":
			return BrowserRequestNavigation
		case "xhr":
			return BrowserRequestXHR
		}
	}
	return BrowserRequestAuto
}

func isJSONContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	return ct == "application/json" || strings.HasPrefix(ct, "application/json;")
}

func (c *browserNavigationConfig) effectiveRequestKind(h http.Header) BrowserRequestKind {
	kind := browserRequestKindFromHeader(h)
	if kind != BrowserRequestAuto {
		return kind
	}
	if c != nil && c.xhrForJSON && isJSONContentType(h.Get("Content-Type")) {
		return BrowserRequestXHR
	}
	return BrowserRequestNavigation
}

func (c *browserNavigationConfig) refererFor(prev, current *url.URL) string {
	if prev == nil || c == nil || c.referrerPolicy == ReferrerPolicyNoReferrer {
		return ""
	}
	if c.referrerPolicy == ReferrerPolicyOrigin {
		return originReferer(prev)
	}
	if sameOrigin(prev, current) {
		return prev.String()
	}
	return originReferer(prev)
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
