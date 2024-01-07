package pinger

import (
	"net"
	"time"

	"github.com/c1emon/gcommon/util"
)

type tcpingOption util.Option[tcpingOptions]

type tcpingOptions struct {
	Timeout time.Duration
	Dialer  *net.Dialer
	Tls     bool
}

func WithTimeout(timeout time.Duration) tcpingOption {
	return util.WrapFuncOption[tcpingOptions](
		func(so *tcpingOptions) {
			so.Timeout = timeout
		})
}

func WithNetResolver(resolver *net.Resolver) tcpingOption {
	return util.WrapFuncOption[tcpingOptions](
		func(so *tcpingOptions) {
			so.Dialer = &net.Dialer{
				Resolver: resolver,
			}
		})
}

func EnableTls() tcpingOption {
	return util.WrapFuncOption[tcpingOptions](
		func(so *tcpingOptions) {
			so.Tls = true
		})
}

type Stats struct {
	Connected bool          `json:"connected"`
	Error     error         `json:"error"`
	Rtt       time.Duration `json:"rtt"`
	DnsRtt    time.Duration `json:"dns_rtt"`
	Address   string        `json:"address"`
}
