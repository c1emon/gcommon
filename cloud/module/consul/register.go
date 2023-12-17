package consul

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c1emon/gcommon/cloud/registry"
	"github.com/hashicorp/consul/api"
)

func NewRegisterClient() (*RegisterClient, error) {
	return &RegisterClient{}, nil
}

// RegisterClient 定义一个RegisterClient结构体，其内部有一个`*api.Client`字段。
type RegisterClient struct {
	client   *Client
	svcInfos []*registry.RemoteSvcRegInfo

	// resolve service entry endpoints
	// resolver ServiceResolver

	// serviceChecks user custom checks
	serviceChecks api.AgentServiceChecks
}

// Register register service instance to consul
func (c *RegisterClient) Register(ctx context.Context, infos []*registry.RemoteSvcRegInfo) error {
	c.svcInfos = infos
	for _, info := range infos {

		addresses := make(map[string]api.ServiceAddress, len(info.Endpoints))

		for _, endpoint := range info.Endpoints {
			addresses[endpoint.Schema.String()] = api.ServiceAddress{Address: endpoint.Host, Port: endpoint.Port}
		}

		// 不确定仅仅用`TaggedAddresses`可不可以
		// save registration to re-register
		// when ttl error(likely consul server restart?)
		registration := &api.AgentServiceRegistration{
			ID:              info.ID,
			Name:            info.Name,
			TaggedAddresses: addresses,
			Tags:            info.Tags,
			Meta:            info.Metadata,
		}

		// check := &api.AgentServiceCheck{
		// 	HTTP:     fmt.Sprintf("http://%s:%d/health", c.ip, c.port), // 这里一定是外部可以访问的地址
		// 	Timeout:  "10s",                                            // 超时时间
		// 	Interval: "10s",                                            // 运行检查的频率
		// 	// 指定时间后自动注销不健康的服务节点
		// 	// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
		// 	DeregisterCriticalServiceAfter: "1m",
		// }

		if info.HealthEndpoint != nil {

			if info.HealthEndpoint.Heartbeat {
				c.serviceChecks = append(c.serviceChecks, &api.AgentServiceCheck{
					// maybe modify?
					CheckID:                        "service:" + info.ID,
					TTL:                            fmt.Sprintf("%ds", info.HealthEndpoint.HealthCheckInterval*2),
					DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", info.HealthEndpoint.DeregisterCriticalServiceAfter),
				})
			}
		}

		registration.Checks = c.serviceChecks
		err := c.client.RegisterSvc(registration)
		if err != nil {
			return err
		}

	}

	// start heartbeat here
	return nil
}

func heartbeat(ctx context.Context, client *Client, id string, interval int) {
	// do not deregister here
	time.Sleep(time.Second)
	// need?
	err := client.UpdateTTL("service:"+id, "pass", "pass")
	if err != nil {
		// c.logger.Error("[Consul]update ttl heartbeat to consul failed!err:=%v", err)
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			// exit
			return
		case <-ticker.C:
			// ensure that unregistered services will not be re-registered by mistake
			if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				// not here
				// _ = c.cli.Agent().ServiceDeregister(svc.ID)
				return
			}
			err = client.UpdateTTL("service:"+id, "pass", "pass")

			// need! if consul server fail over
			// if err != nil {
			// 	// c.logger.Error("[Consul] update ttl heartbeat to consul failed! err=%v", err)
			// 	// when the previous report fails, try to re register the service
			// 	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
			// 	if err := c.cli.Agent().ServiceRegister(asr); err != nil {
			// 		c.logger.Error("[Consul] re registry service failed!, err=%v", err)
			// 	} else {
			// 		c.logger.Warn("[Consul] re registry of service occurred success")
			// 	}
			// }
		}
	}

}

// Deregister service by service ID
func (c *RegisterClient) Deregister(_ context.Context, ids ...string) error {
	for _, id := range ids {
		err := c.client.DeregisterSvc(id)
		if err != nil {
			return err
		}
	}
	return nil
}
