package registry

import (
	"context"

	"github.com/c1emon/gcommon/cloud"
)

// Registrar is service registrar.
// from https://github.com/go-kratos/kratos/blob/main/contrib/registry/consul/registry.go
type Registrar interface {
	// Register the registration.
	Register(ctx context.Context, infos []*RemoteSvcRegInfo) error
	// Deregister the registration(s) by id, no `ids` will deregiste all the service.
	Deregister(ctx context.Context, ids ...string) error
}

// RemoteSvcRegInfo is information of local service to registe
type RemoteSvcRegInfo struct {
	cloud.RemoteService
	HealthEndpoint *cloud.HealthEndpoint
}

func BuildRegistrationInfo(svc cloud.RemoteService, endpoint *cloud.HealthEndpoint) *RemoteSvcRegInfo {
	return &RemoteSvcRegInfo{
		RemoteService:  svc,
		HealthEndpoint: endpoint,
	}
}
