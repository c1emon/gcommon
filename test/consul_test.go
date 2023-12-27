package test

import (
	"testing"
	"time"

	"github.com/c1emon/gcommon/cloud"
	"github.com/c1emon/gcommon/cloud/module/consul"
	"github.com/c1emon/gcommon/cloud/registry"
	"github.com/c1emon/gcommon/logx"
)

func TestConsulRegister(t *testing.T) {
	logger := logx.NewLogrusLoggerFactory(logx.LevelDebug).Get("test")
	consulClient, err := consul.New("docker.dev.clemon:8500")
	if err != nil {
		logger.Fatal("%s", err)
	}

	regClient, err := consul.NewRegisterClient(consulClient)
	if err != nil {
		logger.Fatal("%s", err)
	}

	svc := &registry.RemoteSvcRegInfo{
		RemoteService: cloud.RemoteService{
			ID:   "svc1",
			Name: "svc1",
			Endpoint: &cloud.Endpoint{
				Schema: cloud.HTTP,
				Host:   "baidu.com",
				Port:   80,
			},
		},
		HealthEndpoint: &cloud.HealthEndpoint{
			Endpoint:                       cloud.Endpoint{},
			Heartbeat:                      true,
			HeartbeatInterval:              time.Duration(1) * time.Second,
			HealthCheckInterval:            time.Duration(5) * time.Second,
			Timeout:                        time.Duration(5) * time.Second,
			DeregisterCriticalServiceAfter: time.Duration(5) * time.Second,
		},
	}

	regClient.Register([]*registry.RemoteSvcRegInfo{svc})
	time.Sleep(time.Duration(20) * time.Second)
	regClient.Deregister()
}
