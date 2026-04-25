package httpx

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/c1emon/gcommon/httpx/interceptors"
	"github.com/imroc/req/v3"
	"golang.org/x/time/rate"
)

// Manager owns multiple named [Client] instances with shared defaults (limiter, interceptors, retry, browser).
type Manager struct {
	mu sync.RWMutex

	clients map[string]*Client

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

// NewManager constructs a manager and applies [ManagerOption] values in order.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		clients: make(map[string]*Client),
	}
	for _, opt := range opts {
		opt.Apply(m)
	}
	return m
}

var (
	defaultManagerMu sync.RWMutex
	defaultManager   *Manager
)

// InitDefaultManager replaces the package default [Manager], built with [NewManager](opts...).
func InitDefaultManager(opts ...ManagerOption) {
	m := NewManager(opts...)
	defaultManagerMu.Lock()
	defaultManager = m
	defaultManagerMu.Unlock()
}

// GetDefaultManager returns the [Manager] set by [InitDefaultManager], or nil if [InitDefaultManager] has never been called.
func GetDefaultManager() *Manager {
	defaultManagerMu.RLock()
	defer defaultManagerMu.RUnlock()
	return defaultManager
}

// Register creates or replaces a named client. Options are applied after manager defaults.
func (m *Manager) Register(name string, opts ...ClientOption) *Client {
	o := newClientRegisterOpts()
	for _, opt := range opts {
		opt.Apply(&o)
	}
	rc := m.buildReqClient(name, &o)
	hc := &Client{Client: rc, name: name}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clients == nil {
		m.clients = make(map[string]*Client)
	}
	m.clients[name] = hc
	return hc
}

// Client returns a previously registered client.
func (m *Manager) Client(name string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[name]
	return c, ok
}

// MustClient returns a registered client or panics.
func (m *Manager) MustClient(name string) *Client {
	c, ok := m.Client(name)
	if !ok {
		panic(fmt.Sprintf("httpx: unknown client %q", name))
	}
	return c
}

// Names returns registered client names (unordered).
func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.clients))
	for n := range m.clients {
		out = append(out, n)
	}
	return out
}

func (m *Manager) buildReqClient(name string, o *clientRegisterOpts) *req.Client {
	c := req.C()

	if lg := o.effectiveLogger(m); lg != nil {
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
	for k, v := range m.globalHeaders {
		c.SetCommonHeader(k, v)
	}
	for k, v := range o.headers {
		c.SetCommonHeader(k, v)
	}

	hasBrowserProfile := m.hasGlobalBrowserProfile
	browserProfile := m.globalBrowserProfile
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
			if lg := o.effectiveLogger(m); lg != nil {
				lg.Debug("httpx: WithUserAgent ignored when browser profile is set", slog.String("client", name))
			}
		}
	} else if o.ua != "" {
		c.SetUserAgent(o.ua)
	}

	effRetry := m.globalRetry
	if o.retryDisabled {
		effRetry = RetryPolicy{}
	} else if o.retryExplicit {
		effRetry = o.retry
	}
	applyRetryPolicy(c, effRetry)

	m.applyLimiterMiddleware(c, o)

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

	for _, ri := range m.globalReqInterceptors {
		wrapReq(ri)
	}
	for _, ri := range o.clientReqInterceptors {
		wrapReq(ri)
	}

	if lg := o.effectiveLogger(m); lg != nil {
		c.OnBeforeRequest(interceptors.RequestLogger(lg))
	}

	for _, ii := range m.globalRespInterceptors {
		wrapResp(ii)
	}
	for _, ii := range o.clientRespInterceptors {
		wrapResp(ii)
	}
	if lg := o.effectiveLogger(m); lg != nil {
		c.OnAfterResponse(interceptors.ResponseLogger(lg))
	}
	strictJSONType := m.strictJSONType
	if o.strictJSONTypeSet {
		strictJSONType = o.strictJSONType
	}
	c.OnAfterResponse(interceptors.Error(strictJSONType))

	return c
}
