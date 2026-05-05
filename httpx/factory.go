package httpx

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"github.com/c1emon/gcommon/httpx/interceptors"
	"github.com/imroc/req/v3"
	"golang.org/x/time/rate"
)

// ClientFactory stores named client profiles with shared defaults.
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
}

// NewClientFactory constructs a client factory and applies [FactoryOption] values in order.
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

// InitDefaultClientFactory replaces the package default [ClientFactory].
func InitDefaultClientFactory(opts ...FactoryOption) {
	f := NewClientFactory(opts...)
	defaultClientFactoryMu.Lock()
	defaultClientFactory = f
	defaultClientFactoryMu.Unlock()
}

// GetDefaultClientFactory returns the [ClientFactory] set by [InitDefaultClientFactory], or nil if it has never been called.
func GetDefaultClientFactory() *ClientFactory {
	defaultClientFactoryMu.RLock()
	defer defaultClientFactoryMu.RUnlock()
	return defaultClientFactory
}

// RegisterProfile creates or replaces a named client profile.
func (f *ClientFactory) RegisterProfile(name string, opts ...ClientOption) {
	o := newClientRegisterOptsFrom(opts...)

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.profiles == nil {
		f.profiles = make(map[string]clientRegisterOpts)
	}
	f.profiles[name] = o
}

// NewClient creates a fresh client for a registered profile.
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

// MustNewClient creates a fresh client for a registered profile or panics.
func (f *ClientFactory) MustNewClient(name string, opts ...ClientOption) *Client {
	c, ok := f.NewClient(name, opts...)
	if !ok {
		panic(fmt.Sprintf("httpx: unknown client profile %q", name))
	}
	return c
}

// ProfileNames returns registered profile names in deterministic order.
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

func (f *ClientFactory) buildReqClient(name string, o *clientRegisterOpts) *req.Client {
	c := req.C()

	if lg := o.effectiveLogger(f); lg != nil {
		c.SetLogger(reqSlogLogger{log: lg})
	} else {
		c.SetLogger(nil)
	}

	if o.baseURL != "" {
		c.SetBaseURL(o.baseURL)
	}
	if o.timeout > 0 {
		c.SetTimeout(o.timeout)
	}
	for k, v := range f.globalHeaders {
		c.SetCommonHeader(k, v)
	}
	for k, v := range o.headers {
		c.SetCommonHeader(k, v)
	}

	hasBrowserProfile := f.hasGlobalBrowserProfile
	browserProfile := f.globalBrowserProfile
	if o.hasClientBrowserProfile {
		if o.clientBrowserProfile == BrowserNone {
			hasBrowserProfile = false
			browserProfile = BrowserNone
		} else {
			hasBrowserProfile = true
			browserProfile = o.clientBrowserProfile
		}
	}
	if hasBrowserProfile {
		applyBrowserProfile(c, browserProfile)
		if o.ua != "" {
			if lg := o.effectiveLogger(f); lg != nil {
				lg.Debug("httpx: WithUserAgent ignored when browser profile is set", slog.String("client", name))
			}
		}
	} else if o.ua != "" {
		c.SetUserAgent(o.ua)
	}

	effRetry := f.globalRetry
	if o.retryDisabled {
		effRetry = RetryPolicy{}
	} else if o.retryExplicit {
		effRetry = o.retry
	}
	applyRetryPolicy(c, effRetry)

	f.applyLimiterMiddleware(c, o)

	wrapReq := func(ri ReqInterceptor) {
		h := ri
		c.OnBeforeRequest(func(rc *req.Client, rq *req.Request) error {
			return h(&Client{Client: rc, name: name}, &Request{rq})
		})
	}
	wrapResp := func(ii RespInterceptor) {
		h := ii
		c.OnAfterResponse(func(rc *req.Client, rr *req.Response) error {
			return h(&Client{Client: rc, name: name}, &Response{rr})
		})
	}

	for _, ri := range f.globalReqInterceptors {
		wrapReq(ri)
	}
	for _, ri := range o.clientReqInterceptors {
		wrapReq(ri)
	}

	if lg := o.effectiveLogger(f); lg != nil {
		c.OnBeforeRequest(interceptors.RequestLogger(lg))
	}

	for _, ii := range f.globalRespInterceptors {
		wrapResp(ii)
	}
	for _, ii := range o.clientRespInterceptors {
		wrapResp(ii)
	}
	if lg := o.effectiveLogger(f); lg != nil {
		c.OnAfterResponse(interceptors.ResponseLogger(lg))
	}
	strictJSONType := f.strictJSONType
	if o.strictJSONTypeSet {
		strictJSONType = o.strictJSONType
	}
	c.OnAfterResponse(interceptors.Error(strictJSONType))

	return c
}
