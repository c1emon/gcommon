package consul

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/c1emon/gcommon/cloud"
	"github.com/hashicorp/consul/api"
)

// Registry is consul registry
type Registry struct {
	cli               *Client
	enableHealthCheck bool
	// registry          map[string]*serviceSet
	lock    sync.RWMutex
	timeout time.Duration
}

// RegisterClient 定义一个RegisterClient结构体，其内部有一个`*api.Client`字段。
type RegisterClient struct {
	client      *Client
	id          string
	serviceName string
	ip          string
	port        int

	// resolve service entry endpoints
	resolver ServiceResolver
	// healthcheck time interval in seconds
	healthcheckInterval int
	// heartbeat enable heartbeat
	heartbeat bool
	// deregisterCriticalServiceAfter time interval in seconds
	deregisterCriticalServiceAfter int
	// serviceChecks  user custom checks
	serviceChecks api.AgentServiceChecks
}

// RegisterService
func (c *RegisterClient) RegisterService(serviceName, ip string, port int) error {

	c.id = fmt.Sprintf("%s-%s-%d", serviceName, ip, port)
	c.serviceName = serviceName
	c.ip = ip
	c.port = port

	// ginx.GetGinEng().GET("/health", gin.WrapH(health.Handler(c.serviceName, "0.1")))

	// 健康检查
	check := &api.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d/health", c.ip, c.port), // 这里一定是外部可以访问的地址
		Timeout:  "10s",                                            // 超时时间
		Interval: "10s",                                            // 运行检查的频率
		// 指定时间后自动注销不健康的服务节点
		// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
		DeregisterCriticalServiceAfter: "1m",
	}

	srv := &api.AgentServiceRegistration{
		ID:      c.id,                        // 服务唯一ID
		Name:    c.serviceName,               // 服务名称
		Tags:    []string{"clemon", "hello"}, // 为服务打标签
		Address: c.ip,
		Port:    c.port,
		Check:   check,
	}

	return c.Agent().ServiceRegister(srv)
}

// Register register service instance to consul
func (c *RegisterClient) Register(_ context.Context, svc *cloud.RemoteService, enableHealthCheck bool) error {
	addresses := make(map[string]api.ServiceAddress, len(svc.Endpoints))
	checkAddresses := make([]string, 0, len(svc.Endpoints))
	for _, endpoint := range svc.Endpoints {
		raw, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		addr := raw.Hostname()
		port, _ := strconv.ParseUint(raw.Port(), 10, 16)

		checkAddresses = append(checkAddresses, net.JoinHostPort(addr, strconv.FormatUint(port, 10)))
		addresses[raw.Scheme] = api.ServiceAddress{Address: endpoint, Port: int(port)}
	}
	asr := &api.AgentServiceRegistration{
		ID:              svc.ID,
		Name:            svc.Name,
		Meta:            svc.Metadata,
		Tags:            []string{fmt.Sprintf("version=%s", svc.Version)},
		TaggedAddresses: addresses,
	}
	if len(checkAddresses) > 0 {
		host, portRaw, _ := net.SplitHostPort(checkAddresses[0])
		port, _ := strconv.ParseInt(portRaw, 10, 32)
		asr.Address = host
		asr.Port = int(port)
	}
	if enableHealthCheck {
		for _, address := range checkAddresses {
			asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
				TCP:                            address,
				Interval:                       fmt.Sprintf("%ds", c.healthcheckInterval),
				DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.deregisterCriticalServiceAfter),
				Timeout:                        "5s",
			})
		}
		// custom checks
		asr.Checks = append(asr.Checks, c.serviceChecks...)
	}
	if c.heartbeat {
		asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
			CheckID:                        "service:" + svc.ID,
			TTL:                            fmt.Sprintf("%ds", c.healthcheckInterval*2),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.deregisterCriticalServiceAfter),
		})
	}

	err := c.cli.Agent().ServiceRegister(asr)
	if err != nil {
		return err
	}
	if c.heartbeat {
		go func() {
			time.Sleep(time.Second)
			err = c.cli.Agent().UpdateTTL("service:"+svc.ID, "pass", "pass")
			if err != nil {
				c.logger.Error("[Consul]update ttl heartbeat to consul failed!err:=%v", err)
			}
			ticker := time.NewTicker(time.Second * time.Duration(c.healthcheckInterval))
			defer ticker.Stop()
			for {
				select {
				case <-c.ctx.Done():
					_ = c.cli.Agent().ServiceDeregister(svc.ID)
					return
				default:
				}
				select {
				case <-c.ctx.Done():
					_ = c.cli.Agent().ServiceDeregister(svc.ID)
					return
				case <-ticker.C:
					// ensure that unregistered services will not be re-registered by mistake
					if errors.Is(c.ctx.Err(), context.Canceled) || errors.Is(c.ctx.Err(), context.DeadlineExceeded) {
						_ = c.cli.Agent().ServiceDeregister(svc.ID)
						return
					}
					err = c.cli.Agent().UpdateTTL("service:"+svc.ID, "pass", "pass")
					if err != nil {
						c.logger.Error("[Consul] update ttl heartbeat to consul failed! err=%v", err)
						// when the previous report fails, try to re register the service
						time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
						if err := c.cli.Agent().ServiceRegister(asr); err != nil {
							c.logger.Error("[Consul] re registry service failed!, err=%v", err)
						} else {
							c.logger.Warn("[Consul] re registry of service occurred success")
						}
					}
				}
			}
		}()
	}
	return nil
}

// Deregister service by service ID
func (c *RegisterClient) Deregister(_ context.Context, serviceID string) error {
	defer c.cancel()
	return c.cli.Agent().ServiceDeregister(serviceID)
}
