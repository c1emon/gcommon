package consul

import (
	"context"

	"github.com/c1emon/gcommon/logx"
	"github.com/hashicorp/consul/api"
)

// Client is consul client config
type Client struct {
	cli *api.Client

	ctx    context.Context
	cancel context.CancelFunc

	logger logx.Logger
}

func New(addr string) (*Client, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		cli: c,
	}, nil
}

func (c *Client) NewAgent() *api.Agent {
	return c.cli.Agent()
}

func (c *Client) NewHealth() {
	// c.cli.Agent()
}
