package consul

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c1emon/gcommon/cloud"
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

func registrationResolver(info *registry.RemoteSvcRegInfo) *api.AgentServiceRegistration {
	serviceChecks := make(api.AgentServiceChecks, 0)

	// save registration to re-register
	// when ttl error(likely consul server restart?)
	registration := &api.AgentServiceRegistration{
		ID:      info.ID,
		Name:    info.Name,
		Address: info.Endpoint.Host,
		Port:    info.Endpoint.Port,
		Tags:    info.Tags,
		Meta:    info.Metadata,
	}

	if info.HealthEndpoint != nil {
		if info.HealthEndpoint.Heartbeat {
			serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
				// maybe modify?
				CheckID:                        "service:" + info.ID,
				TTL:                            fmt.Sprintf("%ds", int(info.HealthEndpoint.HeartbeatInterval.Seconds())*2),
				DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", int(info.HealthEndpoint.DeregisterCriticalServiceAfter.Seconds())),
			})
		}

		if info.HealthEndpoint.Enable {
			switch info.HealthEndpoint.Schema {
			case cloud.HTTP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  "",
					HTTP:     info.HealthEndpoint.Uri(), // 这里一定是外部可以访问的地址
					Timeout:  fmt.Sprintf("%ds", int(info.HealthEndpoint.Timeout.Seconds())),
					Interval: fmt.Sprintf("%ds", int(info.HealthEndpoint.HealthCheckInterval.Seconds())),
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					// CheckID:                        "service:" + info.ID,
					DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", info.HealthEndpoint.DeregisterCriticalServiceAfter),
				})
			case cloud.TCP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  "",
					TCP:      fmt.Sprintf("%s:%d", info.HealthEndpoint.Host, info.HealthEndpoint.Port),
					Timeout:  fmt.Sprintf("%ds", int(info.HealthEndpoint.Timeout.Seconds())),
					Interval: fmt.Sprintf("%ds", int(info.HealthEndpoint.HealthCheckInterval.Seconds())),
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					// CheckID:                        "service:" + info.ID,
					DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", int(info.HealthEndpoint.DeregisterCriticalServiceAfter.Seconds())),
				})
			default:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  "",
					TCP:      fmt.Sprintf("%s:%d", info.Endpoint.Host, info.Endpoint.Port),
					Timeout:  fmt.Sprintf("%ds", int(info.HealthEndpoint.Timeout.Seconds())),
					Interval: fmt.Sprintf("%ds", int(info.HealthEndpoint.HealthCheckInterval.Seconds())),
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					// CheckID:                        "service:" + info.ID,
					DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", int(info.HealthEndpoint.DeregisterCriticalServiceAfter.Seconds())),
				})
			}
		}

	}

	registration.Checks = serviceChecks

	return registration
}

// Register register service instance to consul
func (c *RegisterClient) Register(ctx context.Context, infos []*registry.RemoteSvcRegInfo) error {
	c.svcInfos = infos
	for _, info := range infos {

		err := c.client.RegisterSvc(registrationResolver(info))
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
