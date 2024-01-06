package cloud

import (
	"crypto/tls"
	"fmt"
	"time"
)

type Schema int

const (
	TCP Schema = iota
	UDP
	RPC
	HTTP
)

func (s Schema) String() string {
	switch s {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	case RPC:
		return "rpc"
	case HTTP:
		return "http"
	default:
		return ""
	}
}

// Endpoint is where remote service(http/rpc) can be called
type Endpoint struct {
	Enable bool
	Schema Schema `json:"schema"`
	// Host is the ip/domain of endpoint
	Host string `json:"host"`
	Port int    `json:"port"`
	// Path of http url
	Path string `json:"path"`
	// Secure control if use ssl
	Secure bool `json:"secure"`
	// TlsCfg tls/ssl config
	TlsCfg *tls.Config
}

func (e Endpoint) Uri() string {
	var uri string
	if e.Schema == HTTP && e.Secure && e.TlsCfg == nil {
		uri = fmt.Sprintf("%ss://%s:%d", e.Schema, e.Host, e.Port)
	} else {
		uri = fmt.Sprintf("%s://%s:%d", e.Schema, e.Host, e.Port)
	}
	if len(e.Path) > 0 {
		uri = fmt.Sprintf("%s/%s", uri, e.Path)
	}
	return uri
}

// HealthEndpoint define a endpoint for health check
type HealthEndpoint struct {
	// Endpoint is where register center will check
	Endpoint
	// Timeout register center consider the service fail without repley after timeout
	Timeout time.Duration
	// HealthCheckInterval is the interval register center should check endpoint
	HealthCheckInterval time.Duration

	// DeregisterCriticalServiceAfter register center will deregister the service after some interval
	DeregisterCriticalServiceAfter time.Duration

	// Heartbeat means ttl check
	Heartbeat bool
	// HeartbeatInterval update heartbeat interval
	HeartbeatInterval time.Duration
}
