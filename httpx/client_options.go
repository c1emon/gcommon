package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/c1emon/gcommon/util"
	"golang.org/x/time/rate"
)

// ClientOption configures a single named client at registration time.
type ClientOption util.Option[clientRegisterOpts]

type clientRegisterOpts struct {
	baseURL string
	timeout time.Duration
	headers map[string]string

	ua string

	hasClientBrowserProfile bool
	clientBrowserProfile    BrowserProfile

	clientLimiter   *rate.Limiter
	noClientLimiter bool

	clientReqInterceptors  []ReqInterceptor
	clientRespInterceptors []RespInterceptor

	retry             RetryPolicy
	retryExplicit     bool
	retryDisabled     bool
	strictJSONTypeSet bool
	strictJSONType    bool
	businessErrorSet  bool
	businessError     bool

	cookieJar           http.CookieJar
	cookieJarSet        bool
	cookieJarFactory    CookieJarFactory
	cookieJarFactorySet bool

	logDisabled  bool
	clientLogger *slog.Logger
	clientLogSet bool
}

func (o *clientRegisterOpts) effectiveLogger(f *ClientFactory) *slog.Logger {
	if !o.clientLogSet {
		return f.logger
	}
	if o.logDisabled {
		return nil
	}
	return o.clientLogger
}

func (o *clientRegisterOpts) addHeader(k, v string) {
	if o.headers == nil {
		o.headers = make(map[string]string)
	}
	o.headers[k] = v
}

func newClientRegisterOpts() clientRegisterOpts {
	return clientRegisterOpts{}
}

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
	return out
}

func newClientRegisterOptsFrom(opts ...ClientOption) clientRegisterOpts {
	o := newClientRegisterOpts()
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return o
}

func WithBaseURL(url string) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.baseURL = url
	})
}

func WithTimeout(t time.Duration) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.timeout = t
	})
}

func WithHeader(key, val string) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.addHeader(key, val)
	})
}

func WithHeaders(h map[string]string) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		for k, v := range h {
			o.addHeader(k, v)
		}
	})
}

func WithUserAgent(ua string) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.ua = ua
	})
}

func WithBrowser(p BrowserProfile) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.hasClientBrowserProfile = true
		o.clientBrowserProfile = p
	})
}

func WithLimiter(l *rate.Limiter) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.noClientLimiter = false
		o.clientLimiter = l
	})
}

func DisableLimiter() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.noClientLimiter = true
		o.clientLimiter = nil
	})
}

func WithReqInterceptor(i ReqInterceptor) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.clientReqInterceptors = append(o.clientReqInterceptors, i)
	})
}

func WithRespInterceptor(i RespInterceptor) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.clientRespInterceptors = append(o.clientRespInterceptors, i)
	})
}

func WithRetry(p RetryPolicy) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.retryExplicit = true
		o.retry = p
	})
}

func DisableRetry() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.retryDisabled = true
	})
}

// WithStrictJSONContentType enables strict mode for this client: ErrorInterceptor
// only parses when Content-Type indicates JSON.
func WithStrictJSONContentType() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.strictJSONTypeSet = true
		o.strictJSONType = true
	})
}

// WithoutStrictJSONContentType disables strict JSON Content-Type mode for this client.
func WithoutStrictJSONContentType() ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.strictJSONTypeSet = true
		o.strictJSONType = false
	})
}

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

// WithLogger overrides the manager logger for this client. Pass nil to disable logging for this client.
func WithLogger(l *slog.Logger) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.clientLogSet = true
		if l == nil {
			o.logDisabled = true
			o.clientLogger = nil
			return
		}
		o.logDisabled = false
		o.clientLogger = l
	})
}
