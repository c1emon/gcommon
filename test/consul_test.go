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
	loggerFactory := logx.NewLogrusLoggerFactory(logx.LevelDebug)
	logger := loggerFactory.Get("test")
	consulClient, err := consul.New("docker.dev.clemon:8500", loggerFactory)
	if err != nil {
		logger.Fatal("%s", err)
	}

	regClient, err := consul.NewRegisterClient(consulClient)
	if err != nil {
		logger.Fatal("%s", err)
	}
	svc := registry.BuildRegistrationInfo(cloud.RemoteService{
		ID:   "svc1",
		Name: "svc1",
		Endpoint: &cloud.Endpoint{
			Schema: cloud.HTTP,
			Host:   "baidu.com",
			Port:   80,
		},
	}, registry.WithTTL(time.Duration(5)*time.Second),
		registry.WithTimeout(time.Second*time.Duration(10)),
		registry.WithHttpEndpoint("baidu.com", 80))

	regClient.Register([]*registry.RemoteSvcRegInfo{svc})
	time.Sleep(time.Duration(20) * time.Second)
	regClient.Deregister()
}
