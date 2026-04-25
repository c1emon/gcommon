package httpx

import (
	"log/slog"

	"github.com/c1emon/gcommon/util"
	"golang.org/x/time/rate"
)

// ManagerOption configures defaults inherited by every registered client.
type ManagerOption util.Option[Manager]

func WithGlobalLimiter(l *rate.Limiter) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.globalLimiter = l
	})
}

func WithGlobalLogger(l *slog.Logger) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.logger = l
	})
}

func WithGlobalReqInterceptor(i ReqInterceptor) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.globalReqInterceptors = append(m.globalReqInterceptors, i)
	})
}

func WithGlobalRespInterceptor(i RespInterceptor) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.globalRespInterceptors = append(m.globalRespInterceptors, i)
	})
}

func WithGlobalRetry(p RetryPolicy) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.globalRetry = p
	})
}

func WithGlobalBrowser(p BrowserProfile) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		if p == BrowserNone {
			m.hasGlobalBrowserProfile = false
			m.globalBrowserProfile = BrowserNone
			return
		}
		m.hasGlobalBrowserProfile = true
		m.globalBrowserProfile = p
	})
}

func WithGlobalHeader(key, val string) ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		if m.globalHeaders == nil {
			m.globalHeaders = make(map[string]string)
		}
		m.globalHeaders[key] = val
	})
}

// WithGlobalStrictJSONContentType enables strict mode globally: ErrorInterceptor
// only parses when Content-Type indicates JSON.
func WithGlobalStrictJSONContentType() ManagerOption {
	return util.WrapFuncOption(func(m *Manager) {
		m.strictJSONType = true
	})
}
