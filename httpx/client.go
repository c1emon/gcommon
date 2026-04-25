package httpx

import (
	"github.com/imroc/req/v3"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"

type (
	ReqInterceptor func(client *Client, req *Request) error

	RespInterceptor func(client *Client, resp *Response) error
)

// Client wraps a single imroc/req [req.Client], usually created by [Manager.Register].
type Client struct {
	*req.Client
	name string
}

type Request struct {
	*req.Request
}

type Response struct {
	*req.Response
}

// Name is the key passed to [Manager.Register].
func (c *Client) Name() string {
	return c.name
}

func (c *Client) Req() *Request {
	return &Request{c.R()}
}
