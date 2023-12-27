package registry

import (
	"time"

	"github.com/c1emon/gcommon/cloud"
	"github.com/c1emon/gcommon/util"
)

type healthOption util.Option[healthOptions]

func buildHealthEndpoint(opts ...healthOption) *cloud.HealthEndpoint {
	defaultOpts := &healthOptions{
		ttl:            0,
		timeout:        time.Duration(5) * time.Second,
		interval:       time.Duration(5) * time.Second,
		deregisterTime: time.Duration(1) * time.Minute,
		endpoint:       nil,
	}

	for _, opt := range opts {
		opt.Apply(defaultOpts)
	}

	he := &cloud.HealthEndpoint{}
	he.Timeout = defaultOpts.timeout
	he.HealthCheckInterval = defaultOpts.interval
	he.DeregisterCriticalServiceAfter = defaultOpts.deregisterTime

	if defaultOpts.ttl > 0 {
		he.Heartbeat = true
		he.HeartbeatInterval = defaultOpts.ttl
	}

	if defaultOpts.endpoint != nil {
		he.Endpoint = *defaultOpts.endpoint
	}

	if defaultOpts.ttl > 0 || defaultOpts.endpoint != nil {
		return he
	}

	return nil
}

type healthOptions struct {
	ttl            time.Duration
	timeout        time.Duration
	interval       time.Duration
	deregisterTime time.Duration
	endpoint       *cloud.Endpoint
}

// WithTTL enable heartbeat
func WithTTL(ttl time.Duration) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			if ttl < time.Second*time.Duration(1) {
				ttl = time.Second * time.Duration(1)
			}
			ho.ttl = ttl
		})
}

// WithTimeout set timeout for health check
func WithTimeout(timeout time.Duration) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			if timeout < time.Second*time.Duration(1) {
				timeout = time.Second * time.Duration(1)
			}
			ho.timeout = timeout
		})
}

// WithInterval set health check interval
func WithInterval(interval time.Duration) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			if interval < time.Second*time.Duration(1) {
				interval = time.Second * time.Duration(1)
			}
			ho.interval = interval
		})
}

// WithDeregisterCriticalServiceTime deregister the bad service after timeout
func WithDeregisterCriticalServiceTime(deregisterTime time.Duration) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			if deregisterTime < time.Minute*time.Duration(1) {
				deregisterTime = time.Minute * time.Duration(1)
			}
			ho.deregisterTime = deregisterTime
		})
}

func WithHttpEndpoint(host string, port int) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			ho.endpoint = &cloud.Endpoint{
				Enable: true,
				Schema: cloud.HTTP,
				Host:   host,
				Port:   port,
				Secure: false,
				Path:   "",
				TlsCfg: nil,
			}
		})
}

func WithHttpsEndpoint(host string, port int) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			ho.endpoint = &cloud.Endpoint{
				Enable: true,
				Schema: cloud.HTTP,
				Host:   host,
				Port:   port,
				Secure: true,
				Path:   "",
				TlsCfg: nil,
			}
		})
}

func WithTcpEndpoint(host string, port int) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			ho.endpoint = &cloud.Endpoint{
				Enable: true,
				Schema: cloud.TCP,
				Host:   host,
				Port:   port,
				Secure: false,
				Path:   "",
				TlsCfg: nil,
			}
		})
}

func WithUdpEndpoint(host string, port int) healthOption {
	return util.WrapFuncOption[healthOptions](
		func(ho *healthOptions) {
			ho.endpoint = &cloud.Endpoint{
				Enable: true,
				Schema: cloud.UDP,
				Host:   host,
				Port:   port,
				Secure: false,
				Path:   "",
				TlsCfg: nil,
			}
		})
}
