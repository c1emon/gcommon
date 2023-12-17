package cloud

import (
	"crypto/tls"
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

type HealthEndpoint struct {
	Endpoint
	Heartbeat                      bool
	HealthCheckInterval            int
	DeregisterCriticalServiceAfter int
}
