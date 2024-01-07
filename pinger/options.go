package pinger

import (
	"net"
	"time"

	"github.com/c1emon/gcommon/util"
)

const (
	DefaultCount    = 1
	DefaultInterval = time.Millisecond * time.Duration(500)
	DefaultTimeout  = time.Second * time.Duration(3)
)

type pingOption util.Option[pingOptions]

type pingOptions struct {
	Timeout  time.Duration
	Interval time.Duration
	Count    int
	Dialer   *net.Dialer
	Tls      bool
}

func WithTimeout(timeout time.Duration) pingOption {
	return util.WrapFuncOption[pingOptions](
		func(po *pingOptions) {
			po.Timeout = timeout
		})
}

func WithNetResolver(resolver *net.Resolver) pingOption {
	return util.WrapFuncOption[pingOptions](
		func(po *pingOptions) {
			po.Dialer = &net.Dialer{
				Resolver: resolver,
			}
		})
}

func WithInterval(interval time.Duration) pingOption {
	return util.WrapFuncOption[pingOptions](
		func(po *pingOptions) {
			po.Interval = interval
		})
}

func WithCount(count int) pingOption {
	return util.WrapFuncOption[pingOptions](
		func(po *pingOptions) {
			if count >= 1 {
				po.Count = count
			} else {
				po.Count = 1
			}
		})
}

func EnableTls() pingOption {
	return util.WrapFuncOption[pingOptions](
		func(po *pingOptions) {
			po.Tls = true
		})
}

type Stats struct {
	Reachable bool          `json:"reachable"`
	Address   string        `json:"address"`
	Error     error         `json:"error"`
	Rtt       time.Duration `json:"rtt"`
	// DnsRtt    time.Duration `json:"dns_rtt"`
	Loss float64 `json:"loss"`
}
