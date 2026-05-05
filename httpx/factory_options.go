package httpx

import (
	"log/slog"

	"github.com/c1emon/gcommon/util"
	"golang.org/x/time/rate"
)

// FactoryOption configures defaults inherited by every created client.
type FactoryOption util.Option[ClientFactory]

func WithGlobalLimiter(l *rate.Limiter) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalLimiter = l
	})
}

func WithGlobalLogger(l *slog.Logger) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.logger = l
	})
}

func WithGlobalReqInterceptor(i ReqInterceptor) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalReqInterceptors = append(f.globalReqInterceptors, i)
	})
}

func WithGlobalRespInterceptor(i RespInterceptor) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalRespInterceptors = append(f.globalRespInterceptors, i)
	})
}

func WithGlobalRetry(p RetryPolicy) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.globalRetry = p
	})
}

func WithGlobalBrowser(p BrowserProfile) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		if p == BrowserNone {
			f.hasGlobalBrowserProfile = false
			f.globalBrowserProfile = BrowserNone
			return
		}
		f.hasGlobalBrowserProfile = true
		f.globalBrowserProfile = p
	})
}

func WithGlobalHeader(key, val string) FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		if f.globalHeaders == nil {
			f.globalHeaders = make(map[string]string)
		}
		f.globalHeaders[key] = val
	})
}

// WithGlobalStrictJSONContentType enables strict mode globally: ErrorInterceptor
// only parses when Content-Type indicates JSON.
func WithGlobalStrictJSONContentType() FactoryOption {
	return util.WrapFuncOption(func(f *ClientFactory) {
		f.strictJSONType = true
	})
}
