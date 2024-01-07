package pinger

import (
	"context"

	probing "github.com/prometheus-community/pro-bing"
)

type Pinger interface {
	Ping(ctx context.Context) *Stats
}

func NewPinger(host string, opts ...pingOption) *ICMPPinger {
	o := &pingOptions{
		Timeout:  DefaultTimeout,
		Interval: DefaultInterval,
		Count:    DefaultCount,
		Dialer:   nil,
		Tls:      false,
	}

	for _, v := range opts {
		v.Apply(o)
	}

	return &ICMPPinger{
		host:   host,
		option: o,
	}
}

// TODO: icmp ping
// https://github.com/prometheus-community/pro-bing/blob/main/cmd/ping/ping.go
type ICMPPinger struct {
	host   string
	option *pingOptions
}

func (p *ICMPPinger) Ping(ctx context.Context) *Stats {
	var stats = &Stats{}
	stats.Reachable = false

	pinger, err := probing.NewPinger(p.host)
	if err != nil {
		stats.Error = err
		return stats
	}
	pinger.Timeout = p.option.Timeout
	pinger.Count = p.option.Count
	pinger.Interval = p.option.Interval

	ctx, cancel := context.WithTimeout(ctx, p.option.Timeout)
	defer cancel()

	err = pinger.RunWithContext(ctx) // Blocks until finished.
	if err != nil {
		stats.Error = err
		return stats
	}

	statistics := pinger.Statistics() // get send/receive/duplicate/rtt stats
	stats.Rtt = statistics.AvgRtt
	stats.Address = statistics.Addr
	stats.Loss = statistics.PacketLoss

	stats.Reachable = true
	return stats
}
