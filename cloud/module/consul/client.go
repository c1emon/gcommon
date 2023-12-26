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

func (c *Client) RegisterSvc(svc *api.AgentServiceRegistration) error {
	return c.cli.Agent().ServiceRegister(svc)
}

func (c *Client) DeregisterSvc(id string) error {
	return c.cli.Agent().ServiceDeregister(id)
}

func (c *Client) UpdateTTL(checkID, output, status string) error {
	return c.cli.Agent().UpdateTTL(checkID, output, status)
}

func (c *Client) Catalog() *api.Catalog {
	return c.cli.Catalog()
}

func (c *Client) Services() (map[string]*api.AgentService, error) {
	return c.cli.Agent().Services()
}
