package registry

import "github.com/c1emon/gcommon/util"

type registryOption util.Option[registryOptions]

type registryOptions struct {
	ttl int
}

// WithTTL enable heartbeat of agent
func WithTTL(ttl int) registryOption {
	return util.WrapFuncOption[registryOptions](
		func(ro *registryOptions) {
			ro.ttl = ttl
		})
}

// WithHealth enable http health endpoint
func WithHttpHealth(address, timeout, interval string) registryOption {
	return util.WrapFuncOption[registryOptions](
		func(ro *registryOptions) {

		})
}

// WithTcpHealth enable tcp health endpoint
func WithTcpHealth() registryOption {
	return util.WrapFuncOption[registryOptions](
		func(ro *registryOptions) {
			// TODO: tcp health params
		})
}

// WithDeregisterCriticalServiceTime deregister the bad service after timeout
func WithDeregisterCriticalServiceTime(timeout string) registryOption {
	return util.WrapFuncOption[registryOptions](
		func(ro *registryOptions) {

		})
}
