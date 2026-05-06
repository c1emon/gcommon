package httpx

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"github.com/c1emon/gcommon/v2/util"
)

// CookieJarFactory creates a cookie jar for a newly built client.
type CookieJarFactory func() *cookiejar.Jar

// WithCookieJar sets a concrete cookie jar for created clients.
func WithCookieJar(jar http.CookieJar) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.cookieJar = jar
		o.cookieJarSet = true
		o.cookieJarFactory = nil
		o.cookieJarFactorySet = false
	})
}

// WithCookieJarFactory sets a factory that is called for every created client.
func WithCookieJarFactory(factory CookieJarFactory) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.cookieJar = nil
		o.cookieJarSet = false
		o.cookieJarFactory = factory
		o.cookieJarFactorySet = true
	})
}

// SetCookieJar sets a concrete cookie jar on this client.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	if c == nil {
		return nil
	}
	if c.Client != nil {
		c.Client.SetCookieJar(jar)
	}
	return c
}

// SetCookieJarFactory sets a cookie jar factory on this client.
func (c *Client) SetCookieJarFactory(factory CookieJarFactory) *Client {
	if c == nil {
		return nil
	}
	if c.Client != nil {
		c.Client.SetCookieJarFactory(factory)
	}
	return c
}

// GetCookies returns cookies from this client's cookie jar for rawURL.
func (c *Client) GetCookies(rawURL string) ([]*http.Cookie, error) {
	if c == nil || c.Client == nil {
		return nil, errors.New("httpx: nil client")
	}
	return c.Client.GetCookies(rawURL)
}
