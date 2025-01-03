package consul

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
		logger:        client.logger,
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

	logger *slog.Logger
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
		interval := info.HealthEndpoint.HealthCheckInterval.String()
		timeout := info.HealthEndpoint.Timeout.String()
		deregisterAfter := info.HealthEndpoint.DeregisterCriticalServiceAfter.String()
		// AgentServiceCheck's `CheckID`` should be used other then `ID``
		if info.HealthEndpoint.Heartbeat {
			serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
				CheckID:                        fmt.Sprintf("%s-ttl", info.ID),
				TTL:                            fmt.Sprintf("%ds", int(info.HealthEndpoint.HeartbeatInterval.Seconds()+1)),
				DeregisterCriticalServiceAfter: deregisterAfter,
			})
		}

		if info.HealthEndpoint.Enable {
			switch info.HealthEndpoint.Schema {
			case cloud.HTTP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  fmt.Sprintf("%s-http", info.ID),
					HTTP:     info.HealthEndpoint.Uri(), // 这里一定是外部可以访问的地址
					Timeout:  timeout,
					Interval: interval,
					// 指定时间后自动注销不健康的服务节点
					// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
					DeregisterCriticalServiceAfter: deregisterAfter,
				})
			case cloud.TCP:
				serviceChecks = append(serviceChecks, &api.AgentServiceCheck{
					CheckID:  fmt.Sprintf("%s-tcp", info.ID),
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

		c.logger.Info(fmt.Sprintf("consul register server %s", registration.Name))
		if info.HealthEndpoint != nil && info.HealthEndpoint.Heartbeat {
			ttlFn := func() {
				c.ttlWg.Add(1)
				defer c.ttlWg.Done()
				ticker := time.NewTicker(info.HealthEndpoint.HeartbeatInterval * 2)
				defer ticker.Stop()
				id := fmt.Sprintf("%s-ttl", info.ID)
				for {
					select {
					case <-c.ctx.Done():
						if !errors.Is(c.ctx.Err(), context.Canceled) {
							c.logger.Error("consul heartbeat handler exited", "error", c.ctx.Err())
						} else {
							c.logger.Info(fmt.Sprintf("stop heartbeat handler %s success", id))
						}
						return
					case <-ticker.C:
						err := c.client.UpdateTTL(id, "pass", "pass")
						if err != nil {
							time.Sleep(info.HealthEndpoint.HeartbeatInterval)
							// try register again
							c.logger.Warn("update heartbeat failed", "error", err)
							c.logger.Info(fmt.Sprintf("try re-register service %s", registration.ID))
							err := c.client.RegisterSvc(registration)
							if err != nil {
								c.logger.Error("re-register service failed", "id", registration.ID, "error", err)
							} else {
								c.logger.Info("re-register service success", "id", registration.ID)
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
func (c *RegisterClient) Deregister(ids ...string) error {
	c.cancelFn()
	if len(ids) == 0 {
		ids = make([]string, 0)
		for id := range c.registrations {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		c.logger.Info("de-register service %s", id)
		err := c.client.DeregisterSvc(id)
		if err != nil {
			return err
		}
		delete(c.registrations, id)
	}
	c.ttlWg.Wait()
	return nil
}
