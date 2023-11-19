package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

// ConsulClient 定义一个ConsulClient结构体，其内部有一个`*api.Client`字段。
type ConsulClient struct {
	*api.Client
	id          string
	serviceName string
	ip          string
	port        int
}

// New 连接至consul服务返回一个consul对象
func New(addr string) (*ConsulClient, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{
		Client: c,
	}, nil
}

// RegisterService
func (c *ConsulClient) RegisterService(serviceName, ip string, port int) error {

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

func (c *ConsulClient) Deregister() error {
	return c.Agent().ServiceDeregister(c.id)
}
