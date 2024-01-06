package registry

import (
	"github.com/c1emon/gcommon/cloud"
)

// Registrar is service registrar.
// from https://github.com/go-kratos/kratos/blob/main/contrib/registry/consul/registry.go
type Registrar interface {
	// Register the registration.
	Register(infos []*RemoteSvcRegInfo) error
	// Deregister the registration(s) by id, no `ids` will deregiste all the service.
	Deregister(ids ...string) error
}

// RemoteSvcRegInfo is information of local service to registe
type RemoteSvcRegInfo struct {
	cloud.RemoteService
	HealthEndpoint *cloud.HealthEndpoint
}

func BuildRegistrationInfo(svc cloud.RemoteService, healOpts ...healthOption) *RemoteSvcRegInfo {
	return &RemoteSvcRegInfo{
		RemoteService:  svc,
		HealthEndpoint: buildHealthEndpoint(healOpts...),
	}
}
