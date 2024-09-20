package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	transport *http.Transport
	dialer    *net.Dialer

	initOnce sync.Once

	clientTimeout time.Duration
	client        *http.Client
}

// NewBuildableClient returns an initialized client for invoking HTTP
// requests.
func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

// Do implements the HTTPClient interface's Do method to invoke a HTTP request,
// and receive the response. Uses the BuildableClient's current
// configuration to invoke the http.Request.
//
// If connection pooling is enabled (aka HTTP KeepAlive) the client will only
// share pooled connections with its own instance. Copies of the
// BuildableClient will have their own connection pools.
//
// Redirect (3xx) responses will not be followed, the HTTP response received
// will returned instead.
func (b *Client) Do(req *http.Request) (*http.Response, error) {
	// b.initOnce.Do(b.build)

	return b.client.Do(req)
}
