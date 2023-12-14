package registry

import (
	"context"

	"github.com/c1emon/gcommon/cloud"
)

// Registrar is service registrar.
// from https://github.com/go-kratos/kratos/blob/main/contrib/registry/consul/registry.go
type Registrar interface {
	// Register the registration.
	Register(ctx context.Context, service *cloud.RemoteService) error
	// Deregister the registration.
	Deregister(ctx context.Context, service *cloud.RemoteService) error
}
