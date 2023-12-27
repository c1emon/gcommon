package consul

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/c1emon/gcommon/cloud"
	"github.com/c1emon/gcommon/cloud/registry"
	"github.com/hashicorp/consul/api"
)

func NewRegisterClient(client *Client) (*RegisterClient, error) {
	ctx, calFn := context.WithCancel(client.ctx)
	return &RegisterClient{
		client:        client,
		registrations: make(map[string]*api.AgentServiceRegistration),
		ctx:           ctx,
		cancelFn:      calFn,
		ttlWg:         &sync.WaitGroup{},
	}, nil
}

// RegisterClient 定义一个RegisterClient结构体，其内部有一个`*api.Client`字段。
type RegisterClient struct {
	client *Client
	// svcInfos []*registry.RemoteSvcRegInfo
	ctx      context.Context
	cancelFn context.CancelFunc
	ttlWg    *sync.WaitGroup

	registrations map[string]*api.AgentServiceRegistration
}

// registrationResolver resolve RemoteSvcRegInfo to consul's AgentServiceRegistration
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
		interval := fmt.Sprintf("%ds", int(info.HealthEndpoint.HealthCheckInterval.Seconds()))
		timeout := fmt.Sprintf("%ds", int(info.HealthEndpoint.Timeout.Seconds()))
		deregisterAfter := fmt.Sprintf("%ds", info.HealthEndpoint.DeregisterCriticalServiceAfter)
		// AgentServiceCheck's `CheckID`` should be used other then `ID``
		if info.HealthEndpoint.Heartbeat {
			serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
				CheckID:                        fmt.Sprintf("%s-ttl", info.ID),
				TTL:                            fmt.Sprintf("%ds", int(info.HealthEndpoint.HeartbeatInterval.Seconds())*2),
				DeregisterCriticalServiceAfter: deregisterAfter,
			})
		}

		if info.HealthEndpoint.Enable {
			switch info.HealthEndpoint.Schema {
			case cloud.HTTP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  fmt.Sprintf("%s-health-http", info.ID),
					HTTP:     info.HealthEndpoint.Uri(), // 这里一定是外部可以访问的地址
					Timeout:  timeout,
					Interval: interval,
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					DeregisterCriticalServiceAfter: deregisterAfter,
				})
			case cloud.TCP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  fmt.Sprintf("%s-health-tcp", info.ID),
					TCP:      fmt.Sprintf("%s:%d", info.HealthEndpoint.Host, info.HealthEndpoint.Port),
					Timeout:  timeout,
					Interval: interval,
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					DeregisterCriticalServiceAfter: deregisterAfter,
				})
			default:
				// no default check?
			}
		}
	}

	registration.Checks = serviceChecks
	return registration
}

// Register register service instance to consul
func (c *RegisterClient) Register(regInfos []*registry.RemoteSvcRegInfo) error {
	ttlFns := make([]func(), 0)
	for _, info := range regInfos {
		registration := registrationResolver(info)
		c.registrations[registration.ID] = registration

		err := c.client.RegisterSvc(registration)
		if err != nil {
			return err
		}
		if info.HealthEndpoint.Heartbeat {
			ttlFn := func() {
				c.ttlWg.Add(1)
				defer c.ttlWg.Done()
				ticker := time.NewTicker(info.HealthEndpoint.HeartbeatInterval * 2)
				defer ticker.Stop()

				for {
					select {
					case <-c.ctx.Done():
						return
					case <-ticker.C:
						if errors.Is(c.ctx.Err(), context.Canceled) || errors.Is(c.ctx.Err(), context.DeadlineExceeded) {
							// _ = c.cli.Agent().ServiceDeregister(svc.ID)
							return
						} else {
							err := c.client.UpdateTTL(fmt.Sprintf("%s-ttl", info.ID), "pass", "pass")
							if err != nil {
								time.Sleep(info.HealthEndpoint.HeartbeatInterval)
								// try register again
								err := c.client.RegisterSvc(registration)
								if err != nil {
									// handle failure of register
								}
							}
						}

					}
				}
			}
			ttlFns = append(ttlFns, ttlFn)
		}

	}

	// start heartbeat here
	for _, ttlFn := range ttlFns {
		go ttlFn()
	}
	return nil
}

// Deregister service by service ID
func (c *RegisterClient) Deregister(_ context.Context, ids ...string) error {
	c.cancelFn()
	if len(ids) == 0 {
		ids = make([]string, len(c.registrations))
		for id := range c.registrations {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		err := c.client.DeregisterSvc(id)
		if err != nil {
			return err
		}
	}
	c.ttlWg.Wait()
	return nil
}
