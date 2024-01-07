package pinger

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptrace"
	"time"
)

func NewTcpinger(host string, port int, opts ...pingOption) *Tcpinger {
	o := &pingOptions{
		Timeout: DefaultTimeout,
		Dialer:  &net.Dialer{},
		Tls:     false,
	}

	for _, v := range opts {
		v.Apply(o)
	}

	return &Tcpinger{
		host:   host,
		port:   port,
		option: o,
	}
}

type Tcpinger struct {
	host   string
	port   int
	option *pingOptions
}

func (p *Tcpinger) Ping(ctx context.Context) *Stats {

	// Statistics
	var stats Stats

	var dnsStart time.Time

	start := time.Now()
	var (
		conn    net.Conn
		err     error
		tlsConn net.Conn
		tlsErr  error
	)

	if p.option.Tls {
		// build tls connection
		tlsDialer := &tls.Dialer{
			NetDialer: p.option.Dialer,
			Config: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		tlsCtx, cancel := context.WithTimeout(ctx, p.option.Timeout)
		defer cancel()
		// trace dns query
		tlsCtx = httptrace.WithClientTrace(tlsCtx, &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				dnsStart = time.Now()
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				stats.DnsRtt = time.Since(dnsStart)
			},
		})

		tlsConn, tlsErr = tlsDialer.DialContext(tlsCtx, "tcp", fmt.Sprintf("%s:%d", p.host, p.port))

		if tlsErr == nil {
			conn = tlsConn
		}
	}
	if !p.option.Tls || tlsErr != nil {
		// if tls failed, downgrade to plain tcp
		tcpCtx, cancel := context.WithTimeout(ctx, p.option.Timeout)
		defer cancel()
		// trace dns query
		tcpCtx = httptrace.WithClientTrace(tcpCtx, &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				dnsStart = time.Now()
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				stats.DnsRtt = time.Since(dnsStart)
			},
		})
		conn, err = p.option.Dialer.DialContext(tcpCtx, "tcp", fmt.Sprintf("%s:%d", p.host, p.port))
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	stats.Rtt = time.Since(start)

	if err != nil {
		stats.Error = err
		if oe, ok := err.(*net.OpError); ok && oe.Addr != nil {
			stats.Address = oe.Addr.String()
		}
	} else {
		stats.Reachable = true
		stats.Address = conn.RemoteAddr().String()
		stats.Error = tlsErr
	}
	return &stats
}
