package discovery

import (
	"context"

	"github.com/c1emon/gcommon/cloud"
)

// Watcher is service watcher.
type Watcher interface {
	// Next returns services in the following two cases:
	// 1.the first time to watch and the service instance list is not empty.
	// 2.any service instance changes found.
	// if the above two conditions are not met, it will block until context deadline exceeded or canceled
	Next() ([]*cloud.RemoteService, error)
	// Stop close the watcher.
	Stop() error
}

type Discoverer interface {
	// 根据 serviceName 直接拉取实例列表
	GetService(ctx context.Context, serviceName string) ([]*cloud.RemoteService, error)
	// 根据 serviceName 阻塞式订阅一个服务的实例列表信息
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}
