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
	Port string `json:"port"`
	// deal with local ca
	Secure bool `json:"secure"`
	TlsCfg *tls.Config
}
