package httpx

import (
	"github.com/c1emon/gcommon/util"
	"github.com/imroc/req/v3"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"

type (
	ReqInterceptor func(client *Client, req *Request) error

	RespInterceptor func(client *Client, resp *Response) error
)

type Client struct {
	*req.Client
}

type Request struct {
	*req.Request
}

type Response struct {
	*req.Response
}

func (c *Client) Req() *Request {
	return &Request{c.R()}
}

func NewClient(opts ...*util.FuncOption[Client]) *Client {
	client := &Client{
		req.NewClient(),
	}

	for _, opt := range opts {
		opt.Apply(client)
	}
	return client
}
