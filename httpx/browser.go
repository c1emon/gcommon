package httpx

import (
	"github.com/imroc/req/v3"
)

// BrowserProfile selects TLS/HTTP fingerprint impersonation via imroc/req.
type BrowserProfile int

const (
	BrowserNone BrowserProfile = iota
	BrowserChrome
	BrowserFirefox
	BrowserSafari
)

func applyBrowserProfile(c *req.Client, p BrowserProfile) {
	switch p {
	case BrowserChrome:
		c.ImpersonateChrome()
	case BrowserFirefox:
		c.ImpersonateFirefox()
	case BrowserSafari:
		c.ImpersonateSafari()
	default:
		// BrowserNone: no-op
	}
}
